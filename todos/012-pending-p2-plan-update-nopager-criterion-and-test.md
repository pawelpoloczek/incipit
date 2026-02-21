---
status: pending
priority: p2
issue_id: "012"
tags: [code-review, plan, agent-native, testing]
---

# Update Plan: Add --no-pager Acceptance Criterion and Test

## Problem Statement

The plan's acceptance criteria and test list do not include any verification that language labels are absent in `--no-pager` / piped output. Once the guard from todo/009 is implemented (labels only injected in interactive mode), this behavior should be explicitly tested and captured in the plan.

## Proposed Solution

Update `docs/plans/2026-02-19-feat-code-block-background-and-language-label-plan.md`:

**Add to Acceptance Criteria:**
```markdown
- [ ] In `--no-pager` / piped mode, no language label lines appear in stdout
- [ ] `addLanguageLabels` is NOT called from inside `renderMarkdown`
```

**Add to New Tests:**
```markdown
- `TestRenderMarkdown_NoLabelInjection` — calling `renderMarkdown` directly on a fenced
  ` ```go ` block does NOT produce a `` `go` `` label line in the output (confirms the
  call site is in main.go, not renderMarkdown)
```

**Also update the Implementation section** to reflect that `addLanguageLabels` is called in `main.go` before `newModel`, not inside `renderMarkdown`.

## Acceptance Criteria

- [ ] Plan file updated with the `--no-pager` acceptance criterion
- [ ] Plan file updated with `TestRenderMarkdown_NoLabelInjection` test
- [ ] Plan file implementation section reflects `main.go` call site

## Work Log

- 2026-02-19: Found during agent-native review of the code block styling plan.
