package mmap

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const sampleDocument = `<?xml version="1.0" encoding="UTF-8"?>
<ap:Map xmlns:ap="http://schemas.mindjet.com/MindManager/Application/2003">
  <ap:OneTopic>
    <ap:Topic OId="root">
      <ap:Text PlainText="Project *Plan*" />
      <ap:NotesGroup><ap:NotesXhtmlData><p>Root note</p></ap:NotesXhtmlData></ap:NotesGroup>
      <ap:Hyperlink Uri="https://example.com/docs" Title="Documentation" />
      <ap:Hyperlink Uri="javascript:alert(1)" Title="Unsafe" />
      <ap:SubTopics>
        <ap:Topic OId="child-1">
          <ap:Text PlainText="Research" />
          <ap:SubTopics>
            <ap:Topic OId="grandchild"><ap:Text PlainText="Users" /></ap:Topic>
          </ap:SubTopics>
        </ap:Topic>
        <ap:Topic OId="child-2"><ap:Text PlainText="Delivery" /></ap:Topic>
      </ap:SubTopics>
    </ap:Topic>
  </ap:OneTopic>
</ap:Map>`

func TestParseFile(t *testing.T) {
	filename := createArchive(t, "Document.xml", sampleDocument)

	mindMap, err := ParseFile(filename)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	if got := mindMap.Root.Title; got != "Project *Plan*" {
		t.Fatalf("root title = %q", got)
	}
	if got := mindMap.Root.Notes; got != "Root note" {
		t.Fatalf("root notes = %q", got)
	}
	if got := len(mindMap.Root.Links); got != 1 {
		t.Fatalf("root links = %d, want 1", got)
	}
	if got := mindMap.Root.Links[0].URL; got != "https://example.com/docs" {
		t.Fatalf("root link URL = %q", got)
	}
	if got := len(mindMap.Root.Children); got != 2 {
		t.Fatalf("root children = %d, want 2", got)
	}
	if got := mindMap.Root.Children[0].Children[0].Title; got != "Users" {
		t.Fatalf("grandchild title = %q", got)
	}
}

func TestParseFileFindsNestedDocument(t *testing.T) {
	filename := createArchive(t, "map/Document.xml", sampleDocument)
	if _, err := ParseFile(filename); err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
}

func TestParseFileWithoutDocument(t *testing.T) {
	filename := createArchive(t, "Preview.xml", "<preview />")
	_, err := ParseFile(filename)
	if !errors.Is(err, ErrDocumentNotFound) {
		t.Fatalf("ParseFile() error = %v, want ErrDocumentNotFound", err)
	}
}

func TestParseFileWithMalformedXML(t *testing.T) {
	filename := createArchive(t, "Document.xml", "<Map><Topic></Map>")
	if _, err := ParseFile(filename); err == nil {
		t.Fatal("ParseFile() error = nil, want malformed XML error")
	}
}

func createArchive(t *testing.T, entryName, content string) string {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "sample.mmap")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}

	archive := zip.NewWriter(file)
	entry, err := archive.Create(entryName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return filename
}
