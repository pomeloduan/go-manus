package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go-manus/llm"
	"go-manus/logger"
	"go-manus/schema"
)

// BaseAgent Agent 基础结构
type BaseAgent struct {
	Name        string
	Description string

	SystemPrompt    string
	NextStepPrompt string

	LLM    *llm.Client
	Memory *schema.Memory
	State  schema.AgentState

	MaxSteps     int
	CurrentStep  int
	DuplicateThreshold int

	mu sync.RWMutex
}

// NewBaseAgent 创建基础 Agent
func NewBaseAgent(name string) *BaseAgent {
	return &BaseAgent{
		Name:        name,
		LLM:         llm.NewClient("default"),
		Memory:      schema.NewMemory(),
		State:       schema.AgentStateIDLE,
		MaxSteps:    10,
		DuplicateThreshold: 2,
	}
}

// UpdateMemory 更新记忆
func (a *BaseAgent) UpdateMemory(role schema.MessageRole, content string, toolCallID ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var msg schema.Message
	switch role {
	case schema.RoleUser:
		msg = schema.NewUserMessage(content)
	case schema.RoleSystem:
		msg = schema.NewSystemMessage(content)
	case schema.RoleAssistant:
		msg = schema.NewAssistantMessage(content)
	case schema.RoleTool:
		if len(toolCallID) >= 2 {
			msg = schema.NewToolMessage(content, toolCallID[0], toolCallID[1])
		}
	default:
		logger.Errorf("Unsupported message role: %s", role)
		return
	}

	a.Memory.AddMessage(msg)
}

// Run 执行 Agent 主循环
func (a *BaseAgent) Run(ctx context.Context, request string) (string, error) {
	if a.State != schema.AgentStateIDLE {
		return "", fmt.Errorf("cannot run agent from state: %s", a.State)
	}

	if request != "" {
		a.UpdateMemory(schema.RoleUser, request)
	}

	results := make([]string, 0)
	a.State = schema.AgentStateRUNNING

	for a.CurrentStep < a.MaxSteps && a.State != schema.AgentStateFINISHED {
		a.CurrentStep++
		logger.Infof("Executing step %d/%d", a.CurrentStep, a.MaxSteps)

		stepResult, err := a.Step(ctx)
		if err != nil {
			logger.Errorf("Step %d failed: %v", a.CurrentStep, err)
			a.State = schema.AgentStateERROR
			return "", err
		}

		// 检查是否卡住
		if a.IsStuck() {
			a.HandleStuckState()
		}

		results = append(results, fmt.Sprintf("Step %d: %s", a.CurrentStep, stepResult))
	}

	if a.CurrentStep >= a.MaxSteps {
		results = append(results, fmt.Sprintf("Terminated: Reached max steps (%d)", a.MaxSteps))
	}

	if len(results) == 0 {
		return "No steps executed", nil
	}

	return strings.Join(results, "\n"), nil
}

// Step 执行单步（子类实现）
func (a *BaseAgent) Step(ctx context.Context) (string, error) {
	return "", fmt.Errorf("Step method must be implemented by subclass")
}

// IsStuck 检查是否卡住
func (a *BaseAgent) IsStuck() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(a.Memory.Messages) < 2 {
		return false
	}

	lastMsg := a.Memory.Messages[len(a.Memory.Messages)-1]
	if lastMsg.Content == nil {
		return false
	}

	// 检查是否有重复内容
	duplicateCount := 0
	for i := len(a.Memory.Messages) - 2; i >= 0; i-- {
		msg := a.Memory.Messages[i]
		if msg.Role == schema.RoleAssistant && msg.Content != nil && *msg.Content == *lastMsg.Content {
			duplicateCount++
			if duplicateCount >= a.DuplicateThreshold {
				return true
			}
		}
	}

	return false
}

// HandleStuckState 处理卡住状态
func (a *BaseAgent) HandleStuckState() {
	stuckPrompt := "Observed duplicate responses. Consider new strategies and avoid repeating ineffective paths already attempted."
	a.NextStepPrompt = stuckPrompt + "\n" + a.NextStepPrompt
	logger.Warningf("Agent detected stuck state. Added prompt: %s", stuckPrompt)
}

// GetMessages 获取消息列表
func (a *BaseAgent) GetMessages() []schema.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Memory.Messages
}

