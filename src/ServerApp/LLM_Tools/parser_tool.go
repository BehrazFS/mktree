package llms

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Node represents a file or directory
type Node struct {
	Name     string
	Type     string // "dir" or "file"
	Content  string
	Children []*Node
}

// countLeadingSpaces counts spaces at the start of a line
func countLeadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// parseEntry splits a line into name and optional content
func parseEntry(line string) (string, string) {
	if strings.Contains(line, ":|") {
		parts := strings.SplitN(line, ":|", 2)
		return strings.TrimSpace(parts[0]), ":|"
	} else if strings.Contains(line, ":") {
		parts := strings.SplitN(line, ":", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	} else {
		return strings.TrimSpace(line), ""
	}
}

// readTreeFile reads a .tree file into a slice of lines
func readTreeFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// readMultilineContent reads lines of multiline content
func readMultilineContent(lines []string, startIndex int, baseIndent int) (string, int) {
	contentLines := []string{}
	i := startIndex
	for i < len(lines) {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}
		indent := countLeadingSpaces(line) / 2
		if indent <= baseIndent {
			break
		}
		contentLines = append(contentLines, line[baseIndent*2+2:])
		i++
	}
	return strings.Join(contentLines, "\n"), i
}

// detectType determines if a line is a file or dir
func detectType(name, content string) string {
	if content != "" || strings.Contains(name, ".") {
		return "file"
	}
	return "dir"
}

// generateTreeOfNodesFromListOfLines converts lines to a Node tree
func generateTreeOfNodesFromListOfLines(lines []string) *Node {
	root := &Node{Name: "$ROOT", Type: "dir"}
	stack := []*Node{root}

	for i := 0; i < len(lines); {
		line := lines[i]
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			i++
			continue
		}

		indent := countLeadingSpaces(line) / 2
		name, contentIndicator := parseEntry(line)
		nodeType := detectType(name, contentIndicator)
		node := &Node{Name: name, Type: nodeType}

		if contentIndicator == ":|" {
			content, nextI := readMultilineContent(lines, i+1, indent)
			node.Content = content
			i = nextI
		} else if contentIndicator != "" {
			node.Content = contentIndicator
			i++
		} else {
			i++
		}

		// Adjust stack to current parent
		for indent < len(stack)-1 {
			stack = stack[:len(stack)-1]
		}
		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, node)

		if node.Type == "dir" {
			stack = append(stack, node)
		}
	}
	return root
}

// GenerateTreeOfNodesFromTreeFile parses a .tree file into a Node tree
func GenerateTreeOfNodesFromString(treeString string) *Node {
	lines := strings.Split(treeString, "\n")
	return generateTreeOfNodesFromListOfLines(lines)
}

// GenerateRichTreeFromTreeOfNodes prints the tree structure in a readable format
func GenerateRichTreeFromTreeOfNodes(node *Node, prefix string) {
	icon := "📁"
	if node.Type == "file" {
		icon = "📄"
	}
	fmt.Printf("%s%s %s\n", prefix, icon, node.Name)
	for _, child := range node.Children {
		GenerateRichTreeFromTreeOfNodes(child, prefix+"  ")
		if child.Content != "" && child.Type == "file" {
			contentLines := strings.Split(child.Content, "\n")
			for i, line := range contentLines {
				fmt.Printf("%s    📝 %d | %s\n", prefix, i+1, line)
			}
		}
	}
}
