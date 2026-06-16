package store

import (
	"database/sql"
	"errors"
	"strings"
)

const threadSelect = `
	SELECT t.id, t.category_id, t.user_id, t.title, t.content, t.pinned, t.locked,
	       t.views, t.created_at, t.updated_at,
	       u.username, u.avatar_color, u.role,
	       c.name, c.slug, c.icon, c.color,
	       (SELECT COUNT(*) FROM posts p WHERE p.thread_id = t.id) AS reply_count,
	       (SELECT COUNT(*) FROM likes l WHERE l.thread_id = t.id) AS like_count,
	       EXISTS(SELECT 1 FROM likes l WHERE l.thread_id = t.id AND l.user_id = ?) AS liked
	FROM threads t
	JOIN users u ON t.user_id = u.id
	JOIN categories c ON t.category_id = c.id`

func scanThread(row interface{ Scan(...any) error }) (*Thread, error) {
	t := &Thread{Author: &User{}}
	var liked int
	err := row.Scan(&t.ID, &t.CategoryID, &t.UserID, &t.Title, &t.Content,
		&t.Pinned, &t.Locked, &t.Views, &t.CreatedAt, &t.UpdatedAt,
		&t.Author.Username, &t.Author.AvatarColor, &t.Author.Role,
		&t.CategoryName, &t.CategorySlug, &t.CategoryIcon, &t.CategoryColor,
		&t.ReplyCount, &t.LikeCount, &liked)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.Author.ID, t.LikedByMe, t.LastActivity = t.UserID, liked == 1, t.UpdatedAt
	return t, nil
}

func (s *Store) queryThreads(viewerID int64, where string, args ...any) ([]*Thread, error) {
	rows, err := s.db.Query(threadSelect+" "+where, append([]any{viewerID}, args...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Thread
	for rows.Next() {
		t, err := scanThread(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) CreateThread(categoryID, userID int64, title, content string) (*Thread, error) {
	res, err := s.db.Exec(`INSERT INTO threads (category_id, user_id, title, content)
		VALUES (?, ?, ?, ?)`, categoryID, userID, title, content)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetThread(id, userID)
}

func (s *Store) GetThread(id, viewerID int64) (*Thread, error) {
	return scanThread(s.db.QueryRow(threadSelect+` WHERE t.id = ?`, viewerID, id))
}

func (s *Store) ListThreadsByCategory(categoryID, viewerID int64) ([]*Thread, error) {
	return s.queryThreads(viewerID, `WHERE t.category_id = ? ORDER BY t.pinned DESC, t.updated_at DESC`, categoryID)
}

func (s *Store) ListRecentThreads(limit int, viewerID int64) ([]*Thread, error) {
	return s.queryThreads(viewerID, `ORDER BY t.updated_at DESC LIMIT ?`, limit)
}

func (s *Store) ListThreadsByUser(userID, viewerID int64) ([]*Thread, error) {
	return s.queryThreads(viewerID, `WHERE t.user_id = ? ORDER BY t.created_at DESC`, userID)
}

func (s *Store) SearchThreads(q string, viewerID int64) ([]*Thread, error) {
	like := "%" + strings.TrimSpace(q) + "%"
	return s.queryThreads(viewerID, `WHERE t.title LIKE ? OR t.content LIKE ?
		ORDER BY t.updated_at DESC LIMIT 50`, like, like)
}
