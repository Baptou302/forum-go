package store

import "time"

type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	Role         string
	Bio          string
	AvatarColor  string
	CreatedAt    time.Time

	ThreadCount int
	PostCount   int
}

func (u *User) IsAdmin() bool     { return u.Role == "admin" }
func (u *User) IsModerator() bool { return u.Role == "admin" || u.Role == "moderator" }

type Category struct {
	ID          int64
	Name        string
	Slug        string
	Description string
	Icon        string
	Color       string
	Position    int

	ThreadCount    int
	PostCount      int
	LastThread     *Thread
	LastActivityAt time.Time
}

type Thread struct {
	ID         int64
	CategoryID int64
	UserID     int64
	Title      string
	Content    string
	Pinned     bool
	Locked     bool
	Views      int
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Author        *User
	CategoryName  string
	CategorySlug  string
	CategoryIcon  string
	CategoryColor string
	ReplyCount    int
	LikeCount     int
	LikedByMe     bool
	LastActivity  time.Time
}

type Post struct {
	ID        int64
	ThreadID  int64
	UserID    int64
	Content   string
	CreatedAt time.Time

	Author    *User
	LikeCount int
	LikedByMe bool
}

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
}

type Stats struct {
	Members      int
	Threads      int
	Posts        int
	NewestMember string
}
