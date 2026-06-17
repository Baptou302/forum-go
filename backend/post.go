package backend

import (
	"net/http"
	"strconv"
	"strings"
)

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

func postViewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.Split(strings.Trim(r.URL.Path, "/"), "/")[1])
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
		if db.QueryRow(`SELECT id FROM categories WHERE name=?`, catName).Scan(&catID); catID > 0 {
			db.Exec(`INSERT INTO post_categories (post_id, category_id) VALUES (?,?)`, pid, catID)
		}
	}
	http.Redirect(w, r, "/post/"+strconv.FormatInt(pid, 10), http.StatusSeeOther)
}
