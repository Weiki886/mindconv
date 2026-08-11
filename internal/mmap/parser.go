package mmap

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"unicode"

	"github.com/Weiki886/mindconv/internal/model"
)

const (
	maxArchiveEntries = 10_000
	maxDocumentSize   = 64 << 20
	maxXMLDepth       = 512
	maxXMLNodes       = 500_000
)

var (
	// ErrDocumentNotFound means the MMAP archive does not contain Document.xml.
	ErrDocumentNotFound = errors.New("mmap: Document.xml not found")
	// ErrRootTopicNotFound means Document.xml has no MindManager Topic element.
	ErrRootTopicNotFound = errors.New("mmap: root topic not found")
)

// ParseFile parses a MindManager .mmap archive into a format-independent map.
func ParseFile(filename string) (*model.Map, error) {
	archive, document, err := openDocument(filename)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	reader, err := document.Open()
	if err != nil {
		return nil, fmt.Errorf("open Document.xml: %w", err)
	}
	defer reader.Close()

	root, err := parseXML(io.LimitReader(reader, maxDocumentSize+1))
	if err != nil {
		return nil, err
	}

	rootTopic := findRootTopic(root)
	if rootTopic == nil {
		return nil, ErrRootTopicNotFound
	}

	return &model.Map{Root: convertTopic(rootTopic)}, nil
}

// WriteDocumentXML writes the original Document.xml from a MindManager archive.
func WriteDocumentXML(filename string, writer io.Writer) error {
	archive, document, err := openDocument(filename)
	if err != nil {
		return err
	}
	defer archive.Close()

	reader, err := document.Open()
	if err != nil {
		return fmt.Errorf("open Document.xml: %w", err)
	}
	defer reader.Close()

	written, err := io.Copy(writer, io.LimitReader(reader, maxDocumentSize+1))
	if err != nil {
		return fmt.Errorf("write Document.xml: %w", err)
	}
	if written > maxDocumentSize {
		return fmt.Errorf("mmap Document.xml exceeds %d bytes", maxDocumentSize)
	}
	return nil
}

func openDocument(filename string) (*zip.ReadCloser, *zip.File, error) {
	archive, err := zip.OpenReader(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("open mmap archive: %w", err)
	}

	if len(archive.File) > maxArchiveEntries {
		archive.Close()
		return nil, nil, fmt.Errorf("mmap archive contains too many entries: %d", len(archive.File))
	}

	document, err := findDocument(archive.File)
	if err != nil {
		archive.Close()
		return nil, nil, err
	}
	if document.UncompressedSize64 > maxDocumentSize {
		archive.Close()
		return nil, nil, fmt.Errorf("mmap Document.xml exceeds %d bytes", maxDocumentSize)
	}

	return archive, document, nil
}

func findDocument(files []*zip.File) (*zip.File, error) {
	var candidate *zip.File
	for _, file := range files {
		cleanName := path.Clean(strings.ReplaceAll(file.Name, "\\", "/"))
		if strings.EqualFold(cleanName, "Document.xml") {
			return file, nil
		}
		if strings.EqualFold(path.Base(cleanName), "Document.xml") && candidate == nil {
			candidate = file
		}
	}
	if candidate != nil {
		return candidate, nil
	}
	return nil, ErrDocumentNotFound
}

type xmlNode struct {
	name     string
	attrs    map[string]string
	text     string
	children []*xmlNode
}

func parseXML(reader io.Reader) (*xmlNode, error) {
	decoder := xml.NewDecoder(reader)
	var root *xmlNode
	stack := make([]*xmlNode, 0, 32)
	nodeCount := 0

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse Document.xml: %w", err)
		}

		switch value := token.(type) {
		case xml.StartElement:
			if len(stack) >= maxXMLDepth {
				return nil, fmt.Errorf("Document.xml exceeds maximum depth %d", maxXMLDepth)
			}
			nodeCount++
			if nodeCount > maxXMLNodes {
				return nil, fmt.Errorf("Document.xml exceeds maximum node count %d", maxXMLNodes)
			}

			node := &xmlNode{name: value.Name.Local, attrs: make(map[string]string, len(value.Attr))}
			for _, attribute := range value.Attr {
				node.attrs[strings.ToLower(attribute.Name.Local)] = attribute.Value
			}
			if len(stack) == 0 {
				if root != nil {
					return nil, errors.New("Document.xml contains multiple root elements")
				}
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, node)
			}
			stack = append(stack, node)

		case xml.CharData:
			if len(stack) > 0 {
				current := stack[len(stack)-1]
				current.text += string(value)
			}

		case xml.EndElement:
			if len(stack) == 0 {
				return nil, errors.New("Document.xml contains an unmatched closing element")
			}
			stack = stack[:len(stack)-1]
		}
	}

	if root == nil {
		return nil, errors.New("Document.xml is empty")
	}
	if len(stack) != 0 {
		return nil, errors.New("Document.xml contains unclosed elements")
	}
	return root, nil
}

func findRootTopic(root *xmlNode) *xmlNode {
	if equalName(root.name, "OneTopic") {
		return firstDirectChild(root, "Topic")
	}
	for _, child := range root.children {
		if equalName(child.name, "OneTopic") {
			if topic := firstDirectChild(child, "Topic"); topic != nil {
				return topic
			}
		}
		if topic := findRootTopic(child); topic != nil {
			return topic
		}
	}
	return nil
}

func convertTopic(source *xmlNode) *model.Topic {
	topic := &model.Topic{
		Title: topicTitle(source),
		Notes: topicNotes(source),
		Links: topicLinks(source),
	}
	if topic.Title == "" {
		topic.Title = "Untitled"
	}

	for _, child := range source.children {
		if !equalName(child.name, "SubTopics") {
			continue
		}
		for _, candidate := range child.children {
			if equalName(candidate.name, "Topic") {
				topic.Children = append(topic.Children, convertTopic(candidate))
			}
		}
	}
	return topic
}

func topicTitle(topic *xmlNode) string {
	for _, child := range topic.children {
		if !equalName(child.name, "Text") {
			continue
		}
		if plainText := cleanText(child.attrs["plaintext"]); plainText != "" {
			return plainText
		}
		if text := cleanText(collectText(child, false)); text != "" {
			return text
		}
	}
	return cleanText(topic.attrs["text"])
}

func topicNotes(topic *xmlNode) string {
	for _, child := range topic.children {
		if !isNoteElement(child.name) {
			continue
		}
		if plainText := findAttribute(child, "plaintext", true); plainText != "" {
			return cleanText(plainText)
		}
		if text := cleanText(collectText(child, true)); text != "" {
			return text
		}
	}
	return ""
}

// isNoteElement reports whether name is a known MindManager notes container element.
// Using exact matching instead of substring contains avoids false positives from
// elements whose names incidentally contain "note" (e.g., FooterNoteGroup).
func isNoteElement(name string) bool {
	switch strings.ToLower(name) {
	case "notesgroup", "notesxhtmldata", "notesplain", "notes":
		return true
	default:
		return false
	}
}

func topicLinks(topic *xmlNode) []model.Link {
	seen := make(map[string]struct{})
	links := make([]model.Link, 0, 1)

	var walk func(*xmlNode)
	walk = func(node *xmlNode) {
		if node != topic && equalName(node.name, "Topic") {
			return
		}
		for _, key := range []string{"url", "uri", "href"} {
			if raw := strings.TrimSpace(node.attrs[key]); safeURL(raw) {
				if _, exists := seen[raw]; !exists {
					seen[raw] = struct{}{}
					title := cleanText(node.attrs["title"])
					if title == "" {
						title = raw
					}
					links = append(links, model.Link{Title: title, URL: raw})
				}
			}
		}
		for _, child := range node.children {
			walk(child)
		}
	}
	walk(topic)
	return links
}

func safeURL(raw string) bool {
	if raw == "" {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if parsed.Scheme == "" {
		return strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "./") || strings.HasPrefix(raw, "../")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "mailto":
		return true
	default:
		return false
	}
}

func findAttribute(node *xmlNode, key string, recursive bool) string {
	if value := node.attrs[key]; value != "" {
		return value
	}
	if recursive {
		for _, child := range node.children {
			if value := findAttribute(child, key, true); value != "" {
				return value
			}
		}
	}
	return ""
}

func collectText(node *xmlNode, skipTopics bool) string {
	var builder strings.Builder
	var walk func(*xmlNode)
	walk = func(current *xmlNode) {
		if current != node && skipTopics && equalName(current.name, "Topic") {
			return
		}
		if text := strings.TrimSpace(current.text); text != "" {
			if builder.Len() > 0 {
				builder.WriteByte(' ')
			}
			builder.WriteString(text)
		}
		for _, child := range current.children {
			walk(child)
		}
	}
	walk(node)
	return builder.String()
}

func cleanText(value string) string {
	return strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " ")
}

func firstDirectChild(node *xmlNode, name string) *xmlNode {
	for _, child := range node.children {
		if equalName(child.name, name) {
			return child
		}
	}
	return nil
}

func equalName(left, right string) bool {
	return strings.EqualFold(left, right)
}
