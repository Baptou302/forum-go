package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"forum"
)

func (a *App) parseTemplates() error {
	a.templates = map[string]*template.Template{}
	partials, _ := fs.Glob(forum.TemplatesFS, "templates/partials/*.html")
	shared := append([]string{"templates/base.html"}, partials...)
	pages, err := fs.Glob(forum.TemplatesFS, "templates/pages/*.html")
	if err != nil {
		return err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		files := append(append([]string{}, shared...), page)
		t, err := template.New(name).Funcs(funcMap()).ParseFS(forum.TemplatesFS, files...)
		if err != nil {
			return fmt.Errorf("parse %s: %w", name, err)
		}
		a.templates[name] = t
	}
	return nil
}

func (a *App) render(w http.ResponseWriter, r *http.Request, page string, data map[string]any) {
	t, ok := a.templates[page]
	if !ok {
		http.Error(w, "template introuvable: "+page, http.StatusInternalServerError)
		return
	}
	if data == nil {
		data = map[string]any{}
	}
	if _, ok := data["Title"]; !ok {
		data["Title"] = "Forum des 4 Couleurs"
	}
	if _, ok := data["Active"]; !ok {
		data["Active"] = ""
	}
	data["User"] = currentUser(r)
	data["Flash"] = consumeFlash(w, r)
	data["Year"] = time.Now().Year()

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		http.Error(w, "erreur de rendu: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"timeAgo": timeAgo, "dateFR": dateFR, "nl2br": nl2br,
		"excerpt": excerpt, "plural": plural,
		"roleLabel": roleLabel, "roleBadge": roleBadge,
		"initial": initialOf, "dict": dict,
	}
}
