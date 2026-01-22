package tool

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// SearchEngine 搜索引擎接口
type SearchEngine interface {
	Search(ctx context.Context, query string, numResults int) ([]SearchResult, error)
	Name() string
}

// SearchResult 搜索结果
type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

// BaseSearch 基础搜索工具
type BaseSearch struct {
	client *http.Client
}

func NewBaseSearch() *BaseSearch {
	return &BaseSearch{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (b *BaseSearch) makeRequest(ctx context.Context, searchURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	return b.client.Do(req)
}

func (b *BaseSearch) parseHTMLResults(resp *http.Response, selector string, maxResults int) ([]SearchResult, error) {
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0)
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		if len(results) >= maxResults {
			return
		}

		title := strings.TrimSpace(s.Text())
		link, exists := s.Attr("href")
		if !exists {
			return
		}

		// Get snippet (next sibling or parent's text)
		snippet := ""
		if s.Parent() != nil {
			snippet = strings.TrimSpace(s.Parent().Text())
			if len(snippet) > 200 {
				snippet = snippet[:200] + "..."
			}
		}

		results = append(results, SearchResult{
			Title:   title,
			URL:     link,
			Snippet: snippet,
		})
	})

	return results, nil
}
