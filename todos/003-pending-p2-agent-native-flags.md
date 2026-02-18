---
status: pending
priority: p2
issue_id: "003"
tags: [code-review, agent-native, flags]
---

# Add --no-pager and --no-color Flags for Agent/Pipeline Use

## Problem Statement

The non-TTY auto-detection handles most cases, but agents running in pseudo-TTY environments (Docker, `ssh -t`, `script`, some CI runners) may have TTY detection return `true` even when interactive use is not intended. Without an explicit `--no-pager` flag, the process hangs waiting for keyboard input with no way to recover except SIGKILL.

Additionally, ANSI escape codes in output break text-processing pipelines. The `NO_COLOR` env var is a widely-adopted convention for suppressing color output.

## Required Implementation

**Flag additions in `main.go`:**
```go
noPager  := flag.Bool("no-pager", false, "print rendered output without interactive pager")
noColor  := flag.Bool("no-color", false, "disable ANSI colors (also respects NO_COLOR env var)")
```

**Style selection:**
```go
func chooseStyle(dark, light, noColor bool) string {
    if noColor || os.Getenv("NO_COLOR") != "" {
        return "notty"
    }
    if dark  { return "dark" }
    if light { return "light" }
    return "dark"
}
```

**Non-pager gate:**
```go
if *noPager || !term.IsTerminal(int(os.Stdout.Fd())) {
    // render and print, then exit 0
}
```

## Acceptance Criteria

- [ ] `cli-md --no-pager README.md` prints rendered output and exits without interactive pager
- [ ] `cli-md --no-color README.md` prints plain text output with no ANSI sequences
- [ ] `NO_COLOR=1 cli-md README.md` also suppresses colors
- [ ] Both flags work in non-TTY mode too

## Work Log

- 2026-02-18: Finding from agent-native reviewer. Plan updated. Implementation reminder.
