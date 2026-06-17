package backend

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getUser(r *http.Request) *User {
	c, err := r.Cookie("session")
	if err != nil {
		return nil
	}
	var u User
	var exp time.Time
	err = db.QueryRow(
		`SELECT u.id, u.email, u.username, s.expires_at
		 FROM sessions s JOIN users u ON s.user_id = u.id
		 WHERE s.token = ?`, c.Value,
	).Scan(&u.ID, &u.Email, &u.Username, &exp)
	if err != nil || time.Now().After(exp) {
		return nil
	}
	return &u
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		render(w, "auth.html", PageData{Tab: "register"})
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	if email == "" || username == "" || password == "" {
		render(w, "auth.html", PageData{Tab: "register", Error: "All fields are required"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err = db.Exec(`INSERT INTO users (email, username, password_hash) VALUES (?,?,?)`,
		email, username, string(hash)); err != nil {
		render(w, "auth.html", PageData{Tab: "register", Error: "Email or username already taken"})
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		render(w, "auth.html", PageData{Tab: "login"})
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	var u User
	var hash string
	err := db.QueryRow(`SELECT id, email, username, password_hash FROM users WHERE email=?`, email).
		Scan(&u.ID, &u.Email, &u.Username, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		render(w, "auth.html", PageData{Tab: "login", Error: "Invalid email or password"})
		return
	}
	token := generateToken()
	exp := time.Now().Add(24 * time.Hour)
	db.Exec(`INSERT INTO sessions (token, user_id, expires_at) VALUES (?,?,?)`, token, u.ID, exp)
	http.SetCookie(w, &http.Cookie{Name: "session", Value: token, Expires: exp, HttpOnly: true, Path: "/"})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		db.Exec(`DELETE FROM sessions WHERE token=?`, c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Expires: time.Unix(0, 0), Path: "/"})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
