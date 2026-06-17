# Forum — Go + SQLite

Forum web développé en Go avec une base de données SQLite. Permet aux utilisateurs de créer des posts, laisser des commentaires, liker/disliker du contenu et filtrer les publications par catégorie.

---

## Sommaire

- [Fonctionnalités](#fonctionnalités)
- [Technologies utilisées](#technologies-utilisées)
- [Structure du projet](#structure-du-projet)
- [Prérequis](#prérequis)
- [Installation et lancement](#installation-et-lancement)
- [Utilisation](#utilisation)
- [Routes HTTP](#routes-http)
- [Schéma de la base de données](#schéma-de-la-base-de-données)
- [Design — palette 4 couleurs](#design--palette-4-couleurs)

---

## Fonctionnalités

### Authentification
- Inscription avec email, nom d'utilisateur et mot de passe
- Connexion / déconnexion
- Sessions sécurisées via cookie HTTP-only (durée : 24h)
- Mots de passe chiffrés avec **bcrypt**
- Un seul compte par email (contrôle d'unicité)

### Posts
- Création de posts (titre + contenu) réservée aux utilisateurs connectés
- Association d'une ou plusieurs **catégories** à chaque post
- Lecture des posts accessible à tous (connecté ou non)
- Extrait automatique du contenu sur la page d'accueil (180 caractères)

### Commentaires
- Ajout de commentaires sur un post (utilisateurs connectés uniquement)
- Affichage des commentaires visible par tous

### Likes / Dislikes
- Like ou dislike sur les posts et les commentaires
- Système toggle : cliquer une deuxième fois annule la réaction
- Changement de réaction : le dislike remplace automatiquement le like, et vice-versa
- Compteurs visibles par tous les visiteurs

### Filtres
- **Toutes les publications** — vue par défaut
- **Par catégorie** — filtre dans la sidebar (Tech, Science, Art, Music, Sport, Gaming, Other)
- **Mes posts** — posts créés par l'utilisateur connecté
- **Posts likés** — posts aimés par l'utilisateur connecté

---

## Technologies utilisées

| Couche      | Technologie                        |
|-------------|------------------------------------|
| Langage     | Go 1.21+                           |
| Base de données | SQLite 3 (`mattn/go-sqlite3`)  |
| Chiffrement | bcrypt (`golang.org/x/crypto`)     |
| Templates   | `html/template` (stdlib Go)        |
| Frontend    | HTML5 + CSS3 (aucun framework JS)  |

---

## Structure du projet

```
forum-js/
│
├── main.go                  # Point d'entrée — 12 lignes
│
├── backend/
│   └── handlers.go          # Tout le backend : modèles, BDD, handlers, router
│
├── templates/
│   ├── layout.html          # Header et footer partagés (define "header" / "footer")
│   ├── home.html            # Page d'accueil : liste des posts + formulaire nouveau post
│   ├── post.html            # Page détail d'un post + commentaires
│   └── auth.html            # Page login / register (onglets)
│
├── static/
│   └── style.css            # Feuille de style complète — palette 4 couleurs
│
├── go.mod                   # Module Go + dépendances
├── go.sum                   # Checksums des dépendances (auto-généré)
└── forum.db                 # Base de données SQLite (créée au premier lancement)
```

---

## Prérequis

- **Go 1.21** ou supérieur → [golang.org/dl](https://golang.org/dl/)
- **GCC** (nécessaire pour `go-sqlite3` qui utilise CGo)
  - Windows : installer [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) ou [MSYS2](https://www.msys2.org/)
  - Linux : `sudo apt install gcc`
  - macOS : `xcode-select --install`

Vérifier l'installation :

```bash
go version
gcc --version
```

---

## Installation et lancement

### 1. Cloner / récupérer le projet

```bash
git clone <url-du-repo>
cd forum-js
```

### 2. Télécharger les dépendances

```bash
go mod tidy
```

Cette commande télécharge automatiquement :
- `github.com/mattn/go-sqlite3` — driver SQLite pour Go
- `golang.org/x/crypto` — package bcrypt pour le chiffrement des mots de passe

### 3. Lancer le serveur

```bash
go run main.go
```

Le terminal affiche :

```
Forum running → http://localhost:8080
```

### 4. Ouvrir dans le navigateur

```
http://localhost:8080
```

> La base de données `forum.db` est créée automatiquement au premier lancement avec toutes les tables et les catégories par défaut.

---

## Utilisation

### S'inscrire
1. Cliquer sur **Register** en haut à droite
2. Renseigner : email, nom d'utilisateur, mot de passe
3. Redirection automatique vers la page de connexion

### Se connecter
1. Cliquer sur **Login**
2. Entrer email + mot de passe
3. Session active pendant **24 heures**

### Créer un post
1. Être connecté
2. Cliquer sur **+ New Post** dans la barre de navigation
3. Renseigner un titre, un contenu, et sélectionner une ou plusieurs catégories
4. Cliquer sur **Publish**

### Commenter
1. Ouvrir un post en cliquant sur son titre
2. Écrire dans le champ en bas de page
3. Cliquer sur **Post Comment**

### Liker / Disliker
- Sur la page d'accueil ou la page d'un post, cliquer sur **▲** (like) ou **▼** (dislike)
- Cliquer à nouveau sur le même bouton pour annuler la réaction

### Filtrer les posts
Utiliser la sidebar à gauche :
- **All Posts** — tous les posts
- **My Posts** — vos posts uniquement (connecté requis)
- **Liked Posts** — posts que vous avez likés (connecté requis)
- **Categories** — filtrer par thème (Tech, Science, Art, etc.)

---

## Routes HTTP

| Méthode | Route                   | Description                                 | Auth requise |
|---------|-------------------------|---------------------------------------------|:------------:|
| GET     | `/`                     | Page d'accueil (liste des posts + filtres)  | Non          |
| GET     | `/?new=1`               | Afficher le formulaire de création de post  | Oui          |
| GET     | `/post/{id}`            | Page détail d'un post et ses commentaires   | Non          |
| POST    | `/new-post`             | Créer un nouveau post                       | Oui          |
| POST    | `/post/{id}/comment`    | Ajouter un commentaire                      | Oui          |
| POST    | `/post/{id}/react`      | Liker ou disliker un post                   | Oui          |
| POST    | `/comment/{id}/react`   | Liker ou disliker un commentaire            | Oui          |
| GET     | `/login`                | Afficher le formulaire de connexion         | Non          |
| POST    | `/login`                | Traiter la connexion                        | Non          |
| GET     | `/register`             | Afficher le formulaire d'inscription        | Non          |
| POST    | `/register`             | Traiter l'inscription                       | Non          |
| POST    | `/logout`               | Déconnecter l'utilisateur                   | Oui          |
| GET     | `/static/*`             | Fichiers statiques (CSS)                    | Non          |

---

## Schéma de la base de données

```
users
├── id            INTEGER  PRIMARY KEY AUTOINCREMENT
├── email         TEXT     UNIQUE NOT NULL
├── username      TEXT     UNIQUE NOT NULL
├── password_hash TEXT     NOT NULL          ← bcrypt
└── created_at    DATETIME DEFAULT CURRENT_TIMESTAMP

sessions
├── token         TEXT     PRIMARY KEY       ← token aléatoire 64 hex
├── user_id       INTEGER  → users.id
└── expires_at    DATETIME                   ← 24h après la connexion

categories
├── id            INTEGER  PRIMARY KEY AUTOINCREMENT
└── name          TEXT     UNIQUE NOT NULL

posts
├── id            INTEGER  PRIMARY KEY AUTOINCREMENT
├── user_id       INTEGER  → users.id
├── title         TEXT     NOT NULL
├── content       TEXT     NOT NULL
└── created_at    DATETIME DEFAULT CURRENT_TIMESTAMP

post_categories                              ← table de liaison N-N
├── post_id       INTEGER  → posts.id
└── category_id   INTEGER  → categories.id

comments
├── id            INTEGER  PRIMARY KEY AUTOINCREMENT
├── post_id       INTEGER  → posts.id
├── user_id       INTEGER  → users.id
├── content       TEXT     NOT NULL
└── created_at    DATETIME DEFAULT CURRENT_TIMESTAMP

post_reactions                               ← une réaction par (user, post)
├── user_id       INTEGER  → users.id
├── post_id       INTEGER  → posts.id
└── type          TEXT     CHECK IN ('like', 'dislike')

comment_reactions                            ← une réaction par (user, comment)
├── user_id       INTEGER  → users.id
├── comment_id    INTEGER  → comments.id
└── type          TEXT     CHECK IN ('like', 'dislike')
```

Les clés étrangères sont activées avec `ON DELETE CASCADE` : supprimer un utilisateur supprime automatiquement ses posts, commentaires et réactions.

---

## Design — palette 4 couleurs

Le design utilise exactement **4 couleurs** définies comme variables CSS :

| Variable | Valeur      | Usage                              |
|----------|-------------|------------------------------------|
| `--c1`   | `#1a1a2e`   | Fond principal (background)        |
| `--c2`   | `#16213e`   | Fond des cartes et surfaces        |
| `--c3`   | `#e94560`   | Accent rouge-rosé (boutons, liens) |
| `--c4`   | `#f5f5f5`   | Texte clair                        |

Toutes les autres valeurs (bordures, opacités, états hover) sont dérivées de ces 4 couleurs via `rgba()` et des variantes légèrement plus claires.
