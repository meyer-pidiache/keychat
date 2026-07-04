package templates

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed base.html pages/*.html
var TemplatesFS embed.FS

func ServePage(name string) http.HandlerFunc {
	patterns := []string{"base.html", "pages/" + name + ".html"}
	tmpl, err := template.ParseFS(TemplatesFS, patterns...)
	if err != nil {
		log.Printf("ERROR parsing template %s: %v", name, err)
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, nil); err != nil {
			log.Printf("ERROR executing template %s: %v", name, err)
		}
	}
}
