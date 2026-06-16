package store

import (
	"database/sql"
	"errors"
	"time"
)

func (s *Store) CreateSession(userID int64) (*Session, error) {
	token, err := newToken()
	if err != nil {
		return nil, err
	}
	expires := time.Now().Add(30 * 24 * time.Hour)
	if _, err := s.db.Exec(`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expires); err != nil {
		return nil, err
	}
	return &Session{ID: token, UserID: userID, ExpiresAt: expires}, nil
}

func (s *Store) UserBySession(token string) (*User, error) {
	var userID int64
	var expires time.Time
	err := s.db.QueryRow(`SELECT user_id, expires_at FROM sessions WHERE id = ?`, token).Scan(&userID, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if time.Now().After(expires) {
		s.DeleteSession(token)
		return nil, ErrNotFound
	}
	return s.GetUserByID(userID)
}

func (s *Store) DeleteSession(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, token)
	return err
}
