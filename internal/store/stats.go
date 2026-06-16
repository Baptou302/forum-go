package store

func (s *Store) GlobalStats() *Stats {
	st := &Stats{}
	s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&st.Members)
	s.db.QueryRow(`SELECT COUNT(*) FROM threads`).Scan(&st.Threads)
	s.db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&st.Posts)
	s.db.QueryRow(`SELECT username FROM users ORDER BY created_at DESC LIMIT 1`).Scan(&st.NewestMember)
	return st
}
