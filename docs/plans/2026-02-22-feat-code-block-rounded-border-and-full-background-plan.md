---
title: "feat: Code block rounded border and full background"
type: feat
status: completed
date: 2026-02-22
---

# feat: Code block rounded border and full background

Code blocks should render with a full-width background and a rounded border with the
language name embedded in the top border line:

```
╭── go ──────────────────────────────────────────────────────────╮
│                                                                │
│  func main() {                                                 │
│      fmt.Println("hello")                                      │
│  }                                                             │
│                                                                │
╰────────────────────────────────────────────────────────────────╯
```

This is a replacement of all prior code block visual work (Chroma hex backgrounds,
indent strip) which proved ineffective due to glamour's `terminal256` formatter
stripping background colors and the indent writer not padding to full width.

---

## Problem Statement

Glamour renders code blocks by:
1. Running Chroma's `terminal256` formatter (which strips `Background.BackgroundColor`)
2. Using an `indent.Writer` (which only adds left-margin spaces — no full-width padding)

The result: no visible background can be applied to code blocks through the glamour
style API. The language label (`addLanguageLabels`) also produces a dangling backtick span
above the block instead of an integrated design.

---

## Proposed Solution: Pre-Extraction Pipeline

Before glamour renders, extract all fenced code blocks from the raw markdown. Replace
each with a unique placeholder paragraph. Render the placeholder prose through glamour.
After rendering, find each placeholder line in the output and replace it with a custom
code block rendered independently using chroma + manual box-drawing.

```
Raw markdown
     │
     ▼
extractCodeBlocks()  ──► blocks []codeBlock
     │
     ▼
prose markdown (with INCIPIT_CODEBLOCK_N placeholders)
     │
     ▼
glamour render  ──► ANSI prose string
     │
     ▼
injectCodeBlocks()
     │  for each placeholder line:
     │    renderCodeBlock(block, width, style)
     │      ├── chroma syntax highlight
     │      ├── background fill technique
     │      └── manual box-drawing
     ▼
final rendered string
```

---

## Technical Approach

### `renderCodeBlock` — Background Fill Technique

Chroma's `terminal256` formatter emits `\x1b[0m` (reset) after every token, which
clears any background set before the token. To maintain a consistent background across
the entire line:

1. Get chroma output as a string using `quick.Highlight`
2. Replace every `\x1b[0m` in the output with `\x1b[0;48;5;Nm` (reset + background on)
   — where N is the 256-color index for the block background (235 dark / 254 light)
3. Prefix each content line with `\x1b[48;5;Nm` to ensure background starts on
4. Pad the visible portion to `innerWidth` with background-colored spaces
5. Append `\x1b[0m` to close

This ensures every character cell on every content line has the background color,
including cells between syntax tokens.

### Box Drawing

```
╭── go ─────────────────────────╮    ← top border, title embedded
│                               │    ← blank line (top padding)
│  func main() {                │    ← content line (bg filled)
│  }                            │    ← content line (bg filled)
│                               │    ← blank line (bottom padding)
╰───────────────────────────────╯    ← bottom border
```

**Long lines:** `outerWidth = width` equals the full terminal width. Lines wider than
`innerWidth` overflow the right border — acceptable because the terminal itself would
truncate them at the same point. No truncation or ellipsis needed.

Dimensions:
- `outerWidth = width` (the renderMarkdown width parameter = terminal width)
- `innerWidth = outerWidth - 4` (2 for `│` + 1 space padding each side)
- Top border with title: `╭── ` + lang + ` ` + `─` × (outerWidth - 6 - len(lang)) + `╮`
- Top border without title: `╭` + `─` × (outerWidth - 2) + `╮`
- Content line: `│ ` + [chroma line, bg-filled to innerWidth] + ` │`
- Blank line: `│ ` + [bg-filled spaces, innerWidth] + ` │`
- Bottom: `╰` + `─` × (outerWidth - 2) + `╯`

Border characters are styled with lipgloss `Foreground` using the same accent color as
the heading styles (dark: `"23"` teal, light: `"27"` blue). In `notty` style, borders are
rendered without color.

### Chroma Integration

```go
// import path: github.com/alecthomas/chroma/v2/quick
// Chroma is a transitive dep via glamour. Promote to direct with:
//   go get github.com/alecthomas/chroma/v2

import "github.com/alecthomas/chroma/v2/quick"

func syntaxHighlight(code, lang, style string) string {
    var buf strings.Builder
    chromaStyle := chromaStyleName(style)      // "monokai" dark, "github" light
    formatter := "terminal256"
    _ = quick.Highlight(&buf, code, lang, formatter, chromaStyle)
    return buf.String()
}

func chromaStyleName(style string) string {
    switch style {
    case "light":
        return "github"
    default:
        return "monokai"
    }
}
```

### Placeholder Detection

`extractCodeBlocks` inserts `INCIPIT_CODEBLOCK_0`, `INCIPIT_CODEBLOCK_1`, etc. as
standalone paragraphs. Glamour renders each as a paragraph line (typically with
document margin spaces). Detection in `injectCodeBlocks`:

```go
for i, line := range lines {
    plain := stripANSI(line)
    for j, cb := range blocks {
        if strings.Contains(plain, fmt.Sprintf("INCIPIT_CODEBLOCK_%d", j)) {
            lines[i] = renderCodeBlock(cb, width, style)
        }
    }
}
```

---

## Files to Change

### `model.go`

#### Remove
- `fencedLang` package-level regex (line 22)
- `addLanguageLabels` function (lines 31-33)
- `md = addLanguageLabels(md)` call at top of `renderMarkdown` (line 36)
- `s.CodeBlock.StylePrimitive.BackgroundColor = sp("23")` in `customizeHeaders` (line 91)
- `s.CodeBlock.StylePrimitive.BackgroundColor = sp("195")` in `customizeHeadersLight` (line 119)
- Chroma hex background overrides in both customize functions (the `chromaCopy` blocks) —
  the new renderer handles code blocks entirely outside glamour

#### Add
- `type codeBlock struct` with `lang string` and `code string` fields
- `extractCodeBlocks(md string) (prose string, blocks []codeBlock)` — uses the regex:
  ```
  (?ms)^`{3}([a-zA-Z][a-zA-Z0-9_+-]*)?\n(.*?)^`{3}[^\S\r\n]*$
  ```
  Flags: `m` (multiline `^`/`$`) + `s` (dotall `.` matches `\n`). Group 1 = language, group 2 = code content.
- `syntaxHighlight(code, lang, chromaStyle string) string` — calls `quick.Highlight`
- `chromaStyleName(style string) string` — maps dark/light/notty to chroma style names
- `renderCodeBlock(cb codeBlock, width int, style string) string` — assembles the full bordered box
- `injectCodeBlocks(rendered string, blocks []codeBlock, width int, style string) string`

#### Modify
- `renderMarkdown`: call `extractCodeBlocks`, render prose, call `injectCodeBlocks`
- Add `"github.com/alecthomas/chroma/v2/quick"` import
- `customizeHeaders` / `customizeHeadersLight`: remove all CodeBlock overrides (no longer needed)

### `model_test.go`

#### Remove / update
- `TestAddLanguageLabels_WithLang` — `addLanguageLabels` is removed
- `TestAddLanguageLabels_NoLang` — same
- `TestAddLanguageLabels_NoFence` — same
- `TestCustomizeHeaders_CodeBlockDarkIndentStrip` — indent strip is removed
- `TestCustomizeHeadersLight_CodeBlockLightIndentStrip` — same
- `TestRenderMarkdown_CodeBlockDarkBg` / `TestRenderMarkdown_CodeBlockLightBg` — update assertions

#### Add
- `TestExtractCodeBlocks_WithLang` — verify lang and code extracted, placeholder inserted
- `TestExtractCodeBlocks_NoLang` — verify no-lang block extracted, no label
- `TestExtractCodeBlocks_Multiple` — two code blocks get distinct placeholders
- `TestExtractCodeBlocks_NoBlocks` — prose-only markdown unchanged
- `TestRenderCodeBlock_ContainsBorder` — output contains `╭` and `╰`
- `TestRenderCodeBlock_ContainsCode` — output contains the code text (strip ANSI)
- `TestRenderCodeBlock_WithLang_TitleInBorder` — output contains `── go ──`
- `TestRenderCodeBlock_NoLang_PlainBorder` — top border is plain `╭─…─╮`
- `TestInjectCodeBlocks_ReplacesPlaceholder` — placeholder line is replaced
- `TestRenderMarkdown_CodeBlock_EndToEnd` — full render of markdown with code block contains `╭`

---

## Acceptance Criteria

- [ ] Dark mode: rounded border visible around all code blocks (`╭` / `╰` present)
- [ ] Dark mode: full background fill inside the box (`"235"` 256-color dark gray)
- [ ] Light mode: corresponding border and background (`"254"`)
- [ ] Language label embedded in top border line: `╭── go ──────╮`
- [ ] Code block with no language: plain top border `╭──────╮`, no label
- [ ] Syntax highlighting preserved inside the box (chroma `"monokai"` dark / `"github"` light)
- [ ] 1 blank line padding above and below code content inside the box
- [ ] Box spans full `width` passed to `renderMarkdown`
- [ ] Inline code spans unaffected
- [ ] Prose (paragraphs, headings, lists) unaffected
- [ ] `notty` style: border rendered without color (box-drawing characters preserved)
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] `addLanguageLabels` and `fencedLang` removed (no dead code)

---

## Dependencies & Risks

| Risk | Mitigation |
|---|---|
| `chroma/v2/quick` not a direct dep | `go get github.com/alecthomas/chroma/v2` promotes it before adding import |
| Placeholder survives glamour word-wrap | Placeholder is 20 chars max — well within typical width; glamour renders it as its own paragraph line |
| Nested backticks in code content | Full-block regex uses lazy match and requires closing fence on its own line |
| `\x1b[0m` replacement changes non-code ANSI | Replacement only applied to the chroma output string, not to the full rendered prose |
| `notty` mode chroma output | Use `"noop"` chroma formatter for plain text; draw border without color |

---

## References

- `model.go:22` — `fencedLang` regex (to remove)
- `model.go:31-33` — `addLanguageLabels` (to remove)
- `model.go:35-61` — `renderMarkdown` (to restructure)
- `model.go:69-97` — `customizeHeaders` (remove CodeBlock overrides)
- `model.go:99-127` — `customizeHeadersLight` (remove CodeBlock overrides)
- `~/go/pkg/mod/github.com/alecthomas/chroma/v2@v2.14.0/quick/quick.go` — `quick.Highlight` API
- `~/go/pkg/mod/github.com/alecthomas/chroma/v2@v2.14.0/formatters/tty_indexed.go` — `"terminal256"` formatter
- `docs/brainstorms/2026-02-22-code-block-full-background-and-border-brainstorm.md` — design decisions
