# incipit

A terminal markdown reader with consistent rendering for all elements — headings, code blocks, tables, links.

## Usage

```
incipit [options] <file.md>
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
go install github.com/pawelpoloczek/incipit@latest
```

Or clone and build:

```bash
git clone https://github.com/pawelpoloczek/incipit
cd incipit
make install
```

## Examples

```bash
incipit README.md
incipit --light CHANGELOG.md
incipit --no-pager README.md | head -20
NO_COLOR=1 incipit README.md
```
