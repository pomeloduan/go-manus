package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"go-manus/logger"
	"go-manus/schema"
	"go-manus/tool"
)

// ToolCallAgent 支持工具调用的 Agent
type ToolCallAgent struct {
	*ReActAgent

	AvailableTools *tool.ToolCollection
	ToolChoices    string // "none", "auto", "required"
	SpecialToolNames []string
	ToolCalls      []schema.ToolCall
}

// NewToolCallAgent 创建工具调用 Agent
func NewToolCallAgent(name string) *ToolCallAgent {
	tc := &ToolCallAgent{
		ReActAgent:      NewReActAgent(name),
		ToolChoices:     "auto",
		SpecialToolNames: []string{"terminate"},
		AvailableTools:  tool.NewToolCollection(tool.NewTerminate()),
	}
	tc.BaseAgent.MaxSteps = 30
	return tc
}

// Think 思考下一步行动
func (a *ToolCallAgent) Think(ctx context.Context) (bool, error) {
	if a.NextStepPrompt != "" {
		userMsg := schema.NewUserMessage(a.NextStepPrompt)
		a.Memory.AddMessage(userMsg)
	}

	// 准备系统消息
	systemMsgs := make([]schema.Message, 0)
	if a.SystemPrompt != "" {
		systemMsgs = append(systemMsgs, schema.NewSystemMessage(a.SystemPrompt))
	}

	// 转换工具为 OpenAI 格式
	openAITools := make([]openai.Tool, 0)
	for _, t := range a.AvailableTools.ToOpenAITools() {
		toolMap := t.(map[string]interface{})
		if funcMap, ok := toolMap["function"].(map[string]interface{}); ok {
			params, _ := funcMap["parameters"].(map[string]interface{})
			openAITools = append(openAITools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        funcMap["name"].(string),
					Description: funcMap["description"].(string),
					Parameters:  params,
				},
			})
		}
	}

	// 调用 LLM
	response, err := a.LLM.AskTool(ctx, a.Memory.Messages, systemMsgs, openAITools, a.ToolChoices)
	if err != nil {
		logger.Errorf("LLM request failed: %v", err)
		a.Memory.AddMessage(schema.NewAssistantMessage("Error encountered while processing: " + err.Error()))
		return false, err
	}

	logger.Infof("✨ %s's thoughts: %s", a.Name, response.Content)
	logger.Infof("🛠️ %s selected %d tools to use", a.Name, len(response.ToolCalls))

	if len(response.ToolCalls) > 0 {
		toolNames := make([]string, 0, len(response.ToolCalls))
		for _, tc := range response.ToolCalls {
			toolNames = append(toolNames, tc.Function.Name)
		}
		logger.Infof("🧰 Tools being prepared: %v", toolNames)
	}

	// 保存工具调用
	a.ToolCalls = response.ToolCalls

	// 创建助手消息
	var assistantMsg schema.Message
	if len(response.ToolCalls) > 0 {
		assistantMsg = schema.NewMessageFromToolCalls(response.Content, response.ToolCalls)
	} else {
		assistantMsg = schema.NewAssistantMessage(response.Content)
	}
	a.Memory.AddMessage(assistantMsg)

	// 处理不同的工具选择模式
	if a.ToolChoices == "none" {
		if len(response.ToolCalls) > 0 {
			logger.Warningf("🤔 Hmm, %s tried to use tools when they weren't available!", a.Name)
		}
		return response.Content != "", nil
	}

	if a.ToolChoices == "required" && len(response.ToolCalls) == 0 {
		return true, nil // 将在 act() 中处理
	}

	if a.ToolChoices == "auto" && len(response.ToolCalls) == 0 {
		return response.Content != "", nil
	}

	return len(response.ToolCalls) > 0, nil
}

// Act 执行工具调用
func (a *ToolCallAgent) Act(ctx context.Context) (string, error) {
	if len(a.ToolCalls) == 0 {
		if a.ToolChoices == "required" {
			return "", fmt.Errorf("tool calls required but none provided")
		}

		// 返回最后一条消息内容
		if len(a.Memory.Messages) > 0 {
			lastMsg := a.Memory.Messages[len(a.Memory.Messages)-1]
			if lastMsg.Content != nil {
				return *lastMsg.Content, nil
			}
		}
		return "No content or commands to execute", nil
	}

	results := make([]string, 0)
	for _, toolCall := range a.ToolCalls {
		result, err := a.ExecuteTool(ctx, toolCall)
		if err != nil {
			logger.Errorf("Tool execution failed: %v", err)
			result = fmt.Sprintf("Error: %v", err)
		} else {
			logger.Infof("🎯 Tool '%s' completed its mission! Result: %s", toolCall.Function.Name, result)
		}

		// 添加工具响应到记忆
		toolMsg := schema.NewToolMessage(result, toolCall.Function.Name, toolCall.ID)
		a.Memory.AddMessage(toolMsg)
		results = append(results, result)

		// 处理特殊工具（如 terminate）
		if a.isSpecialTool(toolCall.Function.Name) {
			if a.shouldFinishExecution(toolCall.Function.Name, result) {
				logger.Infof("🏁 Special tool '%s' has completed the task!", toolCall.Function.Name)
				a.State = schema.AgentStateFINISHED
			}
		}
	}

	return strings.Join(results, "\n\n"), nil
}

// ExecuteTool 执行单个工具调用
func (a *ToolCallAgent) ExecuteTool(ctx context.Context, toolCall schema.ToolCall) (string, error) {
	if toolCall.Function.Name == "" {
		return "Error: Invalid command format", nil
	}

	// 解析参数
	args, err := tool.ParseToolArgs(toolCall.Function.Arguments)
	if err != nil {
		return fmt.Sprintf("Error parsing arguments for %s: Invalid JSON format", toolCall.Function.Name), nil
	}

	// 执行工具
	logger.Infof("🔧 Activating tool: '%s'...", toolCall.Function.Name)
	result, err := a.AvailableTools.Execute(ctx, toolCall.Function.Name, args)
	if err != nil {
		return fmt.Sprintf("⚠️ Tool '%s' encountered a problem: %v", toolCall.Function.Name, err), nil
	}

	if result.Error != "" {
		return fmt.Sprintf("Error: %s", result.Error), nil
	}

	observation := fmt.Sprintf("Observed output of cmd `%s` executed:\n%s", toolCall.Function.Name, result.Output)
	return observation, nil
}

// isSpecialTool 检查是否是特殊工具
func (a *ToolCallAgent) isSpecialTool(name string) bool {
	for _, specialName := range a.SpecialToolNames {
		if name == specialName {
			return true
		}
	}
	return false
}

// shouldFinishExecution 判断是否应该结束执行
func (a *ToolCallAgent) shouldFinishExecution(name string, result string) bool {
	return true // 默认 terminate 工具会结束执行
}

