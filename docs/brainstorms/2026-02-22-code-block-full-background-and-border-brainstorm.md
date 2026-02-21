---
title: "Code block full background and rounded border"
date: 2026-02-22
status: ready-for-planning
---

# Code block full background and rounded border

## What We're Building

Code blocks in `incipit` should render with:
1. A **full background** covering the entire code block area (not just the left indent strip)
2. A **rounded border** drawn around the whole code block, with the language name embedded in the top border line

Desired visual result:
```
╭── go ──────────────────────────────────────────────────────────╮
│                                                                │
│  func main() {                                                 │
│      fmt.Println("hello")                                      │
│  }                                                             │
│                                                                │
╰────────────────────────────────────────────────────────────────╯
```

## Why This Approach (Pre-extraction)

Glamour renders the entire document as one ANSI string. Its code block renderer uses an `indent.Writer` which cannot pad lines to terminal width or draw borders — glamour has no native support for bordered code blocks.

**Chosen approach: pre-extraction with placeholders**

1. Before passing markdown to glamour, extract all fenced code blocks via regex into a slice
2. Replace each in the markdown with a unique plaintext placeholder paragraph (e.g., `CODEBLOCK_PLACEHOLDER_0`)
3. Glamour renders the placeholder markdown (prose only — no code blocks)
4. Each extracted code block is rendered independently:
   - Syntax highlighting via chroma (same library glamour already uses)
   - Language label embedded in the top border line
   - Background + padding applied to each content line
   - Rounded border drawn using box-drawing characters
5. In the glamour output, find placeholder lines and replace them with the styled code block

This gives full control over code block appearance without fighting glamour's rendering pipeline.

## Key Decisions

| Decision | Choice | Reason |
|---|---|---|
| Approach | Pre-extraction with placeholders | Full rendering control; follows existing preprocessor pattern |
| Border style | Rounded (`╭╮╰╯`) | User's preference; softer aesthetic in terminal |
| Language label position | Embedded in top border line (`╭── go ──╮`) | User's preference; avoids extra line above block |
| Code block width | Fixed to `renderMarkdown` width parameter | Consistent with viewport; avoids variable-width boxes |
| Background color | Neutral: dark `"235"` (#262626), light `"254"` (#e4e4e4) | Code-editor feel; readable; avoids color distraction inside the box |
| Syntax highlighting | Chroma (already a glamour transitive dependency) | No new dependencies; same output quality |
| `addLanguageLabels` | Remove or bypass for code blocks | The new renderer handles language labels internally; no more injected backtick spans |

## Architecture

### New function: `extractCodeBlocks(md string) (prose string, blocks []codeBlock)`

Extracts fenced code blocks from raw markdown, replaces each with a placeholder paragraph, and returns the modified markdown plus a slice of extracted blocks.

```go
type codeBlock struct {
    lang string  // language specifier (empty string if none)
    code string  // raw code content (without the fence lines)
}
```

### New function: `renderCodeBlock(cb codeBlock, width int, style string) string`

Renders a single code block with:
- Chroma syntax highlighting for the code content
- Manual box-drawing for the border (language label in top line)
- Background color applied per-line

### Updated `renderMarkdown`

1. Call `extractCodeBlocks(md)` → get `prose` + `blocks`
2. Render `prose` with glamour as before
3. For each placeholder line in the output, replace with `renderCodeBlock(...)`
4. Return reassembled output

### Placeholder detection

The placeholder `CODEBLOCK_PLACEHOLDER_0` is inserted as a standalone paragraph in the prose markdown. Glamour renders it as a line like:

```
  CODEBLOCK_PLACEHOLDER_0
```

Detection: split rendered output by `\n`, strip ANSI from each line, match lines containing `CODEBLOCK_PLACEHOLDER_N`.

## Open Questions

All resolved:
- Border style → rounded ✓
- Label position → inside top border (`╭── go ──╮`) ✓
- Approach → pre-extraction ✓
- Background color → neutral dark `"235"` / light `"254"` ✓
- Padding → 1 blank line top and bottom, 1 space horizontal (as shown in mockup) ✓
- No-language blocks → plain top border, no label: `╭────────╮` ✓

**For plan/implementation:**
- Whether to call chroma's Go API directly or via glamour's internal helpers (check availability)

## Files to Change

- `model.go` — all changes here:
  - Add `codeBlock` struct
  - Add `extractCodeBlocks()` function (replaces / subsumes `addLanguageLabels`)
  - Add `renderCodeBlock()` function
  - Update `renderMarkdown()` to use the new pipeline
  - Remove `addLanguageLabels` and `fencedLang` regex (superseded by `extractCodeBlocks` + `renderCodeBlock`)
  - Keep `customizeHeaders` / `customizeHeadersLight` (still needed for heading styles)
  - Remove `CodeBlock.StylePrimitive.BackgroundColor` lines (the indent strip is no longer needed)

## References

- `model.go:21-33` — `fencedLang` regex and `addLanguageLabels` (will be superseded)
- `model.go:35-61` — `renderMarkdown` (entry point to update)
- `model.go:69-95` — `customizeHeaders` (remove CodeBlock StylePrimitive change)
- `model.go:97-123` — `customizeHeadersLight` (remove CodeBlock StylePrimitive change)
- `~/go/pkg/mod/github.com/charmbracelet/glamour@v0.10.0/ansi/codeblock.go` — glamour code block renderer (for reference only)
- `~/go/pkg/mod/github.com/alecthomas/chroma/v2@v2.14.0/` — chroma library (for direct use in `renderCodeBlock`)
- `lipgloss` — already imported in `model.go` for header/footer styling
