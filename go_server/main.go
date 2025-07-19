// go_server/main.go
package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Importer le package Go généré
	pb "healthsync/go_server/proto"
)

func main() {
	// Adresse du serveur gRPC Python
	addr := "localhost:50051"

	// Établir une connexion gRPC avec le serveur.
	// On utilise WithTransportCredentials(insecure.NewCredentials()) car on n'a pas de SSL/TLS pour l'instant.
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Impossible de se connecter : %v", err)
	}
	// S'assurer que la connexion est fermée à la fin
	defer conn.Close()

	// Créer un client pour notre service
	c := pb.NewSymptomAnalysisServiceClient(conn)

	// Créer un contexte avec un timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Le texte à analyser
	textToAnalyze := "j'ai une forte fièvre et mal à la tête"
	log.Printf("Envoi de la requête avec le texte : %s", textToAnalyze)

	// Appeler la méthode Analyze sur le serveur distant
	r, err := c.Analyze(ctx, &pb.SymptomAnalysisRequest{Text: textToAnalyze})
	if err != nil {
		log.Fatalf("Impossible d'appeler Analyze : %v", err)
	}

	// Afficher la réponse
	log.Printf("Réponse reçue:")
	for _, symptom := range r.GetSymptoms() {
		log.Printf("- Nom: %s, Durée: %s, Intensité: %s", symptom.GetName(), symptom.GetDuration(), symptom.GetIntensity())
	}
}
