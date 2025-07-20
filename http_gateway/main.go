// http_gateway/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"strings"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Importer le package proto généré
	pb "gemini_api_test/http_gateway/proto"
)

// Structure pour la requête entrante de notre API
type AskRequest struct {
	Text string `json:"text"`
}

// Structure pour la réponse de notre API
type AskResponse struct {
	Response string `json:"response"`
}

// Fonction pour appeler le service NLP via gRPC
func callNlpService(text string) (*pb.SymptomAnalysisResponse, error) {
	// Adresse du serveur gRPC Python
	addr := "localhost:50051"
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("impossible de se connecter au service NLP: %w", err)
	}
	defer conn.Close()

	c := pb.NewSymptomAnalysisServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &pb.SymptomAnalysisRequest{Text: text}
	log.Printf("Envoi de la requête gRPC au service NLP: %v", req)

	return c.Analyze(ctx, req)
}

func askHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Lire et parser la requête de l'utilisateur
	var req AskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête JSON invalide", http.StatusBadRequest)
		return
	}

	// 2. Appeler le service NLP pour l'analyse de symptômes
	nlpResponse, err := callNlpService(req.Text)
	if err != nil {
		log.Printf("Erreur lors de l'appel au service NLP: %v", err)
		http.Error(w, "Erreur interne lors de l'analyse des symptômes", http.StatusInternalServerError)
		return
	}
	log.Printf("Réponse reçue du service NLP: %v", nlpResponse)

	// 3. Construire un prompt intelligent pour Gemini
	var symptoms []string
	for _, s := range nlpResponse.Symptoms {
		symptoms = append(symptoms, fmt.Sprintf("- %s (intensité: %s, durée: %s)", s.Name, s.Intensity, s.Duration))
	}
	promptText := fmt.Sprintf("Un patient décrit les symptômes suivants:\n%s\n\nRédige des conseils de premiers soins clairs et sécuritaires. Mentionne explicitement quand il est impératif de consulter un médecin.", strings.Join(symptoms, "\n"))

	// 4. Appeler l'API Gemini avec le prompt enrichi
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Printf("Erreur lors de la création du client GenAI: %v", err)
		http.Error(w, "Erreur de communication avec le service d'IA", http.StatusInternalServerError)
		return
	}

	result, err := client.Models.GenerateContent(ctx, "gemini-1.5-flash", genai.Text(promptText), nil)
	if err != nil {
		log.Printf("Erreur lors de la génération de contenu: %v", err)
		http.Error(w, "Erreur lors de la génération de la réponse", http.StatusInternalServerError)
		return
	}

	// 5. Renvoyer la réponse finale
	response := AskResponse{Response: result.Text()}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Charger les variables d'environnement
	if err := godotenv.Load(); err != nil {
		log.Println("Attention: Fichier .env non trouvé.")
	}

	http.HandleFunc("/api/ask", askHandler)
	port := "8081"
	log.Printf("Passerelle HTTP démarrée sur http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur HTTP: %v", err)
	}
}
