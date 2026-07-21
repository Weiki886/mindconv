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
