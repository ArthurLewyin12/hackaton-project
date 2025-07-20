# Projet HealthSync

Le parcours utilisateur final sera le suivant :
1.  L'utilisateur décrit ses symptômes dans l'application, soit par texte, soit par la voix.
2.  L'application analyse ces informations en temps réel.
3.  L'IA (propulsée par Gemini) pose des questions de suivi si nécessaire, puis fournit des conseils de premiers soins pertinents et sécuritaires.
4.  Simultanément, l'application géolocalise l'utilisateur et lui propose une liste de cliniques, pharmacies ou services d'urgence à proximité via une carte interactive.

## 2. Architecture Technique

Pour réaliser cette vision, le projet s'articule autour d'une architecture de microservices robuste et évolutive.

![Architecture Diagram](https://i.imgur.com/your-diagram-image.png) <!-- Vous pouvez générer un diagramme et mettre le lien ici -->

### a. Service NLP (Python)
- **Rôle** : Traitement du Langage Naturel.
- **Technologie** : Serveur gRPC en Python.
- **Fonctionnalités** :
  - **Transcription Voix-Texte (Cible)** : Recevoir un flux audio de l'utilisateur et le transcrire en texte brut.
  - **Analyse de Symptômes** : Analyser le texte pour en extraire des entités structurées (symptômes, durée, intensité, etc.).
- **Pourquoi ce choix ?** : Pour capitaliser sur l'écosystème Python, leader en matière d'IA, avec des bibliothèques comme Hugging Face Transformers (pour le NLP) et Whisper (pour la transcription).

### b. Passerelle HTTP & Orchestration (Go)
- **Rôle** : Point d'entrée principal de l'application et cerveau de l'orchestration.
- **Technologie** : Serveur HTTP en Go.
- **Fonctionnalités** :
  - **API Utilisateur** : Exposer une API REST (ex: `/api/ask`) que les clients (web/mobile) appelleront.
  - **Orchestration des appels** : Gérer le dialogue de bout en bout.
    1. Reçoit la requête de l'utilisateur.
    2. Appelle le **Service NLP** via gRPC pour l'analyse de texte.
    3. Construit un prompt enrichi à partir des résultats du NLP.
    4. Appelle directement l'**API Google Gemini** pour obtenir des conseils.
    5. Formate et renvoie la réponse finale.
  - **Intégration Google Maps (à venir)** : Contacter l'API Google Maps pour récupérer les informations sur les cliniques.

### c. Client (Web/Mobile)
- **Rôle** : Interface utilisateur.
- **Technologie** : Next.js pour le web, Flutter pour le mobile (si nécessaire).
- **Fonctionnalités** :
  - Enregistrer la voix ou saisir le texte.
  - Envoyer la requête à la passerelle HTTP.
  - Afficher les conseils de l'IA et la carte avec les cliniques.

## 3. État Actuel

L'architecture de base est fonctionnelle.
- La **Passerelle HTTP (Go)** est le point d'entrée et orchestre les appels.
- Le **Service NLP (Python)** expose un service gRPC pour l'analyse de texte (actuellement simulée).
- La communication entre la passerelle Go et le service Python via gRPC est **opérationnelle**.
- La passerelle communique avec succès avec l'**API Gemini**.

Le projet est prêt pour l'implémentation de la logique NLP réelle et l'intégration de la géolocalisation.