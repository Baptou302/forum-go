package web

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (a *App) handleLoginForm(w http.ResponseWriter, r *http.Request) {
	if currentUser(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	a.render(w, r, "login.html", map[string]any{
		"Title": "Connexion", "Active": "login", "Next": r.URL.Query().Get("next"),
	})
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	identifier := strings.TrimSpace(r.FormValue("identifier"))
	password, next := r.FormValue("password"), r.FormValue("next")

	user, err := a.store.GetUserByUsername(identifier)
	if err != nil {
		user, err = a.store.GetUserByEmail(strings.ToLower(identifier))
	}
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		a.render(w, r, "login.html", map[string]any{
			"Title": "Connexion", "Active": "login",
			"Error": "Identifiants incorrects. Réessayez.", "Identifier": identifier, "Next": next,
		})
		return
	}
	a.startSession(w, user.ID)
	setFlash(w, "success", "Content de vous revoir, "+user.Username+" !")
	if next == "" || !strings.HasPrefix(next, "/") {
		next = "/"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		a.store.DeleteSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(0, 0)})
	setFlash(w, "success", "Vous êtes déconnecté. À bientôt !")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) startSession(w http.ResponseWriter, userID int64) {
	sess, err := a.store.CreateSession(userID)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: sess.ID, Path: "/",
		Expires: sess.ExpiresAt, HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}
