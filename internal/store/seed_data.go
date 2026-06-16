package store

import "time"

func (s *Store) seedThreads(users, cats map[string]int64) error {
	type reply struct{ author, content string }
	type thread struct {
		cat, author, title, content string
		replies                     []reply
	}
	list := []thread{
		{"annonces", "herve", "Bienvenue sur le Forum des puristes des 4 Couleurs !",
			"Bonjour à tous et bienvenue !\n\nCe forum est dédié à tous les passionnés des jeux de cartes traditionnelles. Que vous soyez collectionneur, joueur ou simple curieux, vous êtes ici chez vous.\n\nMerci de rester respectueux et de lire le règlement avant de poster. Bonne visite !",
			[]reply{{"julia", "Merci Hervé pour ce nouvel espace, hâte d'échanger avec vous tous !"}, {"marc", "Enfin un forum sérieux sur le sujet. Bravo pour l'initiative."}}},
		{"collections", "julia", "Vous cherchez la perle rare ?",
			"Si vous recherchez des éditions limitées, c'est ici qu'il faut regarder. Postez vos recherches et vos trouvailles.\n\nJe commence : je cherche un jeu complet de l'édition 1965 en bon état.",
			[]reply{{"sophie", "Magnifique ! J'ai justement un doublon de cette édition, je t'envoie un message privé."}}},
		{"general", "sophie", "Partagez votre passion ici",
			"Si vous voulez simplement discuter avec d'autres personnes qui partagent votre passion, vous êtes au bon endroit ! Racontez-nous comment vous avez commencé.",
			[]reply{{"marc", "J'ai commencé grâce à mon grand-père qui m'a transmis sa collection. Que de souvenirs !"}, {"herve", "Bienvenue dans la grande famille des passionnés Sophie !"}}},
		{"presentations", "marc", "Bonjour à tous, je me présente",
			"Salut ! Je m'appelle Marc, je collectionne depuis 5 ans. Ravi de rejoindre cette communauté.", nil},
		{"entraide", "sophie", "Comment bien conserver ses cartes anciennes ?",
			"Bonjour, j'ai récupéré de vieux jeux et je voudrais les préserver de l'humidité. Avez-vous des conseils de conservation ?",
			[]reply{{"julia", "Range-les dans des pochettes sans acide et évite la lumière directe du soleil. Ça change tout !"}}},
	}

	base := time.Now().Add(-72 * time.Hour)
	for i, t := range list {
		th, err := s.CreateThread(cats[t.cat], users[t.author], t.title, t.content)
		if err != nil {
			return err
		}
		created := base.Add(time.Duration(i*6) * time.Hour)
		s.backdate(th.ID, created)
		for j, r := range t.replies {
			p, err := s.CreatePost(th.ID, users[r.author], r.content)
			if err != nil {
				return err
			}
			s.backdatePost(p.ID, created.Add(time.Duration(j+1)*time.Hour))
		}
	}
	return nil
}
