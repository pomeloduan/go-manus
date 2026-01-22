package tool

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

type WebCrawler struct{}

func NewWebCrawler() *WebCrawler {
	return &WebCrawler{}
}

func (w *WebCrawler) Name() string {
	return "web_crawler"
}

func (w *WebCrawler) Description() string {
	return `Web crawler that extracts clean, AI-ready content from web pages.

Features:
- Extracts clean text content optimized for LLMs
- Handles basic HTML parsing
- Supports multiple URLs in a single request
- Fast and reliable with built-in error handling

Perfect for content analysis, research, and feeding web content to AI models.`
}

func (w *WebCrawler) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"urls": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "(required) List of URLs to crawl. Can be a single URL or multiple URLs.",
				"minItems":    1,
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "(optional) Timeout in seconds for each URL. Default is 30.",
				"default":     30,
				"minimum":     5,
				"maximum":     120,
			},
		},
		"required": []string{"urls"},
	}
}

func (w *WebCrawler) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	urlsInterface, ok := args["urls"].([]interface{})
	if !ok {
		// Try single URL string
		if urlStr, ok := args["urls"].(string); ok {
			urlsInterface = []interface{}{urlStr}
		} else {
			return &ToolResult{Error: "urls parameter is required and must be an array or string"}, nil
		}
	}

	timeout := 30
	if t, ok := args["timeout"].(float64); ok {
		timeout = int(t)
	}

	// Convert to string slice
	urls := make([]string, 0, len(urlsInterface))
	for _, u := range urlsInterface {
		if urlStr, ok := u.(string); ok {
			if w.isValidURL(urlStr) {
				urls = append(urls, urlStr)
			} else {
				logrus.Warnf("Invalid URL skipped: %s", urlStr)
			}
		}
	}

	if len(urls) == 0 {
		return &ToolResult{Error: "No valid URLs provided"}, nil
	}

	results := make([]map[string]interface{}, 0)
	successfulCount := 0
	failedCount := 0

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Process each URL
	for _, urlStr := range urls {
		result := w.crawlURL(ctx, client, urlStr, timeout)
		results = append(results, result)

		if result["success"].(bool) {
			successfulCount++
		} else {
			failedCount++
		}
	}

	// Format output
	var output strings.Builder
	output.WriteString("🕷️ Web Crawler Results Summary:\n")
	output.WriteString(fmt.Sprintf("📊 Total URLs: %d\n", len(urls)))
	output.WriteString(fmt.Sprintf("✅ Successful: %d\n", successfulCount))
	output.WriteString(fmt.Sprintf("❌ Failed: %d\n\n", failedCount))

	for i, result := range results {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, result["url"]))

		if result["success"].(bool) {
			output.WriteString(fmt.Sprintf("   ✅ Status: Success (HTTP %v)\n", result["status_code"]))
			if title, ok := result["title"].(string); ok && title != "" {
				output.WriteString(fmt.Sprintf("   📄 Title: %s\n", title))
			}
			if content, ok := result["content"].(string); ok {
				preview := content
				if len(preview) > 300 {
					preview = preview[:300] + "..."
				}
				output.WriteString(fmt.Sprintf("   📝 Content: %s\n", preview))
			}
			if wordCount, ok := result["word_count"].(int); ok {
				output.WriteString(fmt.Sprintf("   📊 Word Count: %d\n", wordCount))
			}
		} else {
			output.WriteString("   ❌ Status: Failed\n")
			if errMsg, ok := result["error_message"].(string); ok {
				output.WriteString(fmt.Sprintf("   🚫 Error: %s\n", errMsg))
			}
		}
		output.WriteString("\n")
	}

	return &ToolResult{Output: output.String()}, nil
}

func (w *WebCrawler) crawlURL(ctx context.Context, client *http.Client, urlStr string, timeout int) map[string]interface{} {
	startTime := time.Now()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return map[string]interface{}{
			"url":           urlStr,
			"success":       false,
			"error_message": fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	// Set User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"url":           urlStr,
			"success":       false,
			"error_message": fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"url":           urlStr,
			"success":       false,
			"error_message": fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"url":           urlStr,
			"success":       false,
			"error_message": fmt.Sprintf("Failed to read response: %v", err),
		}
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return map[string]interface{}{
			"url":           urlStr,
			"success":       false,
			"error_message": fmt.Sprintf("Failed to parse HTML: %v", err),
		}
	}

	// Extract title
	title := doc.Find("title").First().Text()
	title = strings.TrimSpace(title)

	// Extract text content (remove script and style tags)
	doc.Find("script, style").Remove()
	content := doc.Find("body").Text()
	content = strings.TrimSpace(content)

	// Clean up whitespace
	lines := strings.Split(content, "\n")
	cleanedLines := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}
	content = strings.Join(cleanedLines, "\n")

	wordCount := len(strings.Fields(content))
	executionTime := time.Since(startTime).Seconds()

	logrus.Infof("✅ Successfully crawled %s in %.2fs", urlStr, executionTime)

	return map[string]interface{}{
		"url":          urlStr,
		"success":      true,
		"status_code":  resp.StatusCode,
		"title":        title,
		"content":      content,
		"word_count":   wordCount,
		"execution_time": executionTime,
	}
}

func (w *WebCrawler) isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
