package tool

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type DuckDuckGoSearch struct {
	*BaseSearch
}

func NewDuckDuckGoSearch() *DuckDuckGoSearch {
	return &DuckDuckGoSearch{
		BaseSearch: NewBaseSearch(),
	}
}

func (d *DuckDuckGoSearch) Name() string {
	return "duckduckgo_search"
}

func (d *DuckDuckGoSearch) Description() string {
	return "Perform a DuckDuckGo search and return a list of relevant links. Use this tool when you need to find information on the web with privacy-focused search."
}

func (d *DuckDuckGoSearch) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "(required) The search query to submit to DuckDuckGo.",
			},
			"num_results": map[string]interface{}{
				"type":        "integer",
				"description": "(optional) The number of search results to return. Default is 10.",
				"default":     10,
			},
		},
		"required": []string{"query"},
	}
}

func (d *DuckDuckGoSearch) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, ok := args["query"].(string)
	if !ok {
		return &ToolResult{Error: "query parameter is required"}, nil
	}

	numResults := 10
	if n, ok := args["num_results"].(float64); ok {
		numResults = int(n)
	}

	results, err := d.Search(ctx, query, numResults)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("Search failed: %v", err)}, nil
	}

	if len(results) == 0 {
		return &ToolResult{Output: "No search results found"}, nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("DuckDuckGo Search Results for: %s\n\n", query))
	for i, result := range results {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Title))
		output.WriteString(fmt.Sprintf("   URL: %s\n", result.URL))
		if result.Snippet != "" {
			output.WriteString(fmt.Sprintf("   %s\n", result.Snippet))
		}
		output.WriteString("\n")
	}

	return &ToolResult{Output: output.String()}, nil
}

func (d *DuckDuckGoSearch) Search(ctx context.Context, query string, numResults int) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s",
		url.QueryEscape(query))

	resp, err := d.makeRequest(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Parse DuckDuckGo results
	return d.parseHTMLResults(resp, ".result__a", numResults)
}
