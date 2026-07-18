# Fiche descriptive de projet Hub
---
L'app de buzzer en réseaux

## Contexte et but du projet
Détailler l’origine du projet, les éléments motivants sa réalisation et un 
descriptif de l’usage final de celui-ci

Afin d’obtenir mon année je dois faire un free project. Après qu’un camarade de 
promotion fait des jeux inspiré de jeux televise en début ou en fin d’année il 
manquait des buzzers functionnel a large echelle. Le projet consiste à 
développer une application web permettant de créer des lobbys dans lesquels 
plusieurs joueurs peuvent participer à un système de buzzer, inspiré des jeux 
télévisés comme Questions pour un champion. Chaque joueur rejoint un lobby et 
dispose d'un bouton de buzzer. Lorsqu'une manche est lancée, le serveur 
détermine quel participant a appuyé en premier et annonce immédiatement le 
gagnant à tous les joueurs. L'objectif est de proposer une solution simple, 
rapide et équitable, tout en mettant en œuvre des technologies de communication 
en temps réel entre les clients et le serveur.

## Porteur(s) du projet
Détailler ici les membres du groupe projet, leur rôle si défini ainsi que les 
éventuels partenaires externes [entreprise, communauté open-source, …] et leur 
rôle

Clément DEVAUX - Tek2 BDFL du projet

## Objectif fonctionnel
Lister les fonctionnalités majeures de chacune des parties du projet sous la 
forme de user story https://fr.wikipedia.org/wiki/R%C3%A9cit_utilisateur

- L'utilisateur hôte tombe sur la homepage du site
- L'utilisateur hôte creer un lobby
- L'hôte partage un lien à ses amis
- Ses amis tombe sur une app avec un buzzer
  - Buzzer vert : clickable
  - Buzzer gris : non clickable
  - Buzzer jaune : tu as gagner
  - Buzzer rouge : tu as perdu


## Environnement technique / technologique
Exposer le contexte technique et technologique (matériel, langage, 
environnement d’exécution, ressources, …) dans lequel le projet s’inscrit.

Back:
  - Environnement : Docker
  - Langage : 
    - Go avec Chi
    - Utilisation de websocket
  - Loadbalancer (nginx ou traefik)
  - Valkey pour mise en cache

Front:
  - Vite
  - React.js
  - Typescript

## Description du livrable
Détailler chaque élément (programmes, librairies, modules, assets, …) du 
livrable et leur niveau de finition (déploiement, documentation, …)

2 OCI images: back et front
Page d’accueil
  - avec tous les lobby (public)
  - possibilite de crée un lobby
Dans un lobby:
  - Simple buzzer (couleur qui change selon, si il est pressable, gagner ou 
perdu)

## Organisation et temporalité
Exposer le plan de réalisation du projet : parties, dépendances, planification

Backend (20h):
  - Initialisation backend (2h)
  - Système de lobby (6h)
  - API REST (4h)
  - WebSocket (5h)
  - Préparation Valkey (3h)

Frontend (20h):
  - Initialisation React (2h)
  - Accueil et création lobby (4h)
  - Interface lobby (5h)
  - Buzzer temps réel (6h)
  - Tests et UX (3h)
