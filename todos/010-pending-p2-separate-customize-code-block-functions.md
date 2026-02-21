---
status: pending
priority: p2
issue_id: "010"
tags: [code-review, architecture, plan]
---

# Extract CodeBlock Customization into Separate Functions

## Problem Statement

The plan proposes adding `s.CodeBlock` overrides inside `customizeHeaders` and `customizeHeadersLight`. These functions are named and scoped for heading styles. Adding code block behavior to them violates their stated responsibility, makes them harder to read, and creates a maintenance hazard when the header styles are modified in future iterations.

## Proposed Solution

Add two new sibling functions called from the same `renderMarkdown` switch arms:

```go
// model.go — in renderMarkdown switch:
case "dark":
    s := glamourstyles.DarkStyleConfig
    customizeHeaders(&s)
    customizeCodeBlock(&s)       // add this
    styleOpt = glamour.WithStyles(s)
case "light":
    s := glamourstyles.LightStyleConfig
    customizeHeadersLight(&s)
    customizeCodeBlockLight(&s)  // add this
    styleOpt = glamour.WithStyles(s)

// New functions:
func customizeCodeBlock(s *glamouransi.StyleConfig) {
    s.CodeBlock.Chroma = &glamouransi.Chroma{}
    // copy existing Chroma fields, override Background only:
    s.CodeBlock.Chroma.Background = glamouransi.StylePrimitive{
        BackgroundColor: sp("#1e2030"),
    }
}

func customizeCodeBlockLight(s *glamouransi.StyleConfig) {
    s.CodeBlock.Chroma = &glamouransi.Chroma{}
    s.CodeBlock.Chroma.Background = glamouransi.StylePrimitive{
        BackgroundColor: sp("#f6f8fa"),
    }
}
```

**Note:** The existing `DarkStyleConfig.CodeBlock.Chroma` already has all token fields populated. Replacing the entire `Chroma` pointer would erase all syntax highlighting colors. The correct approach is to copy the existing Chroma struct and override only the Background field, or mutate `s.CodeBlock.Chroma.Background.BackgroundColor` directly.

```go
// Safer: mutate only the Background field
func customizeCodeBlock(s *glamouransi.StyleConfig) {
    sp := func(v string) *string { return &v }
    s.CodeBlock.Chroma.Background = glamouransi.StylePrimitive{
        BackgroundColor: sp("#1e2030"),
    }
}
```

Effort: Small | Risk: None

## Acceptance Criteria

- [ ] `customizeHeaders` and `customizeHeadersLight` contain only heading (H2-H6) overrides
- [ ] `customizeCodeBlock` and `customizeCodeBlockLight` contain only code block overrides
- [ ] Existing syntax highlighting token colors (keywords, strings, etc.) are preserved
- [ ] `go test ./...` passes

## Work Log

- 2026-02-19: Found during architecture and simplicity review of the code block styling plan.
