package render

import (
	"strings"
	"testing"

	"github.com/Weiki886/mindconv/internal/model"
)

func testMap() *model.Map {
	return &model.Map{Root: &model.Topic{
		Title: "Project *Plan* <script>",
		Notes: "A short note.",
		Links: []model.Link{{Title: "Docs", URL: "https://example.com"}},
		Children: []*model.Topic{{
			Title: "Research",
		}},
	}}
}

func TestMarkdown(t *testing.T) {
	var output strings.Builder
	if err := Markdown(&output, testMap()); err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	result := output.String()
	for _, expected := range []string{
		"# Project \\*Plan\\* \\<script\\>",
		"A short note.",
		"- [Docs](<https://example.com>)",
		"## Research",
	} {
		if !strings.Contains(result, expected) {
			t.Errorf("Markdown output does not contain %q:\n%s", expected, result)
		}
	}
}

func TestHTML(t *testing.T) {
	var output strings.Builder
	if err := HTML(&output, testMap()); err != nil {
		t.Fatalf("HTML() error = %v", err)
	}

	result := output.String()
	if strings.Contains(result, "<script>") {
		t.Fatalf("HTML output contains an unescaped script element:\n%s", result)
	}
	for _, expected := range []string{"&lt;script&gt;", "https://example.com", "Research"} {
		if !strings.Contains(result, expected) {
			t.Errorf("HTML output does not contain %q", expected)
		}
	}
}

func TestRenderRejectsEmptyMap(t *testing.T) {
	if err := Markdown(&strings.Builder{}, nil); err == nil {
		t.Fatal("Markdown() error = nil, want an error")
	}
	if err := HTML(&strings.Builder{}, &model.Map{}); err == nil {
		t.Fatal("HTML() error = nil, want an error")
	}
}

func TestMarkdownDeepTopicWithMetadata(t *testing.T) {
	// Build a map where depth 7 (first list-item level) has notes and links.
	// All children live under that deep node so its metadata must be indented
	// to stay inside the list item, otherwise child topics break hierarchy.
	deep := &model.Topic{
		Title: "Deep Node",
		Notes: "Deep note content",
		Links: []model.Link{{Title: "Deep Link", URL: "https://example.com/deep"}},
		Children: []*model.Topic{
			{
				Title: "Child of Deep",
				Notes: "Child note",
				Links: []model.Link{{Title: "Child Link", URL: "https://example.com/child"}},
				Children: []*model.Topic{
					{Title: "Grandchild of Deep"},
				},
			},
		},
	}

	// Construct a map with 5 levels of ancestors so "Deep Node" is at depth 7.
	ancestors := make([]*model.Topic, 5)
	current := deep
	for i := 4; i >= 0; i-- {
		ancestors[i] = &model.Topic{Title: "Level " + string(rune('A'+i)), Children: []*model.Topic{current}}
		current = ancestors[i]
	}

	mindMap := &model.Map{Root: &model.Topic{Title: "Root", Children: []*model.Topic{ancestors[0]}}}

	var output strings.Builder
	if err := Markdown(&output, mindMap); err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	result := output.String()
	lines := strings.Split(result, "\n")

	// Find the "Deep Node" list item line and its subsequent lines.
	deepIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "**Deep Node**") {
			deepIdx = i
			break
		}
	}
	if deepIdx < 0 {
		t.Fatal("Deep Node not found in output")
	}

	// Notes must be indented so they become list-item content (one level deeper
	// than the list-item title itself). indent = strings.Repeat("  ", depth-6)
	// = strings.Repeat("  ", 7-6) = "  "
	if !strings.HasPrefix(lines[deepIdx+1], "  ") || !strings.Contains(lines[deepIdx+1], "Deep note content") {
		t.Errorf("deep notes not properly indented; got line %d: %q", deepIdx+1, lines[deepIdx+1])
	}

	// The first non-empty line after notes must be the first link. Both the
	// prefix and content must match, otherwise a missing link silently passes.
	linkIdx := deepIdx + 2
	for linkIdx < len(lines) && strings.TrimSpace(lines[linkIdx]) == "" {
		linkIdx++
	}
	if linkIdx >= len(lines) {
		t.Fatal("no link found after deep notes")
	}
	if !strings.Contains(lines[linkIdx], "Deep Link") {
		t.Errorf("first link not found after deep notes; got line %d: %q", linkIdx, lines[linkIdx])
	}
	if !strings.HasPrefix(lines[linkIdx], "  - ") {
		t.Errorf("deep link not properly indented; got line %d: %q", linkIdx, lines[linkIdx])
	}

	// "Child of Deep" at depth 8: indent = strings.Repeat("  ", 8-7) = "  "
	childIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "**Child of Deep**") {
			childIdx = i
			break
		}
	}
	if childIdx < 0 {
		t.Fatal("Child of Deep not found in output")
	}
	if !strings.HasPrefix(lines[childIdx], "  - **") {
		t.Errorf("child topic not at correct depth; got line %d: %q", childIdx, lines[childIdx])
	}

	// Grandchild at depth 9: indent = strings.Repeat("  ", 9-7) = "    "
	grandchildIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "**Grandchild of Deep**") {
			grandchildIdx = i
			break
		}
	}
	if grandchildIdx < 0 {
		t.Fatal("Grandchild of Deep not found in output")
	}
	if !strings.HasPrefix(lines[grandchildIdx], "    - **") {
		t.Errorf("grandchild topic not at correct depth; got line %d: %q", grandchildIdx, lines[grandchildIdx])
	}
}
