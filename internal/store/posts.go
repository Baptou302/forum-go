package store

import (
	"database/sql"
	"errors"
)

const postSelect = `
	SELECT p.id, p.thread_id, p.user_id, p.content, p.created_at,
	       u.username, u.avatar_color, u.role,
	       (SELECT COUNT(*) FROM likes l WHERE l.post_id = p.id) AS like_count,
	       EXISTS(SELECT 1 FROM likes l WHERE l.post_id = p.id AND l.user_id = ?) AS liked
	FROM posts p JOIN users u ON p.user_id = u.id`

func scanPost(row interface{ Scan(...any) error }) (*Post, error) {
	p := &Post{Author: &User{}}
	var liked int
	err := row.Scan(&p.ID, &p.ThreadID, &p.UserID, &p.Content, &p.CreatedAt,
		&p.Author.Username, &p.Author.AvatarColor, &p.Author.Role, &p.LikeCount, &liked)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Author.ID, p.LikedByMe = p.UserID, liked == 1
	return p, nil
}

func (s *Store) CreatePost(threadID, userID int64, content string) (*Post, error) {
	res, err := s.db.Exec(`INSERT INTO posts (thread_id, user_id, content) VALUES (?, ?, ?)`,
		threadID, userID, content)
	if err != nil {
		return nil, err
	}
	s.touchThread(threadID)
	id, _ := res.LastInsertId()
	return s.GetPost(id)
}

func (s *Store) GetPost(id int64) (*Post, error) {
	return scanPost(s.db.QueryRow(postSelect+` WHERE p.id = ?`, int64(0), id))
}

func (s *Store) ListPostsByThread(threadID, viewerID int64) ([]*Post, error) {
	rows, err := s.db.Query(postSelect+` WHERE p.thread_id = ? ORDER BY p.created_at ASC`, viewerID, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) DeletePost(id int64) error {
	_, err := s.db.Exec(`DELETE FROM posts WHERE id = ?`, id)
	return err
}
