# Brainstorm: Pill-Shaped Header Backgrounds

**Date:** 2026-02-22
**Status:** Ready for planning

---

## What We're Building

Replace the current full-width colored backgrounds on markdown headers (H1–H6) with tight, pill-shaped backgrounds that hug the text. The colored background should only cover the header text plus a small amount of horizontal padding on each side — no full-width fill, no box-drawing borders.

**Visual target:**

```
  ██ My Header ██
  (background fits the text tightly)
```

---

## Why This Approach

### Chosen: Custom header rendering (Approach A)

Extract headers from markdown before Glamour processes them, render each header with `lipgloss` using inline padding, then inject the result back into the document. This is the same pipeline already established for code blocks.

**Pros:**
- True pill shape — lipgloss renders tight inline backgrounds naturally
- Follows the existing architectural pattern in the codebase (consistent)
- Full per-level control over padding, colors, bold, etc.
- Decoupled from Glamour internals; won't break on Glamour updates

**Alternatives considered:**
- Post-processing Glamour output: fragile due to raw ANSI escape parsing
- Glamour Prefix/Suffix trick: Glamour fills background to line-end regardless

---

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Rendering pipeline | Pre-extraction + injection (like code blocks) | Consistent architecture |
| Styling library | `lipgloss` | Already a dependency; designed for tight inline styling |
| Border characters | None (background fill only) | User explicitly chose pill shape over box-drawing borders |
| Color scheme | Preserve existing per-level colors from `customizeHeaders` / `customizeHeadersLight` | Minimal visual disruption |
| H1 | Styled with pill background | Consistent across all heading levels |
| Horizontal padding | 2 spaces each side | `  Header Text  ` — balanced visual weight |
| Indentation | Flush left for all levels | Clean, uniform left edge |
| Vertical spacing | Keep current Glamour spacing | No change to surrounding blank lines |

---

## Open Questions

*(None — all resolved)*

---

## References

- Existing code block custom renderer: `renderer.go` (lines 74–159)
- Header style configuration: `renderer.go` `customizeHeaders()` / `customizeHeadersLight()`
- lipgloss: `github.com/charmbracelet/lipgloss`
