---
title: feat: Pill-Shaped Header Backgrounds
type: feat
status: completed
date: 2026-02-22
---

# feat: Pill-Shaped Header Backgrounds

## Overview

Replace Glamour's full-width colored header backgrounds with tight, pill-shaped backgrounds rendered via `lipgloss`. The colored fill hugs the header text (with 2 spaces of horizontal padding on each side) rather than stretching to the terminal edge.

## Problem Statement / Motivation

Current H2–H6 backgrounds flood the entire terminal line — an artifact of how Glamour emits ANSI background resets at line-end. This makes headers look like solid ribbons spanning the full width. A tight pill shape is more readable and visually modern, and is consistent with how badge/label elements look in graphical UIs.

H1 is currently unstyled (inherits Glamour default). This feature brings it into the unified styled system along with H2–H6.

## Proposed Solution

Follow the pre-extraction → custom rendering → injection pipeline already established for code blocks (`model.go:39–175`):

1. **Extract** — scan raw markdown for all headers before Glamour runs; replace each with a unique placeholder (`INCIPIT_HEADER_0`, `INCIPIT_HEADER_1`, …)
2. **Render prose** — pass placeholder-substituted markdown to Glamour as normal
3. **Render headers** — for each extracted header, produce a `lipgloss`-styled ANSI string with tight background + padding
4. **Inject** — replace placeholder lines in Glamour's output with the rendered header strings

## Technical Considerations

### Regex

Match all ATX heading lines in multiline mode:

```go
// model.go (new, near codeBlockRe)
var headerRe = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
```

- Group 1: `#` characters → heading level (1–6)
- Group 2: heading text (may contain inline markdown)

### Struct

```go
// model.go
type headerBlock struct {
    level int
    text  string
}
```

### Placeholder scheme

`INCIPIT_HEADER_0`, `INCIPIT_HEADER_1`, … — consistent with `INCIPIT_CODEBLOCK_N`.

### lipgloss rendering

```go
// model.go — renderHeader()
func renderHeader(h headerBlock, isDark bool) string {
    fg, bg, bold := headerColors(h.level, isDark)
    s := lipgloss.NewStyle().
        Foreground(lipgloss.Color(fg)).
        Background(lipgloss.Color(bg)).
        Padding(0, 2).
        Bold(bold)
    return s.Render(stripInlineMarkdown(h.text))
}
```

### Inline markdown stripping

Headers can contain inline markup (`**bold**`, `_italic_`, `` `code` ``). Feeding raw markdown to lipgloss would expose literal delimiters. Strip common patterns with a simple regex before rendering:

```go
// model.go — stripInlineMarkdown()
func stripInlineMarkdown(s string) string {
    // Remove **, __, *, _, `, ~~
    s = regexp.MustCompile(`[*_~` + "`" + `]{1,2}`).ReplaceAllString(s, "")
    return strings.TrimSpace(s)
}
```

### Color mapping

#### Dark theme

| Level | fg  | bg  | Bold  |
|-------|-----|-----|-------|
| H1    | 15  | 57  | true  |
| H2    | 51  | 23  | true  |
| H3    | 48  | 22  | true  |
| H4    | 75  | 17  | true  |
| H5    | 67  | 236 | false |
| H6    | 60  | 235 | false |

#### Light theme

| Level | fg  | bg  | Bold  |
|-------|-----|-----|-------|
| H1    | 0   | 105 | true  |
| H2    | 27  | 195 | true  |
| H3    | 28  | 194 | true  |
| H4    | 19  | 189 | true  |
| H5    | 17  | 153 | false |
| H6    | 59  | 188 | false |

### Glamour cleanup

Once headers are pre-extracted, Glamour never sees them. Remove the H2–H6 assignments from `customizeHeaders()` and `customizeHeadersLight()`, then delete both functions.

## Acceptance Criteria

- [x] H1–H6 backgrounds are tight to the text — no full-width fill
- [x] 2 spaces of horizontal padding on each side (`  Header Text  `)
- [x] All header pills flush left (no level-based indentation)
- [x] H1–H4 are bold; H5–H6 are not bold
- [x] Dark and light themes both render correctly
- [x] Headers with inline markdown (`**bold**`, `` `code` ``, `_italic_`) render without visible delimiters
- [x] Long headers that exceed terminal width wrap without ANSI corruption
- [x] Headers adjacent to code blocks coexist without placeholder or injection errors
- [x] `go test ./...` passes — existing tests updated, new tests added
- [x] `go vet ./...` passes

## Dependencies & Risks

| Risk | Mitigation |
|------|-----------|
| **Inline markdown in header text** — raw `**Bold**` shown to lipgloss | Strip delimiters with `stripInlineMarkdown()` before calling `Render()` |
| **Vertical whitespace drift** — Glamour adds different margins around a placeholder paragraph vs. a heading element | Verify spacing in integration test; adjust placeholder format if needed |
| **H1 color choice** — H1 was previously unstyled so no existing colors to reuse | Colors specified in the table above; adjust during implementation if contrast is insufficient |
| **Placeholder collision** — user content containing `INCIPIT_HEADER_N` | Unique enough prefix for a markdown reader; document as known limitation |

## Implementation Files

- **`model.go`** — all new functions (`headerBlock`, `extractHeaders`, `renderHeader`, `injectHeaders`, `headerColors`, `stripInlineMarkdown`); updated `renderMarkdown()` pipeline; simplified `customizeHeaders` / `customizeHeadersLight`
- **`model_test.go`** — new tests following code block test patterns

## Suggested Test Coverage

```
TestExtractHeaders_SingleH2
TestExtractHeaders_MultipleLevelsReturnsCorrectOrder
TestExtractHeaders_NoHeaders
TestExtractHeaders_WithInlineMarkdown
TestRenderHeader_DarkH2_PillShape
TestRenderHeader_LightH1_PillShape
TestInjectHeaders_ReplacesPlaceholder
TestRenderMarkdown_HeaderH1DarkPill  (integration)
TestRenderMarkdown_HeaderH3LightPill (integration)
TestStripInlineMarkdown_Bold
TestStripInlineMarkdown_InlineCode
```

## References & Research

- Code block pipeline (reference implementation): `model.go:39–175`
- Current header color config: `model.go:214–256`
- Existing header tests: `model_test.go:33–61`
- Existing code block tests (test style reference): `model_test.go:156–333`
- Prior plan (code block custom renderer): `docs/plans/2026-02-22-feat-code-block-rounded-border-and-full-background-plan.md`
- lipgloss `Padding` API: `.Padding(topBottom, leftRight)` — `leftRight` gives symmetric horizontal padding
