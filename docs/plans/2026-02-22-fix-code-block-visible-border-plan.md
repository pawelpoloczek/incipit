---
title: "fix: Code block visible left border via indent strip"
type: fix
status: completed
date: 2026-02-22
---

# fix: Code block visible left border via indent strip

Code blocks in dark (and light) mode show no visible background or border. The previous PR
set `Chroma.Background.BackgroundColor` to themed hex colors (`#1e2030` dark, `#f6f8fa`
light), but these are **silently stripped** by glamour's default `terminal256` Chroma
formatter before any rendering occurs. The result: code blocks look identical to surrounding
prose — no visual distinction.

---

## Root Cause

### Why `Chroma.Background.BackgroundColor` has no effect

Glamour's default Chroma formatter is `"terminal256"` (hardcoded in
`ansi/codeblock.go:17`). The Chroma `terminal256` formatter in
`tty_indexed.go:226-235` calls `clearBackground()` before building its escape-sequence
map, which zeroes out the `chroma.Background` entry:

```go
// alecthomas/chroma/v2 formatters/tty_indexed.go
func clearBackground(style *chroma.Style) *chroma.Style {
    builder := style.Builder()
    bg := builder.Get(chroma.Background)
    bg.Background = 0      // ← background color is discarded
    bg.NoInherit = true
    builder.AddEntry(chroma.Background, bg)
    style, _ = builder.Build()
    return style
}
```

Every theme with `Chroma.Background.BackgroundColor` (including glamour's built-in
Dracula, Tokyo Night, and our customised dark/light themes) is affected.

### Why `StyleCodeBlock.StyleBlock.StylePrimitive.BackgroundColor` has no effect

When Chroma is active (i.e., `rules.Chroma != nil`), glamour's `codeblock.go` goes
through the Chroma highlight path and **never calls** `BaseElement.Render` — the
fallback path that uses `StylePrimitive`. So setting a background on the StylePrimitive
of a CodeBlock is dead code whenever Chroma is configured.

### Why full-width backgrounds are not achievable in the Chroma path

Glamour's code block renderer uses an `indent.Writer` (not the `padding.Writer` used
for other block elements). The indent writer adds left-margin spaces but does not pad
each line to terminal width. Even if a Chroma formatter applied a background color,
short lines would not be filled to the right edge.

---

## The Reliable Fix: Coloured Indent Strip

The one place where `StylePrimitive` IS applied during Chroma rendering is the
**indent writer callback** (`codeblock.go:130-132`):

```go
iw := indent.NewWriterPipe(w, indentation+margin, func(_ io.Writer) {
    renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, " ")
})
```

`bs.Current().Style.StylePrimitive` is `StyleCodeBlock.StyleBlock.StylePrimitive`.
When `BackgroundColor` is set here (using a 256-color index), every left-margin
space drawn on every line of the code block gets that background — producing a
**vertical coloured stripe** that runs the full height of the block.

This approach:
- Works on any 256-color terminal (no TrueColor required)
- Spans the full height of the code block (every line, including blank lines)
- Requires zero changes to the rendering pipeline
- Does not break syntax highlighting

---

## Proposed Solution

### Part 1 — Left border strip (reliable, all terminals)

In `customizeHeaders` and `customizeHeadersLight` in `model.go`, set
`StyleCodeBlock.StyleBlock.StylePrimitive.BackgroundColor` and
`StyleCodeBlock.StyleBlock.StylePrimitive.Color` using 256-color indices.

**Dark theme** — a dark teal strip that echoes the H2 heading accent:
```go
s.CodeBlock.StyleBlock.StylePrimitive.BackgroundColor = sp("23")  // dark teal
s.CodeBlock.StyleBlock.StylePrimitive.Color = sp("244")           // existing text color
```

**Light theme** — a pale blue strip that echoes the H2 light heading accent:
```go
s.CodeBlock.StyleBlock.StylePrimitive.BackgroundColor = sp("195") // pale blue-cyan
s.CodeBlock.StyleBlock.StylePrimitive.Color = sp("242")           // existing text color
```

Color candidates (pick during implementation with visual verification):

| Theme | 256-color | Hex approx. | Character |
|-------|-----------|-------------|-----------|
| dark  | `"23"`    | `#005f5f`   | distinct teal strip  |
| dark  | `"237"`   | `#3a3a3a`   | subtle dark gray     |
| dark  | `"17"`    | `#00005f`   | dark navy blue       |
| light | `"195"`   | `#d7ffff`   | pale cyan            |
| light | `"189"`   | `#d7d7ff`   | pale blue            |
| light | `"153"`   | `#afd7ff`   | soft blue            |

> **Visual verification**: run `go run . --dark README.md` and `go run . --light README.md`
> after each change to confirm the strip is visible at the left edge of code blocks.

### Part 2 — TrueColor Chroma formatter (optional enhancement)

For terminals that support 24-bit colour, passing `glamour.WithChromaFormatter("terminal16m")`
to the renderer enables Chroma's TrueColor formatter, which does NOT call `clearBackground()`.
This makes the existing `Chroma.Background.BackgroundColor` values (`#1e2030`, `#f6f8fa`)
render on token character cells.

Since this only runs for `"dark"` and `"light"` styles, non-TrueColor terminals are unaffected.

```go
case "dark":
    s := glamourstyles.DarkStyleConfig
    customizeHeaders(&s)
    styleOpt = glamour.WithStyles(s)
    // new:
    return renderMarkdownWith(md, width, styleOpt, glamour.WithChromaFormatter("terminal16m"))
```

**Risk:** On terminals that don't support 24-bit color codes, escape sequences like
`\033[38;2;R;G;Bm` may be ignored (no visible regression) or misinterpreted (rare,
old terminals). This can be deferred if there is any concern — Part 1 alone satisfies
the requirement.

---

## Files to Change

### `model.go`

1. In `customizeHeaders` — set `s.CodeBlock.StyleBlock.StylePrimitive.BackgroundColor` and
   `s.CodeBlock.StyleBlock.StylePrimitive.Color` for dark theme
2. In `customizeHeadersLight` — set the same fields for light theme
3. (Part 2, optional) In `renderMarkdown` — add `glamour.WithChromaFormatter("terminal16m")`
   to the renderer options for `"dark"` and `"light"` cases

No other files need changes.

---

## Acceptance Criteria

- [x] Dark mode: a distinct coloured strip is visible along the left edge of every code block
- [x] The strip spans the full vertical height of the code block (all lines, including blank lines inside the block)
- [x] Works on 256-color terminals — no TrueColor terminal required
- [x] Light mode: a corresponding strip is visible (doesn't break light rendering)
- [x] Blocks with no language specifier also show the strip (strip is applied regardless of language label)
- [x] Inline code spans are unaffected
- [x] `go test ./...` passes
- [x] `go vet ./...` passes
- [x] Existing tests `TestRenderMarkdown_CodeBlockDarkBg` and `TestRenderMarkdown_CodeBlockLightBg`
      still pass or are updated if assertions need adjustment

---

## References

- `model.go:86-96` — `customizeHeaders` (add `StylePrimitive` background here)
- `model.go:114-122` — `customizeHeadersLight` (add `StylePrimitive` background here)
- `model.go:35-61` — `renderMarkdown` (Part 2: add `WithChromaFormatter` here)
- `~/go/pkg/mod/github.com/charmbracelet/glamour@v0.10.0/ansi/codeblock.go:130` — indent writer that
  uses `StylePrimitive` when rendering left-margin spaces
- `~/go/pkg/mod/github.com/alecthomas/chroma/v2@v2.14.0/formatters/tty_indexed.go:226-235` — `clearBackground()` that strips Chroma backgrounds in `terminal256`
- `~/go/pkg/mod/github.com/charmbracelet/glamour@v0.10.0/ansi/style.go:38-59` — `StylePrimitive` struct definition
