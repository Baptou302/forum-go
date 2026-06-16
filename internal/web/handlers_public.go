package web

import (
	"net/http"
	"strings"

	"forum/internal/store"

	"github.com/go-chi/chi/v5"
)

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	cats, err := a.store.ListCategories()
	if err != nil {
		a.serverError(w, err)
		return
	}
	recent, err := a.store.ListRecentThreads(6, viewerID(r))
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "home.html", map[string]any{
		"Title": "Accueil", "Active": "accueil",
		"Categories": cats, "Recent": recent, "Stats": a.store.GlobalStats(),
	})
}

func (a *App) handleCategory(w http.ResponseWriter, r *http.Request) {
	cat, err := a.store.GetCategoryBySlug(chi.URLParam(r, "slug"))
	if err != nil {
		a.handleNotFound(w, r)
		return
	}
	threads, err := a.store.ListThreadsByCategory(cat.ID, viewerID(r))
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "category.html", map[string]any{"Title": cat.Name, "Category": cat, "Threads": threads})
}

func (a *App) handleThread(w http.ResponseWriter, r *http.Request) {
	id := idParam(r)
	thread, err := a.store.GetThread(id, viewerID(r))
	if err != nil {
		a.handleNotFound(w, r)
		return
	}
	a.store.IncrementViews(id)
	posts, err := a.store.ListPostsByThread(id, viewerID(r))
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, r, "thread.html", map[string]any{"Title": thread.Title, "Thread": thread, "Posts": posts})
}

func (a *App) handleProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := a.store.GetUserByUsername(chi.URLParam(r, "username"))
	if err != nil {
		a.handleNotFound(w, r)
		return
	}
	a.store.LoadUserStats(profile)
	threads, _ := a.store.ListThreadsByUser(profile.ID, viewerID(r))
	a.render(w, r, "profile.html", map[string]any{"Title": "@" + profile.Username, "Profile": profile, "Threads": threads})
}

func (a *App) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	var results []*store.Thread
	if q != "" {
		results, _ = a.store.SearchThreads(q, viewerID(r))
	}
	a.render(w, r, "search.html", map[string]any{
		"Title": "Recherche", "Active": "recherche", "Query": q, "Results": results,
	})
}
