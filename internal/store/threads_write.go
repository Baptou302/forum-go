package store

import "time"

func (s *Store) IncrementViews(id int64) {
	s.db.Exec(`UPDATE threads SET views = views + 1 WHERE id = ?`, id)
}

func (s *Store) touchThread(threadID int64) {
	s.db.Exec(`UPDATE threads SET updated_at = ? WHERE id = ?`, time.Now(), threadID)
}

func (s *Store) SetPinned(id int64, pinned bool) error {
	_, err := s.db.Exec(`UPDATE threads SET pinned = ? WHERE id = ?`, b2i(pinned), id)
	return err
}

func (s *Store) SetLocked(id int64, locked bool) error {
	_, err := s.db.Exec(`UPDATE threads SET locked = ? WHERE id = ?`, b2i(locked), id)
	return err
}

func (s *Store) DeleteThread(id int64) error {
	_, err := s.db.Exec(`DELETE FROM threads WHERE id = ?`, id)
	return err
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
