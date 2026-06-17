package backend

import (
	"net/http"
	"strconv"
	"strings"
)

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
