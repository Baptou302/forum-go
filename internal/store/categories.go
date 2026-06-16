package store

import (
	"database/sql"
	"errors"
)

func (s *Store) ListCategories() ([]*Category, error) {
	rows, err := s.db.Query(`SELECT id, name, slug, description, icon, color, position
		FROM categories ORDER BY position, name`)
	if err != nil {
		return nil, err
	}

	var cats []*Category
	for rows.Next() {
		c := &Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.Icon, &c.Color, &c.Position); err != nil {
			rows.Close()
			return nil, err
		}
		cats = append(cats, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, c := range cats {
		s.countCategory(c)
		c.LastThread = s.lastThreadOf(c.ID)
		if c.LastThread != nil {
			c.LastActivityAt = c.LastThread.LastActivity
		}
	}
	return cats, nil
}

func (s *Store) countCategory(c *Category) {
	s.db.QueryRow(`SELECT COUNT(*) FROM threads WHERE category_id = ?`, c.ID).Scan(&c.ThreadCount)
	s.db.QueryRow(`SELECT COUNT(*) FROM posts p JOIN threads t ON p.thread_id = t.id
		WHERE t.category_id = ?`, c.ID).Scan(&c.PostCount)
}

func (s *Store) GetCategoryBySlug(slug string) (*Category, error) {
	c := &Category{}
	err := s.db.QueryRow(`SELECT id, name, slug, description, icon, color, position
		FROM categories WHERE slug = ?`, slug).
		Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.Icon, &c.Color, &c.Position)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (s *Store) lastThreadOf(categoryID int64) *Thread {
	t := &Thread{Author: &User{}}
	err := s.db.QueryRow(`SELECT t.id, t.title, t.updated_at, u.username, u.avatar_color
		FROM threads t JOIN users u ON t.user_id = u.id
		WHERE t.category_id = ? ORDER BY t.updated_at DESC LIMIT 1`, categoryID).
		Scan(&t.ID, &t.Title, &t.LastActivity, &t.Author.Username, &t.Author.AvatarColor)
	if err != nil {
		return nil
	}
	return t
}
