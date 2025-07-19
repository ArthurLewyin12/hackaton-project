// go_server/main.go
package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "healthsync/go_server/proto"
)

// gRPCClient est une structure pour contenir notre client gRPC.
// Cela nous permet de le rendre disponible pour le handler de l'outil.
type gRPCClient struct {
	client pb.SymptomAnalysisServiceClient
}

// analyzeSymptomsHandler est la fonction qui sera exécutée lorsque le LLM appelle notre outil.
func (c *gRPCClient) analyzeSymptomsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 1. Récupérer le paramètre "text" de la requête de l'outil
	text, err := request.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	log.Printf("[MCP Tool] Appel de l'outil avec le texte : %s", text)

	// 2. Appeler le serveur Python via gRPC
	r, err := c.client.Analyze(ctx, &pb.SymptomAnalysisRequest{Text: text})
	if err != nil {
		log.Printf("[gRPC Client] Erreur lors de l'appel à Analyze : %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	symptoms := r.GetSymptoms()
	log.Printf("[gRPC Client] Réponse reçue du service NLP : %v", symptoms)

	// 3. Sérialiser la réponse en JSON pour la renvoyer au LLM
	jsonData, err := json.Marshal(symptoms)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(jsonData)),
		},
	}, nil
}

func main() {
	// --- Configuration du client gRPC vers le service Python ---
	grpcAddr := "localhost:50051"
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Impossible de se connecter au serveur gRPC : %v", err)
	}
	defer conn.Close()

	// Créer une instance de notre client gRPC
	grpcClient := &gRPCClient{
		client: pb.NewSymptomAnalysisServiceClient(conn),
	}

	// --- Configuration du serveur MCP ---
	// 1. Créer le serveur MCP
	s := server.NewMCPServer(
		"HealthSync MCP Server",
		"1.0.0",
	)

	// 2. Définir le schéma de notre outil
	analyzeTool := mcp.NewTool("analyzeSymptoms",
		mcp.WithDescription("Extrait des informations structurées sur les symptômes à partir d'un texte brut."),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("Le texte brut décrivant les symptômes du patient."),
		),
	)

	// 3. Ajouter l'outil et son handler au serveur
	s.AddTool(analyzeTool, grpcClient.analyzeSymptomsHandler)

	// 4. Démarrer le serveur MCP sur stdio
	log.Println("Serveur MCP démarré sur stdio")

	// démarrer le serveur MCP sur stdio
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur MCP: %v", err)
	}
}
