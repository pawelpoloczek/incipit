---
status: pending
priority: p2
issue_id: "011"
tags: [code-review, architecture, glamour, documentation]
---

# Document Glamour's Chroma Style Registry Write-Once Constraint

## Problem Statement

Glamour's `codeblock.go` registers the custom chroma style `"charm"` into a global singleton registry at most once per process lifetime:

```go
mutex.Lock()
_, ok := styles.Registry["charm"]
if !ok {
    styles.Register(chroma.MustNewStyle("charm", chroma.StyleEntries{
        chroma.Background: chromaStyle(rules.Chroma.Background),
        // ...
    }))
}
mutex.Unlock()
```

Once registered, the background color cannot be changed in the same process. If a future feature adds runtime theme switching (e.g., a `t` key to toggle dark/light), the second theme's code block background would be silently ignored — no error, just wrong background color.

This is not a current bug (style is fixed at startup) but is a latent trap for future development.

## Proposed Solution

Add a comment at `customizeCodeBlock` explaining the constraint:

```go
// customizeCodeBlock overrides the code block background color in the dark theme.
//
// Note: glamour registers the chroma style "charm" into a global singleton
// (styles.Registry) at most once per process. If the application ever adds
// runtime theme switching, the second theme's background color will be silently
// ignored. Re-creating the glamour.TermRenderer per render call (as renderMarkdown
// already does) does NOT help — the registry is process-global, not renderer-scoped.
// Theme switching would require either separate chroma style names per theme or
// forking glamour's codeblock renderer.
func customizeCodeBlock(s *glamouransi.StyleConfig) {
```

## Acceptance Criteria

- [ ] Comment added to `customizeCodeBlock` (and/or `customizeCodeBlockLight`) explaining the singleton constraint
- [ ] `go vet ./...` passes

## Work Log

- 2026-02-19: Found during architecture review of the code block styling plan.
