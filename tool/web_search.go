package tool

import (
	"context"
	"fmt"
	"strings"
)

type WebSearch struct {
	engines map[string]SearchEngine
}

func NewWebSearch() *WebSearch {
	ws := &WebSearch{
		engines: make(map[string]SearchEngine),
	}

	// Register search engines
	ws.engines["google"] = NewGoogleSearch()
	ws.engines["baidu"] = NewBaiduSearch()
	ws.engines["bing"] = NewBingSearch()
	ws.engines["duckduckgo"] = NewDuckDuckGoSearch()

	return ws
}

func (w *WebSearch) Name() string {
	return "web_search"
}

func (w *WebSearch) Description() string {
	return `Unified web search tool that supports multiple search engines.
Available engines: google, baidu, bing, duckduckgo.
Use this tool when you need to find information on the web. The tool will try multiple engines if one fails.`
}

func (w *WebSearch) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "(required) The search query.",
			},
			"engine": map[string]interface{}{
				"type":        "string",
				"description": "(optional) Search engine to use. Options: google, baidu, bing, duckduckgo. Default is google.",
				"enum":        []string{"google", "baidu", "bing", "duckduckgo"},
				"default":     "google",
			},
			"num_results": map[string]interface{}{
				"type":        "integer",
				"description": "(optional) The number of search results to return. Default is 10.",
				"default":     10,
			},
			"fallback_engines": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "(optional) Fallback engines to try if primary engine fails.",
			},
		},
		"required": []string{"query"},
	}
}

func (w *WebSearch) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, ok := args["query"].(string)
	if !ok {
		return &ToolResult{Error: "query parameter is required"}, nil
	}

	engineName := "google"
	if e, ok := args["engine"].(string); ok && e != "" {
		engineName = e
	}

	numResults := 10
	if n, ok := args["num_results"].(float64); ok {
		numResults = int(n)
	}

	// Get primary engine
	engine, exists := w.engines[engineName]
	if !exists {
		return &ToolResult{Error: fmt.Sprintf("Unknown search engine: %s", engineName)}, nil
	}

	// Try primary engine
	result, err := w.trySearch(ctx, engine, query, numResults)
	if err == nil {
		return result, nil
	}

	// Try fallback engines
	var fallbackEngines []string
	if fe, ok := args["fallback_engines"].([]interface{}); ok {
		for _, e := range fe {
			if engineStr, ok := e.(string); ok {
				fallbackEngines = append(fallbackEngines, engineStr)
			}
		}
	} else {
		// Default fallback order
		fallbackEngines = []string{"bing", "duckduckgo", "baidu"}
	}

	errors := []string{fmt.Sprintf("%s: %v", engineName, err)}
	for _, fallbackName := range fallbackEngines {
		if fallbackName == engineName {
			continue
		}
		fallbackEngine, exists := w.engines[fallbackName]
		if !exists {
			continue
		}

		result, err := w.trySearch(ctx, fallbackEngine, query, numResults)
		if err == nil {
			return &ToolResult{
				Output: fmt.Sprintf("Primary engine (%s) failed, but fallback engine (%s) succeeded:\n\n%s",
					engineName, fallbackName, result.Output),
			}, nil
		}
		errors = append(errors, fmt.Sprintf("%s: %v", fallbackName, err))
	}

	return &ToolResult{
		Error: fmt.Sprintf("All search engines failed:\n%s", strings.Join(errors, "\n")),
	}, nil
}

func (w *WebSearch) trySearch(ctx context.Context, engine SearchEngine, query string, numResults int) (*ToolResult, error) {
	// Try to use Search method if available
	if searcher, ok := engine.(interface {
		Search(ctx context.Context, query string, numResults int) ([]SearchResult, error)
	}); ok {
		results, err := searcher.Search(ctx, query, numResults)
		if err != nil {
			return nil, err
		}

		var output strings.Builder
		output.WriteString(fmt.Sprintf("%s Search Results for: %s\n\n", engine.Name(), query))
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

	// Fallback to Execute method (engine must also implement Tool interface)
	if toolEngine, ok := engine.(Tool); ok {
		args := map[string]interface{}{
			"query":       query,
			"num_results": numResults,
		}
		return toolEngine.Execute(ctx, args)
	}

	return nil, fmt.Errorf("engine does not implement Search or Tool interface")
}
