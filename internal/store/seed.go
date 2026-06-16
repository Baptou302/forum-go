package store

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Store) seedIfEmpty() error {
	if s.CountUsers() > 0 {
		return nil
	}
	users, err := s.seedUsers()
	if err != nil {
		return err
	}
	cats, err := s.seedCategories()
	if err != nil {
		return err
	}
	return s.seedThreads(users, cats)
}

func hashPw(pw string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h)
}

func (s *Store) seedUsers() (map[string]int64, error) {
	type su struct{ name, email, pw, role, bio string }
	list := []su{
		{"herve", "herve@forum.fr", "admin1234", "admin", "Fondateur et administrateur du forum. Collectionneur depuis 30 ans."},
		{"julia", "julia@forum.fr", "forum1234", "moderator", "Modératrice passionnée par les éditions limitées."},
		{"marc", "marc@forum.fr", "forum1234", "member", "Joueur de belote du dimanche, à la recherche de la perle rare."},
		{"sophie", "sophie@forum.fr", "forum1234", "member", "J'adore les jeux de cartes anciens et les designs vintage."},
	}
	ids := map[string]int64{}
	for _, u := range list {
		created, err := s.CreateUser(u.name, u.email, hashPw(u.pw))
		if err != nil {
			return nil, err
		}
		if created.Role != u.role {
			s.SetUserRole(created.ID, u.role)
		}
		s.UpdateProfile(created.ID, u.bio)
		ids[u.name] = created.ID
	}
	return ids, nil
}

func (s *Store) seedCategories() (map[string]int64, error) {
	type sc struct {
		name, slug, desc, icon, color string
	}
	list := []sc{
		{"Annonces & Règlement", "annonces", "Les annonces officielles de l'équipe et le règlement du forum.", "megaphone", "#ef4444"},
		{"Présentations", "presentations", "Nouveau venu ? Présente-toi à la communauté !", "hand", "#f59e0b"},
		{"Discussions générales", "general", "Tout ce qui touche à la passion des 4 couleurs.", "messages-square", "#6366f1"},
		{"Collections & Éditions limitées", "collections", "Montrez vos plus belles pièces et trouvez la perle rare.", "gem", "#ec4899"},
		{"Achat / Vente / Échange", "marche", "La place de marché entre membres de confiance.", "shopping-cart", "#10b981"},
		{"Entraide & Questions", "entraide", "Une question ? La communauté est là pour vous aider.", "life-buoy", "#06b6d4"},
	}
	ids := map[string]int64{}
	for i, c := range list {
		res, err := s.db.Exec(`INSERT INTO categories (name, slug, description, icon, color, position)
			VALUES (?, ?, ?, ?, ?, ?)`, c.name, c.slug, c.desc, c.icon, c.color, i+1)
		if err != nil {
			return nil, err
		}
		ids[c.slug], _ = res.LastInsertId()
	}
	return ids, nil
}

func (s *Store) backdate(threadID int64, t time.Time) {
	s.db.Exec(`UPDATE threads SET created_at = ?, updated_at = ? WHERE id = ?`, t, t, threadID)
}

func (s *Store) backdatePost(postID int64, t time.Time) {
	s.db.Exec(`UPDATE posts SET created_at = ? WHERE id = ?`, t, postID)
	s.db.Exec(`UPDATE threads SET updated_at = ? WHERE id = (SELECT thread_id FROM posts WHERE id = ?)`, t, postID)
}
