# Header Styling Brainstorm

**Date:** 2026-02-18
**Status:** Draft

---

## Problem Statement

The first header (`# H1`) is rendered with a background color block and no markdown syntax visible. All subsequent headers (`## H2`, `### H3`, etc.) display the raw `##` markdown syntax characters with only a colored font — no background color, and the prefix markers are retained in the output. This is visually inconsistent.

**Root cause:** Glamour's built-in `dark`/`light` themes define this distinction intentionally. H1 is treated as a document title (styled prominently, marker stripped), while H2–H6 retain their markdown syntax as a visual hierarchy hint. There is no custom styling in `cli-md` — the app calls `glamour.WithStandardStyle("dark")` directly.

---

## What We're Building

Custom Glamour style JSON files (`dark.json` and `light.json`) embedded in the binary that override the default header rendering so that:

- **All** headers (H1–H6) strip their markdown prefix markers (`#`, `##`, etc.)
- **All** headers render with a background color block
- Visual hierarchy is expressed through decreasing background/foreground weight across levels H1→H6
- The `--dark`, `--light`, and `--no-color` flags continue to work as before

---

## Why This Approach

**Custom Glamour JSON styles** is the correct mechanism:
- Glamour officially supports `glamour.WithStylesFromJSONBytes()` and `glamour.WithStylePath()` for customization
- No need to rewrite the renderer or post-process ANSI output
- Both dark and light themes can be independently tuned
- Changes are isolated to style files — `model.go` needs only minor updates
- The binary stays self-contained by embedding the JSON files with `//go:embed`

---

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Customization mechanism | Custom JSON style files, embedded | Official Glamour API, no runtime file deps |
| H1–H6 prefix | Stripped (empty `""`) for all levels | Consistency — no raw markdown syntax visible |
| Visual hierarchy method | Background color blocks with decreasing weight | User's preference; most visually prominent |
| Scope | Dark and light themes both updated | `--light` flag should get equivalent treatment |
| `--no-color` / `notty` | Unchanged | Plain text mode unaffected |

---

## Proposed Visual Design (Dark Theme)

```
# Title         ███ yellow on indigo/purple ███      (H1 — current behavior preserved)
## Section      ██ bright-cyan on dark-teal ██       (H2 — new)
### Subsection  █ green on dark-green █              (H3 — new)
#### Topic      dim-blue on near-black               (H4 — new)
##### Item      muted on near-black                  (H5 — new)
###### Minor    very dim on near-black               (H6 — new)
```

Exact color values TBD during implementation — the palette should feel cohesive with the existing H1 style.

---

## Open Questions

None.

---

## Resolved Questions

- **Why is H1 different?** → Glamour built-in theme design. H1 strips `#` and adds background; H2+ retain `##` prefix and use foreground-only color.
- **Fix or rethink?** → Fix the inconsistency (not a full redesign).
- **Approach?** → Custom Glamour JSON style files embedded in binary.
- **Visual hierarchy without ## markers?** → Background color blocks for all levels with decreasing weight.
- **Dark theme palette for H2–H6?** → Cool tones: H2 bold cyan-on-teal, H3 green-on-dark-green, H4–H6 progressively dimmer blue-grays on near-black.
- **Light theme approach?** → Same background-block structure as dark theme, with pastel/lighter backgrounds appropriate for light terminals.
- **Spacing/margin?** → Keep existing spacing — only prefix and background color change.

---

## Scope / Out of Scope

**In scope:**
- Update H2–H6 rendering in dark and light Glamour style overrides
- Embed custom style JSON in the binary
- Wire `glamour.WithStylesFromJSONBytes()` into `model.go`

**Out of scope:**
- Changing H1 behavior (it already works correctly)
- New CLI flags for custom style paths
- Changing the markdown parser (goldmark)
