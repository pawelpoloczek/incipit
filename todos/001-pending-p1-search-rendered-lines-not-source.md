---
status: pending
priority: p1
issue_id: "001"
tags: [code-review, architecture, search]
---

# Search: Use Rendered Lines, Not Raw Markdown

## Problem Statement

The original plan described searching `rawMarkdown` source lines and mapping source line indices to rendered viewport lines. This approach is **architecturally unsound** and will navigate to wrong lines on any document with code blocks, tables, headings, or long paragraphs.

Glamour does not preserve a 1:1 mapping between source lines and rendered lines:
- A heading (`## Foo`) expands to ~4 rendered lines (blank + text + rule + blank)
- Code fence delimiters (` ``` `) produce zero rendered lines
- Table separator rows (`|---|---|`) produce zero rendered lines
- Long paragraphs word-wrap to N rendered lines (changes on every resize)
- Glamour injects blank lines between elements that don't correspond to source lines

## Correct Implementation

Search the **ANSI-stripped rendered output**. Store rendered line indices directly.

```go
// After each glamour render:
m.renderedContent = rendered
m.viewport.SetContent(rendered)
m.searchLines = strings.Split(stripANSI(rendered), "\n")

// On search submit:
func computeMatches(lines []string, query string) []int {
    lower := strings.ToLower(query)
    var result []int
    for i, line := range lines {
        if strings.Contains(strings.ToLower(line), lower) {
            result = append(result, i)
        }
    }
    return result
}
m.matchLines = computeMatches(m.searchLines, m.searchQuery)

// To jump to match:
m.viewport.GotoTop()
m.viewport.LineDown(m.matchLines[m.matchIdx])
```

**ANSI stripping:** Check if `github.com/muesli/reflow/ansi` is already a transitive dep of glamour before adding a new dep. It almost certainly is.

## Acceptance Criteria

- [ ] Search finds matches at the correct viewport line in documents with headings, code blocks, and tables
- [ ] n/N cycling navigates accurately across all element types
- [ ] Search results remain correct after terminal resize

## Work Log

- 2026-02-18: Finding identified during plan review. Plan updated to reflect correct approach.
