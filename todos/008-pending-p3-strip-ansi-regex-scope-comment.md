---
status: pending
priority: p3
issue_id: "008"
tags: [code-review, quality, search]
---

# Add Scope Comment to stripANSI Regex

## Problem Statement

`model.go:21` defines `ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)` which covers CSI (Control Sequence Introducer) sequences — the standard SGR color/style codes produced by glamour and lipgloss. However, it does not cover:

- OSC sequences (`\x1b]...ST` or `\x1b]...\x07`) — used for terminal hyperlinks, window titles
- DCS sequences (`\x1bP...ST`) — rare in CLI output
- Other non-CSI escapes

If glamour ever produces OSC hyperlinks (a feature that has been discussed upstream), `stripANSI` will leave partial escape sequences in `searchLines`, causing search results to contain garbage characters and potentially causing false-negative matches on content adjacent to a hyperlink.

This is not an active bug at glamour v0.10.0, but it is a silent assumption worth documenting.

## Proposed Solution

Add a one-line comment above the regex:

```go
// ansiEscape matches CSI sequences (SGR colors/styles from glamour/lipgloss).
// Does not strip OSC sequences (e.g. terminal hyperlinks) — acceptable for glamour v0.10.0.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
```

Effort: Trivial (2 lines) | Risk: None

## Acceptance Criteria

- [ ] Comment added above `ansiEscape` declaration in `model.go`
- [ ] `go vet ./...` passes
- [ ] No behavior change

## Work Log

- 2026-02-19: Found during architecture review of the pending-todos close plan.
