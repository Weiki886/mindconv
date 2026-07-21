package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Weiki886/mindconv/internal/mmap"
	"github.com/Weiki886/mindconv/internal/render"
)

const Version = "0.1.0"

// Run executes the mindconv command and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("mindconv", flag.ContinueOnError)
	flags.SetOutput(stderr)

	var format string
	var output string
	var force bool
	var toStdout bool
	var showVersion bool
	flags.StringVar(&format, "format", "md", "output format: md or html")
	flags.StringVar(&format, "f", "md", "output format: md or html")
	flags.StringVar(&output, "output", "", "output file path; use - for stdout")
	flags.StringVar(&output, "o", "", "output file path; use - for stdout")
	flags.BoolVar(&force, "force", false, "overwrite an existing output file")
	flags.BoolVar(&toStdout, "stdout", false, "write converted content to stdout")
	flags.BoolVar(&showVersion, "version", false, "print version and exit")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: mindconv [options] <input.mmap>")
		fmt.Fprintln(stderr)
		flags.PrintDefaults()
	}

	if err := flags.Parse(reorderArgs(args)); errors.Is(err, flag.ErrHelp) {
		return 0
	} else if err != nil {
		return 2
	}
	if showVersion {
		fmt.Fprintf(stdout, "mindconv %s\n", Version)
		return 0
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return 2
	}
	if toStdout {
		if output != "" && output != "-" {
			fmt.Fprintln(stderr, "mindconv: --stdout cannot be combined with --output")
			return 2
		}
		output = "-"
	}

	format = normalizeFormat(format)
	if format == "" {
		fmt.Fprintln(stderr, "mindconv: unsupported format; use md or html")
		return 2
	}

	input := flags.Arg(0)
	mindMap, err := mmap.ParseFile(input)
	if err != nil {
		fmt.Fprintf(stderr, "mindconv: %v\n", err)
		return 1
	}

	var converted bytes.Buffer
	switch format {
	case "md":
		err = render.Markdown(&converted, mindMap)
	case "html":
		err = render.HTML(&converted, mindMap)
	}
	if err != nil {
		fmt.Fprintf(stderr, "mindconv: %v\n", err)
		return 1
	}

	if output == "-" {
		if _, err := converted.WriteTo(stdout); err != nil {
			fmt.Fprintf(stderr, "mindconv: write stdout: %v\n", err)
			return 1
		}
		return 0
	}
	if output == "" {
		output = defaultOutputPath(input, format)
	}
	if err := writeFile(output, converted.Bytes(), force); err != nil {
		fmt.Fprintf(stderr, "mindconv: %v\n", err)
		return 1
	}

	fmt.Fprintf(stderr, "Wrote %s\n", output)
	return 0
}

func reorderArgs(args []string) []string {
	options := make([]string, 0, len(args))
	positionals := make([]string, 0, 1)
	valueFlags := map[string]bool{
		"-f": true, "--format": true,
		"-o": true, "--output": true,
	}

	for index := 0; index < len(args); index++ {
		argument := args[index]
		if argument == "--" {
			positionals = append(positionals, args[index+1:]...)
			break
		}
		if !strings.HasPrefix(argument, "-") || argument == "-" {
			positionals = append(positionals, argument)
			continue
		}

		options = append(options, argument)
		name := argument
		if separator := strings.IndexByte(argument, '='); separator >= 0 {
			name = argument[:separator]
		}
		if valueFlags[name] && !strings.Contains(argument, "=") && index+1 < len(args) {
			index++
			options = append(options, args[index])
		}
	}

	return append(options, positionals...)
}

func normalizeFormat(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "md", "markdown":
		return "md"
	case "html", "htm":
		return "html"
	default:
		return ""
	}
}

func defaultOutputPath(input, format string) string {
	extension := filepath.Ext(input)
	return strings.TrimSuffix(input, extension) + "." + format
}

func writeFile(filename string, content []byte, force bool) error {
	flags := os.O_WRONLY | os.O_CREATE
	if force {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	file, err := os.OpenFile(filename, flags, 0o644)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("output file already exists: %s (use --force to overwrite)", filename)
	}
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}

	if _, err := file.Write(content); err != nil {
		file.Close()
		return fmt.Errorf("write output file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close output file: %w", err)
	}
	return nil
}
