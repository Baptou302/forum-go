package backend

import (
	"net/http"
	"strings"
)

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
