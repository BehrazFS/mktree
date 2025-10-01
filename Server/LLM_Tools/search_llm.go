package llms

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const SystemPrompt = `

You are a codebase analysis assistant.
Given a user query and a current .tree file representing a codebase structure, generate exactly 3 meaningful search queries that connect the user's request with the codebase structure:

"task_query" – The main coding task or feature the user wants to implement or modify.

"technology_query" – The specific technologies, frameworks, or programming languages relevant to the task.

"codebase_query" – The file or directory structure patterns, filenames, or components in the .tree that need to be modified or used.

The .tree file contains the hierarchical structure of the current project, using these conventions:

Indentation with two spaces = subfolder depth.

A "." in the name indicates a file.

":|" indicates the content block of that file.

Preserve and include essential imports and boilerplate needed for the technology (e.g., Python imports, VHDL entity declarations).

Use the .tree context to make the search queries specific to the actual codebase layout.

Output the three queries strictly in JSON format with the keys exactly as above.
Do NOT output markdown formatting, explanations, or commentary.
`

// SearchResult is the scraped/search result used by SearchLLM
type SearchResult struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

// SearchLLM wraps a base LLM and provides search orchestration
type SearchLLM struct {
	Base *LLM
}

// NewSearchLLM creates a SearchLLM using the base LLM constructor
func NewSearchLLM(model string) *SearchLLM {
	return &SearchLLM{Base: NewLLM(model, SystemPrompt, true)}
}

// webToSearchResults converts WebSearchResponse into SearchResult slice
func webToSearchResults(web *WebSearchResponse) []SearchResult {
	if web == nil {
		return nil
	}
	out := make([]SearchResult, 0, len(web.Results))
	for _, r := range web.Results {
		out = append(out, SearchResult{
			URL:   r.URL,
			Title: r.Title,
			Text:  r.Snippet,
		})
	}
	return out
}

// selectResultsFromLLM asks the LLM which indices to scrape. It expects a JSON array reply like [0,2]
func (s *SearchLLM) selectResultsFromLLM(candidates []SearchResult, userQuery string) ([]int, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	var b strings.Builder
	b.WriteString("You are given a user query and a numbered list of search results.\n")
	b.WriteString("Return a JSON array of integers (0-based) indicating which results should be scraped for best evidence.\n")
	b.WriteString("Only return the JSON array, e.g. [0,2].\n\n")
	b.WriteString(fmt.Sprintf("User query: %s\n\nResults:\n", userQuery))
	for i, r := range candidates {
		b.WriteString(fmt.Sprintf("%d) %s - %s\n", i, r.Title, r.URL))
		if r.Text != "" {
			// include snippet so LLM can decide
			snippet := r.Text
			if len(snippet) > 300 {
				snippet = snippet[:300]
			}
			b.WriteString(fmt.Sprintf("   snippet: %s\n", snippet))
		}
	}

	resp, err := s.Base.Call(b.String())
	if err != nil {
		return nil, fmt.Errorf("llm selection call failed: %w", err)
	}

	// Try to parse JSON array
	var indices []int
	if err := json.Unmarshal([]byte(resp), &indices); err == nil {
		return indices, nil
	}

	// Fallback: extract numbers from the response text
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(resp, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	seen := map[int]struct{}{}
	for _, m := range matches {
		n, convErr := strconv.Atoi(m)
		if convErr != nil {
			continue
		}
		if n >= 0 && n < len(candidates) {
			seen[n] = struct{}{}
		}
	}
	out := make([]int, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out, nil
}

// SearchTool orchestrates web search, LLM selection of results, and scraping
func (s *SearchLLM) SearchTool(query string) ([]SearchResult, error) {
	webResp, err := WebSearch(query)
	if err != nil {
		return nil, fmt.Errorf("web search failed: %w", err)
	}

	candidates := webToSearchResults(webResp)

	selected, err := s.selectResultsFromLLM(candidates, query)
	if err != nil {
		return nil, fmt.Errorf("selection error: %w", err)
	}

	// Default to top 3 if none selected
	if len(selected) == 0 {
		max := 3
		if len(candidates) < max {
			max = len(candidates)
		}
		selected = make([]int, max)
		for i := 0; i < max; i++ {
			selected[i] = i
		}
	}

	var results []SearchResult
	for _, idx := range selected {
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		c := candidates[idx]
		text, err := GetPageText(c.URL)
		if err != nil {
			// fallback to snippet
			text = c.Text
		}
		results = append(results, SearchResult{
			URL:   c.URL,
			Title: c.Title,
			Text:  text,
		})
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
			sb.WriteString(fmt.Sprintf("Result %d:\nTitle: %s\nURL: %s\nType: %s\n Context: %s\n\n", i+1, r.Title, r.URL, key, r.Text))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
