package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const (
	headerLines = 1
	footerLines = 1
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// codeBlockRe matches fenced code blocks. Group 1 = language (optional), group 2 = code content.
// Flags: m (multiline ^/$) and s (dotall — . matches \n).
var codeBlockRe = regexp.MustCompile("(?ms)^`{3}([a-zA-Z][a-zA-Z0-9_+-]*)?\n(.*?)^`{3}[^\\S\\r\\n]*$")

// headerRe matches ATX headings. Group 1 = '#' characters (level), group 2 = heading text.
var headerRe = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)

// inlineMarkdownRe strips common inline markdown delimiters from heading text.
var inlineMarkdownRe = regexp.MustCompile("[*_~`]{1,2}")

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

type codeBlock struct {
	lang string
	code string
}

type headerBlock struct {
	level int
	text  string
}

// extractCodeBlocks pulls fenced code blocks out of md, replacing each with a
// unique placeholder paragraph, and returns the modified prose plus the blocks.
func extractCodeBlocks(md string) (string, []codeBlock) {
	var blocks []codeBlock
	prose := codeBlockRe.ReplaceAllStringFunc(md, func(match string) string {
		sub := codeBlockRe.FindStringSubmatch(match)
		lang, code := "", ""
		if len(sub) > 1 {
			lang = sub[1]
		}
		if len(sub) > 2 {
			code = sub[2]
		}
		placeholder := fmt.Sprintf("INCIPIT_CODEBLOCK_%d", len(blocks))
		blocks = append(blocks, codeBlock{lang: lang, code: code})
		return placeholder
	})
	return prose, blocks
}

func chromaStyleName(style string) string {
	switch style {
	case "light":
		return "github"
	default:
		return "monokai"
	}
}

func syntaxHighlight(code, lang, chromaStyle string) string {
	var buf strings.Builder
	_ = quick.Highlight(&buf, code, lang, "terminal256", chromaStyle)
	return buf.String()
}

// renderCodeBlock renders a single code block with a rounded border, syntax
// highlighting, and a full background fill across all content lines.
func renderCodeBlock(cb codeBlock, width int, style string) string {
	outerWidth := width
	innerWidth := outerWidth - 4 // 1 char border + 1 space padding on each side
	if innerWidth < 1 {
		innerWidth = 1
	}

	useColor := style != "notty"

	var bgIndex, borderColor string
	switch style {
	case "light":
		bgIndex = "254"
		borderColor = "27"
	default: // dark and anything else
		bgIndex = "235"
		borderColor = "23"
	}

	bgOn := fmt.Sprintf("\x1b[48;5;%sm", bgIndex)
	resetToBg := fmt.Sprintf("\x1b[0;48;5;%sm", bgIndex)
	reset := "\x1b[0m"

	var bs lipgloss.Style
	if useColor {
		bs = lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))
	} else {
		bs = lipgloss.NewStyle()
	}

	// Top border — language label embedded when present.
	var top string
	if cb.lang != "" {
		dashes := outerWidth - 6 - len(cb.lang)
		if dashes < 0 {
			dashes = 0
		}
		top = bs.Render("╭── " + cb.lang + " " + strings.Repeat("─", dashes) + "╮")
	} else {
		top = bs.Render("╭" + strings.Repeat("─", outerWidth-2) + "╮")
	}
	bottom := bs.Render("╰" + strings.Repeat("─", outerWidth-2) + "╯")
	lbar := bs.Render("│")
	rbar := bs.Render("│")

	// Blank padding line (top and bottom inside the box).
	var blank string
	if useColor {
		blank = lbar + " " + bgOn + strings.Repeat(" ", innerWidth) + reset + " " + rbar
	} else {
		blank = lbar + " " + strings.Repeat(" ", innerWidth) + " " + rbar
	}

	// Obtain syntax-highlighted (or plain) code lines.
	var raw string
	if useColor {
		raw = syntaxHighlight(cb.code, cb.lang, chromaStyleName(style))
	} else {
		raw = cb.code
	}
	raw = strings.TrimRight(raw, "\n")
	codeLines := strings.Split(raw, "\n")

	var out []string
	out = append(out, top, blank)

	for _, line := range codeLines {
		visible := stripANSI(line)
		pad := innerWidth - len([]rune(visible))
		if pad < 0 {
			pad = 0
		}
		var cl string
		if useColor {
			// Replace every reset with reset+background so the bg persists across tokens.
			colored := strings.ReplaceAll(line, "\x1b[0m", resetToBg)
			cl = lbar + " " + bgOn + colored + strings.Repeat(" ", pad) + reset + " " + rbar
		} else {
			cl = lbar + " " + visible + strings.Repeat(" ", pad) + " " + rbar
		}
		out = append(out, cl)
	}

	out = append(out, blank, bottom)
	return strings.Join(out, "\n")
}

// injectCodeBlocks replaces INCIPIT_CODEBLOCK_N placeholder lines in rendered
// with the fully-rendered code block for each corresponding block.
func injectCodeBlocks(rendered string, blocks []codeBlock, width int, style string) string {
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		plain := stripANSI(line)
		for j, cb := range blocks {
			if strings.Contains(plain, fmt.Sprintf("INCIPIT_CODEBLOCK_%d", j)) {
				lines[i] = renderCodeBlock(cb, width, style)
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

// extractHeaders pulls ATX headings out of md, replacing each with a unique
// placeholder, and returns the modified prose plus the extracted headers.
func extractHeaders(md string) (string, []headerBlock) {
	var headers []headerBlock
	prose := headerRe.ReplaceAllStringFunc(md, func(match string) string {
		sub := headerRe.FindStringSubmatch(match)
		level := len(sub[1])
		text := sub[2]
		placeholder := fmt.Sprintf("INCIPIT_HEADER_%d", len(headers))
		headers = append(headers, headerBlock{level: level, text: text})
		return placeholder
	})
	return prose, headers
}

// stripInlineMarkdown removes common inline markdown delimiters (**, *, _, `, ~~)
// from a heading text string so lipgloss receives clean plain text.
func stripInlineMarkdown(s string) string {
	return strings.TrimSpace(inlineMarkdownRe.ReplaceAllString(s, ""))
}

// headerColors returns the 256-color fg/bg indices and bold flag for a heading
// level in the given style ("dark", "light", or anything else → dark defaults).
func headerColors(level int, style string) (fg, bg string, bold bool) {
	if style == "light" {
		switch level {
		case 1:
			return "0", "105", true
		case 2:
			return "27", "195", true
		case 3:
			return "28", "194", true
		case 4:
			return "19", "189", true
		case 5:
			return "17", "153", false
		default:
			return "59", "188", false
		}
	}
	switch level {
	case 1:
		return "15", "57", true
	case 2:
		return "51", "23", true
	case 3:
		return "48", "22", true
	case 4:
		return "75", "17", true
	case 5:
		return "67", "236", false
	default:
		return "60", "235", false
	}
}

// renderHeader renders a single heading as a pill-shaped lipgloss string.
// For "notty" style it returns plain text with no ANSI codes.
func renderHeader(h headerBlock, style string) string {
	text := stripInlineMarkdown(h.text)
	if style == "notty" {
		return text
	}
	fg, bg, bold := headerColors(h.level, style)
	s := lipgloss.NewStyle().
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(bg)).
		Padding(0, 2).
		Bold(bold)
	return s.Render(text)
}

// injectHeaders replaces INCIPIT_HEADER_N placeholder lines in rendered with
// the pill-rendered heading for each corresponding header.
func injectHeaders(rendered string, headers []headerBlock, style string) string {
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		plain := stripANSI(line)
		for j, h := range headers {
			if strings.Contains(plain, fmt.Sprintf("INCIPIT_HEADER_%d", j)) {
				lines[i] = renderHeader(h, style)
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

func renderMarkdown(md, style string, width int) string {
	prose, blocks := extractCodeBlocks(md)
	prose, headers := extractHeaders(prose)

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return md
	}
	out, err := r.Render(prose)
	if err != nil {
		return md
	}
	out = strings.TrimRight(out, "\n")
	out = injectCodeBlocks(out, blocks, width, style)
	out = injectHeaders(out, headers, style)
	return out
}


func computeMatches(lines []string, query string) []int {
	lower := strings.ToLower(query)
	var result []int
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), lower) {
			result = append(result, i)
		}
	}
	return result
}

type model struct {
	filename     string
	rawMarkdown  string
	glamourStyle string

	viewport  viewport.Model
	ready     bool
	lastWidth int

	// search state
	searching   bool
	searchQuery string
	searchLines []string // ANSI-stripped rendered lines
	matchLines  []int    // rendered line indices of matches
	matchIdx    int
	noMatches   bool
}

func newModel(filename, rawMarkdown, glamourStyle string) model {
	return model{
		filename:     filename,
		rawMarkdown:  rawMarkdown,
		glamourStyle: glamourStyle,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// applyContent renders markdown at the given width and populates the viewport.
// Preserves scroll position across calls (e.g. on resize).
func (m *model) applyContent(width int) {
	rendered := renderMarkdown(m.rawMarkdown, m.glamourStyle, width)
	m.lastWidth = width
	savedOffset := m.viewport.YOffset
	m.viewport.SetContent(rendered)
	m.viewport.YOffset = savedOffset
	m.searchLines = strings.Split(stripANSI(rendered), "\n")
	if m.searchQuery != "" {
		m.matchLines = computeMatches(m.searchLines, m.searchQuery)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerLines-footerLines)
			m.viewport.YPosition = headerLines
			m.applyContent(msg.Width)
			m.ready = true
		} else {
			m.viewport.Height = msg.Height - headerLines - footerLines
			if msg.Width != m.lastWidth {
				m.applyContent(msg.Width)
			}
			m.viewport.Width = msg.Width
		}

	case tea.KeyMsg:
		if m.searching {
			switch {
			case msg.Type == tea.KeyEnter:
				if m.searchQuery == "" {
					m.searching = false
				} else {
					m.matchLines = computeMatches(m.searchLines, m.searchQuery)
					m.matchIdx = 0
					m.noMatches = len(m.matchLines) == 0
					m.searching = false
					if len(m.matchLines) > 0 {
						m.viewport.GotoTop()
						m.viewport.LineDown(m.matchLines[0])
					}
				}
			case msg.Type == tea.KeyEsc:
				m.searching = false
				m.searchQuery = ""
				m.matchLines = nil
				m.noMatches = false
			case msg.Type == tea.KeyBackspace:
				runes := []rune(m.searchQuery)
				if len(runes) > 0 {
					m.searchQuery = string(runes[:len(runes)-1])
				}
				m.noMatches = false
			default:
				m.searchQuery += string(msg.Runes)
			}
			return m, tea.Batch(cmds...)
		}

		// Normal pager mode
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "g":
			m.viewport.GotoTop()
		case "G":
			m.viewport.GotoBottom()
		case "/":
			m.searching = true
			m.noMatches = false
		case "n":
			if len(m.matchLines) > 0 {
				m.matchIdx = (m.matchIdx + 1) % len(m.matchLines)
				m.viewport.GotoTop()
				m.viewport.LineDown(m.matchLines[m.matchIdx])
			}
		case "N":
			if len(m.matchLines) > 0 {
				m.matchIdx = (m.matchIdx - 1 + len(m.matchLines)) % len(m.matchLines)
				m.viewport.GotoTop()
				m.viewport.LineDown(m.matchLines[m.matchIdx])
			}
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	// Header: bold filename
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	header := lipgloss.NewStyle().
		Width(m.viewport.Width).
		Render(" " + headerStyle.Render(m.filename))

	// Footer
	var footerContent string
	switch {
	case m.searching:
		footerContent = "/" + m.searchQuery + "_"
	case m.noMatches && m.searchQuery != "":
		footerContent = fmt.Sprintf(" no matches: %s", m.searchQuery)
	case len(m.matchLines) > 0:
		footerContent = fmt.Sprintf(" %d/%d: %s", m.matchIdx+1, len(m.matchLines), m.searchQuery)
	default:
		help := " ↑/k ↓/j  g/G  / search  q quit"
		pct := fmt.Sprintf("  %3.f%% ", m.viewport.ScrollPercent()*100)
		gap := m.viewport.Width - lipgloss.Width(help) - lipgloss.Width(pct)
		if gap < 0 {
			gap = 0
		}
		footerContent = help + strings.Repeat(" ", gap) + pct
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(m.viewport.Width).
		Render(footerContent)

	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
}
