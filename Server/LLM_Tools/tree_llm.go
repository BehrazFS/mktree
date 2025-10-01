package llms

import (
	"fmt"
	"strings"
)

const TreeSystemPrompt = `
You are a project tree generator. 
Your task is to analyze:
1. The user's request.
2. Relevant search results about the task, technology, and example codebases.
3. The current .tree file (if provided).

Rules:
- If the current tree already exists, modify or extend it.
- If no tree is given, create a new one.
- Only add imports for new files, do not change imports and code in existing files.
- Always follow proper programming paradigms.
- Tag existing files with <exist> and files to remove with <DELETE>.
- Always output a single .tree file in the following format:

<project_name>
  <folder>
    <file.ext> <exist>: <short inline text if small> 
    <file.ext>: <DELETE>| 
      <multiline file contents if code>
    <file.ext>:|
      <multiline file contents if code for new files>

Conventions:
- Indentation with two spaces = subfolder depth.
- A "." in the name means a file.
- ":|" means the content block of that file.
- Preserve and include essential imports and boilerplate needed for the technology (e.g., Python imports, VHDL entity declarations) only for new files.
- Do NOT output explanations or commentary. Only the .tree file.
`

// TreeLLM represents the LLM that generates or edits project trees
type TreeLLM struct {
	Base *LLM
}

// NewTreeLLM creates a TreeLLM using the base LLM constructor
func NewTreeLLM(model string) *TreeLLM {
	return &TreeLLM{Base: NewLLM(model, TreeSystemPrompt, false)}
}

// GenerateTree analyzes the user input, search results, and current tree,
// then produces a new or updated .tree file.
func (t *TreeLLM) GenerateTree(userInput, searchResults, currentTree string) (string, error) {
	prompt := fmt.Sprintf(`
%s

User Input:
%s

Search Results:
%s

Current Tree:
%s
`, TreeSystemPrompt, userInput, searchResults, currentTree)

	resp, err := t.Base.Call(strings.TrimSpace(prompt))
	if err != nil {
		return "", fmt.Errorf("TreeLLM failed: %w", err)
	}

	// Ensure we only return trimmed output
	return strings.TrimSpace(resp), nil
}
