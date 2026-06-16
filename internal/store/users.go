package store

import (
	"database/sql"
	"errors"
)

var avatarPalette = []string{
	"#6366f1", "#ec4899", "#f59e0b", "#10b981", "#3b82f6",
	"#8b5cf6", "#ef4444", "#14b8a6", "#f97316", "#06b6d4",
}

func colorFor(username string) string {
	var sum int
	for _, r := range username {
		sum += int(r)
	}
	return avatarPalette[sum%len(avatarPalette)]
}

const userCols = `id, username, email, password_hash, role, bio, avatar_color, created_at`

func scanUser(row interface{ Scan(...any) error }) (*User, error) {
	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.Role, &u.Bio, &u.AvatarColor, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (s *Store) CreateUser(username, email, passwordHash string) (*User, error) {
	role := "member"
	if s.CountUsers() == 0 {
		role = "admin"
	}
	res, err := s.db.Exec(`INSERT INTO users (username, email, password_hash, role, avatar_color)
		VALUES (?, ?, ?, ?, ?)`, username, email, passwordHash, role, colorFor(username))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetUserByID(id)
}

func (s *Store) GetUserByID(id int64) (*User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userCols+` FROM users WHERE id = ?`, id))
}

func (s *Store) GetUserByUsername(name string) (*User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userCols+` FROM users WHERE username = ? COLLATE NOCASE`, name))
}

func (s *Store) GetUserByEmail(email string) (*User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userCols+` FROM users WHERE email = ? COLLATE NOCASE`, email))
}

func (s *Store) CountUsers() int {
	var n int
	s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n
}

func (s *Store) LoadUserStats(u *User) {
	s.db.QueryRow(`SELECT COUNT(*) FROM threads WHERE user_id = ?`, u.ID).Scan(&u.ThreadCount)
	s.db.QueryRow(`SELECT COUNT(*) FROM posts WHERE user_id = ?`, u.ID).Scan(&u.PostCount)
}

func (s *Store) UpdateProfile(userID int64, bio string) error {
	_, err := s.db.Exec(`UPDATE users SET bio = ? WHERE id = ?`, bio, userID)
	return err
}

func (s *Store) SetUserRole(userID int64, role string) error {
	_, err := s.db.Exec(`UPDATE users SET role = ? WHERE id = ?`, role, userID)
	return err
}
