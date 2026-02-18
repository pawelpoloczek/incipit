# cli-md

A terminal markdown reader with consistent rendering for all elements — headings, code blocks, tables, links.

## Usage

```
cli-md [options] <file.md>
```

### Options

| Flag | Description |
|------|-------------|
| `--dark` | Force dark color theme (default) |
| `--light` | Force light color theme |
| `--no-pager` | Print rendered output without interactive pager |
| `--no-color` | Disable ANSI colors (also respects `NO_COLOR` env var) |

### Keybindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Scroll up |
| `↓` / `j` | Scroll down |
| `PgUp` / `b` | Page up |
| `PgDn` / `f` / `Space` | Page down |
| `Ctrl+U` | Half page up |
| `Ctrl+D` | Half page down |
| `g` | Go to top |
| `G` | Go to bottom |
| `/` | Search |
| `n` | Next match |
| `N` | Previous match |
| `q` / `Ctrl+C` | Quit |

## Installation

```bash
go install github.com/pawelpoloczek/cli-md@latest
```

Or clone and build:

```bash
git clone https://github.com/pawelpoloczek/cli-md
cd cli-md
make install
```

## Examples

```bash
cli-md README.md
cli-md --light CHANGELOG.md
cli-md --no-pager README.md | head -20
NO_COLOR=1 cli-md README.md
```
