package llms

import (
	"encoding/json"
	"fmt"
)

type TrimLLM struct {
	Base *LLM
}

type TrimmedResult struct {
	Title   string `json:"title"`
	Type    string `json:"type"`    // replaces URL
	Content string `json:"content"` // cleaned content
}

// System prompt for cleaning

const TrimSystemPrompt = `
You are a cleaning assistant. Your task is to process raw search results into a **concise JSON**.

Rules:
1. Read all raw results and focus only on the main concepts relevant to:
   - The task the user wants to perform ("task_query")
   - The technology involved ("technology_query")
   - The project’s codebase or directory structure ("codebase_query")
2. Preserve **directory trees, code snippets, and technical explanations and md desciption** if they exist.
3. Remove all unrelated text such as ads, navigation, cookie banners, or website boilerplate.
4. Output only a JSON array, where each item is in the following format:

{
  "type": "<task_query | technology_query | codebase_query>",
  "title": "<short title>",
  "content": "<core concept or tree/code if available>"
}

5. Do NOT output explanations, notes, or prefixes. Return only valid JSON. Don't include markdown formatting.
6. Ensure the JSON is syntactically correct.
`

func NewTrimLLM(model string) *TrimLLM {
	return &TrimLLM{Base: NewLLM(model, TrimSystemPrompt, true)}
}

// TrimResults uses the LLM to clean up raw scraped results
func (t *TrimLLM) TrimResults(user_query string, results string) ([]TrimmedResult, error) {

	query := fmt.Sprintf("User query: %s\n\nResults:\n", user_query, results)
	resp, err := t.Base.Call(query)
	if err != nil {
		return nil, fmt.Errorf("failed to trim results: %w", err)
	}

	// Parse JSON output into []TrimmedResult
	var cleaned []TrimmedResult
	err = json.Unmarshal([]byte(resp), &cleaned)
	if err != nil {
		return nil, fmt.Errorf("invalid LLM output JSON: %w\nRaw: %s", err, resp)
	}

	return cleaned, nil
}
