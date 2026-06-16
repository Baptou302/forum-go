package store

func (s *Store) toggleLike(col string, userID, targetID int64) (bool, int, error) {
	res, err := s.db.Exec(`DELETE FROM likes WHERE user_id = ? AND `+col+` = ?`, userID, targetID)
	if err != nil {
		return false, 0, err
	}
	liked := false
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := s.db.Exec(`INSERT INTO likes (user_id, `+col+`) VALUES (?, ?)`, userID, targetID); err != nil {
			return false, 0, err
		}
		liked = true
	}
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM likes WHERE `+col+` = ?`, targetID).Scan(&count)
	return liked, count, nil
}

func (s *Store) ToggleThreadLike(userID, threadID int64) (bool, int, error) {
	return s.toggleLike("thread_id", userID, threadID)
}

func (s *Store) TogglePostLike(userID, postID int64) (bool, int, error) {
	return s.toggleLike("post_id", userID, postID)
}
