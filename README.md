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
- **Client** : application en ligne de commande
- **Administrateur** : application en ligne de commande séparée
- Communication via serveur HTTP avec le module `net/http` (aucun framework)

### Persistance des données *(1 pt)*
- Module `database/sql` uniquement (aucun ORM)
- Base de données PostgreSQL via Docker, ou SQLite

### Charmbracelet *(3 pts)*
Utilisation de [charm.land](https://charm.land) pour des interfaces réactives et vivantes :
- [bubbletea](https://github.com/charmbracelet/bubbletea)
- [bubbles](https://github.com/charmbracelet/bubbles)
- [huh](https://github.com/charmbracelet/huh)
- [lipgloss](https://github.com/charmbracelet/lipgloss)

### Interface SSH *(2 pts)*
- Accès client et administrateur via SSH en plus du terminal classique

---

## Lancement

```bash
# Démarrer la base de données
docker compose up -d

# Démarrer le serveur
docker compose exec go air (air sert a lancer le serveur en mode watch, settings dans .air.toml)

test server : curl localhost:8000/health, doit retourner "ok"

# Lancer le client
docker compose exec go go run ./cmd/client
# go run main.go client


# Lancer l'interface admin
docker compose exec go go run ./cmd/admin
# go run main.go admin

# Formatage du code
docker compose exec go gofmt -w .
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

++ utilisation de la suite Charmbracelet (donc Huh, Bubbletea, Wish ? Lipgloss ?) 
++ forward avec ngrok pendant soutenance pour acces global