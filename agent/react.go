package agent

import (
	"context"
)

// ReActAgent ReAct 模式的 Agent
type ReActAgent struct {
	*BaseAgent
}

// NewReActAgent 创建 ReAct Agent
func NewReActAgent(name string) *ReActAgent {
	return &ReActAgent{
		BaseAgent: NewBaseAgent(name),
	}
}

// Think 思考下一步行动（子类实现）
func (a *ReActAgent) Think(ctx context.Context) (bool, error) {
	return false, nil
}

// Act 执行行动（子类实现）
func (a *ReActAgent) Act(ctx context.Context) (string, error) {
	return "", nil
}

// Step 执行单步：思考 + 行动
func (a *ReActAgent) Step(ctx context.Context) (string, error) {
	shouldAct, err := a.Think(ctx)
	if err != nil {
		return "", err
	}

	if !shouldAct {
		return "Thinking complete - no action needed", nil
	}

	return a.Act(ctx)
}

