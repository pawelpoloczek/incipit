---
status: pending
priority: p2
issue_id: "013"
tags: [code-review, plan, quality]
---

# Update Plan: Remove Ambiguous Sibling Function Alternative and Fragile Tests

## Problem Statement

Two simplification issues in the plan document:

**Issue 1 — False optionality on line 29:**
The plan says: "Extend `customizeHeaders` / `customizeHeadersLight` **(or add sibling functions `customizeCode` / `customizeCodeLight`)**"

The parenthetical presents two options without recommending one. The correct answer is clear (separate functions — see todo/010), but leaving both options in the plan creates confusion for the implementer and suggests the decision hasn't been made.

**Issue 2 — Two fragile render-layer tests (lines 120-121):**
`TestRenderMarkdown_CodeBlockDarkBg` and `TestRenderMarkdown_CodeBlockLightBg` test that a config struct field was set, not that user-visible behavior works correctly. These tests:
- Depend on glamour's internal rendering pipeline emitting color values in a predictable string format
- Are sensitive to test ordering due to the Chroma style registry singleton (todo/011)
- Add maintenance burden without catching real bugs

## Proposed Solution

Update `docs/plans/2026-02-19-feat-code-block-background-and-language-label-plan.md`:

1. Line 29: Replace "Extend `customizeHeaders` / `customizeHeadersLight` (or add sibling functions...)" with "Add `customizeCodeBlock` and `customizeCodeBlockLight` functions, called alongside the existing header functions in `renderMarkdown`."

2. Lines 120-121: Remove `TestRenderMarkdown_CodeBlockDarkBg` and `TestRenderMarkdown_CodeBlockLightBg` from the test list. The three `TestAddLanguageLabels_*` tests cover all real risk.

## Acceptance Criteria

- [ ] Plan line 29 states the chosen approach (separate functions), no "or" alternative
- [ ] `TestRenderMarkdown_CodeBlockDarkBg` removed from plan test list
- [ ] `TestRenderMarkdown_CodeBlockLightBg` removed from plan test list

## Work Log

- 2026-02-19: Found during simplicity review of the code block styling plan.
