package mmap

import (
	"archive/zip"
	"bytes"
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

func TestWriteDocumentXML(t *testing.T) {
	const document = `<?xml version="1.0"?><Map><OneTopic/></Map>`
	filename := createArchive(t, "map/Document.xml", document)

	var output bytes.Buffer
	if err := WriteDocumentXML(filename, &output); err != nil {
		t.Fatalf("WriteDocumentXML() error = %v", err)
	}
	if got := output.String(); got != document {
		t.Fatalf("WriteDocumentXML() = %q, want %q", got, document)
	}
}

func TestWriteDocumentXMLWithoutDocument(t *testing.T) {
	filename := createArchive(t, "Preview.xml", "<preview />")
	if err := WriteDocumentXML(filename, &bytes.Buffer{}); !errors.Is(err, ErrDocumentNotFound) {
		t.Fatalf("WriteDocumentXML() error = %v, want ErrDocumentNotFound", err)
	}
}

func TestIsNoteElement(t *testing.T) {
	tests := []struct {
		name string
		elem string
		want bool
		why  string
	}{
		{"notesgroup is accepted", "NotesGroup", true, "standard MindManager notes container"},
		{"notesxhtmldata is accepted", "NotesXhtmlData", true, "inner notes data element"},
		{"notesplain is accepted", "NotesPlain", true, "plain text notes variant"},
		{"notes is accepted", "Notes", true, "short form notes element"},
		{"notesgroup lowercased", "notesgroup", true, "case-insensitive matching"},
		{"notesxhtmldata mixed case", "NotesXHTMLData", true, "case-insensitive matching"},
		{"footernotegroup is rejected", "FooterNoteGroup", false, "incidentally contains note but is not a notes container"},
		{"notesetting is rejected", "NoteSetting", false, "configuration element, not notes content"},
		{"unknown element", "RandomTag", false, "not a known notes container"},
		{"empty string", "", false, "no element name"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isNoteElement(tc.elem)
			if got != tc.want {
				t.Errorf("isNoteElement(%q) = %v, want %v (%s)", tc.elem, got, tc.want, tc.why)
			}
		})
	}
}

func TestTopicNotesWithDifferentElementNames(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{"NotesGroup with plaintext attr", `<Topic><NotesGroup PlainText="note from group"/><Text PlainText="T"/></Topic>`, "note from group"},
		{"NotesXhtmlData with plaintext attr", `<Topic><NotesXhtmlData PlainText="note from xhtml"/><Text PlainText="T"/></Topic>`, "note from xhtml"},
		{"NotesPlain with plaintext attr", `<Topic><NotesPlain PlainText="plain note"/><Text PlainText="T"/></Topic>`, "plain note"},
		{"Notes with plaintext attr", `<Topic><Notes PlainText="short note"/><Text PlainText="T"/></Topic>`, "short note"},
		{"FooterNoteGroup should be ignored", `<Topic><FooterNoteGroup PlainText="should be ignored"/><Text PlainText="T"/></Topic>`, ""},
		{"NoteSetting should be ignored", `<Topic><NoteSetting PlainText="should be ignored"/><Text PlainText="T"/></Topic>`, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := `<Map><OneTopic>` + tc.xml + `</OneTopic></Map>`
			filename := createArchive(t, "Document.xml", doc)
			mindMap, err := ParseFile(filename)
			if err != nil {
				t.Fatalf("ParseFile() error = %v", err)
			}
			if got := mindMap.Root.Notes; got != tc.want {
				t.Errorf("root notes = %q, want %q", got, tc.want)
			}
		})
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
