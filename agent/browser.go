package agent

import (
	"context"
	"fmt"

	"go-manus/logger"
	"go-manus/schema"
	"go-manus/tool"
)

// BrowserContextHelper 浏览器上下文助手
type BrowserContextHelper struct {
	agent interface{} // 可以是 BaseAgent 或 ToolCallAgent
}

func NewBrowserContextHelper(agent interface{}) *BrowserContextHelper {
	return &BrowserContextHelper{
		agent: agent,
	}
}

// GetBrowserState 获取浏览器当前状态
func (b *BrowserContextHelper) GetBrowserState(ctx context.Context) (map[string]interface{}, error) {
	// 需要从 ToolCallAgent 获取工具
	toolCallAgent, ok := b.agent.(*ToolCallAgent)
	if !ok {
		return nil, fmt.Errorf("Agent is not a ToolCallAgent")
	}

	browserTool := toolCallAgent.GetTool("browser_use")
	if browserTool == nil {
		return nil, fmt.Errorf("BrowserUseTool not found")
	}

	// 尝试获取浏览器状态
	// 注意：这需要 BrowserUse 工具支持 get_current_state 方法
	// 目前简化实现，返回空状态
	return map[string]interface{}{
		"url":   "N/A",
		"title": "N/A",
		"tabs":  []string{},
	}, nil
}

// FormatNextStepPrompt 格式化下一步提示词，包含浏览器状态
func (b *BrowserContextHelper) FormatNextStepPrompt(ctx context.Context) (string, error) {
	state, err := b.GetBrowserState(ctx)
	if err != nil {
		logger.Warningf("Failed to get browser state: %v", err)
		state = map[string]interface{}{}
	}

	urlInfo := ""
	titleInfo := ""
	tabsInfo := ""

	if url, ok := state["url"].(string); ok && url != "N/A" {
		urlInfo = fmt.Sprintf("\n   URL: %s", url)
	}
	if title, ok := state["title"].(string); ok && title != "N/A" {
		titleInfo = fmt.Sprintf("\n   Title: %s", title)
	}
	if tabs, ok := state["tabs"].([]string); ok && len(tabs) > 0 {
		tabsInfo = fmt.Sprintf("\n   %d tab(s) available", len(tabs))
	}

	prompt := fmt.Sprintf(`What should I do next to achieve my goal?

When you see [Current state starts here], focus on the following:
- Current URL and page title%s
- Available tabs%s
- Interactive elements and their indices
- Content above%s or below%s the viewport (if indicated)
- Any action results or errors

For browser interactions:
- To navigate: browser_use with action="go_to_url", url="..."
- To click: browser_use with action="click_element", index=N
- To type: browser_use with action="input_text", index=N, text="..."
- To extract: browser_use with action="extract_content", goal="..."
- To scroll: browser_use with action="scroll_down" or "scroll_up"

Consider both what's visible and what might be beyond the current viewport.
Be methodical - remember your progress and what you've learned so far.

If you want to stop the interaction at any point, use the terminate tool/function call.`,
		urlInfo, tabsInfo, "", "")

	return prompt, nil
}

// CleanupBrowser 清理浏览器资源
func (b *BrowserContextHelper) CleanupBrowser(ctx context.Context) error {
	toolCallAgent, ok := b.agent.(*ToolCallAgent)
	if !ok {
		return nil
	}

	browserTool := toolCallAgent.GetTool("browser_use")
	if browserTool == nil {
		return nil
	}

	// 如果工具有 Cleanup 方法，调用它
	if cleanupTool, ok := browserTool.(interface {
		Cleanup()
	}); ok {
		cleanupTool.Cleanup()
	}

	return nil
}

// BrowserAgent 浏览器专用 Agent
type BrowserAgent struct {
	*ToolCallAgent
	browserContextHelper *BrowserContextHelper
}

// NewBrowserAgent 创建浏览器 Agent
func NewBrowserAgent() *BrowserAgent {
	agent := &BrowserAgent{
		ToolCallAgent: NewToolCallAgent("browser"),
	}

	// 设置提示词（来自 Python 版本的 app/prompt/browser.py）
	agent.SystemPrompt = `You are an AI agent designed to automate browser tasks. Your goal is to accomplish the ultimate task following the rules.

# Input Format
Task
Previous steps
Current URL
Open Tabs
Interactive Elements
[index]<type>text</type>
- index: Numeric identifier for interaction
- type: HTML element type (button, input, etc.)
- text: Element description
Example:
[33]<button>Submit Form</button>

- Only elements with numeric indexes in [] are interactive
- elements without [] provide only context

# Response Rules
1. RESPONSE FORMAT: You must ALWAYS respond with valid JSON in this exact format:
{"current_state": {"evaluation_previous_goal": "Success|Failed|Unknown - Analyze the current elements and the image to check if the previous goals/actions are successful like intended by the task. Mention if something unexpected happened. Shortly state why/why not",
"memory": "Description of what has been done and what you need to remember. Be very specific. Count here ALWAYS how many times you have done something and how many remain. E.g. 0 out of 10 websites analyzed. Continue with abc and xyz",
"next_goal": "What needs to be done with the next immediate action"}},
"action":[{"one_action_name": {// action-specific parameter}}, // ... more actions in sequence]}

2. ACTIONS: You can specify multiple actions in the list to be executed in sequence. But always specify only one action name per item. Use maximum actions per sequence.
Common action sequences:
- Form filling: [{"input_text": {"index": 1, "text": "username"}}, {"input_text": {"index": 2, "text": "password"}}, {"click_element": {"index": 3}}}]
- Navigation and extraction: [{"go_to_url": {"url": "https://example.com"}}, {"extract_content": {"goal": "extract the names"}}}]
- Actions are executed in the given order
- If the page changes after an action, the sequence is interrupted and you get the new state.
- Only provide the action sequence until an action which changes the page state significantly.
- Try to be efficient, e.g. fill forms at once, or chain actions where nothing changes on the page
- only use multiple actions if it makes sense.

3. ELEMENT INTERACTION:
- Only use indexes of the interactive elements
- Elements marked with "[]Non-interactive text" are non-interactive

4. NAVIGATION & ERROR HANDLING:
- If no suitable elements exist, use other functions to complete the task
- If stuck, try alternative approaches - like going back to a previous page, new search, new tab etc.
- Handle popups/cookies by accepting or closing them
- Use scroll to find elements you are looking for
- If you want to research something, open a new tab instead of using the current tab
- If captcha pops up, try to solve it - else try a different approach
- If the page is not fully loaded, use wait action

5. TASK COMPLETION:
- Use the done action as the last action as soon as the ultimate task is complete
- Dont use "done" before you are done with everything the user asked you, except you reach the last step of max_steps.
- If you reach your last step, use the done action even if the task is not fully finished. Provide all the information you have gathered so far. If the ultimate task is completly finished set success to true. If not everything the user asked for is completed set success in done to false!
- If you have to do something repeatedly for example the task says for "each", or "for all", or "x times", count always inside "memory" how many times you have done it and how many remain. Don't stop until you have completed like the task asked you. Only call done after the last step.
- Don't hallucinate actions
- Make sure you include everything you found out for the ultimate task in the done text parameter. Do not just say you are done, but include the requested information of the task.

6. VISUAL CONTEXT:
- When an image is provided, use it to understand the page layout
- Bounding boxes with labels on their top right corner correspond to element indexes

7. Form filling:
- If you fill an input field and your action sequence is interrupted, most often something changed e.g. suggestions popped up under the field.

8. Long tasks:
- Keep track of the status and subresults in the memory.

9. Extraction:
- If your task is to find information - call extract_content on the specific pages to get and store the information.
Your responses must be always JSON with the specified format.`

	agent.NextStepPrompt = `What should I do next to achieve my goal?

When you see [Current state starts here], focus on the following:
- Current URL and page title{url_placeholder}
- Available tabs{tabs_placeholder}
- Interactive elements and their indices
- Content above{content_above_placeholder} or below{content_below_placeholder} the viewport (if indicated)
- Any action results or errors{results_placeholder}

For browser interactions:
- To navigate: browser_use with action="go_to_url", url="..."
- To click: browser_use with action="click_element", index=N
- To type: browser_use with action="input_text", index=N, text="..."
- To extract: browser_use with action="extract_content", goal="..."
- To scroll: browser_use with action="scroll_down" or "scroll_up"

Consider both what's visible and what might be beyond the current viewport.
Be methodical - remember your progress and what you've learned so far.

If you want to stop the interaction at any point, use the terminate tool/function call.`

	// 配置工具
	agent.AvailableTools = tool.NewToolCollection(
		tool.NewBrowserUse(),
		tool.NewTerminate(),
	)

	agent.Description = "A browser agent that can control a browser to accomplish tasks"

	// 初始化浏览器上下文助手
	agent.browserContextHelper = NewBrowserContextHelper(agent.ToolCallAgent)

	return agent
}

// Think 思考下一步行动，包含浏览器状态
func (b *BrowserAgent) Think(ctx context.Context) (bool, error) {
	// 更新提示词以包含浏览器状态
	prompt, err := b.browserContextHelper.FormatNextStepPrompt(ctx)
	if err == nil {
		b.NextStepPrompt = prompt
	}

	return b.ToolCallAgent.Think(ctx)
}

// Cleanup 清理资源
func (b *BrowserAgent) Cleanup(ctx context.Context) error {
	if b.browserContextHelper != nil {
		b.browserContextHelper.CleanupBrowser(ctx)
	}
	return nil
}
