---
status: pending
priority: p1
issue_id: "009"
tags: [code-review, agent-native, pipeline, plan]
---

# Language Label Must Not Run in --no-pager / Piped Mode

## Problem Statement

The plan proposes calling `addLanguageLabels(md)` at the top of `renderMarkdown`. Both the interactive pager and the `--no-pager`/non-TTY pipeline path call the same `renderMarkdown` function (`main.go:65`). This means injected `` `lang` `` labels will appear in piped stdout, causing downstream tools (`grep`, `awk`, `sed`, scripts) to see injected lines that were never in the source document.

Concrete example — a user runs `incipit --no-pager README.md | grep pattern`:

```
# Before (today)
func main() {}

# After plan with no guard
go
func main() {}
```

The `go` line is an artifact, not source content. Searching for code snippets in pipeline mode would return false results.

Additionally, if `addLanguageLabels` runs inside `renderMarkdown`, the search index (`m.searchLines`) will contain label artifact lines, so `/` searching for `go` would match label lines, not just code content.

## Proposed Solutions

### Option A: Call in main.go interactive branch only (Recommended)

Move the `addLanguageLabels(content)` call to `main.go`, inside the interactive branch only:

```go
// main.go
if noPagerFlag || !term.IsTerminal(int(os.Stdout.Fd())) {
    out := renderMarkdown(content, style, 80)   // raw content, no labels
    fmt.Print(out)
    return
}

content = addLanguageLabels(content)            // only for interactive pager
m := newModel(filename, content, style)
```

`renderMarkdown` stays a pure renderer. `m.rawMarkdown` stores the pre-injected content, so `applyContent` on resize re-renders from the labeled source consistently. No double-injection because `addLanguageLabels` is called once before model construction.

Pros: Clean separation, `renderMarkdown` stays pure, no label in pipeline output, no search index contamination
Cons: None
Effort: Small | Risk: None

### Option B: Add a boolean parameter to renderMarkdown

```go
func renderMarkdown(md, style string, width int, injectLabels bool) string {
    if injectLabels {
        md = addLanguageLabels(md)
    }
    // ...
}
```

Pros: Explicit at each call site
Cons: Clutters function signature; the decision is already made at the branch in main.go
Effort: Small | Risk: Low

## Acceptance Criteria

- [ ] `addLanguageLabels` is NOT called inside `renderMarkdown`
- [ ] `addLanguageLabels` IS called in `main.go` before `newModel`, in the interactive branch only
- [ ] `incipit --no-pager README.md | cat` output contains no injected language label lines
- [ ] Searching (`/`) for a language name does not match label artifacts
- [ ] New test: verify `renderMarkdown` called directly on a fenced block does not inject labels

## Work Log

- 2026-02-19: Found during agent-native review of the code block styling plan.
