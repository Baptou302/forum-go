package web

import (
	"net/http"
	"strconv"
	"strings"
)

func (a *App) handleNewThreadForm(w http.ResponseWriter, r *http.Request) {
	cats, _ := a.store.ListCategories()
	a.render(w, r, "new_thread.html", map[string]any{
		"Title": "Nouveau sujet", "Active": "nouveau",
		"Categories": cats, "Selected": r.URL.Query().Get("categorie"),
	})
}

func (a *App) handleCreateThread(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	categoryID, _ := strconv.ParseInt(r.FormValue("category_id"), 10, 64)
	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))

	if categoryID == 0 || len(title) < 5 || len(content) < 10 {
		cats, _ := a.store.ListCategories()
		a.render(w, r, "new_thread.html", map[string]any{
			"Title": "Nouveau sujet", "Active": "nouveau", "Categories": cats,
			"Error":     "Choisissez une catégorie, un titre (5+ caractères) et un message (10+ caractères).",
			"FormTitle": title, "FormContent": content,
		})
		return
	}
	thread, err := a.store.CreateThread(categoryID, u.ID, title, content)
	if err != nil {
		a.serverError(w, err)
		return
	}
	setFlash(w, "success", "Votre sujet a été publié !")
	http.Redirect(w, r, threadPath(thread.ID), http.StatusSeeOther)
}

func (a *App) handleReply(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	id := idParam(r)
	thread, err := a.store.GetThread(id, u.ID)
	if err != nil {
		a.handleNotFound(w, r)
		return
	}
	content := strings.TrimSpace(r.FormValue("content"))
	switch {
	case thread.Locked:
		setFlash(w, "error", "Ce sujet est verrouillé, vous ne pouvez pas y répondre.")
	case len(content) < 2:
		setFlash(w, "error", "Votre réponse est trop courte.")
	default:
		if _, err := a.store.CreatePost(id, u.ID, content); err != nil {
			a.serverError(w, err)
			return
		}
		http.Redirect(w, r, threadPath(id)+"#fin", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, threadPath(id), http.StatusSeeOther)
}

func (a *App) handleLikeThread(w http.ResponseWriter, r *http.Request) {
	a.store.ToggleThreadLike(currentUser(r).ID, idParam(r))
	redirectBack(w, r, "/")
}

func (a *App) handleLikePost(w http.ResponseWriter, r *http.Request) {
	a.store.TogglePostLike(currentUser(r).ID, idParam(r))
	redirectBack(w, r, "/")
}

func (a *App) handleSettingsForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "settings.html", map[string]any{"Title": "Paramètres", "Active": "parametres"})
}

func (a *App) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	bio := strings.TrimSpace(r.FormValue("bio"))
	if len(bio) > 280 {
		bio = bio[:280]
	}
	a.store.UpdateProfile(u.ID, bio)
	setFlash(w, "success", "Profil mis à jour.")
	http.Redirect(w, r, "/u/"+u.Username, http.StatusSeeOther)
}
