// http_gateway/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

// Structure pour la requête entrante de notre API
type AskRequest struct {
	Text           string `json:"text"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// Structure pour la réponse de notre API
type AskResponse struct {
	Response       string `json:"response"`
	ConversationID string `json:"conversation_id"`
}

// Structure pour représenter un message dans l'historique
type Message struct {
	Role    string // "user" ou "model"
	Content string
}

// Structure pour représenter une conversation
type Conversation struct {
	ID       string
	Messages []Message
	Created  time.Time
	Updated  time.Time
}

// Stockage en mémoire des conversations
var conversations = sync.Map{}

// Prompt système pour définir le comportement de l'IA
const systemPromptText = `Bonjour Chat, Je suis le Professeur SymptoAI Agrégé et Urgentiste, Professeur à la Faculté de Médecine Générale Avancée, et j'ai traité un cas pratique avec mes étudiants, déjà corrigé, je veux que tu lises exactement ce que dit le patient, ça constitue l'énoncé du sujet, puis tu poses 3 questions ciblées, ensuite il répond et après ça, tu ne fais Aucun Commentaire, Tu fournis Un corrigé Clair sous forme de 3 blocs textuels:

Bloc 1 : Diagnostic le Plus Probable et à Combien de % Exactement
Bloc 2 : Prescription Médicale adaptée précise avec le moins de médicament et le plus d'efficacité + Posologie exacte
Bloc 3 : Niveau d'urgence Avec code couleurs VERT JAUNE ORANGE ROUGE

Instructions strictes :
- Première interaction : Lis le cas du patient puis pose EXACTEMENT 3 questions ciblées
- Après les réponses du patient : Fournis UNIQUEMENT les 3 blocs sans aucun commentaire supplémentaire
- Sois précis dans les pourcentages de diagnostic
- Donne des prescriptions médicales concrètes avec posologies exactes
- Assigne un code couleur d'urgence approprié

Reste dans ce rôle de Professeur SymptoAI tout au long de la conversation.`

// Fonction pour récupérer ou créer une conversation
func getOrCreateConversation(conversationID string) *Conversation {
	if conversationID != "" {
		if conv, exists := conversations.Load(conversationID); exists {
			conversation := conv.(*Conversation)
			conversation.Updated = time.Now()
			return conversation
		}
	}

	// Créer une nouvelle conversation
	newID := uuid.New().String()
	conversation := &Conversation{
		ID:       newID,
		Messages: []Message{},
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	conversations.Store(newID, conversation)
	return conversation
}

// Fonction pour construire l'historique complet pour l'API Gemini
func buildConversationHistory(conversation *Conversation, newUserMessage string) []*genai.Content {
	var contents []*genai.Content

	// Construire le contenu complet pour la génération
	var fullPrompt string

	// Ajouter le prompt système en premier (seulement si c'est le début de la conversation)
	if len(conversation.Messages) == 0 {
		fullPrompt = systemPromptText + "\n\nCas patient: " + newUserMessage
	} else {
		// Reconstruire l'historique complet
		fullPrompt = systemPromptText + "\n\n"

		for i, msg := range conversation.Messages {
			if msg.Role == "user" {
				if i == 0 {
					fullPrompt += "Cas patient: " + msg.Content + "\n\n"
				} else {
					fullPrompt += "Réponses patient: " + msg.Content + "\n\n"
				}
			} else if msg.Role == "model" {
				fullPrompt += "Professeur SymptoAI: " + msg.Content + "\n\n"
			}
		}

		// Ajouter le nouveau message
		fullPrompt += "Réponses patient: " + newUserMessage
	}

	// Créer le contenu pour l'API
	content := &genai.Content{
		Parts: []genai.Part{
			genai.Text(fullPrompt),
		},
	}

	contents = append(contents, content)
	return contents
}

// Fonction pour sauvegarder les messages dans la conversation
func saveMessagesToConversation(conversation *Conversation, userMessage, modelResponse string) {
	conversation.Messages = append(conversation.Messages, Message{
		Role:    "user",
		Content: userMessage,
	})
	conversation.Messages = append(conversation.Messages, Message{
		Role:    "model",
		Content: modelResponse,
	})
	conversation.Updated = time.Now()
	conversations.Store(conversation.ID, conversation)
}

func askHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Lire et parser la requête de l'utilisateur
	var req AskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête JSON invalide", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "Le champ 'text' est obligatoire", http.StatusBadRequest)
		return
	}

	// 2. Récupérer ou créer la conversation
	conversation := getOrCreateConversation(req.ConversationID)
	log.Printf("Traitement de la requête pour la conversation %s", conversation.ID)

	// 3. Construire l'historique complet pour l'API
	conversationContents := buildConversationHistory(conversation, req.Text)

	// 4. Appeler l'API Gemini
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Printf("Erreur lors de la création du client GenAI: %v", err)
		http.Error(w, "Erreur de communication avec le service d'IA", http.StatusInternalServerError)
		return
	}

	// Utiliser GenerateContent avec l'historique complet
	result, err := client.Models.GenerateContent(ctx, "gemini-1.5-flash", conversationContents, nil)
	if err != nil {
		log.Printf("Erreur lors de la génération de contenu: %v", err)
		http.Error(w, "Erreur lors de la génération de la réponse", http.StatusInternalServerError)
		return
	}

	if result == nil || result.Text() == "" {
		log.Printf("Réponse vide reçue de l'API Gemini")
		http.Error(w, "Aucune réponse générée par l'IA", http.StatusInternalServerError)
		return
	}

	responseText := result.Text()

	// 5. Sauvegarder les messages dans la conversation
	saveMessagesToConversation(conversation, req.Text, responseText)

	// 6. Renvoyer la réponse finale avec l'ID de conversation
	response := AskResponse{
		Response:       responseText,
		ConversationID: conversation.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Erreur lors de l'encodage JSON: %v", err)
		http.Error(w, "Erreur lors de la préparation de la réponse", http.StatusInternalServerError)
		return
	}

	log.Printf("Réponse envoyée pour la conversation %s", conversation.ID)
}

// Endpoint pour récupérer l'historique d'une conversation (optionnel, utile pour le debug)
func conversationHandler(w http.ResponseWriter, r *http.Request) {
	conversationID := r.URL.Query().Get("id")
	if conversationID == "" {
		http.Error(w, "Paramètre 'id' manquant", http.StatusBadRequest)
		return
	}

	if conv, exists := conversations.Load(conversationID); exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conv)
	} else {
		http.Error(w, "Conversation non trouvée", http.StatusNotFound)
	}
}

// Endpoint pour lister toutes les conversations (optionnel, utile pour le debug)
func conversationsHandler(w http.ResponseWriter, r *http.Request) {
	var allConversations []*Conversation
	conversations.Range(func(key, value interface{}) bool {
		conv := value.(*Conversation)
		allConversations = append(allConversations, conv)
		return true
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allConversations)
}

func main() {
	// Charger les variables d'environnement
	if err := godotenv.Load(); err != nil {
		log.Println("Attention: Fichier .env non trouvé.")
	}

	// Routes principales
	http.HandleFunc("/api/ask", askHandler)

	// Routes utiles pour le développement et le debug
	http.HandleFunc("/api/conversation", conversationHandler)
	http.HandleFunc("/api/conversations", conversationsHandler)

	port := "8081"
	log.Printf("Passerelle HTTP conversationnelle démarrée sur http://localhost:%s", port)
	log.Println("Endpoints disponibles:")
	log.Println("  POST /api/ask - Envoyer un message")
	log.Println("  GET /api/conversation?id=<conversation_id> - Récupérer une conversation")
	log.Println("  GET /api/conversations - Lister toutes les conversations")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur HTTP: %v", err)
	}
}
