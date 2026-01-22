package flow

import (
	"context"
	"fmt"
	"strings"

	"go-manus/agent"
	"go-manus/logger"
	"go-manus/schema"
	"go-manus/tool"
)

// PlanningFlow 规划执行流程
type PlanningFlow struct {
	*FlowBase
	planningTool *tool.PlanningTool
	activePlanID string
	currentStepIndex int
	executorKeys []string
}

// NewPlanningFlow 创建 Planning Flow
func NewPlanningFlow(agents map[string]*agent.BaseAgent, primaryKey string) *PlanningFlow {
	// 确定可用的 executor keys
	executorKeys := make([]string, 0, len(agents))
	for key := range agents {
		executorKeys = append(executorKeys, key)
	}

	return &PlanningFlow{
		FlowBase:     NewFlowBase(agents, primaryKey),
		planningTool: tool.NewPlanningTool(),
		executorKeys: executorKeys,
	}
}

// Execute 执行规划流程
func (p *PlanningFlow) Execute(ctx context.Context, inputText string) (string, error) {
	logger.Infof("Starting PlanningFlow execution for: %s", inputText)

	// 创建初始计划
	planID := fmt.Sprintf("plan_%d", p.planningTool.GetActivePlan() != nil)
	if err := p.createInitialPlan(ctx, inputText, planID); err != nil {
		return "", fmt.Errorf("failed to create plan: %w", err)
	}

	p.activePlanID = planID

	var result strings.Builder
	for {
		// 获取当前步骤
		stepIndex, stepInfo := p.getCurrentStepInfo()
		if stepIndex == nil {
			// 没有更多步骤，完成
			result.WriteString(p.finalizePlan())
			break
		}

		// 执行当前步骤
		stepType := stepInfo["type"]
		executor := p.getExecutor(stepType)
		if executor == nil {
			result.WriteString(fmt.Sprintf("Step %d: No executor available for type %s\n", *stepIndex, stepType))
			break
		}

		stepResult, err := p.executeStep(ctx, executor, stepInfo)
		if err != nil {
			result.WriteString(fmt.Sprintf("Step %d failed: %v\n", *stepIndex, err))
			break
		}

		result.WriteString(fmt.Sprintf("Step %d: %s\n", *stepIndex, stepResult))

		// 检查 Agent 是否完成
		if toolCallAgent, ok := executor.(*agent.ToolCallAgent); ok {
			if toolCallAgent.State == schema.AgentStateFINISHED {
				break
			}
		}
	}

	return result.String(), nil
}

// createInitialPlan 创建初始计划
func (p *PlanningFlow) createInitialPlan(ctx context.Context, request string, planID string) error {
	// 生成计划步骤（简化实现，实际应该调用 LLM）
	// 这里使用固定的步骤模板
	steps := []interface{}{
		"Analyze the request",
		"Plan the solution",
		"Execute the plan",
		"Verify the results",
	}

	// 创建计划
	args := map[string]interface{}{
		"command": "create",
		"plan_id": planID,
		"title":   fmt.Sprintf("Plan for: %s", request),
		"steps":   steps,
	}

	_, err := p.planningTool.Execute(ctx, args)
	if err != nil {
		return err
	}

	// 设置活动计划
	args = map[string]interface{}{
		"command": "set_active",
		"plan_id": planID,
	}
	_, err = p.planningTool.Execute(ctx, args)
	return err
}

// getCurrentStepInfo 获取当前步骤信息
func (p *PlanningFlow) getCurrentStepInfo() (*int, map[string]interface{}) {
	plan := p.planningTool.GetActivePlan()
	if plan == nil {
		return nil, nil
	}

	// 查找下一个未完成的步骤
	for i, step := range plan.Steps {
		if step.Status == tool.PlanStepNotStarted || step.Status == tool.PlanStepInProgress {
			idx := i
			return &idx, map[string]interface{}{
				"index":       i,
				"description": step.Description,
				"type":        "default", // 可以根据描述判断类型
			}
		}
	}

	return nil, nil
}

// getExecutor 根据步骤类型获取执行器
func (p *PlanningFlow) getExecutor(stepType interface{}) *agent.BaseAgent {
	// 简化实现：根据类型选择 Agent
	// 实际应该根据步骤描述智能选择
	if stepType == "data_analysis" {
		if ag := p.GetAgent("data_analysis"); ag != nil {
			return ag
		}
	}

	// 默认使用主 Agent
	return p.GetPrimaryAgent()
}

// executeStep 执行步骤
func (p *PlanningFlow) executeStep(ctx context.Context, executor *agent.BaseAgent, stepInfo map[string]interface{}) (string, error) {
	stepIndex, ok := stepInfo["index"].(int)
	if !ok {
		return "", fmt.Errorf("invalid step index")
	}

	description, ok := stepInfo["description"].(string)
	if !ok {
		return "", fmt.Errorf("invalid step description")
	}

	// 标记步骤为进行中
	args := map[string]interface{}{
		"command":    "mark_step",
		"step_index": stepIndex,
		"status":      "in_progress",
	}
	p.planningTool.Execute(ctx, args)

	// 执行步骤
	result, err := executor.Run(ctx, description)
	if err != nil {
		// 标记为失败
		args = map[string]interface{}{
			"command":    "mark_step",
			"step_index": stepIndex,
			"status":      "blocked",
			"result":      fmt.Sprintf("Error: %v", err),
		}
		p.planningTool.Execute(ctx, args)
		return "", err
	}

	// 标记为完成
	args = map[string]interface{}{
		"command":    "mark_step",
		"step_index": stepIndex,
		"status":      "completed",
		"result":      result,
	}
	p.planningTool.Execute(ctx, args)

	return result, nil
}

// finalizePlan 完成计划
func (p *PlanningFlow) finalizePlan() string {
	plan := p.planningTool.GetActivePlan()
	if plan == nil {
		return "Plan execution completed."
	}

	completed := 0
	for _, step := range plan.Steps {
		if step.Status == tool.PlanStepCompleted {
			completed++
		}
	}

	return fmt.Sprintf("Plan execution completed. %d/%d steps completed.", completed, len(plan.Steps))
}
