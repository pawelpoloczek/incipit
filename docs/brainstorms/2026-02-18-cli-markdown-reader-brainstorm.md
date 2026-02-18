---
date: 2026-02-18
topic: cli-markdown-reader
---

# CLI Markdown Reader

## What We're Building

A terminal-based markdown reader for the command line, invoked as `cli-md README.md`. It renders markdown with consistent, correct styling for all elements and displays the result in a built-in pager (like `less`). The user can scroll through content with arrow keys and quit with `q`.

The key motivation is that existing tools like `glow` render some elements poorly (e.g. only the first header looks correct). This tool should render all markdown elements — H1-H6, code blocks, tables, links — with distinct, readable styles.

## Why This Approach

We chose the **Charmbracelet stack**: `glamour` for markdown-to-ANSI rendering and `bubbletea` + `bubbles/viewport` for the pager.

This is the standard Go terminal UI ecosystem. Glow's rendering issues appear to be glow-specific rather than glamour itself — using glamour directly gives us full control without glow's wrapper. The bubbletea viewport handles scrolling, keyboard input, and terminal resize cleanly, producing a fully self-contained binary with no system dependencies.

Alternatives considered:
- **Goldmark + custom renderer + bubbletea**: More rendering control, but significantly more code to write for questionable gain when glamour already handles all target elements.
- **Goldmark + system less**: Rendering control but depends on system `less` being present, which is problematic for distribution.

## Key Decisions

- **Language**: Go — good CLI ergonomics, single binary distribution
- **Markdown renderer**: `glamour` (charmbracelet) — handles H1-H6, code blocks, tables, links
- **Pager**: `bubbletea` + `bubbles/viewport` — self-contained scrolling, no system `less` dependency
- **Input**: File path argument (`cli-md README.md`)
- **Navigation**: Pager-style — arrow keys / j/k to scroll, g/G for top/bottom, / to search, `q` to quit
- **Distribution**: Single binary, distributable (GitHub releases)
- **Color scheme**: Dark theme by default, overridable via `--light` / `--dark` flags
- **Terminal width**: Auto-detect from actual terminal width for responsive rendering

## Resolved Questions

- **Color scheme**: Configurable via flags (`--light` / `--dark`), defaults to dark
- **Terminal width**: Auto-detect terminal width
- **Keyboard shortcuts**: arrow keys / j/k scroll, g/G top/bottom, `/` search, `q` quit

## Next Steps

→ `/workflows:plan` for implementation details
