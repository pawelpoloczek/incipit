# incipit

A terminal markdown reader built with Go and the Charmbracelet stack.

## Commands

```bash
# Build
go build -o incipit .

# Run
go run . README.md
go run . --dark README.md
go run . --light README.md
go run . --no-pager README.md
go run . --no-color README.md

# Test
go test ./...

# Vet
go vet ./...

# Install to $GOPATH/bin
go install .
```

## Cross-compile

```bash
make release   # builds dist/ for darwin/linux amd64/arm64
```
