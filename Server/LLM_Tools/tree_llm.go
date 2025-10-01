package llms

import (
	"fmt"
	"strings"
)

const TreeSystemPrompt = `
You are a project tree generator. 

.tree Syntax:
- "<file.ext>:" → for single-line inline code or text.
- "<file.ext>:|" → for multi-line code or content blocks.
- "<file.ext>" (no ":" or ":|") → for empty files.
- Folder names have no "/" or ":|".
- Indentation with two spaces = subfolder depth.

Your task is to analyze:
1. The user's request.
2. Relevant search results about the task, technology, and example codebases.
3. The current .tree file (if provided).
4. The optional Q&A input (if provided). Q&A contains clarifying questions asked by you and answered by the user about missing data required to build the tree.

Rules:
- If the current tree already exists, modify or extend it.
- If no tree is given, create a new one.
- Only add imports for new files, do not change imports and code in existing files.
- Always follow proper programming paradigms.
- Tag existing files with <exist> and files to remove with <DELETE>.
- Always output a single .tree file in the following format:

<project_name>
  <folder>
    <file1.ext> <exist>: short inline text if small
    <file2.ext> <DELETE>
    <file3.ext>:
      inline one-liner content
    <file4.ext>:|
      multiline
      content
      here
    <file5.ext>
      (empty file, no ":" or ":|")

Conventions:
- Do NOT put ":|" after empty files.
- Do NOT put "/" or ":|" after folder names.
- Preserve and include essential imports and boilerplate needed for the technology (e.g., Python imports, VHDL entity declarations) only for new files.
- Always produce the most clean and minimal structure.
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
func (t *TreeLLM) GenerateTree(userInput, searchResults, currentTree string, QA string) (string, error) {
	prompt := fmt.Sprintf(`

User Input:
%s

Q&A:
%s
Search Results:
%s

Current Tree:
%s
`, userInput, QA, searchResults, currentTree)

	resp, err := t.Base.Call(strings.TrimSpace(prompt))
	if err != nil {
		return "", fmt.Errorf("TreeLLM failed: %w", err)
	}

	// Ensure we only return trimmed output
	return strings.TrimSpace(resp), nil
}
