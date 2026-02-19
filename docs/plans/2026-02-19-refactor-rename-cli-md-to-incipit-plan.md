---
title: "refactor: Rename project from cli-md to incipit"
type: refactor
status: active
date: 2026-02-19
brainstorm: docs/brainstorms/2026-02-19-project-name-brainstorm.md
---

# refactor: Rename project from cli-md to incipit

Rename every occurrence of `cli-md` to `incipit` across the codebase. Pure find-and-replace — no logic changes.

## Acceptance Criteria

- [ ] `go build -o incipit .` produces a working binary
- [ ] `./incipit README.md` opens the pager correctly
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] No remaining `cli-md` references in source files (docs/plans and docs/brainstorms excluded — they are historical records)

## Files to Change

### `go.mod`
```
module github.com/pawelpoloczek/cli-md  →  module github.com/pawelpoloczek/incipit
```

### `main.go`
```
"Usage: cli-md [...]"              →  "Usage: incipit [...]"
"cli-md: --dark and --light..."    →  "incipit: --dark and --light..."
"cli-md: %s\n"  (×2)              →  "incipit: %s\n"
```

### `CLAUDE.md`
```
# cli-md             →  # incipit
go build -o cli-md . →  go build -o incipit .
go run . README.md   (unchanged)
go run . --dark ...  (unchanged)
```

### `.gitignore`
```
/cli-md  →  /incipit
```

### `Makefile`
```
go build -o cli-md .              →  go build -o incipit .
dist/cli-md-darwin-amd64          →  dist/incipit-darwin-amd64
dist/cli-md-darwin-arm64          →  dist/incipit-darwin-arm64
dist/cli-md-linux-amd64           →  dist/incipit-linux-amd64
dist/cli-md-linux-arm64           →  dist/incipit-linux-arm64
rm -f cli-md                      →  rm -f incipit
```

### `README.md`
All occurrences: `cli-md` → `incipit`
- Title `# cli-md` → `# incipit`
- Usage line `cli-md [options] <file.md>` → `incipit [options] <file.md>`
- `go install github.com/pawelpoloczek/cli-md@latest` → `go install github.com/pawelpoloczek/incipit@latest`
- Clone URL and `cd cli-md` → `cd incipit`
- All example invocations

### `todos/005-pending-p1-non-pager-hardcoded-80-column-width.md`
Update example invocations in the problem statement and acceptance criteria.

## Post-rename Steps

- [ ] Delete the old built binary `/cli-md` from the project root (if present)
- [ ] Run `go mod tidy` after updating `go.mod`
- [ ] Rename the GitHub repository on GitHub: Settings → Repository name → `incipit`
- [ ] Update the local remote URL: `git remote set-url origin git@github.com:pawelpoloczek/incipit.git`

## Out of Scope

Historical documents (`docs/brainstorms/`, `docs/plans/`) retain `cli-md` references — they are records of past work and do not need updating.
