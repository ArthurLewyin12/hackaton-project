# Projet HealthSync


Le parcours utilisateur final sera le suivant :
1.  L'utilisateur décrit ses symptômes dans l'application, soit par texte, soit par la voix.
2.  L'application analyse ces informations en temps réel.
3.  L'IA (propulsée par Gemini) pose des questions de suivi si nécessaire, puis fournit des conseils de premiers soins pertinents et sécuritaires.
4.  Simultanément, l'application géolocalise l'utilisateur et lui propose une liste de cliniques, pharmacies ou services d'urgence à proximité via une carte interactive.

## 2. Architecture Technique Cible

Pour réaliser cette vision, le projet s'articule autour d'une architecture de microservices robuste et évolutive.

![Architecture Diagram](https://i.imgur.com/your-diagram-image.png) <!-- Vous pouvez générer un diagramme et mettre le lien ici -->

### a. Service NLP (Python)
- **Rôle** : Traitement du Langage Naturel.
- **Technologie** : Serveur gRPC en Python.
- **Fonctionnalités** :
  - **Transcription Voix-Texte** : Recevoir un flux audio de l'utilisateur et le transcrire en texte brut.
  - **Analyse de Symptômes** : Analyser le texte pour en extraire des entités structurées (symptômes, durée, intensité, etc.).
- **Pourquoi ce choix ?** : Pour capitaliser sur l'écosystème Python, leader en matière d'IA, avec des bibliothèques comme Hugging Face Transformers (pour le NLP) et Whisper (pour la transcription).

### b. Serveur MCP (Go)
- **Rôle** : Pont entre notre logique métier et le LLM (Gemini).
- **Technologie** : Serveur Go implémentant le **Model Context Protocol (MCP)**.
- **Fonctionnalités** :
  - Exposer les capacités du service NLP comme un "outil" (`analyzeSymptoms`) que Gemini peut appeler.
  - Gérer la communication gRPC avec le service NLP.
  - Servir de source de contexte fiable pour le LLM.
- **Pourquoi ce choix ?** : Go offre des performances élevées pour les services réseau. L'utilisation du MCP standardise la communication avec le LLM, rendant le système agnostique et prêt pour le futur.

### c. Passerelle HTTP & Orchestration (Go)
- **Rôle** : Point d'entrée principal de l'application et cerveau de l'orchestration.
- **Technologie** : Serveur HTTP en Go.
- **Fonctionnalités** :
  - **API Utilisateur** : Exposer une API REST (ex: `/api/ask`) que les clients (web/mobile) appelleront.
  - **Orchestration des appels** : Gérer le dialogue avec Gemini. C'est lui qui reçoit la requête de l'utilisateur, appelle Gemini, gère les allers-retours pour les appels d'outils au serveur MCP, et formate la réponse finale.
  - **Intégration Google Maps** : Contacter l'API Google Maps pour récupérer les informations sur les cliniques ( à venir).


### d. Client (Web/Mobile)
- **Rôle** : Interface utilisateur.
- **Technologie** :  Next js pour le web,
Flutter pour le mobile ( si necessaire).
- **Fonctionnalités** :
  - Enregistrer la voix ou saisir le texte.
  - Envoyer la requête à la passerelle HTTP.
  - Afficher les conseils de l'IA et la carte avec les cliniques.

## 3. État Actuel

L'infrastructure de base pour les services **NLP (Python)**, **MCP (Go)** et la **Passerelle HTTP (Go)** est en place. La communication entre la passerelle et le serveur MCP via HTTP, et entre le MCP et le service NLP via gRPC, est définie. Le projet est actuellement en phase de débogage de la communication inter-services.
