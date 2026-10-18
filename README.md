# E-Commerce CLI Platform

Plateforme d'achat e-commerce dédiée aux clients sur interface en ligne de commande, développée en Go.

## Fonctionnalités

### Authentification *(1 pt)*
- Connexion avec email et mot de passe
- Inscription avec email et mot de passe
- Réinitialisation du mot de passe (sans être connecté) via email
- Confirmation d'inscription via un code de confirmation unique et aléatoire

### Recherche de produit *(2 pts)*
- Recherche par nom, prix, description, catégorie ou prix TTC

### Panier *(2 pts)*
- Ajout d'un ou plusieurs produits au panier
- Modification ou suppression de la quantité d'un produit
- Affichage du total TTC (livraison toujours gratuite)
- Panier lié à un compte utilisateur, toujours à jour
- Identifiant métier du panier (ex: `BSK-1KH8E7`)

### Paiement *(2 pts)*
- Paiement d'un panier sauvegardé
- Saisie des informations bancaires : numéro de carte, date d'expiration, CVC
- Pas d'intégration Stripe requise

### Commandes *(2 pts)*
- Affichage des commandes et de leur statut :
  - En attente
  - Payé
  - En cours de livraison
  - Livré
  - Annulé (avec raison)
- Chaque commande contient : produits, date, identifiant métier unique (ex: `CMD-1F2S8B`)

### Gestion des commandes — Admin *(1 pt)*
- Changement du statut d'une commande
- Création d'une commande pour n'importe quel client (liaison produits + utilisateur)

### Gestion des utilisateurs — Admin *(1 pt)*
- Créer, modifier, supprimer et confirmer un compte utilisateur

### Gestion des produits — Admin *(1 pt)*
- Ajout de produits avec génération automatique d'un identifiant métier (ex: `PDT-7D2K8N`)

---

## Architecture

### Interfaces *(2 pts)*
- **Client** et **administrateur** : deux applications en ligne de commande séparées (`cmd/client`, `cmd/admin`)
- Elles communiquent avec une API HTTP écrite avec `net/http` uniquement (aucun framework)
- Authentification par JWT (access token + refresh token)

### Persistance des données *(1 pt)*
- Module `database/sql` uniquement (aucun ORM), driver `pgx`
- Base de données PostgreSQL dans un conteneur Docker
- Le schéma (`internal/db/schema.sql`) et les produits de démo (`internal/db/seed.sql`) sont appliqués au démarrage de l'API

### Charmbracelet *(3 pts)*
Utilisation de [charm.land](https://charm.land) pour des interfaces réactives et vivantes :
- [bubbletea](https://github.com/charmbracelet/bubbletea) : boucle de l'application (un `tea.Model` par application)
- [bubbles](https://github.com/charmbracelet/bubbles) : tableaux, champ de recherche
- [huh](https://github.com/charmbracelet/huh) : formulaires (connexion, inscription, paiement, admin)
- [lipgloss](https://github.com/charmbracelet/lipgloss) : styles
- [wish](https://github.com/charmbracelet/wish) : serveur SSH

### Interface SSH *(2 pts)*
- Accès au client et à l'administration via SSH, sans rien installer
- Port `2222` : client, port `2223` : administration
- Chaque connexion SSH a sa propre session (son propre modèle et ses propres tokens)

---

## Arborescence

```
cmd/
  api/        serveur HTTP
  client/     application client (terminal)
  admin/      application d'administration (terminal)
  ssh/        serveur SSH qui sert le client et l'admin
internal/
  auth/ user/ product/ cart/ order/   un package par domaine : handler -> service -> repository
  api/        routes et démarrage du serveur HTTP
  db/         connexion, schéma et données de démo
  mail/       envoi des codes par SMTP
  ref/        génération des identifiants métier (PDT-, BSK-, CMD-)
  apiclient/  client HTTP utilisé par les interfaces
  tui/client/ tui/admin/   interfaces bubbletea
  sshserver/  serveur SSH (wish)
```

---

## Lancement

### 1. Configuration (une seule fois)

```bash
cp .env.example .env
```

Dans `.env`, renseigner `JWT_SECRET` (32 caractères minimum, par exemple `openssl rand -hex 32`).
Le compte administrateur est créé au démarrage de l'API à partir de `ADMIN_EMAIL` et `ADMIN_PASSWORD`.

### 2. Démarrer les conteneurs

```bash
docker compose up -d
```

| Service | Adresse |
|---|---|
| API | http://localhost:8000 |
| Adminer (base de données) | http://localhost:8080 |
| Mailpit (codes de confirmation et de réinitialisation) | http://localhost:8025 |
| SSH client / admin | ports `2222` / `2223` |

### 3. Démarrer l'API

```bash
docker compose exec go air
```

`air` recompile et relance l'API à chaque modification (configuration dans `.air.toml`).
Vérification : `curl localhost:8000/health` doit retourner `{"status":"ok"}`.

### 4. Lancer les interfaces

Dans le terminal :

```bash
docker compose exec go go run ./cmd/client
docker compose exec go go run ./cmd/admin
```

Ou via SSH, après avoir démarré le serveur SSH dans un autre terminal :

```bash
docker compose exec go go run ./cmd/ssh

ssh -p 2222 localhost   # client
ssh -p 2223 localhost   # administration
```

Les interfaces appellent l'API sur `http://localhost:8000` par défaut. Pour en utiliser une autre (par exemple un tunnel ngrok), définir la variable `API_URL`.

### 5. Accès à distance (soutenance)

```bash
ngrok tcp 2222    # client en SSH : ssh -p <port> <hôte> avec l'adresse donnée par ngrok
ngrok tcp 2223    # administration en SSH
ngrok http 8000   # API publique, pour une TUI lancée ailleurs : API_URL=https://xxx.ngrok.app
```

En SSH, rien à installer côté correcteur : le serveur SSH tourne à côté de l'API et l'appelle en local.

---

## Utilisation du client

| Écran | Touches |
|---|---|
| Connexion | `↑/↓` choisir, `entrée` valider, `échap` retour, `ctrl+r` renvoyer un code |
| Catalogue | taper pour rechercher (voir ci-dessous), `↑/↓` naviguer, `entrée` détail, `+` ajouter au panier, `tab` panier, `ctrl+o` commandes |
| Panier | `+ / -` quantité, `suppr` retirer, `p` payer, `o` commandes, `échap` catalogue |
| Commandes | `↑/↓` naviguer, `entrée` détail, `échap` retour |

`ctrl+c` quitte l'application depuis n'importe quel écran.

Un compte non confirmé qui se connecte arrive sur l'écran de confirmation (`ctrl+r` pour recevoir un nouveau code).

### Recherche

Les mots sans préfixe cherchent dans le nom ; les préfixes ajoutent des filtres, combinables :

| Préfixe | Filtre | Exemple |
|---|---|---|
| *(aucun)* | nom | `clavier` |
| `cat:` | catégorie | `cat:livres` |
| `desc:` | description | `desc:bio` |
| `min:` / `max:` | prix TTC en euros | `max:49,99` |
| `minht:` / `maxht:` | prix HT en euros | `minht:10` |

Exemple : `clavier cat:informatique max:100`.

---

## Utilisation de l'administration

| Écran | Touches |
|---|---|
| Menu | `↑/↓` naviguer, `entrée` ouvrir, `q` quitter |
| Utilisateurs | `n` nouveau, `e` modifier (email, rôle), `c` confirmer, `r` basculer le rôle, `d` puis `o` supprimer |
| Produits | `n` nouveau (prix HT en euros, TVA) |
| Commandes | `entrée` détail, `s` changer le statut (seuls les passages autorisés sont proposés), `n` nouvelle commande pour un client |
| Formulaires | `entrée` valider, `échap` annuler, `espace` cocher |

---

## Développement

```bash
docker compose exec go go test ./...   # tests
docker compose exec go go vet ./...    # analyse statique
docker compose exec go gofmt -w .      # formatage
```

---

## Pénalités

- Code non maîtrisé ou inexplicable
- Utilisation abusive de l'IA
- Utilisation de frameworks hors Charmbracelet
- Non-respect des fonctionnalités
- Dépassement du temps imparti en soutenance
- Livrable absent sur MyGES
- Livrable sous forme de lien Git (doit être une archive ZIP)
- Participation inégale des membres du groupe
