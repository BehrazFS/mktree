package llms

import (
	"strings"
	// "fmt"
)

const TrackerSystemPrompt = `
You are a project tree analyzer. Your task is to compare a previous project tree with a new project tree and categorize the files into four categories:

1. ADDED – files that are new in the new tree.
2. MODIFIED – files that existed before but whose content has changed.
3. DELETE – files that existed before but are no longer present.
4. NOT_CHANGED – files that existed before and whose content has not changed.

Rules:
- Use file paths relative to the project root.
- The input tree may include markers:
    - <exist> – indicates that the file exists and its content is unchanged.
    - <DELETE> – indicates that the file was removed.
- Compare the content of files when available. If a file has <exist>, treat it as unchanged.
- If a file has <DELETE>, treat it as deleted.
- Output must be a valid JSON with exactly these four keys: "ADDED", "MODIFIED", "DELETE", "NOT_CHANGED".
- The values of each key must be a list of file paths (strings).
- Do not include any explanations, commentary, or extra text — only output the JSON.

Example input tree snippet:

server/
  app.py <exist>
  old_script.py <DELETE>
  new_module.py

Example output:

{
  "ADDED": ["server/new_module.py"],
  "MODIFIED": [],
  "DELETE": ["server/old_script.py"],
  "NOT_CHANGED": ["server/app.py"]
}
`

// TrackerLLM represents the LLM that tracks changes between project trees
type TrackerLLM struct {
	Base *LLM
}

// NewTrackerLLM creates a TrackerLLM using the base LLM constructor
func NewTrackerLLM(model string) *TrackerLLM {
	return &TrackerLLM{Base: NewLLM(model, TrackerSystemPrompt, false)}
}

// TrackChanges compares the previous and new project trees and categorizes files
func (t *TrackerLLM) TrackChanges(previousTree, newTree string) (string, error) {
	prompt := "Previous Tree: " + strings.TrimSpace(previousTree) + "\n" + "New Tree: " + strings.TrimSpace(newTree)
	response, err := t.Base.Call(prompt)
	if err != nil {
		return "", err
	}
	return response, nil
}
