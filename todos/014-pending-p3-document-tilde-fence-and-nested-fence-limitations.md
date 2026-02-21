---
status: pending
priority: p3
issue_id: "014"
tags: [code-review, documentation, quality]
---

# Document Tilde Fence and Nested Fence Limitations in addLanguageLabels

## Problem Statement

The `addLanguageLabels` regex handles only backtick fences (`` ``` ``). Two known limitations are not documented:

1. **Tilde fences (`~~~go`)** — CommonMark allows `~` as an alternate fence character. The regex silently ignores them (no label is injected). This is acceptable but should be a deliberate, documented decision.

2. **Nested fences in documentation** — A README that *shows examples of fenced code syntax* (like incipit's own README) contains lines matching the fence pattern inside what is semantically an outer code block. The regex has no state for "currently inside a code block," so it will inject a label on those inner fence lines. This edge case exists in any document that contains Markdown documentation of Markdown.

## Proposed Solution

Add a comment to the `addLanguageLabels` function:

```go
// addLanguageLabels inserts an inline code span label before each fenced
// code block that declares a language (e.g. ```go → `go` above the block).
// Blocks with no language specifier (plain ```) are left unchanged.
//
// Known limitations (by design):
//   - Tilde fences (~~~go) are not matched; no label is injected for them.
//   - Documents that show fenced code syntax inside a code block (e.g. a README
//     documenting Markdown) may have labels injected at the inner fence lines.
//     This is a known edge case for documentation-of-documentation content.
//   - The function is not idempotent: calling it twice on the same string
//     produces double labels. Always pass m.rawMarkdown (the original source),
//     never a previously-rendered or previously-preprocessed string.
var fencedLang = regexp.MustCompile(...)
```

## Acceptance Criteria

- [ ] Comment on `addLanguageLabels` lists tilde fence limitation
- [ ] Comment lists nested-fence/documentation edge case
- [ ] Comment lists non-idempotency constraint

## Work Log

- 2026-02-19: Found during architecture and agent-native review of the code block styling plan.
