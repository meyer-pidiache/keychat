package templates

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"
)

func NewTemplateFS(fsys fs.FS) (*template.Template, error) {
	t := template.New("base.html").Funcs(template.FuncMap{
		"safeHTML":   func(s string) template.HTML { return template.HTML(s) },
		"truncate":   truncate,
		"formatTime": formatTime,
	})

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		_, err = t.Parse(string(data))
		return err
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

func formatTime(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}
