---
status: pending
priority: p1
issue_id: "004"
tags: [code-review, agent-native, flags, spec-compliance]
---

# NO_COLOR: Use os.LookupEnv for Spec Compliance

## Problem Statement

`main.go:13` checks `os.Getenv("NO_COLOR") != ""` which misses the case where `NO_COLOR` is exported but set to the empty string (`export NO_COLOR=`). The [no-color.org spec](https://no-color.org/) states the variable should suppress color "regardless of its value" — including an empty value. The spec-compliant check is `os.LookupEnv("NO_COLOR")` which returns a `bool` indicating presence rather than checking the value.

An agent running in a container that sets `NO_COLOR=` (empty string convention) will still receive ANSI-colored output, breaking downstream text processing.

## Findings

- `main.go:13` — `os.Getenv("NO_COLOR") != ""` — misses `NO_COLOR=""`
- `model_test.go:124-153` — no test exercises the `NO_COLOR` env var branch at all
- The flag-based path (`noColor bool` param) is correctly tested but the env path is untested

## Proposed Solutions

### Option A: Switch to os.LookupEnv (Recommended)

```go
// main.go:12-23 — chooseStyle
func chooseStyle(dark, light, noColor bool) string {
    _, noColorEnv := os.LookupEnv("NO_COLOR")
    if noColor || noColorEnv {
        return "notty"
    }
    ...
}
```

Pros: Spec-compliant, one-line change, zero new imports
Cons: None
Effort: Small | Risk: None

### Option B: Check both empty and non-empty

```go
if noColor || os.Getenv("NO_COLOR") != "" || func() bool { _, ok := os.LookupEnv("NO_COLOR"); return ok }() {
```

Pros: None — more verbose, same result
Cons: Ugly
Effort: Small | Risk: Low

## Acceptance Criteria

- [ ] `chooseStyle` uses `os.LookupEnv("NO_COLOR")` not `os.Getenv("NO_COLOR") != ""`
- [ ] New test `TestChooseStyle_NoColorEnv` uses `t.Setenv("NO_COLOR", "1")` and asserts `"notty"`
- [ ] New test `TestChooseStyle_NoColorEnvEmpty` uses `t.Setenv("NO_COLOR", "")` and asserts `"notty"`
- [ ] `go test ./...` passes

## Work Log

- 2026-02-19: Found during agent-native review of the pending-todos close plan. Agent-native reviewer confirmed spec deviation against no-color.org.
