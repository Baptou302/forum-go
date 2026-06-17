package backend

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// ─── Models ───────────────────────────────────────────────────────────────────

type User struct {
	ID       int
	Email    string
	Username string
}

type Post struct {
	ID           int
	UserID       int
	Username     string
	Title        string
	Content      string
	CreatedAt    string
	Categories   []string
	Likes        int
	Dislikes     int
	CommentCount int
	UserReaction string
}

type Comment struct {
	ID           int
	PostID       int
	UserID       int
	Username     string
	Content      string
	CreatedAt    string
	Likes        int
	Dislikes     int
	UserReaction string
}

type Category struct {
	ID   int
	Name string
}

type PageData struct {
	User        *User
	Posts       []Post
	Post        *Post
	Comments    []Comment
	Categories  []Category
	Error       string
	Tab         string
	FilterCat   string
	FilterMine  string
	FilterLiked string
	ShowNewPost bool
	IsHome      bool
}

// ─── Globals ──────────────────────────────────────────────────────────────────

var db *sql.DB
var tmpl *template.Template

var funcMap = template.FuncMap{
	"truncate": func(s string, n int) string {
		r := []rune(s)
		if len(r) <= n {
			return s
		}
		return string(r[:n]) + "…"
	},
}

// ─── Init (appelé depuis main) ────────────────────────────────────────────────

func Init() {
	initDB()
	var err error
	tmpl, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("templates:", err)
	}
}

// ─── Router ───────────────────────────────────────────────────────────────────

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/post/", postRouter)
	mux.HandleFunc("/comment/", commentRouter)
	mux.HandleFunc("/new-post", newPostHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return mux
}

func postRouter(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	switch {
	case len(parts) == 2:
		postViewHandler(w, r)
	case len(parts) == 3 && parts[2] == "comment":
		commentHandler(w, r)
	case len(parts) == 3 && parts[2] == "react":
		reactPostHandler(w, r)
	default:
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}

func commentRouter(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 3 && parts[2] == "react" {
		reactCommentHandler(w, r)
	} else {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}

// ─── Database ─────────────────────────────────────────────────────────────────

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./forum.db?_foreign_keys=on")
	if err != nil {
		log.Fatal(err)
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS post_categories (
			post_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			PRIMARY KEY (post_id, category_id),
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS post_reactions (
			user_id INTEGER NOT NULL,
			post_id INTEGER NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('like','dislike')),
			PRIMARY KEY (user_id, post_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS comment_reactions (
			user_id INTEGER NOT NULL,
			comment_id INTEGER NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('like','dislike')),
			PRIMARY KEY (user_id, comment_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
		)`,
		`INSERT OR IGNORE INTO categories (name) VALUES ('Tech'),('Science'),('Art'),('Music'),('Sport'),('Gaming'),('Other')`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			log.Fatal("DB init:", err)
		}
	}
}

// ─── Auth helpers ─────────────────────────────────────────────────────────────

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

// ─── Template helpers ─────────────────────────────────────────────────────────

func render(w http.ResponseWriter, name string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		log.Println("template error:", err)
	}
}

func allCategories() []Category {
	rows, _ := db.Query(`SELECT id, name FROM categories ORDER BY name`)
	defer rows.Close()
	var cats []Category
	for rows.Next() {
		var c Category
		rows.Scan(&c.ID, &c.Name)
		cats = append(cats, c)
	}
	return cats
}

func loadPost(id, userID int) (*Post, error) {
	p := &Post{}
	err := db.QueryRow(`
		SELECT p.id, p.user_id, u.username, p.title, p.content,
		       strftime('%d/%m/%Y %H:%M', p.created_at),
		       (SELECT COUNT(*) FROM post_reactions WHERE post_id=p.id AND type='like'),
		       (SELECT COUNT(*) FROM post_reactions WHERE post_id=p.id AND type='dislike'),
		       (SELECT COUNT(*) FROM comments WHERE post_id=p.id)
		FROM posts p JOIN users u ON p.user_id=u.id WHERE p.id=?`, id,
	).Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content,
		&p.CreatedAt, &p.Likes, &p.Dislikes, &p.CommentCount)
	if err != nil {
		return nil, err
	}
	rows, _ := db.Query(
		`SELECT c.name FROM categories c JOIN post_categories pc ON c.id=pc.category_id WHERE pc.post_id=?`, id)
	for rows.Next() {
		var cat string
		rows.Scan(&cat)
		p.Categories = append(p.Categories, cat)
	}
	rows.Close()
	if userID > 0 {
		db.QueryRow(`SELECT type FROM post_reactions WHERE user_id=? AND post_id=?`, userID, id).Scan(&p.UserReaction)
	}
	return p, nil
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	user := getUser(r)
	fc, fm, fl := r.URL.Query().Get("cat"), r.URL.Query().Get("mine"), r.URL.Query().Get("liked")
	showNew := r.URL.Query().Get("new") == "1"

	query := `
		SELECT p.id, p.user_id, u.username, p.title, p.content,
		       strftime('%d/%m/%Y %H:%M', p.created_at),
		       (SELECT COUNT(*) FROM post_reactions WHERE post_id=p.id AND type='like'),
		       (SELECT COUNT(*) FROM post_reactions WHERE post_id=p.id AND type='dislike'),
		       (SELECT COUNT(*) FROM comments WHERE post_id=p.id)
		FROM posts p JOIN users u ON p.user_id=u.id`

	var conds []string
	var args []interface{}
	if fc != "" {
		conds = append(conds, `p.id IN (SELECT post_id FROM post_categories pc JOIN categories c ON pc.category_id=c.id WHERE c.name=?)`)
		args = append(args, fc)
	}
	if fm == "1" && user != nil {
		conds = append(conds, `p.user_id=?`)
		args = append(args, user.ID)
	}
	if fl == "1" && user != nil {
		conds = append(conds, `p.id IN (SELECT post_id FROM post_reactions WHERE user_id=? AND type='like')`)
		args = append(args, user.ID)
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY p.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content,
			&p.CreatedAt, &p.Likes, &p.Dislikes, &p.CommentCount)
		catRows, _ := db.Query(
			`SELECT c.name FROM categories c JOIN post_categories pc ON c.id=pc.category_id WHERE pc.post_id=?`, p.ID)
		for catRows.Next() {
			var cat string
			catRows.Scan(&cat)
			p.Categories = append(p.Categories, cat)
		}
		catRows.Close()
		if user != nil {
			db.QueryRow(`SELECT type FROM post_reactions WHERE user_id=? AND post_id=?`, user.ID, p.ID).Scan(&p.UserReaction)
		}
		posts = append(posts, p)
	}

	render(w, "home.html", PageData{
		User: user, Posts: posts, Categories: allCategories(),
		FilterCat: fc, FilterMine: fm, FilterLiked: fl,
		ShowNewPost: showNew, IsHome: fc == "" && fm == "" && fl == "",
	})
}

func postViewHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.Split(strings.Trim(r.URL.Path, "/"), "/")[1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	user := getUser(r)
	uid := 0
	if user != nil {
		uid = user.ID
	}
	p, err := loadPost(id, uid)
	if err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	rows, _ := db.Query(`
		SELECT c.id, c.post_id, c.user_id, u.username, c.content,
		       strftime('%d/%m/%Y %H:%M', c.created_at),
		       (SELECT COUNT(*) FROM comment_reactions WHERE comment_id=c.id AND type='like'),
		       (SELECT COUNT(*) FROM comment_reactions WHERE comment_id=c.id AND type='dislike')
		FROM comments c JOIN users u ON c.user_id=u.id
		WHERE c.post_id=? ORDER BY c.created_at ASC`, id)
	defer rows.Close()
	var comments []Comment
	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &c.CreatedAt, &c.Likes, &c.Dislikes)
		if user != nil {
			db.QueryRow(`SELECT type FROM comment_reactions WHERE user_id=? AND comment_id=?`,
				user.ID, c.ID).Scan(&c.UserReaction)
		}
		comments = append(comments, c)
	}
	render(w, "post.html", PageData{User: user, Post: p, Comments: comments})
}

func newPostHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/?new=1", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	title, content, cats := strings.TrimSpace(r.FormValue("title")), strings.TrimSpace(r.FormValue("content")), r.Form["categories"]
	if title == "" || content == "" {
		render(w, "home.html", PageData{User: user, Categories: allCategories(), ShowNewPost: true, Error: "Title and content are required"})
		return
	}
	res, err := db.Exec(`INSERT INTO posts (user_id, title, content) VALUES (?,?,?)`, user.ID, title, content)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	pid, _ := res.LastInsertId()
	for _, catName := range cats {
		var catID int
		db.QueryRow(`SELECT id FROM categories WHERE name=?`, catName).Scan(&catID)
		if catID > 0 {
			db.Exec(`INSERT INTO post_categories (post_id, category_id) VALUES (?,?)`, pid, catID)
		}
	}
	http.Redirect(w, r, "/post/"+strconv.FormatInt(pid, 10), http.StatusSeeOther)
}

func commentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	postID, err := strconv.Atoi(strings.Split(strings.Trim(r.URL.Path, "/"), "/")[1])
	if err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	if content := strings.TrimSpace(r.FormValue("content")); content != "" {
		db.Exec(`INSERT INTO comments (post_id, user_id, content) VALUES (?,?,?)`, postID, user.ID, content)
	}
	http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}

func reactPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	postID, err := strconv.Atoi(strings.Split(strings.Trim(r.URL.Path, "/"), "/")[1])
	if err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	rtype := r.FormValue("type")
	if rtype != "like" && rtype != "dislike" {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}
	var existing string
	db.QueryRow(`SELECT type FROM post_reactions WHERE user_id=? AND post_id=?`, user.ID, postID).Scan(&existing)
	if existing == rtype {
		db.Exec(`DELETE FROM post_reactions WHERE user_id=? AND post_id=?`, user.ID, postID)
	} else {
		db.Exec(`INSERT INTO post_reactions (user_id, post_id, type) VALUES (?,?,?)
			ON CONFLICT(user_id,post_id) DO UPDATE SET type=excluded.type`, user.ID, postID, rtype)
	}
	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/post/" + strconv.Itoa(postID)
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func reactCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	commentID, err := strconv.Atoi(strings.Split(strings.Trim(r.URL.Path, "/"), "/")[1])
	if err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	rtype := r.FormValue("type")
	if rtype != "like" && rtype != "dislike" {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}
	var existing string
	db.QueryRow(`SELECT type FROM comment_reactions WHERE user_id=? AND comment_id=?`, user.ID, commentID).Scan(&existing)
	if existing == rtype {
		db.Exec(`DELETE FROM comment_reactions WHERE user_id=? AND comment_id=?`, user.ID, commentID)
	} else {
		db.Exec(`INSERT INTO comment_reactions (user_id, comment_id, type) VALUES (?,?,?)
			ON CONFLICT(user_id,comment_id) DO UPDATE SET type=excluded.type`, user.ID, commentID, rtype)
	}
	var postID int
	db.QueryRow(`SELECT post_id FROM comments WHERE id=?`, commentID).Scan(&postID)
	http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
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
		if _, err = db.Exec(`INSERT INTO users (email, username, password_hash) VALUES (?,?,?)`, email, username, string(hash)); err != nil {
			render(w, "auth.html", PageData{Tab: "register", Error: "Email or username already taken"})
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	render(w, "auth.html", PageData{Tab: "register"})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
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
		return
	}
	render(w, "auth.html", PageData{Tab: "login"})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		db.Exec(`DELETE FROM sessions WHERE token=?`, c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Expires: time.Unix(0, 0), Path: "/"})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
