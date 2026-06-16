# Forum des 4 Couleurs ♠ ♥ ♦ ♣

Forum web complet pour les passionnés de jeux de cartes traditionnelles.
Backend **Go** (routeur chi + SQLite), interface **Tailwind CSS** server-rendered
avec mode sombre, icônes Lucide et interactions Alpine.js. Tout est embarqué dans
un seul binaire : aucun build front-end à lancer.

## Lancer le projet

```bash
go run ./cmd/server
```

Puis ouvrir **http://localhost:8080**.

La base SQLite (`data/forum.db`) est créée et remplie automatiquement de données
de démonstration au premier démarrage.

### Variables d'environnement (optionnel)

| Variable     | Défaut           | Rôle                          |
|--------------|------------------|-------------------------------|
| `FORUM_ADDR` | `:8080`          | Adresse d'écoute du serveur   |
| `FORUM_DB`   | `data/forum.db`  | Chemin de la base SQLite      |

## Comptes de démonstration

| Pseudo  | Mot de passe | Rôle          |
|---------|--------------|---------------|
| `herve` | `admin1234`  | Administrateur |
| `julia` | `forum1234`  | Modératrice   |
| `marc`  | `forum1234`  | Membre        |
| `sophie`| `forum1234`  | Membre        |

> Le **premier compte inscrit** sur une base vierge devient automatiquement administrateur.

## Fonctionnalités

- 🔐 **Authentification** : inscription, connexion, déconnexion, mots de passe hashés (bcrypt), sessions en base
- 👥 **Rôles** : membre / modérateur / administrateur
- 🗂️ **Catégories** thématiques avec statistiques (sujets, messages, dernière activité)
- 💬 **Sujets & réponses** avec compteur de vues
- ❤️ **Likes** sur les sujets et les réponses
- 🔍 **Recherche** plein-texte sur les sujets
- 👤 **Profils** publics avec bio et statistiques
- 🛡️ **Modération** : épingler, verrouiller, supprimer sujets et réponses
- 🌙 **Mode sombre**, design responsive et moderne

## Architecture

```
cmd/server/main.go      → point d'entrée
internal/store/         → base de données (schéma, requêtes, données de démo)
internal/web/           → serveur HTTP (routeur, handlers, middleware, rendu)
templates/              → pages HTML (Tailwind), partials et mise en page de base
static/                 → CSS complémentaire
assets.go               → embarque templates + static dans le binaire
```

## Construire un binaire autonome

```bash
go build -o forum.exe ./cmd/server
./forum.exe
```

Le binaire contient tout (templates + CSS) : il est portable et n'a besoin que
du dossier `data/` (créé automatiquement) pour la base.
