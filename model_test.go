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

// extractHeaders tests

func TestExtractHeaders_SingleH2(t *testing.T) {
	md := "## Section\n\nSome prose."
	prose, headers := extractHeaders(md)
	if len(headers) != 1 {
		t.Fatalf("expected 1 header, got %d", len(headers))
	}
	if headers[0].level != 2 {
		t.Errorf("expected level 2, got %d", headers[0].level)
	}
	if headers[0].text != "Section" {
		t.Errorf("expected text 'Section', got %q", headers[0].text)
	}
	if !strings.Contains(prose, "INCIPIT_HEADER_0") {
		t.Errorf("expected placeholder in prose, got %q", prose)
	}
	if strings.Contains(prose, "## ") {
		t.Error("expected ## to be removed from prose")
	}
}

func TestExtractHeaders_MultipleLevelsReturnsCorrectOrder(t *testing.T) {
	md := "# Title\n\n## Section\n\n### Sub\n"
	_, headers := extractHeaders(md)
	if len(headers) != 3 {
		t.Fatalf("expected 3 headers, got %d", len(headers))
	}
	if headers[0].level != 1 || headers[1].level != 2 || headers[2].level != 3 {
		t.Errorf("unexpected levels: %v", []int{headers[0].level, headers[1].level, headers[2].level})
	}
}

func TestExtractHeaders_NoHeaders(t *testing.T) {
	md := "Just plain prose.\n\nAnother paragraph.\n"
	prose, headers := extractHeaders(md)
	if len(headers) != 0 {
		t.Errorf("expected 0 headers, got %d", len(headers))
	}
	if prose != md {
		t.Errorf("expected prose unchanged, got %q", prose)
	}
}

func TestExtractHeaders_WithInlineMarkdown(t *testing.T) {
	md := "## **Bold** Title\n"
	_, headers := extractHeaders(md)
	if len(headers) != 1 {
		t.Fatalf("expected 1 header, got %d", len(headers))
	}
	if headers[0].text != "**Bold** Title" {
		t.Errorf("expected raw text preserved, got %q", headers[0].text)
	}
}

// stripInlineMarkdown tests

func TestStripInlineMarkdown_Bold(t *testing.T) {
	got := stripInlineMarkdown("**Bold** Title")
	if strings.Contains(got, "*") {
		t.Errorf("expected asterisks removed, got %q", got)
	}
	if !strings.Contains(got, "Bold") {
		t.Errorf("expected 'Bold' preserved, got %q", got)
	}
}

func TestStripInlineMarkdown_InlineCode(t *testing.T) {
	got := stripInlineMarkdown("`code` here")
	if strings.Contains(got, "`") {
		t.Errorf("expected backtick removed, got %q", got)
	}
	if !strings.Contains(got, "code") {
		t.Errorf("expected 'code' preserved, got %q", got)
	}
}

// renderHeader tests

func TestRenderHeader_DarkH2_PillShape(t *testing.T) {
	h := headerBlock{level: 2, text: "Section"}
	out := renderHeader(h, "dark")
	plain := stripANSI(out)
	if !strings.Contains(plain, "Section") {
		t.Errorf("expected 'Section' in rendered header, got %q", plain)
	}
	// Pill has 2 spaces padding on each side
	if !strings.Contains(plain, "  Section  ") {
		t.Errorf("expected 2-space padding around text, got %q", plain)
	}
}

func TestRenderHeader_LightH1_PillShape(t *testing.T) {
	h := headerBlock{level: 1, text: "Title"}
	out := renderHeader(h, "light")
	plain := stripANSI(out)
	if !strings.Contains(plain, "  Title  ") {
		t.Errorf("expected 2-space padding around text, got %q", plain)
	}
}

func TestRenderHeader_NottyPlainText(t *testing.T) {
	h := headerBlock{level: 2, text: "Section"}
	out := renderHeader(h, "notty")
	if out != stripANSI(out) {
		t.Error("expected no ANSI codes in notty header output")
	}
	if out != "Section" {
		t.Errorf("expected plain 'Section', got %q", out)
	}
}

// injectHeaders tests

func TestInjectHeaders_ReplacesPlaceholder(t *testing.T) {
	headers := []headerBlock{{level: 2, text: "Section"}}
	rendered := "Some prose\n\n  INCIPIT_HEADER_0\n\nMore prose"
	out := injectHeaders(rendered, headers, "dark")
	if strings.Contains(stripANSI(out), "INCIPIT_HEADER_0") {
		t.Error("expected placeholder to be replaced")
	}
	if !strings.Contains(stripANSI(out), "Section") {
		t.Error("expected header text in output")
	}
}

// extractCodeBlocks tests

func TestExtractCodeBlocks_WithLang(t *testing.T) {
	md := "```go\nfunc main() {}\n```\n"
	prose, blocks := extractCodeBlocks(md)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].lang != "go" {
		t.Errorf("expected lang 'go', got %q", blocks[0].lang)
	}
	if !strings.Contains(blocks[0].code, "func main()") {
		t.Errorf("expected code to contain 'func main()', got %q", blocks[0].code)
	}
	if !strings.Contains(prose, "INCIPIT_CODEBLOCK_0") {
		t.Errorf("expected prose to contain placeholder, got %q", prose)
	}
	if strings.Contains(prose, "```") {
		t.Error("expected prose to not contain fenced block")
	}
}

func TestExtractCodeBlocks_NoLang(t *testing.T) {
	md := "```\nsome code\n```\n"
	prose, blocks := extractCodeBlocks(md)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].lang != "" {
		t.Errorf("expected empty lang, got %q", blocks[0].lang)
	}
	if !strings.Contains(prose, "INCIPIT_CODEBLOCK_0") {
		t.Errorf("expected placeholder in prose, got %q", prose)
	}
}

func TestExtractCodeBlocks_Multiple(t *testing.T) {
	md := "```go\nfunc a() {}\n```\n\nSome prose.\n\n```python\nprint('hi')\n```\n"
	prose, blocks := extractCodeBlocks(md)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].lang != "go" {
		t.Errorf("block 0: expected lang 'go', got %q", blocks[0].lang)
	}
	if blocks[1].lang != "python" {
		t.Errorf("block 1: expected lang 'python', got %q", blocks[1].lang)
	}
	if !strings.Contains(prose, "INCIPIT_CODEBLOCK_0") {
		t.Error("expected INCIPIT_CODEBLOCK_0 in prose")
	}
	if !strings.Contains(prose, "INCIPIT_CODEBLOCK_1") {
		t.Error("expected INCIPIT_CODEBLOCK_1 in prose")
	}
}

func TestExtractCodeBlocks_NoBlocks(t *testing.T) {
	md := "Just plain prose.\n\nAnother paragraph.\n"
	prose, blocks := extractCodeBlocks(md)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(blocks))
	}
	if prose != md {
		t.Errorf("expected prose unchanged, got %q", prose)
	}
}

// renderCodeBlock tests

func TestRenderCodeBlock_ContainsBorder(t *testing.T) {
	cb := codeBlock{lang: "go", code: "func main() {}\n"}
	out := renderCodeBlock(cb, 60, "dark")
	if !strings.Contains(out, "╭") {
		t.Error("expected top-left border character ╭")
	}
	if !strings.Contains(out, "╰") {
		t.Error("expected bottom-left border character ╰")
	}
}

func TestRenderCodeBlock_ContainsCode(t *testing.T) {
	cb := codeBlock{lang: "", code: "hello world\n"}
	out := stripANSI(renderCodeBlock(cb, 60, "dark"))
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected code content in output, got: %q", out)
	}
}

func TestRenderCodeBlock_WithLang_TitleInBorder(t *testing.T) {
	cb := codeBlock{lang: "go", code: "x := 1\n"}
	out := stripANSI(renderCodeBlock(cb, 60, "dark"))
	if !strings.Contains(out, "── go ──") {
		t.Errorf("expected language label in top border, got: %q", out)
	}
}

func TestRenderCodeBlock_NoLang_PlainBorder(t *testing.T) {
	cb := codeBlock{lang: "", code: "x := 1\n"}
	out := stripANSI(renderCodeBlock(cb, 60, "dark"))
	// Top border should be plain ╭───...───╮ with no language label
	if strings.Contains(out, "──  ──") {
		t.Error("expected no language label in plain border")
	}
	firstLine := strings.Split(out, "\n")[0]
	if !strings.HasPrefix(firstLine, "╭") {
		t.Errorf("top border should start with ╭, got: %q", firstLine)
	}
}

func TestRenderCodeBlock_NottyNoBorderColor(t *testing.T) {
	cb := codeBlock{lang: "go", code: "x := 1\n"}
	out := renderCodeBlock(cb, 60, "notty")
	// notty style should produce no ANSI color codes
	if out != stripANSI(out) {
		t.Error("expected no ANSI codes in notty output")
	}
	if !strings.Contains(out, "╭") {
		t.Error("expected border characters even in notty output")
	}
}

// injectCodeBlocks tests

func TestInjectCodeBlocks_ReplacesPlaceholder(t *testing.T) {
	blocks := []codeBlock{{lang: "go", code: "x := 1\n"}}
	rendered := "Some prose\n\n  INCIPIT_CODEBLOCK_0\n\nMore prose"
	out := injectCodeBlocks(rendered, blocks, 60, "dark")
	if strings.Contains(stripANSI(out), "INCIPIT_CODEBLOCK_0") {
		t.Error("expected placeholder to be replaced")
	}
	if !strings.Contains(out, "╭") {
		t.Error("expected code block border in output")
	}
}

// End-to-end tests

func TestRenderMarkdown_CodeBlock_EndToEnd(t *testing.T) {
	md := "Some text.\n\n```go\nfunc main() {}\n```\n\nMore text."
	out := renderMarkdown(md, "dark", 80)
	if out == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(out, "╭") {
		t.Error("expected rounded border ╭ in code block output")
	}
	if !strings.Contains(out, "╰") {
		t.Error("expected rounded border ╰ in code block output")
	}
}

func TestRenderMarkdown_CodeBlockDarkBg(t *testing.T) {
	out := renderMarkdown("```go\nfunc main() {}\n```", "dark", 80)
	if out == "" {
		t.Fatal("expected non-empty output for dark code block")
	}
	// Should use 256-color index 235 background (not the old #373737)
	if strings.Contains(out, "55;55;55") {
		t.Error("dark code block should not use the old #373737 background")
	}
	if !strings.Contains(out, "╭") {
		t.Error("expected border in dark code block output")
	}
}

func TestRenderMarkdown_CodeBlockLightBg(t *testing.T) {
	out := renderMarkdown("```go\nfunc main() {}\n```", "light", 80)
	if out == "" {
		t.Fatal("expected non-empty output for light code block")
	}
	// Should use 256-color index 254 background (not the old #373737)
	if strings.Contains(out, "55;55;55") {
		t.Error("light code block should not use the dark #373737 background")
	}
	if !strings.Contains(out, "╭") {
		t.Error("expected border in light code block output")
	}
}

// Header integration tests

func TestRenderMarkdown_HeaderH1DarkPill(t *testing.T) {
	out := renderMarkdown("# Title", "dark", 80)
	plain := stripANSI(out)
	if !strings.Contains(plain, "  Title  ") {
		t.Errorf("H1 dark: expected pill-padded 'Title', got %q", plain)
	}
	if strings.Contains(plain, "# ") {
		t.Error("H1 dark: expected no '# ' prefix in output")
	}
}

func TestRenderMarkdown_HeaderH3LightPill(t *testing.T) {
	out := renderMarkdown("### Subsection", "light", 80)
	plain := stripANSI(out)
	if !strings.Contains(plain, "  Subsection  ") {
		t.Errorf("H3 light: expected pill-padded 'Subsection', got %q", plain)
	}
	if strings.Contains(plain, "### ") {
		t.Error("H3 light: expected no '### ' prefix in output")
	}
}
