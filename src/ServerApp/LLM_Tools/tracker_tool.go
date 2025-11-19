package llms

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ChangeReport defines the structured output for the file changes.
// Using a struct ensures the JSON output is always valid and correctly formatted.
type ChangeReport struct {
	Added      []string `json:"ADDED"`
	Modified   []string `json:"MODIFIED"`
	Deleted    []string `json:"DELETE"`
	NotChanged []string `json:"NOT_CHANGED"`
}

// HashTracker is the main struct for our algorithmic tracker.
// It doesn't need any internal state like the LLM version.
type HashTracker struct{}

// NewHashTracker creates a new instance of the HashTracker.
func NewHashTracker() *HashTracker {
	return &HashTracker{}
}

// TrackChanges is the main function that compares two project trees.
// It takes string representations of the trees, where each line is "path <hash>".
func (ht *HashTracker) TrackChanges(previousTree, newTree string) (string, error) {
	// 1. Parse the input strings into maps of {path: hash}
	prevMap := parseTreeToMap(previousTree)
	newMap := parseTreeToMap(newTree)

	// 2. Initialize sets to store the results. Using map[string]struct{} is efficient for sets.
	addedSet := make(map[string]struct{})
	modifiedSet := make(map[string]struct{})
	deletedSet := make(map[string]struct{})
	notChangedSet := make(map[string]struct{})

	// 3. Find ADDED and MODIFIED/NOT_CHANGED files by iterating the new map
	for path, newHash := range newMap {
		prevHash, exists := prevMap[path]
		if !exists {
			// File is new
			addedSet[path] = struct{}{}
		} else if newHash != prevHash {
			// File existed but hash is different
			modifiedSet[path] = struct{}{}
		} else {
			// File existed and hash is the same
			notChangedSet[path] = struct{}{}
		}
	}

	// 4. Find DELETED files by iterating the previous map
	for path := range prevMap {
		if _, exists := newMap[path]; !exists {
			// File was in the previous tree but is missing from the new one
			deletedSet[path] = struct{}{}
		}
	}

	// 5. Assemble the final report
	report := ChangeReport{
		Added:      mapKeysToSlice(addedSet),
		Modified:   mapKeysToSlice(modifiedSet),
		Deleted:    mapKeysToSlice(deletedSet),
		NotChanged: mapKeysToSlice(notChangedSet),
	}

	// 6. Marshal the report into a JSON string
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	return string(jsonData), nil
}

// parseTreeToMap converts the flat "path <hash>" string into a map.
func parseTreeToMap(treeString string) map[string]string {
	fileMap := make(map[string]string)
	if treeString == "" {
		return fileMap
	}

	lines := strings.Split(strings.TrimSpace(treeString), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			path := strings.TrimSpace(parts[0])
			hash := strings.TrimSpace(parts[1])
			if path != "" {
				fileMap[path] = hash
			}
		}
	}
	return fileMap
}

// mapKeysToSlice is a helper to convert the keys of a set map into a string slice.
func mapKeysToSlice(m map[string]struct{}) []string {
	slice := make([]string, 0, len(m))
	for key := range m {
		slice = append(slice, key)
	}
	return slice
}
