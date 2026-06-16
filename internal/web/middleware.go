package web

import (
	"context"
	"net/http"
	"net/url"

	"forum/internal/store"
)

type ctxKey string

const userKey ctxKey = "user"

const sessionCookie = "forum_session"

func (a *App) loadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(sessionCookie); err == nil {
			if u, err := a.store.UserBySession(c.Value); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), userKey, u))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func currentUser(r *http.Request) *store.User {
	u, _ := r.Context().Value(userKey).(*store.User)
	return u
}

func (a *App) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if currentUser(r) == nil {
			setFlash(w, "error", "Vous devez être connecté pour accéder à cette page.")
			http.Redirect(w, r, "/login?next="+url.QueryEscape(r.URL.Path), http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (a *App) requireModerator(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := currentUser(r)
		if u == nil || !u.IsModerator() {
			http.Error(w, "Accès réservé à l'équipe de modération.", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
