package llms

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// ChangeReport defines the structured output for the file changes.
type ChangeReport struct {
	Added      []string `json:"ADDED"`
	Modified   []string `json:"MODIFIED"`
	Deleted    []string `json:"DELETE"`
	NotChanged []string `json:"NOT_CHANGED"`
}

// HashTracker compares two .tree files and reports changes.
type HashTracker struct{}

// NewHashTracker creates a new instance of HashTracker.
func NewHashTracker() *HashTracker {
	return &HashTracker{}
}

// TrackChanges compares two .tree files and returns a JSON report of changes.
func (ht *HashTracker) TrackChanges(previousTree, newTree string) (string, error) {
	prevRoot := GenerateTreeOfNodesFromString(previousTree)
	newRoot := GenerateTreeOfNodesFromString(newTree)

	prevMap := map[string]string{}
	newMap := map[string]string{}
	deletedSet := make(map[string]struct{})

	buildPathHashMap(prevRoot, "", prevMap, deletedSet)
	buildPathHashMap(newRoot, "", newMap, deletedSet)

	addedSet := make(map[string]struct{})
	modifiedSet := make(map[string]struct{})
	notChangedSet := make(map[string]struct{})

	// Compare new vs previous
	for path, newHash := range newMap {
		if _, isDeleted := deletedSet[path]; isDeleted {
			continue // skip <DELETE> files
		}

		prevHash, exists := prevMap[path]
		if !exists {
			addedSet[path] = struct{}{}
		} else if newHash != prevHash {
			modifiedSet[path] = struct{}{}
		} else {
			notChangedSet[path] = struct{}{}
		}
	}

	// Detect deleted files
	for path := range prevMap {
		if _, exists := newMap[path]; !exists || strings.Contains(path, "<DELETE>") {
			deletedSet[path] = struct{}{}
		}
	}

	report := ChangeReport{
		Added:      mapKeysToSlice(addedSet),
		Modified:   mapKeysToSlice(modifiedSet),
		Deleted:    mapKeysToSlice(deletedSet),
		NotChanged: mapKeysToSlice(notChangedSet),
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	return string(jsonData), nil
}

// buildPathHashMap recursively builds a {path: sha256(content)} map from a Node tree.
// Files containing "<DELETE>" are only added to deletedSet and not to the normal map.
func buildPathHashMap(node *Node, parentPath string, m map[string]string, deletedSet map[string]struct{}) {
	if node.Name == "$ROOT" {
		for _, child := range node.Children {
			buildPathHashMap(child, "", m, deletedSet)
		}
		return
	}

	currentPath := filepath.Join(parentPath, node.Name)

	// If filename contains "<DELETE>", mark as deleted and skip
	if node.Type == "file" && strings.Contains(node.Name, "<DELETE>") {
		deletedSet[currentPath] = struct{}{}
		return
	}

	if node.Type == "file" {
		m[currentPath] = sha256Hash(node.Content)
	} else if node.Type == "dir" {
		for _, child := range node.Children {
			buildPathHashMap(child, currentPath, m, deletedSet)
		}
	}
}

// sha256Hash generates a SHA256 hash for a string
func sha256Hash(text string) string {
	h := sha256.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

// mapKeysToSlice converts a map[string]struct{} to a slice of keys
func mapKeysToSlice(m map[string]struct{}) []string {
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}
