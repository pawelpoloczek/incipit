---
status: pending
priority: p2
issue_id: "006"
tags: [code-review, testing, agent-native, flags]
---

# Add Missing Tests: NO_COLOR Env Var and Notty Fenced Code Block

## Problem Statement

Two test gaps were identified during the code review:

1. **NO_COLOR env var branch has zero test coverage.** `TestChooseStyle_*` tests (model_test.go:124-153) cover the `noColor bool` parameter but no test exercises `os.LookupEnv("NO_COLOR")` (or the current `os.Getenv`). The spec deviation in todo/004 would never be caught by CI.

2. **Notty style + fenced code block ANSI suppression is untested.** The plan's acceptance criterion "output contains no ANSI escape sequences" for `--no-color` is unverifiable. Glamour delegates syntax highlighting to chroma, which may emit its own ANSI independently of the glamour style. Without a test rendering a fenced code block under `style = "notty"` and asserting zero ANSI output, this is an unvalidated assumption.

## Findings

- `model_test.go:124-153` — five `TestChooseStyle_*` tests; none use `t.Setenv`
- `model_test.go` has no test rendering a fenced code block under any style
- `model.go:40` — `"notty"` path passes through `glamour.WithStandardStyle("notty")` unmodified; chroma behavior under notty is untested

## Proposed Solutions

### Option A: Add tests directly to model_test.go (Recommended)

```go
// Test 1 — NO_COLOR env var (non-empty)
func TestChooseStyle_NoColorEnv(t *testing.T) {
    t.Setenv("NO_COLOR", "1")
    if got := chooseStyle(false, false, false); got != "notty" {
        t.Errorf("expected notty, got %q", got)
    }
}

// Test 2 — NO_COLOR env var (empty string — spec compliance)
func TestChooseStyle_NoColorEnvEmpty(t *testing.T) {
    t.Setenv("NO_COLOR", "")
    if got := chooseStyle(false, false, false); got != "notty" {
        t.Errorf("expected notty when NO_COLOR= (empty), got %q", got)
    }
}

// Test 3 — notty style fenced code block produces no ANSI
func TestRenderMarkdown_NottyCodeBlockNoANSI(t *testing.T) {
    md := "```go\nfunc main() {}\n```\n"
    out := renderMarkdown(md, "notty", 80)
    stripped := stripANSI(out)
    if out != stripped {
        t.Errorf("notty style emitted ANSI sequences in fenced code block output")
    }
}
```

Pros: Uses `t.Setenv` (auto-restores env), minimal boilerplate, directly tests the risk
Cons: Test 2 will fail until todo/004 is fixed (LookupEnv change)
Effort: Small | Risk: None

## Acceptance Criteria

- [ ] `TestChooseStyle_NoColorEnv` passes with `t.Setenv("NO_COLOR", "1")`
- [ ] `TestChooseStyle_NoColorEnvEmpty` passes with `t.Setenv("NO_COLOR", "")` — requires todo/004 fix first
- [ ] `TestRenderMarkdown_NottyCodeBlockNoANSI` passes (zero ANSI in notty fenced code output)
- [ ] `go test ./...` passes

## Dependencies

- todo/004 (NO_COLOR LookupEnv fix) must be completed before `TestChooseStyle_NoColorEnvEmpty` can pass

## Work Log

- 2026-02-19: Found during agent-native and architecture review of the pending-todos close plan.
