---
title: "feat: Code block distinct background and language label"
type: feat
status: completed
date: 2026-02-19
---

# feat: Code block distinct background and language label

Code blocks need two visual improvements:
1. A clearly distinct background color that separates code from prose (both dark and light themes)
2. A language label rendered above each fenced code block that has a language specifier

---

## Current State

- **Dark theme**: `Chroma.Background.BackgroundColor: "#373737"` — a dark gray fill. Visible but not strongly distinct from the surrounding terminal background.
- **Light theme**: `Chroma.Background.BackgroundColor: "#373737"` — **a bug**. The light theme uses the same dark background as dark theme, making code blocks look broken on light terminals.
- **Inline code**: dark=`Color:"203" bg:"236"`, light=`Color:"203" bg:"254"` — acceptable as-is.
- **Language label**: not rendered at all.

---

## Proposed Solution

### Part 1 — Background color (model.go, `customizeHeaders` pattern)

Extend `customizeHeaders` / `customizeHeadersLight` (or add sibling functions `customizeCode` / `customizeCodeLight`) to override `s.CodeBlock.Chroma.Background.BackgroundColor`.

**Dark theme** — use a dark blue-tinted background that reads as "code space":
```
Chroma.Background.BackgroundColor = "#1e2030"   // dark blue-gray (like Tokyo Night)
```

**Light theme** — use a light warm gray (GitHub-style):
```
Chroma.Background.BackgroundColor = "#f6f8fa"   // near-white cool gray
```

This fixes the light theme bug and gives both themes a visually coherent "code area".

### Part 2 — Language label (new `addLanguageLabels` preprocessor)

Glamour has no API for language labels. Approach: inject a markdown inline code span immediately before each fenced block that has a language specifier. Glamour renders the `` `lang` `` with the existing inline `Code` style (colored background), which appears as a small tag above the code block.

**Transformation (applied in `renderMarkdown` before glamour):**

Input markdown:
````
```go
func main() {}
```
````

After preprocessing:
````
`go`
```go
func main() {}
```
````

This renders as:
```
 go                      ← inline code span (colored badge)
  func main() {}         ← syntax-highlighted code block
```

**Implementation in `model.go`:**

```go
// addLanguageLabels inserts an inline code span label before each fenced
// code block that declares a language (e.g. ```go → `go` above the block).
// Blocks with no language specifier (plain ```) are left unchanged.
var fencedLang = regexp.MustCompile(`(?m)^(` + "```" + `)([a-zA-Z][a-zA-Z0-9_+-]*)\n`)

func addLanguageLabels(md string) string {
	return fencedLang.ReplaceAllString(md, "`$2`\n$1$2\n")
}
```

Call it in `renderMarkdown` before passing to glamour:
```go
func renderMarkdown(md, style string, width int) string {
    md = addLanguageLabels(md)
    // ... existing glamour rendering
}
```

---

## Files to Change

### `model.go`

1. Add `fencedLang` compiled regex (package-level, next to `ansiEscape`)
2. Add `addLanguageLabels(md string) string` function
3. Call `addLanguageLabels(md)` at top of `renderMarkdown`
4. In `customizeHeaders` — add `CodeBlock` override (dark theme bg)
5. In `customizeHeadersLight` — add `CodeBlock` override (light theme bg, fixes bug)

No other files need changes.

---

## Acceptance Criteria

- [x] Dark theme: code block background is visually distinct from the terminal background (not `#373737`)
- [x] Light theme: code block background is a light color (not the dark `#373737` from the glamour default)
- [x] A fenced code block with a language (`\`\`\`go`) renders a small `go` label above it
- [x] A fenced code block without a language (`\`\`\``) renders no label
- [x] Inline code (backtick spans) is unaffected by the `addLanguageLabels` transform
- [x] `go test ./...` passes
- [x] `go vet ./...` passes
- [x] New tests:
  - `TestAddLanguageLabels_WithLang` — input ` ```go\n ` → output contains `` `go` ``
  - `TestAddLanguageLabels_NoLang` — input ` ``` ` → output unchanged
  - `TestAddLanguageLabels_NoFence` — plain text → output unchanged
  - `TestRenderMarkdown_CodeBlockDarkBg` — dark render of fenced block contains `#1e2030` or visually distinct bg (via no raw `#373737`)
  - `TestRenderMarkdown_CodeBlockLightBg` — light render does not contain `#373737`

---

## References

- `model.go:27-52` — `renderMarkdown` (add `addLanguageLabels` call here)
- `model.go:57-100` — `customizeHeaders` (add `s.CodeBlock` override)
- `model.go:103-136` — `customizeHeadersLight` (add `s.CodeBlock` override, fix light bg)
- `model.go:21-25` — `ansiEscape` regex (add `fencedLang` alongside it)
- `~/go/pkg/mod/github.com/charmbracelet/glamour@v0.10.0/ansi/style.go:77-81` — `StyleCodeBlock` struct
- `~/go/pkg/mod/github.com/charmbracelet/glamour@v0.10.0/styles/styles.go` — `DarkStyleConfig.CodeBlock`, `LightStyleConfig.CodeBlock`
