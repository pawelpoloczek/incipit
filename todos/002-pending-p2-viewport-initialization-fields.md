---
status: pending
priority: p2
issue_id: "002"
tags: [code-review, bubbletea, viewport]
---

# Viewport Initialization: Three Required Fields

## Problem Statement

Three viewport fields must be set at initialization that are not obvious from the bubbletea docs and are easy to miss:

1. `m.viewport.YPosition = headerLines` — prevents the viewport from visually overlapping the header bar
2. `m.viewport.MouseWheelEnabled = true` — required for mouse wheel scrolling (NOT set by default)
3. `tea.WithMouseCellMotion()` on the program — required in addition to `MouseWheelEnabled`

Missing any of these causes silent failures: overlap rendering or mouse wheel simply not working.

## Required Code

```go
// In Update, inside first tea.WindowSizeMsg:
m.viewport = viewport.New(msg.Width, msg.Height - headerLines - footerLines)
m.viewport.YPosition = headerLines         // <- required: no header overlap
m.viewport.MouseWheelEnabled = true        // <- required: mouse wheel
m.viewport.MouseWheelDelta = 3             // <- optional: lines per tick
m.viewport.SetContent(rendered)
m.ready = true

// In main():
p := tea.NewProgram(
    model,
    tea.WithAltScreen(),
    tea.WithMouseCellMotion(),             // <- required: delivers mouse events
)
```

## Acceptance Criteria

- [ ] Header and viewport content do not overlap
- [ ] Mouse wheel scrolling works in the pager
- [ ] All three fields are set in the implementation

## Work Log

- 2026-02-18: Finding identified during plan review. Added to gotchas list in plan.
