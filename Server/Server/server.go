package Server

import (
	"encoding/json"
	"fmt"
	"log"
	llms "mktree/LLM_Tools"
	"net/http"
)

// Input represents the expected request structure
type Input struct {
	Tree          string `json:"Tree"`
	Prompt        string `json:"Prompt"`
	OperationType string `json:"OperationType"`
}

// Output represents the response structure
type Output struct {
	Tree      interface{} `json:"Tree"`
	ExtraInfo string      `json:"ExtraInfo"`
}

// Server handles HTTP requests for LLM operations
type Server struct{}

// NewServer creates a new server instance
func NewServer() *Server {
	return &Server{}
}

// ProcessTreeHandler handles the main processing endpoint
func (s *Server) ProcessTreeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Call your AgentProcess function - assuming it returns (string, error)
	result, err := llms.AgentProcess(input.Tree, input.Prompt, input.OperationType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Agent processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Process the string result
	var processedTree interface{}
	var extraInfo string

	// Try to parse the result as JSON, if it fails, use it as a plain string
	if err := json.Unmarshal([]byte(result), &processedTree); err != nil {
		processedTree = result // Use as plain string
		extraInfo = "Processing completed with plain text result"
	} else {
		extraInfo = "Processing completed successfully"
	}

	// Prepare output
	output := Output{
		Tree:      processedTree,
		ExtraInfo: extraInfo,
	}

	if err := json.NewEncoder(w).Encode(output); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// StartServer initializes and starts the HTTP server
func (s *Server) StartServer(port string) {
	http.HandleFunc("/process", s.ProcessTreeHandler)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
