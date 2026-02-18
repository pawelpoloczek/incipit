---
title: "feat: Build CLI Markdown Reader"
type: feat
status: completed
date: 2026-02-18
---

# feat: Build CLI Markdown Reader

## Overview

Build `cli-md` — a terminal markdown reader that renders all markdown elements consistently and displays them in a built-in keyboard-driven pager. Replaces `glow`, which renders only the first header correctly. Written in Go with the Charmbracelet stack (glamour + bubbletea + viewport).

## Problem Statement

`glow` fails to render markdown consistently — only the first header (`H1`) looks correct, with subsequent headers (`H2–H6`), code blocks, tables, and links rendering poorly or without distinct styles. A simpler, direct use of `glamour` (which glow wraps) gives full control over rendering quality without glow's overhead.

## Proposed Solution

A single Go binary invoked as `cli-md README.md`. It reads the file, renders it with `glamour`, and displays it in a full-screen `bubbletea` pager with a scrollable `bubbles/viewport`. When stdout is not a TTY (piped output), it prints rendered ANSI text directly without the pager.

## Technical Approach

### Project Structure

```
cli-md/
├── main.go          # entry point: flag parsing, TTY detection, launch
├── model.go         # bubbletea Model: Init, Update, View + render helper (inlined)
├── go.mod
├── go.sum
├── .gitignore
├── CLAUDE.md        # build/test/run commands for AI-assisted development
├── Makefile         # build, install, release (cross-compile) targets
└── README.md        # usage and installation docs
```

> Note: `render.go` is intentionally omitted. The glamour helper is a ~10-line function that lives at the top of `model.go`. A separate file is premature abstraction for this size.

### Dependencies

```
github.com/charmbracelet/bubbletea   v0.27.x  (v2 is beta — use v0.27.x)
github.com/charmbracelet/bubbles     v0.20.x
github.com/charmbracelet/glamour     v0.8.x
github.com/charmbracelet/lipgloss    v1.x      (may be transitive; add if used directly)
golang.org/x/term                    v0.x      (for TTY detection — required in main.go)
```

Module path: `github.com/pawelpoloczek/cli-md`
Go version: 1.22+

### Implementation Phases

#### Phase 1: Project Scaffolding

Create the baseline project infrastructure.

**Tasks:**
- `go mod init github.com/pawelpoloczek/cli-md`
- Create `.gitignore` (Go standard: `/cli-md`, `*.exe`, `dist/`, `vendor/`)
- Create `CLAUDE.md` documenting: `go build ./...`, `go test ./...`, `go run . README.md`
- Create `Makefile` with `build`, `install`, `clean` targets
- Update `README.md` with usage, installation, keybindings

**Acceptance:** `go build ./...` produces a `cli-md` binary with no errors.

---

#### Phase 2: Core Pager

Implement the full-screen pager with markdown rendering.

**`main.go`** — Entry point

```go
// main.go
func main() {
    // 1. Parse flags: --dark, --light, --no-pager, --no-color + positional file argument
    // 2. Validate: exactly one file argument required
    // 3. Read file contents (os.ReadFile)
    // 4. Detect non-interactive mode: --no-pager flag OR !term.IsTerminal(int(os.Stdout.Fd())):
    //      render via glamour with width=80 fallback
    //      print to stdout, exit 0
    // 5. Launch bubbletea: tea.NewProgram(model, WithAltScreen(), WithMouseCellMotion())
}
```

**`model.go`** — bubbletea Model + render helper

```go
// renderMarkdown is inlined here — no separate render.go needed
func renderMarkdown(md, style string, width int) string {
    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle(style),
        glamour.WithWordWrap(width),
    )
    out, _ := r.Render(md)
    return strings.TrimRight(out, "\n")
}

type model struct {
    filename        string
    rawMarkdown     string
    glamourStyle    string     // "dark" | "light" | "ascii" (when --no-color)
    viewport        viewport.Model
    ready           bool
    lastWidth       int        // track width to avoid redundant re-renders on resize
    searching       bool
    searchQuery     string     // plain string, no textinput component needed
    searchLines     []string   // ANSI-stripped rendered lines for search
    matchLines      []int      // rendered line indices of matches
    matchIdx        int
}
```

> Note: `textinput.Model` is replaced with a plain `searchQuery string`. The search bar only needs: append character, backspace, Enter (submit), Escape (cancel). A full textinput component is over-engineering for this use case.

**Key implementation notes:**

- `viewport.New()` must be called inside the first `tea.WindowSizeMsg`, not at model construction
- Re-render via glamour on `WindowSizeMsg` **only when width changed** (`if msg.Width != m.lastWidth`) — guards against redundant renders when only height changes
- Save/restore `m.viewport.YOffset` around `SetContent` calls on resize to preserve scroll position
- `tea.WithMouseCellMotion()` required for mouse wheel scrolling
- Set `m.viewport.MouseWheelEnabled = true` at viewport initialization (does not work without this even with `WithMouseCellMotion()`)
- Set `m.viewport.YPosition = headerLines` after `viewport.New()` to prevent header/viewport overlap

**View layout (3 rows: header + viewport + footer):**

```
┌─ README.md ─────────────────────────────────────────┐  <- header (1 line)
│                                                      │
│  [rendered markdown via viewport]                   │  <- viewport (terminal height - 2)
│                                                      │
└─ ↑/k ↓/j  g/G  / search  q quit          37%  ─────┘  <- footer (1 line)
```

When search is active, footer is replaced by the textinput: `/ query_`

---

#### Phase 3: Keybindings

All keyboard shortcuts in `Update()`:

| Key | Action | Source |
|-----|---------|--------|
| `↑` / `k` | scroll up 1 line | viewport built-in |
| `↓` / `j` | scroll down 1 line | viewport built-in |
| `pgup` / `b` | page up | viewport built-in |
| `pgdn` / `f` / `space` | page down | viewport built-in |
| `ctrl+u` | half page up | viewport built-in |
| `ctrl+d` | half page down | viewport built-in |
| `g` | go to top | `m.viewport.GotoTop()` |
| `G` | go to bottom | `m.viewport.GotoBottom()` |
| `/` | enter search mode | custom |
| `n` | next search match | custom |
| `N` | previous search match | custom |
| `esc` | cancel search, return to normal mode | custom (search mode only) |
| `q` / `ctrl+c` | quit (normal mode only) | `tea.Quit` |

**Critical:** When `m.searching == true`, return from `Update` early (before forwarding to viewport) to prevent `j`/`k`/`q` from firing while typing the search query.

---

#### Phase 4: Search

Search operates on the **ANSI-stripped rendered output** (not raw markdown). This is the only correct approach — glamour does not preserve a 1:1 mapping between source lines and rendered lines (headings expand to 4-6 rendered lines, code fences vanish, table separators disappear, paragraphs word-wrap, etc.).

**Algorithm:**
1. After each `glamour.Render()` call, store the result in `m.renderedContent`
2. Strip ANSI escape codes from `m.renderedContent` → `strippedContent`
3. Split `strippedContent` by `\n` → store as `m.searchLines []string`
4. On search query submit: case-insensitive substring scan of `m.searchLines`
5. Store matching **rendered line indices** in `m.matchLines []int`
6. Jump: `m.viewport.GotoTop()` + `m.viewport.LineDown(m.matchLines[m.matchIdx])`
7. Recompute `m.matchLines` on every resize if `m.searchQuery != ""`

**ANSI stripping:** Use `github.com/muesli/reflow/ansi` (already a transitive dependency of glamour — no extra dep needed). Verify with `go mod graph | grep reflow`.

**Search input handling (no `textinput` component):**
```go
case tea.KeyMsg:
    if m.searching {
        switch {
        case msg.Type == tea.KeyEnter:
            if m.searchQuery == "" { /* treat as cancel */ }
            // compute matches, exit search mode
        case msg.Type == tea.KeyEsc:
            // clear query, exit search mode
        case msg.Type == tea.KeyBackspace:
            if len(m.searchQuery) > 0 {
                m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
            }
        default:
            m.searchQuery += string(msg.Runes)
        }
    }
```

**UX states:**
- **Empty query + Enter:** treat as Escape (cancel)
- **No matches:** show `"no matches: <query>"` in footer
- **Enter with matches:** jump to first match, show `"1/N: <query>"` in footer
- **Escape:** clear query and `m.matchLines`, return to normal footer
- **n/N:** cycle through `m.matchLines` (wraps around); `n/N` remain active after Escape if matches exist
- **Resize while search active:** re-render glamour, recompute `m.searchLines` and `m.matchLines`

> **Note:** Matches are not highlighted in v1. Navigation to the correct viewport line is sufficient. In v2, highlighting can be added by injecting ANSI underline sequences into `m.renderedContent` at match positions.

---

#### Phase 5: Non-TTY Mode

When stdout is not an interactive terminal, or when `--no-pager` is set, skip the pager and print rendered output directly:

```go
// main.go (after flag parsing)
noPager := noPagerFlag || !term.IsTerminal(int(os.Stdout.Fd()))
if noPager {
    glamourStyle := chooseStyle(darkFlag, lightFlag, noColorFlag)
    out, err := renderMarkdown(content, glamourStyle, 80)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    fmt.Print(out)
    os.Exit(0)
}
```

**TTY detection:** `golang.org/x/term` (supplementary module maintained by the Go team, not stdlib). Call `term.IsTerminal(int(os.Stdout.Fd()))`.

**`--no-pager` flag:** Explicit escape hatch for agents and CI environments where TTY detection may be unreliable (Docker, pseudo-TTYs, `ssh -t`). Reuses the same non-interactive code path.

**`--no-color` flag / `NO_COLOR` env var:** Pass `glamour.WithStandardStyle("ascii")` or `"notty"` when set. Produces plain text output with no ANSI escape sequences — required for text-processing pipelines.

```go
func chooseStyle(dark, light, noColor bool) string {
    if noColor || os.Getenv("NO_COLOR") != "" {
        return "notty"
    }
    if dark { return "dark" }
    if light { return "light" }
    return "dark"  // default
}
```

---

#### Phase 6: Distribution

**`Makefile`:**

```makefile
build:
	go build -o cli-md .

install:
	go install .

release:
	mkdir -p dist
	GOOS=darwin  GOARCH=amd64 go build -o dist/cli-md-darwin-amd64  .
	GOOS=darwin  GOARCH=arm64 go build -o dist/cli-md-darwin-arm64  .
	GOOS=linux   GOARCH=amd64 go build -o dist/cli-md-linux-amd64   .
	GOOS=linux   GOARCH=arm64 go build -o dist/cli-md-linux-arm64   .

clean:
	rm -f cli-md
	rm -rf dist/
```

> **No GoReleaser in v1.** A 4-line `release` Makefile target produces all four binaries. GoReleaser adds CI config, a YAML file to maintain, and tooling dependencies — none of which are warranted until a Homebrew tap is actually needed. Add GoReleaser later when automating formula updates.

**GitHub Releases:** Upload the `dist/` binaries manually to a GitHub release for v1.

**Homebrew tap (future v2+):** Add GoReleaser and a tap formula once release automation is worth the complexity.

---

## Acceptance Criteria

### Functional

- [x] `cli-md README.md` opens the file in a full-screen pager
- [x] All heading levels H1–H6 render with distinct, visually different styles
- [x] Code blocks render with background highlighting and language label
- [x] Tables render with aligned columns
- [x] Links render in a distinguishable color/style
- [x] `j`/`k`/arrow keys scroll one line at a time
- [x] `g`/`G` jump to top/bottom of document
- [x] `/` opens search input at the bottom; Enter executes search; Escape cancels
- [x] `n`/`N` cycle through search matches in the document
- [x] `q` and `Ctrl+C` exit cleanly (exit code 0)
- [x] `--dark` flag forces dark theme; `--light` flag forces light theme
- [x] `--no-pager` flag skips the TUI and prints rendered output to stdout
- [x] `--no-color` flag and `NO_COLOR` env var produce plain text output (no ANSI)
- [x] Default theme is dark when no flag is passed
- [x] Terminal resize reflows content to new width without jumping to top

### Non-Functional

- [x] When stdout is not a TTY, output rendered ANSI text without pager
- [x] Missing or unreadable file prints error to stderr and exits with code 1
- [x] No file argument prints usage and exits with code 1
- [x] Binary is self-contained — no system dependencies (no `less`, no `glow`)
- [x] Builds to a single static binary via `go build`

### Quality Gates

- [x] `go vet ./...` passes
- [x] `go build ./...` produces clean binary
- [x] `go test ./...` passes — unit tests for `renderMarkdown()` and search (`computeMatches`)
- [ ] Manually tested with a markdown file containing all element types (headers, code, tables, links, lists, blockquotes)

## Error Handling & Exit Codes

| Situation | stderr message | Exit code |
|-----------|---------------|-----------|
| Clean quit (`q` / Ctrl+C) | — | 0 |
| Non-TTY stdout (piped) or `--no-pager` | — | 0 |
| Missing file argument | `Usage: cli-md [--dark\|--light] [--no-pager] [--no-color] <file.md>` | 1 |
| File not found | `cli-md: <path>: no such file or directory` | 1 |
| File not readable | `cli-md: <path>: permission denied` | 1 |
| Both `--dark` and `--light` | `cli-md: --dark and --light are mutually exclusive` | 1 |

## Implementation Gotchas (from framework research)

These must be handled correctly — they are common mistakes:

1. **`viewport.New()` timing**: Call it inside the first `tea.WindowSizeMsg`, never at `initialModel()` construction. The terminal size is unknown at startup.

2. **`viewport.YPosition`**: Set `m.viewport.YPosition = headerLines` after `viewport.New()`. Without this, the viewport overlaps the header visually.

3. **`viewport.MouseWheelEnabled`**: Set `m.viewport.MouseWheelEnabled = true` at viewport initialization. `tea.WithMouseCellMotion()` alone is not enough — both are required for mouse wheel scrolling.

4. **Glamour render on resize**: Guard the re-render with `if msg.Width != m.lastWidth`. Only re-render when width changes, not on every `WindowSizeMsg` (height-only resizes are common and cheap to ignore).

5. **Scroll position on resize**: Before calling `viewport.SetContent()`, save `m.viewport.YOffset`. Restore it after. Otherwise the user jumps to the top on every resize.

6. **Search on rendered lines, not raw markdown**: Glamour does not preserve 1:1 source-to-rendered line mapping. Search the ANSI-stripped rendered output. Store rendered line indices in `m.matchLines`. See Phase 4 for the full algorithm.

7. **Recompute search on resize**: When `m.searchQuery != ""` and the terminal is resized, recompute `m.matchLines` against the new `m.searchLines` after re-rendering.

8. **Search mode key capture**: When `m.searching == true`, handle keys and `return` early in `Update` before they reach the viewport. Otherwise `q` quits while typing.

9. **ANSI string width**: Always use `lipgloss.Width(s)` instead of `len(s)` when measuring strings for layout alignment. `len()` counts bytes, not display columns.

10. **Glamour trailing newlines**: `strings.TrimRight(out, "\n")` on glamour output prevents a blank first line in the viewport.

11. **`g` vs `G` keybindings**: `msg.String()` returns `"g"` and `"G"` as distinct strings. No extra modifier logic needed; Shift+G delivers `"G"` directly.

## References

### Internal

- Brainstorm: `docs/brainstorms/2026-02-18-cli-markdown-reader-brainstorm.md`

### Libraries

- `github.com/charmbracelet/glamour` — markdown-to-ANSI renderer
- `github.com/charmbracelet/bubbletea` — terminal UI framework (Model/Update/View)
- `github.com/charmbracelet/bubbles/viewport` — scrollable viewport component
- `github.com/charmbracelet/bubbles/textinput` — text input for search bar
- `github.com/charmbracelet/lipgloss` — terminal layout and styling
- `golang.org/x/term` — TTY detection (`term.IsTerminal`)
