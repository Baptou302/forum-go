package web

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func idParam(r *http.Request) int64 {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	return id
}

func threadPath(id int64) string { return "/sujet/" + strconv.FormatInt(id, 10) }

func viewerID(r *http.Request) int64 {
	if u := currentUser(r); u != nil {
		return u.ID
	}
	return 0
}

func redirectBack(w http.ResponseWriter, r *http.Request, fallback string) {
	ref := r.Referer()
	if ref == "" {
		ref = fallback
	}
	http.Redirect(w, r, ref, http.StatusSeeOther)
}

func (a *App) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	a.render(w, r, "error.html", map[string]any{
		"Title": "Page introuvable", "Code": 404,
		"Message": "Oups, cette page n'existe pas (ou plus).",
	})
}

func (a *App) serverError(w http.ResponseWriter, err error) {
	http.Error(w, "Erreur serveur : "+err.Error(), http.StatusInternalServerError)
}
