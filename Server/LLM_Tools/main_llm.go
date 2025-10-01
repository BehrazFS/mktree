package llms

import (
	"fmt"
	"log"
)

func AgentProcess(log *log.Logger, user_query string, curr_tree string, op_type string, QA string, req_id string) (string, string, []string, error) {

	log.Printf("[req_id=%s] The Tree generation is started", req_id)

	Searchllm := NewSearchLLM("gpt-oss-120b")

	log.Printf("[req_id=%s] Searching and selecting relevant data...", req_id)
	// Call the LLM with a query
	response, err := Searchllm.Call(user_query, curr_tree, QA)
	if err != nil {
		log.Printf("[req_id=%s] Error: %v", req_id, err)
		return "", "", nil, err
	}
	log.Printf("[req_id=%s] Search and selection done. Now trimming...", req_id)
	trimllm := NewTrimLLM("gpt-oss-120b")
	trimmed, err := trimllm.TrimResults(user_query, response)
	if err != nil {
		log.Printf("[req_id=%s] Error: %v", req_id, err)
		return "", "", nil, err
	}

	treellm := NewTreeLLM("qwen-3-coder-480b")
	log.Printf("[req_id=%s] Trimming done. Now generating tree...", req_id)
	tree, err := treellm.GenerateTree(user_query, fmt.Sprintf("%v", trimmed), curr_tree, QA)
	if err != nil {
		log.Printf("[req_id=%s] Error: %v", req_id, err)
		return "", "", nil, err
	}

	log.Printf("[req_id=%s] The Tree has been generated", req_id)

	tracker := NewTrackerLLM("qwen-3-235b-a22b-instruct-2507")
	changes, err := tracker.TrackChanges(curr_tree, tree)
	if err != nil {
		log.Printf("[req_id=%s] Error: %v", req_id, err)
		return "", "", nil, err
	}

	log.Printf("[req_id=%s] The changes have been tracked", req_id)

	questionllm := NewQuestionLLM("qwen-3-235b-a22b-instruct-2507")
	log.Printf("[req_id=%s] Now checking if any clarifying questions are needed...", req_id)
	questions, err := questionllm.GenerateQuestions(user_query, curr_tree, "")
	if err != nil {
		log.Printf("[req_id=%s] Error: %v", req_id, err)
		return "", "", nil, err
	}
	log.Printf("[req_id=%s] Clarifying questions generation done", req_id)

	return tree, changes, questions, nil
}
