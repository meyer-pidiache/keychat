package templates

import (
	"embed"
	"net/http"
)

//go:embed pages/*.html
var PagesFS embed.FS

func ServePage(name string) http.HandlerFunc {
	path := "pages/" + name + ".html"
	data, err := PagesFS.ReadFile(path)
	if err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}
}
