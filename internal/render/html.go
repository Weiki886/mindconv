package render

import (
	"errors"
	"html/template"
	"io"

	"github.com/Weiki886/mindconv/internal/model"
)

var htmlDocument = template.Must(template.New("document").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Root.Title}}</title>
  <style>
    :root { color-scheme: light dark; font-family: system-ui, sans-serif; }
    body { margin: 0; background: #f6f7fb; color: #1f2937; }
    main { max-width: 960px; margin: 0 auto; padding: 48px 24px 80px; }
    h1 { margin: 0 0 28px; font-size: clamp(2rem, 6vw, 3.5rem); }
    .tree, .tree ul { list-style: none; margin: 0; padding-left: 24px; }
    .tree { padding-left: 0; }
    .topic { position: relative; margin: 12px 0; padding: 14px 16px; border: 1px solid #dbe1ea; border-radius: 12px; background: #fff; box-shadow: 0 4px 18px rgb(15 23 42 / 6%); }
    .topic-title { font-weight: 700; }
    .notes { margin: 8px 0 0; color: #4b5563; white-space: pre-wrap; }
    .links { margin: 8px 0 0; padding-left: 20px; }
    a { color: #2563eb; overflow-wrap: anywhere; }
    @media (prefers-color-scheme: dark) {
      body { background: #111827; color: #f3f4f6; }
      .topic { background: #1f2937; border-color: #374151; }
      .notes { color: #cbd5e1; }
      a { color: #93c5fd; }
    }
  </style>
</head>
<body>
  <main>
    <h1>{{.Root.Title}}</h1>
    <ul class="tree">{{template "topic" .Root}}</ul>
  </main>
</body>
</html>
{{define "topic"}}
<li>
  <article class="topic">
    <div class="topic-title">{{.Title}}</div>
    {{if .Notes}}<p class="notes">{{.Notes}}</p>{{end}}
    {{if .Links}}<ul class="links">{{range .Links}}<li><a href="{{.URL}}">{{.Title}}</a></li>{{end}}</ul>{{end}}
  </article>
  {{if .Children}}<ul>{{range .Children}}{{template "topic" .}}{{end}}</ul>{{end}}
</li>
{{end}}
`))

// HTML writes a self-contained static HTML representation of a map.
func HTML(writer io.Writer, mindMap *model.Map) error {
	if mindMap == nil || mindMap.Root == nil {
		return errors.New("render html: map has no root topic")
	}
	if err := htmlDocument.ExecuteTemplate(writer, "document", mindMap); err != nil {
		return errors.New("render html: " + err.Error())
	}
	return nil
}
