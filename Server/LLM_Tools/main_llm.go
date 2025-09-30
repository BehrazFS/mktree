package llms

import (
	"fmt"
)

func AgentProcess(user_query string, curr_tree string, op_type string) (string, error) {

	fmt.Printf("The Tree generation is started\n")

	Searchllm := NewSearchLLM("gpt-oss-120b")

	// Call the LLM with a query
	response, err := Searchllm.Call(user_query, curr_tree)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", err
	}
	fmt.Println("Search and selection done. Now trimming...")
	trimllm := NewTrimLLM("gpt-oss-120b")
	trimmed, err := trimllm.TrimResults(user_query, response)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", err
	}

	treellm := NewTreeLLM("qwen-3-coder-480b")
	fmt.Println("Trimming done. Now generating tree...")
	tree, err := treellm.GenerateTree(user_query, fmt.Sprintf("%v", trimmed), curr_tree)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", err
	}

	fmt.Printf("The Tree has been generated\n")

	return tree, nil
}
