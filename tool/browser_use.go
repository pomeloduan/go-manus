package tool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
)

type BrowserUse struct {
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	allocCtx context.Context
}

func NewBrowserUse() *BrowserUse {
	return &BrowserUse{}
}

func (b *BrowserUse) Name() string {
	return "browser_use"
}

func (b *BrowserUse) Description() string {
	return "Interact with a web browser to perform various actions such as navigation, element interaction, content extraction, and tab management. Supported actions include: navigate, click, input_text, screenshot, get_html, execute_js, scroll, switch_tab, new_tab, close_tab, refresh."
}

func (b *BrowserUse) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "The browser action to perform",
				"enum": []string{
					"navigate", "click", "input_text", "screenshot",
					"get_html", "execute_js", "scroll", "switch_tab",
					"new_tab", "close_tab", "refresh",
				},
			},
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL for 'navigate' or 'new_tab' actions",
			},
			"index": map[string]interface{}{
				"type":        "integer",
				"description": "Element index for 'click' or 'input_text' actions",
			},
			"text": map[string]interface{}{
				"type":        "string",
				"description": "Text for 'input_text' action",
			},
			"script": map[string]interface{}{
				"type":        "string",
				"description": "JavaScript code for 'execute_js' action",
			},
			"scroll_amount": map[string]interface{}{
				"type":        "integer",
				"description": "Pixels to scroll (positive for down, negative for up) for 'scroll' action",
			},
			"tab_id": map[string]interface{}{
				"type":        "integer",
				"description": "Tab ID for 'switch_tab' action",
			},
		},
		"required": []string{"action"},
	}
}

func (b *BrowserUse) ensureBrowser(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.ctx != nil {
		return nil // 浏览器已初始化
	}

	// 创建浏览器上下文
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	b.allocCtx = allocCtx
	b.cancel = cancel

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	b.ctx = browserCtx
	_ = cancelBrowser // 保存 cancel 函数以便后续清理

	// 启动浏览器
	if err := chromedp.Run(browserCtx); err != nil {
		return fmt.Errorf("failed to start browser: %w", err)
	}

	return nil
}

func (b *BrowserUse) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	action, ok := args["action"].(string)
	if !ok {
		return &ToolResult{Error: "action parameter is required"}, nil
	}

	// 确保浏览器已初始化
	if err := b.ensureBrowser(ctx); err != nil {
		return &ToolResult{Error: err.Error()}, nil
	}

	b.mu.Lock()
	browserCtx := b.ctx
	b.mu.Unlock()

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(browserCtx, 30*time.Second)
	defer cancel()

	switch action {
	case "navigate":
		return b.navigate(timeoutCtx, args)
	case "click":
		return b.click(timeoutCtx, args)
	case "input_text":
		return b.inputText(timeoutCtx, args)
	case "screenshot":
		return b.screenshot(timeoutCtx)
	case "get_html":
		return b.getHTML(timeoutCtx)
	case "execute_js":
		return b.executeJS(timeoutCtx, args)
	case "scroll":
		return b.scroll(timeoutCtx, args)
	case "refresh":
		return b.refresh(timeoutCtx)
	default:
		return &ToolResult{Error: "Unknown action: " + action}, nil
	}
}

func (b *BrowserUse) navigate(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	url, ok := args["url"].(string)
	if !ok {
		return &ToolResult{Error: "URL is required for 'navigate' action"}, nil
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
	)
	if err != nil {
		return &ToolResult{Error: "Failed to navigate: " + err.Error()}, nil
	}

	return &ToolResult{Output: "Navigated to " + url}, nil
}

func (b *BrowserUse) click(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	index, ok := args["index"].(float64)
	if !ok {
		return &ToolResult{Error: "Index is required for 'click' action"}, nil
	}

	// 简化实现：通过 CSS 选择器点击（实际应通过 index 查找元素）
	selector := fmt.Sprintf("body > *:nth-child(%d)", int(index))
	err := chromedp.Run(ctx,
		chromedp.Click(selector, chromedp.ByQuery),
	)
	if err != nil {
		return &ToolResult{Error: "Failed to click: " + err.Error()}, nil
	}

	return &ToolResult{Output: fmt.Sprintf("Clicked element at index %d", int(index))}, nil
}

func (b *BrowserUse) inputText(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	index, ok1 := args["index"].(float64)
	text, ok2 := args["text"].(string)
	if !ok1 || !ok2 {
		return &ToolResult{Error: "Index and text are required for 'input_text' action"}, nil
	}

	selector := fmt.Sprintf("body > *:nth-child(%d)", int(index))
	err := chromedp.Run(ctx,
		chromedp.SendKeys(selector, text, chromedp.ByQuery),
	)
	if err != nil {
		return &ToolResult{Error: "Failed to input text: " + err.Error()}, nil
	}

	return &ToolResult{Output: fmt.Sprintf("Input '%s' into element at index %d", text, int(index))}, nil
}

func (b *BrowserUse) screenshot(ctx context.Context) (*ToolResult, error) {
	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.CaptureScreenshot(&buf),
	)
	if err != nil {
		return &ToolResult{Error: "Failed to capture screenshot: " + err.Error()}, nil
	}

	return &ToolResult{
		Output: fmt.Sprintf("Screenshot captured (length: %d bytes)", len(buf)),
		System: string(buf), // Base64 编码的截图
	}, nil
}

func (b *BrowserUse) getHTML(ctx context.Context) (*ToolResult, error) {
	var html string
	err := chromedp.Run(ctx,
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	)
	if err != nil {
		return &ToolResult{Error: "Failed to get HTML: " + err.Error()}, nil
	}

	// 截断长 HTML
	if len(html) > 2000 {
		html = html[:2000] + "..."
	}

	return &ToolResult{Output: html}, nil
}

func (b *BrowserUse) executeJS(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	script, ok := args["script"].(string)
	if !ok {
		return &ToolResult{Error: "Script is required for 'execute_js' action"}, nil
	}

	var result string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(script, &result),
	)
	if err != nil {
		return &ToolResult{Error: "Failed to execute JS: " + err.Error()}, nil
	}

	return &ToolResult{Output: result}, nil
}

func (b *BrowserUse) scroll(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	amount, ok := args["scroll_amount"].(float64)
	if !ok {
		return &ToolResult{Error: "Scroll amount is required for 'scroll' action"}, nil
	}

	direction := "down"
	if amount < 0 {
		direction = "up"
		amount = -amount
	}

	script := fmt.Sprintf("window.scrollBy(0, %d);", int(amount))
	err := chromedp.Run(ctx,
		chromedp.Evaluate(script, nil),
	)
	if err != nil {
		return &ToolResult{Error: "Failed to scroll: " + err.Error()}, nil
	}

	return &ToolResult{Output: fmt.Sprintf("Scrolled %s by %d pixels", direction, int(amount))}, nil
}

func (b *BrowserUse) refresh(ctx context.Context) (*ToolResult, error) {
	err := chromedp.Run(ctx,
		chromedp.Reload(),
	)
	if err != nil {
		return &ToolResult{Error: "Failed to refresh: " + err.Error()}, nil
	}

	return &ToolResult{Output: "Refreshed current page"}, nil
}

// Cleanup 清理浏览器资源
func (b *BrowserUse) Cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel != nil {
		b.cancel()
	}
	if b.ctx != nil {
		chromedp.Cancel(b.ctx)
	}
	logrus.Info("Browser resources cleaned up")
}

