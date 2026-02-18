package main

import (
	"fmt"
	"regexp"
	"strings"

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

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func renderMarkdown(md, style string, width int) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return md
	}
	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return strings.TrimRight(out, "\n")
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
