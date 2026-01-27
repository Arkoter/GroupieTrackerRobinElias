# Groupie Tracker

## Description
Groupie Tracker est une application web développée en Go. Elle consiste à consommer une API externe pour recevoir, manipuler et afficher des données concernant des artistes musicaux et des groupes. Le projet met l'accent sur le traitement des données côté serveur, la gestion des requêtes HTTP et la manipulation de formats JSON pour offrir une interface utilisateur dynamique et fonctionnelle.

## Fonctionnalités

### 1. Page d'accueil
* Présentation globale de l'application.
* Navigation claire vers la liste des artistes.

### 2. Liste des artistes
* Affichage de tous les artistes sous forme de blocs ou de cartes.
* Affichage des informations essentielles : image, nom, année de création et nombre de membres.
* Lien direct vers la page détaillée de chaque artiste.

### 3. Page de détails d'un artiste
* Affichage complet : image, nom, année de création, premier album et liste des membres.
* Liste des concerts regroupant les dates et les lieux correspondants.
* Navigation fluide permettant le retour à la liste ou l'accès aux autres pages.

### 4. Recherche (Côté Serveur)
* Barre de recherche fonctionnelle basée sur une requête HTTP (GET).
* Recherche possible par nom d'artiste, membre, lieu ou date.
* Traitement effectué à 100 % en Go, sans dépendance au JavaScript pour la logique de recherche.

### 5. Filtres
* Filtre par intervalle : sélection d'une année de création minimale et maximale.
* Filtre par sélection multiple : tri par nombre de membres ou lieux de concert.
* Combinaison des filtres : possibilité d'appliquer plusieurs critères simultanément.

### 6. Événement interactif
* Système où une action utilisateur déclenche une nouvelle requête vers le serveur.
* Exemple : le clic sur un lieu redirige vers une page listant tous les concerts prévus à cet endroit.
* Gestion intégrale par le serveur, sans logique JavaScript complexe.

### 7. Gestion des erreurs
* Pages d'erreur personnalisées (404 page non trouvée, erreurs de paramètres, etc.).
* Code sécurisé pour éviter tout crash du serveur.
* Gestion propre des erreurs et des exceptions dans le code Go.

## Fonctionnalités bonus

L'application intègre également les extensions suivantes :

* **Favoris :** Marquage d'artistes favoris persistant via l'utilisation de cookies.
* **Comparaison :** Page dédiée permettant de comparer les informations et les lieux de concert de deux artistes.
* **Suggestions de recherche :** Ajout d'une couche JavaScript pour proposer des suggestions automatiques lors de la saisie dans la barre de recherche.