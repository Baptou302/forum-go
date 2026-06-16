package web

import (
	"html/template"
	"io/fs"
	"net/http"

	"forum"
	"forum/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type App struct {
	store     *store.Store
	templates map[string]*template.Template
}

func New(st *store.Store) (*App, error) {
	a := &App{store: st}
	if err := a.parseTemplates(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(a.loadUser)

	staticFS, _ := fs.Sub(forum.StaticFS, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	r.Get("/", a.handleHome)
	r.Get("/categorie/{slug}", a.handleCategory)
	r.Get("/sujet/{id}", a.handleThread)
	r.Get("/u/{username}", a.handleProfile)
	r.Get("/recherche", a.handleSearch)

	r.Get("/inscription", a.handleRegisterForm)
	r.Post("/inscription", a.handleRegister)
	r.Get("/login", a.handleLoginForm)
	r.Post("/login", a.handleLogin)
	r.Post("/logout", a.handleLogout)

	r.Get("/nouveau-sujet", a.requireAuth(a.handleNewThreadForm))
	r.Post("/nouveau-sujet", a.requireAuth(a.handleCreateThread))
	r.Post("/sujet/{id}/repondre", a.requireAuth(a.handleReply))
	r.Post("/sujet/{id}/like", a.requireAuth(a.handleLikeThread))
	r.Post("/reponse/{id}/like", a.requireAuth(a.handleLikePost))
	r.Get("/parametres", a.requireAuth(a.handleSettingsForm))
	r.Post("/parametres", a.requireAuth(a.handleUpdateSettings))

	r.Post("/sujet/{id}/epingler", a.requireModerator(a.handlePinThread))
	r.Post("/sujet/{id}/verrouiller", a.requireModerator(a.handleLockThread))
	r.Post("/sujet/{id}/supprimer", a.requireModerator(a.handleDeleteThread))
	r.Post("/reponse/{id}/supprimer", a.requireModerator(a.handleDeletePost))

	r.NotFound(a.handleNotFound)
	return r
}
