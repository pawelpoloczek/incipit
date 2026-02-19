---
title: "feat: Consistent background-color styling for all markdown headers (H1–H6)"
type: feat
status: completed
date: 2026-02-18
brainstorm: docs/brainstorms/2026-02-18-header-styling-brainstorm.md
---

# feat: Consistent background-color styling for all markdown headers (H1–H6)

## Overview

All markdown headers (H1–H6) should render with a background color block and no raw `##` prefix markers visible. Currently, H1 behaves correctly (background color, marker stripped), while H2–H6 display the raw `##` syntax characters with foreground-only color — a visual inconsistency from Glamour's built-in themes.

The fix: copy Glamour's exported `DarkStyleConfig`/`LightStyleConfig` structs, mutate only the H2–H6 fields, and pass the result to `glamour.WithStyles()`. No new files, no embed directives.

## Problem Statement

Glamour's built-in `dark.json` sets:

```json
"heading": { "color": "39", "bold": true },
"h1":      { "prefix": " ", "suffix": " ", "color": "228", "background_color": "63", "bold": true },
"h2":      { "prefix": "## " },
"h3":      { "prefix": "### " },
"h4":      { "prefix": "#### " },
"h5":      { "prefix": "##### " },
"h6":      { "prefix": "###### ", "color": "35", "bold": false }
```

H2–H6 inherit the base `heading` color (`39` = bright cyan) and bold, but retain their `##` prefix tokens. The result:

```
█ TITLE █          ← H1: yellow on purple bg, marker stripped ✓
## Section         ← H2: cyan text, ## visible ✗
### Subsection     ← H3: cyan text, ### visible ✗
```

## Proposed Solution

Copy Glamour's exported style structs (`styles.DarkStyleConfig`, `styles.LightStyleConfig`) by value, override the H2–H6 fields, and use `glamour.WithStyles()`. This requires zero new files and stays in sync with Glamour's upstream defaults for all non-heading elements automatically.

**Dark theme visual design (cool-tone hierarchy):**

```
H1  █ yellow (228) on purple (63) █        — preserved, warmest/most prominent
H2  █ bright cyan (51) on dark teal (23) █ — cool, high contrast
H3  █ cyan-green (48) on dark green (22) █ — slightly dimmer
H4  █ sky blue (75) on dark navy (17) █    — quieter
H5    dim blue (67) on near-black (236)    — subtle
H6    muted purple (60) on very dark (235) — barely-there
```

**Light theme visual design:**

```
H1  █ yellow (228) on purple (63) █          — preserved (existing behavior)
H2  █ dark blue (27) on light cyan (195) █   — cool, prominent
H3  █ dark green (28) on light green (194) █ — green family
H4  █ dark navy (19) on light lavender (189) █ — blue family
H5  █ dark navy (17) on light blue (153) █   — mid-pastel, perceptible on white terminals
H6  █ dark gray (59) on medium gray (188) █  — medium gray, visible on white terminals
```

> Note: H5/H6 light theme use mid-pastel backgrounds (153–188 range) rather than near-white (229–252) to ensure perceptible contrast on white terminal backgrounds.

All H2–H6 use `Prefix: " "` and `Suffix: " "` to match H1's single-space padding.

## Technical Approach

### Files to modify

#### `model.go` only — no new files

Import the Glamour `styles` and `ansi` sub-packages (already transitive deps, just not directly imported):

```go
import (
    "github.com/charmbracelet/glamour"
    glamouransi "github.com/charmbracelet/glamour/ansi"
    glamourstyles "github.com/charmbracelet/glamour/styles"
)

func renderMarkdown(md, style string, width int) string {
    var styleOpt glamour.TermRendererOption
    switch style {
    case "dark":
        s := glamourstyles.DarkStyleConfig
        customizeHeaders(&s)
        styleOpt = glamour.WithStyles(s)
    case "light":
        s := glamourstyles.LightStyleConfig
        customizeHeadersLight(&s)
        styleOpt = glamour.WithStyles(s)
    default:
        styleOpt = glamour.WithStandardStyle(style) // "notty" unchanged
    }

    r, err := glamour.NewTermRenderer(styleOpt, glamour.WithWordWrap(width))
    if err != nil {
        return md
    }
    out, err := r.Render(md)
    if err != nil {
        return md
    }
    return strings.TrimRight(out, "\n")
}

// customizeHeaders overrides H2-H6 in the given StyleConfig to render with
// a background color block and no raw markdown prefix (e.g., "## ").
// H1 is left unchanged — it already renders correctly in Glamour's built-in themes.
//
// Cascade note: each Hx's explicit non-empty Prefix (" ") wins in
// cascadeStylePrimitive() regardless of what the base heading style's Prefix is.
func customizeHeaders(s *glamouransi.StyleConfig) {
    sp := func(v string) *string { return &v }
    bt := true
    bf := false

    s.H2 = glamouransi.StyleBlock{StylePrimitive: glamouransi.StylePrimitive{
        Prefix: " ", Suffix: " ", Color: sp("51"), BackgroundColor: sp("23"), Bold: &bt,
    }}
    s.H3 = glamouransi.StyleBlock{StylePrimitive: glamouransi.StylePrimitive{
        Prefix: " ", Suffix: " ", Color: sp("48"), BackgroundColor: sp("22"), Bold: &bt,
    }}
    s.H4 = glamouransi.StyleBlock{StylePrimitive: glamouransi.StylePrimitive{
        Prefix: " ", Suffix: " ", Color: sp("75"), BackgroundColor: sp("17"), Bold: &bt,
    }}
    s.H5 = glamouransi.StyleBlock{StylePrimitive: glamouransi.StylePrimitive{
        Prefix: " ", Suffix: " ", Color: sp("67"), BackgroundColor: sp("236"), Bold: &bf,
    }}
    s.H6 = glamouransi.StyleBlock{StylePrimitive: glamouransi.StylePrimitive{
        Prefix: " ", Suffix: " ", Color: sp("60"), BackgroundColor: sp("235"), Bold: &bf,
    }}
}

func customizeHeadersLight(s *glamouransi.StyleConfig) {
    sp := func(v string) *string { return &v }
    bt := true
    bf := false

    s.H2 = glamouransi.StyleBlock{StylePrimitive: glamouransi.StylePrimitive{
        Prefix: " ", Suffix: " ", Color: sp("27"), BackgroundColor: sp("195"), Bold: &bt,
    }}
    s.H3 = glamouransi.StyleBlock{StylePrimitive: glamouransi.StylePrimitive{
        Prefix: " ", Suffix: " ", Color: sp("28"), BackgroundColor: sp("194"), Bold: &bt,
    }}
    s.H4 = glamouransi.StyleBlock{StylePrimitive: glamouransi.StylePrimitive{
        Prefix: " ", Suffix: " ", Color: sp("19"), BackgroundColor: sp("189"), Bold: &bt,
    }}
    s.H5 = glamouransi.StyleBlock{StylePrimitive: glamouransi.StylePrimitive{
        Prefix: " ", Suffix: " ", Color: sp("17"), BackgroundColor: sp("153"), Bold: &bf,
    }}
    s.H6 = glamouransi.StyleBlock{StylePrimitive: glamouransi.StylePrimitive{
        Prefix: " ", Suffix: " ", Color: sp("59"), BackgroundColor: sp("188"), Bold: &bf,
    }}
}
```

**Why not JSON+embed:** Glamour exports `styles.DarkStyleConfig` and `styles.LightStyleConfig` as Go structs. Using `glamour.WithStyles()` with a copied and mutated struct eliminates ~108 lines (100 JSON + 8 embed boilerplate), adds zero files, and auto-syncs non-heading elements with upstream theme changes on version bumps.

## Acceptance Criteria

- [x] `# H1` renders with yellow text on purple background, no `#` marker (behavior unchanged)
- [x] `## H2` through `###### H6` render with background color blocks, no `##` prefix markers visible
- [x] Headers H1–H6 show a visually distinct hierarchy through decreasing color weight
- [x] `--dark` flag (and default) uses the dark cool-tone palette
- [x] `--light` flag uses the light palette; all 6 levels are visually distinct on a white terminal background
- [x] `--no-color` / `NO_COLOR` env var remains completely unaffected (uses `notty`)
- [x] `go test ./...` passes — all existing tests green
- [x] `go vet ./...` passes with no warnings
- [x] `go build -o cli-md .` produces a working binary
- [x] No `styles/` directory created, no `//go:embed` directives added
- [x] New tests:
  - `TestRenderMarkdown_H2DarkNoHashPrefix` — H2 in dark theme: no `## `, heading text present
  - `TestRenderMarkdown_H3DarkNoHashPrefix` — H3 in dark theme: no `### `, heading text present
  - `TestRenderMarkdown_H2LightNoHashPrefix` — H2 in light theme: no `## `, heading text present
- [ ] Visual check: run `cli-md --light README.md` on a light-background terminal; confirm H5/H6 are legible

## Dependencies & Risks

- **No new external dependencies.** `github.com/charmbracelet/glamour/ansi` and `github.com/charmbracelet/glamour/styles` are already transitive deps via the existing glamour import — this just adds direct imports.
- **Cascade behavior:** `customizeHeaders` replaces the entire `StyleBlock` for H2–H6. This means inherited fields from the `heading` base (e.g., `BlockSuffix: "\n"`) are NOT inherited by the new `StyleBlock`. The `heading.block_suffix` applies at the heading renderer level, not per-StyleBlock, so spacing is unaffected. Verify with a rendered output check during implementation.
- **No CLI flag changes** — `chooseStyle()` in `main.go` is untouched. The switch cases `"dark"` and `"light"` must match `chooseStyle()`'s return values; add a comment at the switch documenting this coupling.

## References

### Internal

- `model.go:25-38` — `renderMarkdown()` function to update
- `model_test.go:8-16` — `TestRenderMarkdown_ReturnsContent` (must still pass)
- `model_test.go:18-24` — `TestRenderMarkdown_FallsBackOnBadStyle` (must still pass)
- `main.go:12-23` — `chooseStyle()` — switch cases must match its return values

### External (Glamour source, read-only reference)

- `~/go/pkg/mod/github.com/charmbracelet/glamour@v0.10.0/glamour.go:163` — `WithStyles(ansi.StyleConfig)` API
- `~/go/pkg/mod/github.com/charmbracelet/glamour@v0.10.0/styles/styles.go` — `DarkStyleConfig`, `LightStyleConfig` exported structs
- `~/go/pkg/mod/github.com/charmbracelet/glamour@v0.10.0/ansi/style.go:97-139` — `StyleConfig` and `StyleBlock` struct definitions

### Related Todos (addressed in this revision)

- `todos/004` — switched from JSON+embed to `WithStyles()` + struct copy ✓
- `todos/005` — fixed light theme H5 (`153`) and H6 (`188`) background colors ✓
- `todos/006` — added H3 dark + H2 light tests, bidirectional assertions ✓
- `todos/007` — added cascade comment to `customizeHeaders()` ✓
