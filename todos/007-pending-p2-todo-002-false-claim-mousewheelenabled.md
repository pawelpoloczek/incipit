---
status: pending
priority: p2
issue_id: "007"
tags: [code-review, documentation, bubbletea, viewport]
---

# Correct False Claim in todos/002: MouseWheelEnabled Default

## Problem Statement

`todos/002-pending-p2-viewport-initialization-fields.md` line 16 states:

> `MouseWheelEnabled = true` — required for mouse wheel scrolling (**NOT set by default**)

This claim is **factually incorrect** for `github.com/charmbracelet/bubbles v1.0.0`. The `viewport.New()` constructor calls `setInitialValues()` which explicitly sets `MouseWheelEnabled = true` and `MouseWheelDelta = 3`. The field is affirmatively initialized, not zero-valued.

Leaving this false claim in the repository risks:
- Future maintainers adding redundant explicit assignments "for safety"
- Incorrect documentation being referenced if the pattern is reused
- The plan (2026-02-19-chore-close-pending-todos-plan.md) proposing an unnecessary code change

## Findings

- `todos/002` line 16: "NOT set by default" — incorrect for bubbles v1.0.0
- `bubbles@v1.0.0/viewport/viewport.go` `setInitialValues()` — sets `MouseWheelEnabled = true` explicitly
- `model.go:168` — `viewport.New(...)` already sets the field correctly via constructor
- The plan's proposed fix ("add `m.viewport.MouseWheelEnabled = true`") is redundant and should not be implemented

## Proposed Solution

Update `todos/002` work log with a correction note, then mark it done:

```markdown
## Work Log

- 2026-02-18: Finding identified during plan review. Added to gotchas list in plan.
- 2026-02-19: Closing. All three fields confirmed present. Note: the problem statement's
  claim that MouseWheelEnabled is "NOT set by default" is incorrect for bubbles v1.0.0 —
  viewport.New() calls setInitialValues() which sets MouseWheelEnabled = true explicitly.
  No code change required.
```

Also update `docs/plans/2026-02-19-chore-close-pending-todos-plan.md` to remove the
proposed `m.viewport.MouseWheelEnabled = true` code change (the "⚠️ gap" claim is wrong).

## Acceptance Criteria

- [ ] `todos/002` work log notes the incorrect claim and explains the actual bubbles v1.0.0 behavior
- [ ] `todos/002` status updated to `done`
- [ ] Plan file no longer proposes adding `m.viewport.MouseWheelEnabled = true`
- [ ] No code change made to `model.go` for this item

## Work Log

- 2026-02-19: Found during architecture review of the pending-todos close plan. Reviewer verified against bubbles v1.0.0 source on disk.
