package web

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (a *App) handleRegisterForm(w http.ResponseWriter, r *http.Request) {
	if currentUser(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	a.render(w, r, "register.html", map[string]any{"Title": "Inscription", "Active": "inscription"})
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password, confirm := r.FormValue("password"), r.FormValue("confirm")

	fail := func(msg string) {
		a.render(w, r, "register.html", map[string]any{
			"Title": "Inscription", "Active": "inscription",
			"Error": msg, "Username": username, "Email": email,
		})
	}
	switch {
	case len(username) < 3 || len(username) > 24:
		fail("Le pseudo doit contenir entre 3 et 24 caractères.")
	case !strings.Contains(email, "@") || !strings.Contains(email, "."):
		fail("L'adresse e-mail n'est pas valide.")
	case len(password) < 6:
		fail("Le mot de passe doit contenir au moins 6 caractères.")
	case password != confirm:
		fail("Les deux mots de passe ne correspondent pas.")
	default:
		a.createAccount(w, r, username, email, password, fail)
	}
}

func (a *App) createAccount(w http.ResponseWriter, r *http.Request, username, email, password string, fail func(string)) {
	if _, err := a.store.GetUserByUsername(username); err == nil {
		fail("Ce pseudo est déjà pris.")
		return
	}
	if _, err := a.store.GetUserByEmail(email); err == nil {
		fail("Un compte existe déjà avec cette adresse e-mail.")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.serverError(w, err)
		return
	}
	user, err := a.store.CreateUser(username, email, string(hash))
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.startSession(w, user.ID)
	setFlash(w, "success", "Bienvenue "+user.Username+" ! Votre compte a été créé.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
