package main

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_ReturnsContent(t *testing.T) {
	out := renderMarkdown("# Hello\n\nSome text.", "dark", 80)
	if out == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(stripANSI(out), "Hello") {
		t.Errorf("expected output to contain 'Hello', got: %q", stripANSI(out))
	}
}

func TestRenderMarkdown_FallsBackOnBadStyle(t *testing.T) {
	// An unknown style should still return something (either fallback or raw)
	out := renderMarkdown("# Test", "nonexistent-style", 80)
	if out == "" {
		t.Fatal("expected non-empty fallback output")
	}
}

func TestRenderMarkdown_NoTrailingNewline(t *testing.T) {
	out := renderMarkdown("Hello", "dark", 80)
	if strings.HasSuffix(out, "\n") {
		t.Errorf("expected no trailing newline, got output ending in newline")
	}
}

func TestRenderMarkdown_H2DarkNoHashPrefix(t *testing.T) {
	out := stripANSI(renderMarkdown("## Section", "dark", 80))
	if strings.Contains(out, "## ") {
		t.Error("H2 dark: expected no '## ' prefix in output")
	}
	if !strings.Contains(out, "Section") {
		t.Error("H2 dark: expected heading text 'Section' to be present")
	}
}

func TestRenderMarkdown_H3DarkNoHashPrefix(t *testing.T) {
	out := stripANSI(renderMarkdown("### Subsection", "dark", 80))
	if strings.Contains(out, "### ") {
		t.Error("H3 dark: expected no '### ' prefix in output")
	}
	if !strings.Contains(out, "Subsection") {
		t.Error("H3 dark: expected heading text 'Subsection' to be present")
	}
}

func TestRenderMarkdown_H2LightNoHashPrefix(t *testing.T) {
	out := stripANSI(renderMarkdown("## Section", "light", 80))
	if strings.Contains(out, "## ") {
		t.Error("H2 light: expected no '## ' prefix in output")
	}
	if !strings.Contains(out, "Section") {
		t.Error("H2 light: expected heading text 'Section' to be present")
	}
}

func TestStripANSI(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"\x1b[31mhello\x1b[0m", "hello"},
		{"\x1b[1;32mworld\x1b[0m", "world"},
		{"no escapes", "no escapes"},
		{"", ""},
	}
	for _, tc := range cases {
		got := stripANSI(tc.input)
		if got != tc.want {
			t.Errorf("stripANSI(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestComputeMatches_BasicMatch(t *testing.T) {
	lines := []string{"hello world", "foo bar", "hello again"}
	matches := computeMatches(lines, "hello")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0] != 0 || matches[1] != 2 {
		t.Errorf("expected matches at [0, 2], got %v", matches)
	}
}

func TestComputeMatches_CaseInsensitive(t *testing.T) {
	lines := []string{"Hello World", "HELLO", "hello"}
	matches := computeMatches(lines, "hello")
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d: %v", len(matches), matches)
	}
}

func TestComputeMatches_NoMatches(t *testing.T) {
	lines := []string{"foo", "bar", "baz"}
	matches := computeMatches(lines, "xyz")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestComputeMatches_EmptyQuery(t *testing.T) {
	lines := []string{"foo", "bar"}
	// empty query matches every line
	matches := computeMatches(lines, "")
	if len(matches) != 2 {
		t.Errorf("expected 2 matches for empty query, got %d", len(matches))
	}
}

func TestComputeMatches_EmptyLines(t *testing.T) {
	matches := computeMatches([]string{}, "hello")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches on empty input, got %d", len(matches))
	}
}

func TestChooseStyle_Dark(t *testing.T) {
	if chooseStyle(true, false, false) != "dark" {
		t.Error("expected dark style")
	}
}

func TestChooseStyle_Light(t *testing.T) {
	if chooseStyle(false, true, false) != "light" {
		t.Error("expected light style")
	}
}

func TestChooseStyle_NoColor(t *testing.T) {
	if chooseStyle(false, false, true) != "notty" {
		t.Error("expected notty style for no-color")
	}
}

func TestChooseStyle_Default(t *testing.T) {
	if chooseStyle(false, false, false) != "dark" {
		t.Error("expected dark as default style")
	}
}

func TestChooseStyle_NoColorOverridesDark(t *testing.T) {
	// --no-color takes precedence over --dark
	if chooseStyle(true, false, true) != "notty" {
		t.Error("expected notty when both dark and no-color set")
	}
}

func TestAddLanguageLabels_WithLang(t *testing.T) {
	input := "```go\nfunc main() {}\n```\n"
	out := addLanguageLabels(input)
	if !strings.Contains(out, "`go`") {
		t.Errorf("expected language label `go` in output, got: %q", out)
	}
	if !strings.Contains(out, "```go\n") {
		t.Errorf("expected fenced block still present in output, got: %q", out)
	}
}

func TestAddLanguageLabels_NoLang(t *testing.T) {
	input := "```\nsome code\n```\n"
	out := addLanguageLabels(input)
	if out != input {
		t.Errorf("expected unchanged output for fence with no lang, got: %q", out)
	}
}

func TestAddLanguageLabels_NoFence(t *testing.T) {
	input := "Just some plain text.\n"
	out := addLanguageLabels(input)
	if out != input {
		t.Errorf("expected unchanged output for plain text, got: %q", out)
	}
}

func TestRenderMarkdown_CodeBlockDarkBg(t *testing.T) {
	out := renderMarkdown("```go\nfunc main() {}\n```", "dark", 80)
	if out == "" {
		t.Fatal("expected non-empty output for dark code block")
	}
	// The old background was #373737 (R=55,G=55,B=55); verify it is gone.
	if strings.Contains(out, "55;55;55") {
		t.Error("dark code block should not use the old #373737 background")
	}
}

func TestRenderMarkdown_CodeBlockLightBg(t *testing.T) {
	out := renderMarkdown("```go\nfunc main() {}\n```", "light", 80)
	if out == "" {
		t.Fatal("expected non-empty output for light code block")
	}
	// The light theme previously used the dark #373737 background; verify it is fixed.
	if strings.Contains(out, "55;55;55") {
		t.Error("light code block should not use the dark #373737 background")
	}
}
