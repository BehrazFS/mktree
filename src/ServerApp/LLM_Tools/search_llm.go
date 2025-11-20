package llms

import (
	"encoding/json"
	"fmt"
	"strings"
)

const SystemPrompt = `

You are a codebase analysis assistant.  
Given a user query and a current .tree file representing a codebase structure, generate exactly 3 meaningful search queries that connect the user's request with the codebase structure:

- "task_query" – The main coding task or feature the user wants to implement or modify.  
- "technology_query" – The specific technologies, frameworks, or programming languages relevant to the task.  
- "codebase_query" – The file or directory structure patterns, filenames, or components in the .tree that need to be modified or used.  

Notes:  
1. The optional Q&A input may contain clarifications where the LLM asked questions about missing data, and the user provided answers. Incorporate this information if present.  
2. If you cannot extract all 3 queries directly from the user input (including Q&A), do not invent details. Instead, create search queries that aim to find best practices for the requested task (e.g., if the user asks for a "fast API," search for best practices about FastAPI usage).  

The .tree file contains the hierarchical structure of the current project, using these conventions:  
- Indentation with two spaces = subfolder depth.  
- A "." in the name indicates a file.  
- ":|" indicates the content block of that file.  

Preserve and include essential imports and boilerplate needed for the technology (e.g., Python imports, VHDL entity declarations).  
Use the .tree context to make the search queries specific to the actual codebase layout.  

Output the three queries strictly in JSON format with the keys exactly as above.  
Do NOT output markdown formatting, explanations, or commentary.  

`

// SearchResult is the scraped/search result used by SearchLLM
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

// SearchLLM wraps a base LLM and provides search orchestration
type SearchLLM struct {
	Base *LLM
}

// NewSearchLLM creates a SearchLLM using the base LLM constructor
func NewSearchLLM(model string) *SearchLLM {
	return &SearchLLM{Base: NewLLM(model, SystemPrompt, true)}
}

// SearchTool orchestrates web search, LLM selection of results, and scraping
func (s *SearchLLM) SearchTool(query string) ([]SearchResult, error) {
	results, err := RunSearch(query)
	if err != nil {
		return nil, fmt.Errorf("web search failed: %w", err)
	}

	return results, nil
}

func (s *SearchLLM) ProcessUserInput(input string, curr_tree string) (map[string]string, error) {

	// Compose prompt for LLM
	prompt := fmt.Sprintf("User query: %s\nCurrent project tree: %s", input, curr_tree)
	resp, err := s.Base.Call(prompt) // TODO: hook this to your actual LLM API
	if err != nil {
		return nil, fmt.Errorf("failed to get queries from LLM: %w", err)
	}

	// Parse JSON output from LLM
	queries := map[string]string{}
	err = json.Unmarshal([]byte(resp), &queries)
	if err != nil {
		return nil, fmt.Errorf("invalid LLM output JSON: %w\nRaw: %s", err, resp)
	}

	return queries, nil
}
func (s *SearchLLM) Call(userInput string, curr_tree string) (string, error) {
	// Step 1: Get queries from LLM
	queries, err := s.ProcessUserInput(userInput, curr_tree)
	// fmt.Println(queries)
	if err != nil {
		return "", err
	}

	// Step 2: Run search for each query
	var sb strings.Builder
	for key, q := range queries {
		results, err := s.SearchTool(q)
		if err != nil {
			return "", fmt.Errorf("search failed for %s: %w", key, err)
		}
		sb.WriteString(fmt.Sprintf("=== %s ===\nQuery: %s\n", key, q))
		for i, r := range results {
			sb.WriteString(fmt.Sprintf("Result %d:\nTitle: %s\nURL: %s\nType: %s\n Context: %s\n\n", i+1, r.Title, r.URL, key, r.Content))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
