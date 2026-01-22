package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PlanStepStatus 计划步骤状态
type PlanStepStatus string

const (
	PlanStepNotStarted PlanStepStatus = "not_started"
	PlanStepInProgress PlanStepStatus = "in_progress"
	PlanStepCompleted  PlanStepStatus = "completed"
	PlanStepBlocked    PlanStepStatus = "blocked"
)

// Plan 计划结构
type Plan struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Steps     []PlanStep             `json:"steps"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PlanStep 计划步骤
type PlanStep struct {
	Description string         `json:"description"`
	Status      PlanStepStatus `json:"status"`
	Result      string         `json:"result,omitempty"`
	Error       string         `json:"error,omitempty"`
}

// PlanningTool 计划管理工具
type PlanningTool struct {
	plans      map[string]*Plan
	activePlan string
	mu         sync.RWMutex
	storageDir string
}

func NewPlanningTool() *PlanningTool {
	pt := &PlanningTool{
		plans:      make(map[string]*Plan),
		storageDir: "workspace/plans",
	}

	// 确保存储目录存在
	os.MkdirAll(pt.storageDir, 0755)

	// 加载已存在的计划
	pt.loadPlans()

	return pt
}

func (p *PlanningTool) Name() string {
	return "planning"
}

func (p *PlanningTool) Description() string {
	return `A planning tool that allows the agent to create and manage plans for solving complex tasks.
The tool provides functionality for creating plans, updating plan steps, and tracking progress.`
}

func (p *PlanningTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"description": "The command to execute. Available commands: create, update, list, get, set_active, mark_step, delete.",
				"enum": []string{
					"create",
					"update",
					"list",
					"get",
					"set_active",
					"mark_step",
					"delete",
				},
				"type": "string",
			},
			"plan_id": map[string]interface{}{
				"description": "Unique identifier for the plan. Required for create, update, set_active, and delete commands. Optional for get and mark_step (uses active plan if not specified).",
				"type":        "string",
			},
			"title": map[string]interface{}{
				"description": "Title for the plan. Required for create command, optional for update command.",
				"type":        "string",
			},
			"steps": map[string]interface{}{
				"description": "List of plan steps. Required for create command, optional for update command.",
				"type":        "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"step_index": map[string]interface{}{
				"description": "Index of the step to mark (0-based). Required for mark_step command.",
				"type":        "integer",
			},
			"status": map[string]interface{}{
				"description": "Status to set for the step. Required for mark_step command.",
				"enum": []string{
					"not_started",
					"in_progress",
					"completed",
					"blocked",
				},
				"type": "string",
			},
			"result": map[string]interface{}{
				"description": "Result or error message for the step. Optional for mark_step command.",
				"type":        "string",
			},
		},
		"required": []string{"command"},
	}
}

func (p *PlanningTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	command, ok := args["command"].(string)
	if !ok {
		return &ToolResult{Error: "command parameter is required"}, nil
	}

	switch command {
	case "create":
		return p.createPlan(ctx, args)
	case "update":
		return p.updatePlan(ctx, args)
	case "list":
		return p.listPlans(ctx)
	case "get":
		return p.getPlan(ctx, args)
	case "set_active":
		return p.setActivePlan(ctx, args)
	case "mark_step":
		return p.markStep(ctx, args)
	case "delete":
		return p.deletePlan(ctx, args)
	default:
		return &ToolResult{Error: fmt.Sprintf("Unknown command: %s", command)}, nil
	}
}

func (p *PlanningTool) createPlan(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	planID, ok := args["plan_id"].(string)
	if !ok || planID == "" {
		return &ToolResult{Error: "plan_id is required for create command"}, nil
	}

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return &ToolResult{Error: "title is required for create command"}, nil
	}

	stepsInterface, ok := args["steps"].([]interface{})
	if !ok || len(stepsInterface) == 0 {
		return &ToolResult{Error: "steps is required for create command"}, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查计划是否已存在
	if _, exists := p.plans[planID]; exists {
		return &ToolResult{Error: fmt.Sprintf("Plan with ID %s already exists", planID)}, nil
	}

	// 创建步骤
	steps := make([]PlanStep, len(stepsInterface))
	for i, stepDesc := range stepsInterface {
		steps[i] = PlanStep{
			Description: stepDesc.(string),
			Status:      PlanStepNotStarted,
		}
	}

	plan := &Plan{
		ID:        planID,
		Title:     title,
		Steps:     steps,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	p.plans[planID] = plan
	p.savePlan(plan)

	return &ToolResult{
		Output: fmt.Sprintf("Plan '%s' created successfully with %d steps", title, len(steps)),
	}, nil
}

func (p *PlanningTool) updatePlan(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	planID, ok := args["plan_id"].(string)
	if !ok || planID == "" {
		return &ToolResult{Error: "plan_id is required for update command"}, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	plan, exists := p.plans[planID]
	if !exists {
		return &ToolResult{Error: fmt.Sprintf("Plan with ID %s not found", planID)}, nil
	}

	// 更新标题
	if title, ok := args["title"].(string); ok && title != "" {
		plan.Title = title
	}

	// 更新步骤
	if stepsInterface, ok := args["steps"].([]interface{}); ok && len(stepsInterface) > 0 {
		steps := make([]PlanStep, len(stepsInterface))
		for i, stepDesc := range stepsInterface {
			steps[i] = PlanStep{
				Description: stepDesc.(string),
				Status:      PlanStepNotStarted,
			}
		}
		plan.Steps = steps
	}

	plan.UpdatedAt = time.Now()
	p.savePlan(plan)

	return &ToolResult{Output: fmt.Sprintf("Plan '%s' updated successfully", planID)}, nil
}

func (p *PlanningTool) listPlans(ctx context.Context) (*ToolResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.plans) == 0 {
		return &ToolResult{Output: "No plans found"}, nil
	}

	var output string
	output = "Available plans:\n"
	for id, plan := range p.plans {
		status := "inactive"
		if id == p.activePlan {
			status = "active"
		}
		output += fmt.Sprintf("- %s (%s): %s [%d steps, %s]\n",
			id, status, plan.Title, len(plan.Steps), plan.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	return &ToolResult{Output: output}, nil
}

func (p *PlanningTool) getPlan(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	planID, _ := args["plan_id"].(string)
	if planID == "" {
		planID = p.activePlan
	}

	if planID == "" {
		return &ToolResult{Error: "No plan_id provided and no active plan set"}, nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	plan, exists := p.plans[planID]
	if !exists {
		return &ToolResult{Error: fmt.Sprintf("Plan with ID %s not found", planID)}, nil
	}

	// 格式化输出
	var output string
	output = fmt.Sprintf("Plan: %s\n", plan.Title)
	output += fmt.Sprintf("ID: %s\n", plan.ID)
	output += fmt.Sprintf("Created: %s\n", plan.CreatedAt.Format("2006-01-02 15:04:05"))
	output += fmt.Sprintf("Updated: %s\n", plan.UpdatedAt.Format("2006-01-02 15:04:05"))
	output += fmt.Sprintf("Steps (%d):\n", len(plan.Steps))

	for i, step := range plan.Steps {
		statusMark := p.getStatusMark(step.Status)
		output += fmt.Sprintf("  %d. %s %s\n", i+1, statusMark, step.Description)
		if step.Result != "" {
			output += fmt.Sprintf("     Result: %s\n", step.Result)
		}
		if step.Error != "" {
			output += fmt.Sprintf("     Error: %s\n", step.Error)
		}
	}

	return &ToolResult{Output: output}, nil
}

func (p *PlanningTool) setActivePlan(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	planID, ok := args["plan_id"].(string)
	if !ok || planID == "" {
		return &ToolResult{Error: "plan_id is required for set_active command"}, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.plans[planID]; !exists {
		return &ToolResult{Error: fmt.Sprintf("Plan with ID %s not found", planID)}, nil
	}

	p.activePlan = planID
	return &ToolResult{Output: fmt.Sprintf("Plan '%s' set as active", planID)}, nil
}

func (p *PlanningTool) markStep(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	planID, _ := args["plan_id"].(string)
	if planID == "" {
		planID = p.activePlan
	}

	if planID == "" {
		return &ToolResult{Error: "No plan_id provided and no active plan set"}, nil
	}

	stepIndex, ok := args["step_index"].(float64)
	if !ok {
		return &ToolResult{Error: "step_index is required for mark_step command"}, nil
	}

	statusStr, ok := args["status"].(string)
	if !ok {
		return &ToolResult{Error: "status is required for mark_step command"}, nil
	}

	status := PlanStepStatus(statusStr)
	if status != PlanStepNotStarted && status != PlanStepInProgress &&
		status != PlanStepCompleted && status != PlanStepBlocked {
		return &ToolResult{Error: fmt.Sprintf("Invalid status: %s", statusStr)}, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	plan, exists := p.plans[planID]
	if !exists {
		return &ToolResult{Error: fmt.Sprintf("Plan with ID %s not found", planID)}, nil
	}

	idx := int(stepIndex)
	if idx < 0 || idx >= len(plan.Steps) {
		return &ToolResult{Error: fmt.Sprintf("Invalid step_index: %d (plan has %d steps)", idx, len(plan.Steps))}, nil
	}

	plan.Steps[idx].Status = status

	if result, ok := args["result"].(string); ok {
		plan.Steps[idx].Result = result
	}

	plan.UpdatedAt = time.Now()
	p.savePlan(plan)

	return &ToolResult{
		Output: fmt.Sprintf("Step %d marked as %s", idx+1, status),
	}, nil
}

func (p *PlanningTool) deletePlan(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	planID, ok := args["plan_id"].(string)
	if !ok || planID == "" {
		return &ToolResult{Error: "plan_id is required for delete command"}, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.plans[planID]; !exists {
		return &ToolResult{Error: fmt.Sprintf("Plan with ID %s not found", planID)}, nil
	}

	delete(p.plans, planID)

	if p.activePlan == planID {
		p.activePlan = ""
	}

	// 删除文件
	planFile := filepath.Join(p.storageDir, planID+".json")
	os.Remove(planFile)

	return &ToolResult{Output: fmt.Sprintf("Plan '%s' deleted successfully", planID)}, nil
}

func (p *PlanningTool) getStatusMark(status PlanStepStatus) string {
	marks := map[PlanStepStatus]string{
		PlanStepCompleted:  "[✓]",
		PlanStepInProgress: "[→]",
		PlanStepBlocked:    "[!]",
		PlanStepNotStarted: "[ ]",
	}
	if mark, ok := marks[status]; ok {
		return mark
	}
	return "[ ]"
}

func (p *PlanningTool) savePlan(plan *Plan) error {
	planFile := filepath.Join(p.storageDir, plan.ID+".json")
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(planFile, data, 0644)
}

func (p *PlanningTool) loadPlans() {
	files, err := os.ReadDir(p.storageDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() || !filepath.Ext(file.Name()) == ".json" {
			continue
		}

		planFile := filepath.Join(p.storageDir, file.Name())
		data, err := os.ReadFile(planFile)
		if err != nil {
			continue
		}

		var plan Plan
		if err := json.Unmarshal(data, &plan); err != nil {
			continue
		}

		p.plans[plan.ID] = &plan
	}
}

// GetActivePlan 获取当前活动计划
func (p *PlanningTool) GetActivePlan() *Plan {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.activePlan == "" {
		return nil
	}

	return p.plans[p.activePlan]
}

// GetPlan 获取指定计划
func (p *PlanningTool) GetPlan(planID string) *Plan {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.plans[planID]
}
