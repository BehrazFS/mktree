package Server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	llms "mktree/LLM_Tools"

	"github.com/google/uuid"
)

// Input represents the expected request structure
type Input struct {
	Tree          string `json:"Tree"`
	Prompt        string `json:"Prompt"`
	OperationType string `json:"OperationType"`
	QA            string `json:"QA,omitempty"`
}

// Output represents the response structure
type Output struct {
	Tree      interface{} `json:"Tree"`
	ExtraInfo string      `json:"ExtraInfo"`
}

// Server handles HTTP requests for LLM operations
type Server struct {
	log *log.Logger
}

// NewServer creates a new server instance
func NewServer() *Server {
	return &Server{
		log: log.Default(), // or log.New(os.Stdout, "", log.LstdFlags)
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

	// Log the incoming request with req_id
	s.log.Printf("[REQ_ID: %s] Incoming request from <%s>", reqID, remoteAddr)

	var input Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("[REQ_ID: %s] Invalid JSON input: %v", reqID, err)
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Optional: Create a context with req_id for deeper logging or tracing

	// Call your AgentProcess function - now accepts req_id and context
	result, changes, questions, err := llms.AgentProcess(s.log, input.Tree, input.Prompt, input.OperationType, input.QA, reqID)
	if err != nil {
		s.log.Printf("[REQ_ID: %s] Agent processing failed: %v", reqID, err)
		http.Error(w, fmt.Sprintf("Agent processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Process the string result
	var processedTree interface{}
	var extraInfo string

	// Try to parse the result as JSON, if it fails, use it as a plain string
	if err := json.Unmarshal([]byte(result), &processedTree); err != nil {
		processedTree = result // Use as plain string
		extraInfo = fmt.Sprintf("{changes: %s, questions: %s}", changes, questions)
	} else {
		extraInfo = fmt.Sprintf("{changes: %s, questions: %s}", changes, questions)

	}

	// Prepare output with req_id
	output := Output{
		Tree:      processedTree,
		ExtraInfo: extraInfo,
	}
	s.log.Printf("[REQ_ID: %s] The result has been sent %v", reqID, err)

	if err := json.NewEncoder(w).Encode(output); err != nil {
		s.log.Printf("[REQ_ID: %s] Error encoding response: %v", reqID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// StartServer initializes and starts the HTTP server
func (s *Server) StartServer(port string) {
	http.HandleFunc("/process", s.ProcessTreeHandler)

	s.log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		s.log.Fatalf("Server failed to start: %v", err)
	}
}
