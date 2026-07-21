package render

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Weiki886/mindconv/internal/model"
)

// Markdown writes a map as a heading-based Markdown document.
func Markdown(writer io.Writer, mindMap *model.Map) error {
	if mindMap == nil || mindMap.Root == nil {
		return errors.New("render markdown: map has no root topic")
	}
	return writeMarkdownTopic(writer, mindMap.Root, 1)
}

func writeMarkdownTopic(writer io.Writer, topic *model.Topic, depth int) error {
	if depth <= 6 {
		if _, err := fmt.Fprintf(writer, "%s %s\n\n", strings.Repeat("#", depth), escapeMarkdown(topic.Title)); err != nil {
			return err
		}
	} else {
		indent := strings.Repeat("  ", depth-7)
		if _, err := fmt.Fprintf(writer, "%s- **%s**\n\n", indent, escapeMarkdown(topic.Title)); err != nil {
			return err
		}
	}

	if topic.Notes != "" {
		if _, err := fmt.Fprintf(writer, "%s\n\n", escapeMarkdown(topic.Notes)); err != nil {
			return err
		}
	}

	for _, link := range topic.Links {
		url := strings.ReplaceAll(link.URL, ">", "%3E")
		if _, err := fmt.Fprintf(writer, "- [%s](<%s>)\n", escapeMarkdown(link.Title), url); err != nil {
			return err
		}
	}
	if len(topic.Links) > 0 {
		if _, err := io.WriteString(writer, "\n"); err != nil {
			return err
		}
	}

	for _, child := range topic.Children {
		if err := writeMarkdownTopic(writer, child, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func escapeMarkdown(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"<", "\\<",
		">", "\\>",
		"#", "\\#",
	)
	return replacer.Replace(value)
}
