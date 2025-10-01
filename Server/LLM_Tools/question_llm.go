package llms

import (
	// "fmt"
	"encoding/json"
	"strings"
)

const QuestionSystemPrompt = `
You are a project clarification question generator.

Your task is to analyze:
1. The user's request (prompt).
2. The current .tree file (if provided).
3. The previous Q&A list (if provided).

Goal:
- Identify missing, unclear, or ambiguous information that would help the project tree generator produce a more complete and accurate .tree structure.
- Focus on clarifying three key areas:
  1. "task_query" – The main coding task or feature the user wants to implement or modify.
  2. "technology_query" – The specific technologies, frameworks, or programming languages relevant to the task.
  3. "codebase_query" – The file or directory structure patterns, filenames, or components in the .tree that need to be modified or used.
- Generate clarifying questions only if they are strictly necessary to improve understanding of these three areas.
- If no additional clarification is needed, output an empty list.

Output Format:
- Always output a valid JSON list of strings.
- Example:
  ["What programming language should be used for the server file?", 
   "Should the new authentication logic be placed in auth.py or a new file?"]

- If no questions are needed:
  []

Rules:
- Do NOT explain or comment, only output the JSON list.
- Keep questions specific and concise.
- Avoid repeating questions that already exist in the previous Q&A.
- Only ask for information that directly affects understanding of task, technology, or codebase structure for building the .tree.

`

// QuestionLLM represents the LLM that generates clarification questions
type QuestionLLM struct {
	Base *LLM
}

// NewQuestionLLM creates a QuestionLLM using the base LLM constructor
func NewQuestionLLM(model string) *QuestionLLM {
	return &QuestionLLM{Base: NewLLM(model, QuestionSystemPrompt, false)}
}

// call the LLM to generate clarification questions
func (q *QuestionLLM) GenerateQuestions(prompt, tree string, previousQA string) ([]string, error) {
	// Prepare the input by combining prompt, tree, and previous Q&A
	var inputBuilder strings.Builder
	inputBuilder.WriteString("User Request:\n")
	inputBuilder.WriteString(prompt)
	inputBuilder.WriteString("\n\n")

	if tree != "" {
		inputBuilder.WriteString("Current .tree file:\n")
		inputBuilder.WriteString(tree)
		inputBuilder.WriteString("\n\n")
	}

	if previousQA != "" {
		inputBuilder.WriteString("Previous Q&A:\n")
		inputBuilder.WriteString(previousQA)
		inputBuilder.WriteString("\n\n")
	}

	// Call the LLM with the prepared input
	response, err := q.Base.Call(inputBuilder.String())
	if err != nil {
		return nil, err
	}

	// Parse the JSON response into a slice of strings
	var questions []string
	err = json.Unmarshal([]byte(response), &questions)
	if err != nil {
		return nil, err
	}

	return questions, nil
}
