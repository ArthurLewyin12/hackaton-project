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

// analyzeSymptomsHandler est la fonction qui sera exécutée lorsque le LLM appelle notre outil.
func analyzeSymptomsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// --- Connexion gRPC ---
	grpcAddr := "localhost:50051"
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[gRPC Client] Erreur de connexion : %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer conn.Close()
	c := pb.NewSymptomAnalysisServiceClient(conn)

	// 1. Récupérer le paramètre "text"
	text, err := request.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	log.Printf("[MCP Tool] Appel de l'outil avec le texte : %s", text)

	// 2. Appeler le serveur Python
	r, err := c.Analyze(ctx, &pb.SymptomAnalysisRequest{Text: text})
	if err != nil {
		log.Printf("[gRPC Client] Erreur lors de l'appel à Analyze : %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	symptoms := r.GetSymptoms()
	log.Printf("[gRPC Client] Réponse reçue du service NLP : %v", symptoms)

	// 3. Sérialiser la réponse en JSON
	jsonData, err := json.Marshal(symptoms)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 4. Construire la réponse de l'outil
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(jsonData)),
		},
	}, nil
}

func main() {
	// 1. Créer le serveur MCP logique
	mcpServer := server.NewMCPServer("HealthSync MCP Server", "1.0.0")
	analyzeTool := mcp.NewTool("analyzeSymptoms",
		mcp.WithDescription("Extrait des informations structurées sur les symptômes à partir d'un texte brut."),
		mcp.WithString("text", mcp.Required(), mcp.Description("Le texte brut décrivant les symptômes du patient.")),
	)
	mcpServer.AddTool(analyzeTool, analyzeSymptomsHandler)

	// 2. Démarrer le serveur en utilisant la méthode StreamableHTTP de la documentation
	port := ":8090"
	log.Printf("Serveur MCP démarré sur http://localhost%s", port)
	httpServer := server.NewStreamableHTTPServer(mcpServer)
    if err := httpServer.Start(port); err != nil {
        log.Fatal(err)
    }
}
