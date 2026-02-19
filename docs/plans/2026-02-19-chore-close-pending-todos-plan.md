---
title: "chore: Verify and close 3 pending todos (search, viewport, agent-native flags)"
type: chore
status: active
date: 2026-02-19
todos: [todos/001, todos/002, todos/003]
---

# chore: Verify and close 3 pending todos

## Overview

All 3 pending todos are already fully implemented. No code changes required. Close all three by updating their `status` field to `done`.

---

## Todo 001 — Search: Use Rendered Lines Not Raw Markdown

**Status: DONE.** Verified by `go test ./...`. No changes required.

---

## Todo 002 — Viewport Initialization: Three Required Fields

**Status: DONE.** All three fields are satisfied:

| Field | Where | Status |
|---|---|---|
| `m.viewport.YPosition = headerLines` | `model.go:169` | ✅ Explicit |
| `m.viewport.MouseWheelEnabled = true` | set by `viewport.New()` in bubbles v1.0.0 | ✅ Constructor default |
| `tea.WithMouseCellMotion()` | `main.go:71` | ✅ Explicit |

Note: the todo's problem statement incorrectly claims `MouseWheelEnabled` is "NOT set by default." In bubbles v1.0.0, `viewport.New()` calls `setInitialValues()` which sets it to `true` explicitly. No code change needed. Update the work log in `todos/002` to note this when closing.

---

## Todo 003 — Add --no-pager and --no-color Flags

**Status: DONE.** Verified by `go test ./...`. No changes required.

---

## Implementation Tasks

- [ ] Update `todos/001` status to `done`
- [ ] Update `todos/002` status to `done` and add work log note correcting the `MouseWheelEnabled` claim
- [ ] Update `todos/003` status to `done`
