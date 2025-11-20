package Server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	llms "ServerApp/LLM_Tools"

	"github.com/google/uuid"
)

// Input represents the expected request structure
type Input struct {
	Tree          string `json:"Tree"`
	Prompt        string `json:"Prompt"`
	OperationType string `json:"OperationType"`
}

// Output represents the response structure
type Output struct {
	Tree    interface{} `json:"Tree"`
	Changes string      `json:"Changes"`
}

// Server handles HTTP requests for LLM operations
type Server struct {
	log *log.Logger
}

// NewServer creates a new server instance
func NewServer() *Server {
	return &Server{
		log: log.Default(),
	}
}

// ProcessTreeHandler handles the main processing endpoint
func (s *Server) ProcessTreeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate a unique request ID
	reqID := uuid.New().String()
	remoteAddr := r.RemoteAddr

	s.log.Printf("[REQ_ID: %s] Incoming request from <%s>", reqID, remoteAddr)

	var input Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("[REQ_ID: %s] Invalid JSON input: %v", reqID, err)
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Call AgentProcess function
	result, changes, err := llms.AgentProcess(s.log, input.Tree, input.Prompt, input.OperationType, reqID)
	if err != nil {
		s.log.Printf("[REQ_ID: %s] Agent processing failed: %v", reqID, err)
		http.Error(w, fmt.Sprintf("Agent processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Process the string result
	var processedTree interface{}
	if err := json.Unmarshal([]byte(result), &processedTree); err != nil {
		processedTree = result // Use as plain string
	}

	// Prepare output with just Tree and Changes
	output := Output{
		Tree:    processedTree,
		Changes: changes,
	}

	s.log.Printf("[REQ_ID: %s] The result has been sent", reqID)

	if err := json.NewEncoder(w).Encode(output); err != nil {
		s.log.Printf("[REQ_ID: %s] Error encoding response: %v", reqID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// StartServer initializes and starts the HTTP server
func (s *Server) StartServer(port string) {
	http.HandleFunc("/mktree", s.ProcessTreeHandler)

	s.log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		s.log.Fatalf("Server failed to start: %v", err)
	}
}
