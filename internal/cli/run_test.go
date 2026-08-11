package cli

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"--version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Run() code = %d, stderr = %s", code, stderr.String())
	}
	if got := stdout.String(); got != "mindconv "+Version+"\n" {
		t.Fatalf("version output = %q", got)
	}
}

func TestHelpReturnsSuccess(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Run() code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Usage: mindconv") {
		t.Fatalf("help output = %q", stderr.String())
	}
}

func TestUnsupportedFormat(t *testing.T) {
	var stderr bytes.Buffer
	code := Run([]string{"--format", "pdf", "sample.mmap"}, &bytes.Buffer{}, &stderr)
	if code != 2 {
		t.Fatalf("Run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "unsupported format") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestDefaultOutputPath(t *testing.T) {
	input := filepath.Join("maps", "project.mmap")
	for _, format := range []string{"md", "html", "xml"} {
		want := filepath.Join("maps", "project."+format)
		if got := defaultOutputPath(input, format); got != want {
			t.Errorf("defaultOutputPath(%q) = %q, want %q", format, got, want)
		}
	}
}

func TestWriteFileRefusesOverwrite(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "output.md")
	if err := writeFile(filename, []byte("first"), false); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filename, []byte("second"), false); err == nil {
		t.Fatal("writeFile() error = nil, want overwrite error")
	}
	if err := writeFile(filename, []byte("second"), true); err != nil {
		t.Fatal(err)
	}
}

func TestRunConvertsWithOptionsAfterInput(t *testing.T) {
	input := createMMAP(t)
	output := filepath.Join(t.TempDir(), "result.html")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{input, "--format", "html", "--output", output}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %s", code, stderr.String())
	}
	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Sample Map") {
		t.Fatalf("output does not contain the root topic: %s", content)
	}
}

func TestRunWritesMarkdownToStdout(t *testing.T) {
	input := createMMAP(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--stdout", input}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Sample Map") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunWritesXMLToStdout(t *testing.T) {
	input := createMMAP(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--format", "xml", "--stdout", input}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %s", code, stderr.String())
	}
	if got := stdout.String(); got != sampleXML {
		t.Fatalf("stdout = %q, want %q", got, sampleXML)
	}
}

const sampleXML = `<Map><OneTopic><Topic><Text PlainText="Sample Map"/><SubTopics><Topic><Text PlainText="Child"/></Topic></SubTopics></Topic></OneTopic></Map>`

func createMMAP(t *testing.T) string {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "sample.mmap")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	document, err := archive.Create("Document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := document.Write([]byte(sampleXML)); err != nil {
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
