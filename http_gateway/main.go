// http_gateway/main.go
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// Structure pour la requête entrante sur notre serveur HTTP
type AskRequest struct {
	Text string `json:"text"`
}

// Structures pour construire la requête JSON-RPC
type MCPParams struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments"`
}

type MCPRequest struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  MCPParams   `json:"params"`
}

func askHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Lire le corps de la requête HTTP
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erreur de lecture de la requête", http.StatusBadRequest)
		return
	}

	// 2. Désérialiser le JSON de la requête
	var askReq AskRequest
	if err := json.Unmarshal(body, &askReq); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	// 3. Construire la requête JSON-RPC complète, comme dans la doc
	mcpReq := MCPRequest{
		JsonRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: MCPParams{
			Name:      "analyzeSymptoms",
			Arguments: map[string]string{"text": askReq.Text},
		},
	}
	mcpBodyBytes, err := json.Marshal(mcpReq)
	if err != nil {
		http.Error(w, "Erreur de sérialisation MCP", http.StatusInternalServerError)
		return
	}

	log.Printf("Envoi de la requête à http://localhost:8090/mcp/tools/call: %s", string(mcpBodyBytes))

	// 4. Appeler le bon point de terminaison du serveur MCP
	resp, err := http.Post("http://localhost:8090/mcp/tools/call", "application/json", bytes.NewBuffer(mcpBodyBytes))
	if err != nil {
		http.Error(w, "Erreur lors de l'appel au serveur MCP: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 5. Lire la réponse du serveur MCP
	mcpResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Erreur de lecture de la réponse MCP", http.StatusInternalServerError)
		return
	}

	log.Printf("Réponse reçue du MCP : %s", string(mcpResponse))

	// 6. Renvoyer la réponse du MCP au client final
	w.Header().Set("Content-Type", "application/json")
	w.Write(mcpResponse)
}

func main() {
	http.HandleFunc("/api/ask", askHandler)
	log.Println("Passerelle HTTP démarrée sur http://localhost:8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur HTTP: %v", err)
	}
}