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
