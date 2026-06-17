package backend

import (
	"database/sql"
	"html/template"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func Init() {
	initDB()
	var err error
	tmpl, err = template.New("").Funcs(FuncMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("templates:", err)
	}
}

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
