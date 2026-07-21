package handler

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	templateDir string
}

func NewRenderer(templateDir string) *Renderer {
	return &Renderer{templateDir: templateDir}
}

func (r *Renderer) Render(w http.ResponseWriter, tmpl string, data interface{}) {
	files := []string{
		filepath.Join(r.templateDir, "base.html"),
		filepath.Join(r.templateDir, tmpl),
	}

	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("template parse error: %v", err)
		http.Error(w, "failed to compile template", http.StatusInternalServerError)
		return
	}

	// Buffer output to catch execution errors before writing headers to ResponseWriter
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		log.Printf("template execute error: %v", err)
		http.Error(w, "failed to render template", http.StatusInternalServerError)
		return
	}

	// Output buffer contents upon successful rendering
	_, _ = buf.WriteTo(w)
}