---
status: pending
priority: p1
issue_id: "005"
tags: [code-review, agent-native, pipeline, rendering]
---

# Non-Pager Path: Hardcoded 80-Column Width Breaks Pipeline Output

## Problem Statement

`main.go:65` hardcodes `renderMarkdown(content, style, 80)` in the non-pager/non-TTY path. When an agent or CI job pipes output — `incipit README.md | grep ...` or `incipit --no-pager README.md > out.txt` — the render width is always 80 columns regardless of what the consumer needs.

Glamour's word-wrapper is not reversible. A pipeline consumer receiving 80-column-wrapped paragraphs will see artificial line breaks mid-sentence. For language model ingestion, grep, or text file output, this corrupts semantic structure.

## Findings

- `main.go:65` — `out := renderMarkdown(content, style, 80)` — hardcoded 80
- `golang.org/x/term` is already imported (used for TTY detection at `main.go:64`) — `term.GetSize` is available but not called here
- In non-TTY stdout, `term.GetSize(int(os.Stdout.Fd()))` returns an error — so a fallback wide value is needed
- Interactive pager path correctly uses `msg.Width` from bubbletea's `tea.WindowSizeMsg` — the non-pager path has no equivalent

## Proposed Solutions

### Option A: Use wide fallback width (Recommended, minimal change)

```go
// main.go — non-pager path
width := 4096 // no-wrap sentinel; glamour treats 0 or large values as unlimited
if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
    width = w
}
out := renderMarkdown(content, style, width)
```

Pros: One-time fix, sensible behavior for all consumers, TTY width still used when available
Cons: Glamour behavior with very large widths should be verified (likely fine — it is word-wrap max, not allocation)
Effort: Small | Risk: Low

### Option B: Add --width flag

```go
flag.IntVar(&widthFlag, "width", 0, "render width (0 = auto-detect or no-wrap)")
```

Then use `widthFlag` if set, otherwise fall through to Option A logic.

Pros: Explicit control for pipeline consumers
Cons: Adds a flag, more surface area, overkill for now
Effort: Small | Risk: Low

### Option C: Keep 80 for --no-color / notty style (plain text), use wide for styled

Only apply wide width when color is on; plain text at 80 columns is often fine for terminal width assumptions.

Pros: Preserves historical behavior for plain text consumers
Cons: Inconsistent — agent still gets 80-wrapped colored output if --no-color is not set
Effort: Small | Risk: Medium

## Acceptance Criteria

- [ ] `incipit --no-pager README.md | cat` produces output without artificial line wraps in paragraphs
- [ ] When stdout is a TTY, terminal width is used (unchanged behavior in interactive mode)
- [ ] When stdout is not a TTY, render width is ≥ 4096 (or effectively unlimited)
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes

## Work Log

- 2026-02-19: Found during agent-native review of the pending-todos close plan. Agent-native reviewer identified 80-column hardcode as the most impactful agent ergonomics issue.
