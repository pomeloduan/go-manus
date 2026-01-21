package tool

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type GoogleSearch struct{}

func NewGoogleSearch() *GoogleSearch {
	return &GoogleSearch{}
}

func (g *GoogleSearch) Name() string {
	return "google_search"
}

func (g *GoogleSearch) Description() string {
	return "Perform a Google search and return a list of relevant links. Use this tool when you need to find information on the web, get up-to-date data, or research specific topics. The tool returns a list of URLs that match the search query."
}

func (g *GoogleSearch) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "(required) The search query to submit to Google.",
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

func (g *GoogleSearch) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, ok := args["query"].(string)
	if !ok {
		return &ToolResult{Error: "query parameter is required"}, nil
	}

	numResults := 10
	if n, ok := args["num_results"].(float64); ok {
		numResults = int(n)
	}

	// 构造 Google 搜索 URL
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&num=%d",
		url.QueryEscape(query), numResults)

	// 发送 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return &ToolResult{Error: "Failed to create request: " + err.Error()}, nil
	}

	// 设置 User-Agent 以避免被阻止
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &ToolResult{Error: "Failed to execute search: " + err.Error()}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ToolResult{Error: fmt.Sprintf("Search failed with status: %d", resp.StatusCode)}, nil
	}

	// 读取响应
	body := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	// 简单的 HTML 解析提取链接（实际项目中应使用 HTML 解析库）
	links := g.extractLinks(string(body), numResults)

	if len(links) == 0 {
		return &ToolResult{
			Output: "No search results found. Note: Google may require more sophisticated parsing or API access.",
		}, nil
	}

	result := strings.Join(links, "\n")
	return &ToolResult{Output: result}, nil
}

// extractLinks 从 HTML 中提取链接（简化版）
func (g *GoogleSearch) extractLinks(html string, maxResults int) []string {
	links := make([]string, 0)
	
	// 匹配 Google 搜索结果中的链接模式
	// 这是一个简化的实现，实际应使用 HTML 解析库
	re := regexp.MustCompile(`href="(https?://[^"]+)"`)
	matches := re.FindAllStringSubmatch(html, -1)
	
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			link := match[1]
			// 过滤掉 Google 内部链接
			if !strings.Contains(link, "google.com") && !seen[link] {
				links = append(links, link)
				seen[link] = true
				if len(links) >= maxResults {
					break
				}
			}
		}
	}
	
	return links
}

