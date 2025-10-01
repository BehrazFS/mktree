package llms

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// GetPageText fetches a URL and returns the text content of the page
func GetPageText(url string) (string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (LLM_Tools/1.0) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Parse HTML and extract text
	text, err := extractTextFromHTML(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	return text, nil
}

// extractTextFromHTML parses HTML and extracts all text content
func extractTextFromHTML(body io.Reader) (string, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return "", err
	}

	var text strings.Builder

	// Recursive function to traverse the HTML tree
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			// Add text content, but filter out script and style tags
			text.WriteString(n.Data)
		}

		// Skip script and style tags entirely
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}

		// Traverse child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	// Clean up the text
	return cleanText(text.String()), nil
}

// cleanText removes extra whitespace and normalizes text
func cleanText(rawText string) string {
	// Remove excessive whitespace
	text := strings.TrimSpace(rawText)

	// Replace multiple consecutive newlines with single newlines
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// Split by lines and clean each line
	lines := strings.Split(text, "\n")
	var cleanedLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			cleanedLines = append(cleanedLines, trimmedLine)
		}
	}

	// Join lines with proper spacing
	return strings.Join(cleanedLines, "\n\n")
}

// GetPageTextSimple is a simpler version that just returns all text
func GetPageTextSimple(url string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}

	return extractText(doc), nil
}

// extractText recursively extracts text from HTML nodes
func extractText(n *html.Node) string {
	if n == nil {
		return ""
	}

	var text strings.Builder

	// Skip script and style elements
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		return ""
	}

	// Add text content
	if n.Type == html.TextNode {
		text.WriteString(n.Data)
	}

	// Recursively process child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(extractText(c))
	}

	return text.String()
}
