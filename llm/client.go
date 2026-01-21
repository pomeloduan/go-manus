package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"go-manus/config"
	"go-manus/schema"
)

type Client struct {
	client      *openai.Client
	model       string
	maxTokens   int
	temperature float64
}

// NewClient 创建新的 LLM 客户端
func NewClient(configName string) *Client {
	cfg := config.GetInstance()
	settings := cfg.GetLLM(configName)

	clientConfig := openai.DefaultConfig(settings.APIKey)
	clientConfig.BaseURL = settings.BaseURL

	return &Client{
		client:      openai.NewClientWithConfig(clientConfig),
		model:       settings.Model,
		maxTokens:   settings.MaxTokens,
		temperature: settings.Temperature,
	}
}

// FormatMessages 格式化消息为 OpenAI 格式
func FormatMessages(messages []schema.Message) []openai.ChatCompletionMessage {
	formatted := make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, msg := range messages {
		formattedMsg := openai.ChatCompletionMessage{
			Role: string(msg.Role),
		}
		if msg.Content != nil {
			formattedMsg.Content = *msg.Content
		}
		if len(msg.ToolCalls) > 0 {
			toolCalls := make([]openai.ToolCall, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				toolCall := openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
				toolCalls = append(toolCalls, toolCall)
			}
			formattedMsg.ToolCalls = toolCalls
		}
		if msg.Name != nil {
			formattedMsg.Name = *msg.Name
		}
		if msg.ToolCallID != nil {
			formattedMsg.ToolCallID = *msg.ToolCallID
		}
		formatted = append(formatted, formattedMsg)
	}
	return formatted
}

// Ask 发送消息并获取响应（无工具调用）
func (c *Client) Ask(ctx context.Context, messages []schema.Message, systemMsgs []schema.Message) (string, error) {
	allMessages := make([]schema.Message, 0)
	if len(systemMsgs) > 0 {
		allMessages = append(allMessages, systemMsgs...)
	}
	allMessages = append(allMessages, messages...)

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    FormatMessages(allMessages),
		MaxTokens:   c.maxTokens,
		Temperature: float32(c.temperature),
		Stream:      false,
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty response from LLM")
	}

	return resp.Choices[0].Message.Content, nil
}

// AskTool 发送消息并获取响应（支持工具调用）
func (c *Client) AskTool(ctx context.Context, messages []schema.Message, systemMsgs []schema.Message, tools []openai.Tool, toolChoice string) (*ChatCompletionMessage, error) {
	allMessages := make([]schema.Message, 0)
	if len(systemMsgs) > 0 {
		allMessages = append(allMessages, systemMsgs...)
	}
	allMessages = append(allMessages, messages...)

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    FormatMessages(allMessages),
		MaxTokens:   c.maxTokens,
		Temperature: float32(c.temperature),
		Tools:       tools,
	}

	// 设置工具选择策略
	switch toolChoice {
	case "none":
		req.ToolChoice = "none"
	case "required":
		req.ToolChoice = "required"
	case "auto", "":
		req.ToolChoice = "auto"
	default:
		req.ToolChoice = "auto"
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from LLM")
	}

	msg := resp.Choices[0].Message
	result := &ChatCompletionMessage{
		Content: msg.Content,
	}

	// 转换工具调用
	if len(msg.ToolCalls) > 0 {
		toolCalls := make([]schema.ToolCall, 0, len(msg.ToolCalls))
		for _, tc := range msg.ToolCalls {
		// Arguments 已经是字符串类型（JSON 格式）
		argsJSON := tc.Function.Arguments
		if argsJSON == "" {
			argsJSON = "{}"
		}
			toolCall := schema.ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: schema.Function{
					Name:      tc.Function.Name,
					Arguments: argsJSON,
				},
			}
			toolCalls = append(toolCalls, toolCall)
		}
		result.ToolCalls = toolCalls
	}

	return result, nil
}

// ChatCompletionMessage LLM 响应消息
type ChatCompletionMessage struct {
	Content   string
	ToolCalls []schema.ToolCall
}

// AskWithRetry 带重试的请求
func (c *Client) AskWithRetry(ctx context.Context, messages []schema.Message, systemMsgs []schema.Message, maxRetries int) (string, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			waitTime := time.Duration(i) * time.Second
			logrus.Warnf("Retrying after %v...", waitTime)
			time.Sleep(waitTime)
		}
		result, err := c.Ask(ctx, messages, systemMsgs)
		if err == nil {
			return result, nil
		}
		lastErr = err
		logrus.Errorf("Attempt %d failed: %v", i+1, err)
	}
	return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// AskToolWithRetry 带重试的工具调用请求
func (c *Client) AskToolWithRetry(ctx context.Context, messages []schema.Message, systemMsgs []schema.Message, tools []openai.Tool, toolChoice string, maxRetries int) (*ChatCompletionMessage, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			waitTime := time.Duration(i) * time.Second
			logrus.Warnf("Retrying after %v...", waitTime)
			time.Sleep(waitTime)
		}
		result, err := c.AskTool(ctx, messages, systemMsgs, tools, toolChoice)
		if err == nil {
			return result, nil
		}
		lastErr = err
		logrus.Errorf("Attempt %d failed: %v", i+1, err)
	}
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// ToolToOpenAI 将工具定义转换为 OpenAI 格式
func ToolToOpenAI(name, description string, parameters map[string]interface{}) openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
	}
}

