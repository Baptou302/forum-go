package backend

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

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

var db *sql.DB
var tmpl *template.Template

var FuncMap = template.FuncMap{
	"truncate": func(s string, n int) string {
		r := []rune(s)
		if len(r) <= n {
			return s
		}
		return string(r[:n]) + "…"
	},
}

func render(w http.ResponseWriter, name string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		log.Println("template error:", err)
	}
}
