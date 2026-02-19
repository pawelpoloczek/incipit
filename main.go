package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func chooseStyle(dark, light, noColor bool) string {
	if noColor || os.Getenv("NO_COLOR") != "" {
		return "notty"
	}
	if dark {
		return "dark"
	}
	if light {
		return "light"
	}
	return "dark"
}

func main() {
	var (
		darkFlag    bool
		lightFlag   bool
		noPagerFlag bool
		noColorFlag bool
	)

	flag.BoolVar(&darkFlag, "dark", false, "force dark color theme (default)")
	flag.BoolVar(&lightFlag, "light", false, "force light color theme")
	flag.BoolVar(&noPagerFlag, "no-pager", false, "print rendered output without interactive pager")
	flag.BoolVar(&noColorFlag, "no-color", false, "disable ANSI colors")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: incipit [--dark|--light] [--no-pager] [--no-color] <file.md>\n")
	}
	flag.Parse()

	if darkFlag && lightFlag {
		fmt.Fprintf(os.Stderr, "incipit: --dark and --light are mutually exclusive\n")
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	filename := args[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "incipit: %s\n", err)
		os.Exit(1)
	}

	style := chooseStyle(darkFlag, lightFlag, noColorFlag)
	content := string(data)

	// Non-interactive mode: --no-pager flag or stdout is not a TTY
	if noPagerFlag || !term.IsTerminal(int(os.Stdout.Fd())) {
		out := renderMarkdown(content, style, 80)
		fmt.Print(out)
		return
	}

	m := newModel(filename, content, style)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "incipit: %s\n", err)
		os.Exit(1)
	}
}
