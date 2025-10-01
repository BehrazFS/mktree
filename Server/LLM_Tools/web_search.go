package llms

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// WebSearchResult stores title, URL and snippet for web search results
type WebSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// WebSearchResponse stores all web search results
type WebSearchResponse struct {
	Results []WebSearchResult `json:"results"`
}

// WebSearch fetches only the first page of DuckDuckGo results
func WebSearch(query string) (*WebSearchResponse, error) {
	baseURL := "https://duckduckgo.com/html/"

	// Build search URL
	params := url.Values{}
	params.Add("q", query)
	searchURL := baseURL + "?" + params.Encode()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent
	req.Header.Set("User-Agent", "LLM_Tools/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch search results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var results []WebSearchResult

	// Parse each result on the first page
	doc.Find("div.result__body").Each(func(i int, s *goquery.Selection) {
		title := s.Find("a.result__a").Text()
		urlAttr, exists := s.Find("a.result__a").Attr("href")
		if !exists {
			return
		}

		// Extract real URL from DuckDuckGo redirect
		parsedURL, err := url.Parse(urlAttr)
		if err == nil {
			if uddg := parsedURL.Query().Get("uddg"); uddg != "" {
				urlAttr = uddg
			}
		}

		snippet := s.Find("a.result__snippet, div.result__snippet").Text()

		if title != "" && urlAttr != "" {
			results = append(results, WebSearchResult{
				Title:   title,
				URL:     urlAttr,
				Snippet: snippet,
			})
		}
	})

	return &WebSearchResponse{
		Results: results,
	}, nil
}
