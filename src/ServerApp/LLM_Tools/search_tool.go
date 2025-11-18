package llms

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"
)

// ---------------- DUCKDUCKGO SEARCH ----------------
func WebSearch(q string) ([]SearchResult, error) {
	base := "https://duckduckgo.com/html/"
	params := url.Values{}
	params.Set("q", q)

	reqURL := base + "?" + params.Encode()
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("User-Agent", "SearchTool/2.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	doc.Find("div.result__body").Each(func(i int, s *goquery.Selection) {
		title := s.Find("a.result__a").Text()
		rawURL, ok := s.Find("a.result__a").Attr("href")
		if !ok {
			return
		}

		parsed, err := url.Parse(rawURL)
		if err == nil {
			if uddg := parsed.Query().Get("uddg"); uddg != "" {
				rawURL = uddg
			}
		}

		results = append(results, SearchResult{
			Title:   title,
			URL:     rawURL,
		})
	})

	return results, nil
}

// ---------------- CLEAR TEXT ----------------
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


// ---------------- MAIN CONTENT SCRAPER ----------------
func ExtractMainContent(pageURL string) (string, error) {
	article, err := readability.FromURL(pageURL, 20*time.Second)
	if err != nil {
		return "", err
	}

	text := cleanText(article.TextContent)
	return text, nil
}

// ---------------- CONCURRENT SEARCH TOOL ----------------
func RunSearch(query string) ([]SearchResult, error) {
	results, err := WebSearch(query)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Limit concurrency
	maxWorkers := 6
	sem := make(chan struct{}, maxWorkers)

	for i := range results {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			sem <- struct{}{} // Acquire slot
			defer func() { <-sem }() // Release slot
			content, err := ExtractMainContent(results[i].URL)
			mu.Lock()
			if err != nil {
				// Remove result at index i safely
				results = append(results[:i], results[i+1:]...)
			} else {
				results[i].Content = content
			}
			mu.Unlock()

		}(i)
	}

	wg.Wait()

	return results[:3] , nil
}
