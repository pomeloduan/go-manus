package tool

import (
	"context"
	"encoding/json"
)

// ToolResult 工具执行结果
type ToolResult struct {
	Output string
	Error  string
	System string
}

// IsSuccess 检查是否成功
func (r *ToolResult) IsSuccess() bool {
	return r.Error == ""
}

// String 返回字符串表示
func (r *ToolResult) String() string {
	if r.Error != "" {
		return "Error: " + r.Error
	}
	return r.Output
}

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

// ToolCollection 工具集合
type ToolCollection struct {
	tools map[string]Tool
}

// NewToolCollection 创建工具集合
func NewToolCollection(tools ...Tool) *ToolCollection {
	tc := &ToolCollection{
		tools: make(map[string]Tool),
	}
	for _, t := range tools {
		tc.AddTool(t)
	}
	return tc
}

// AddTool 添加工具
func (tc *ToolCollection) AddTool(t Tool) {
	tc.tools[t.Name()] = t
}

// GetTool 获取工具
func (tc *ToolCollection) GetTool(name string) (Tool, bool) {
	t, ok := tc.tools[name]
	return t, ok
}

// Execute 执行工具
func (tc *ToolCollection) Execute(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	t, ok := tc.GetTool(name)
	if !ok {
		return &ToolResult{
			Error: "Tool " + name + " is invalid",
		}, nil
	}

	return t.Execute(ctx, args)
}

// ToOpenAITools 转换为 OpenAI 工具格式
func (tc *ToolCollection) ToOpenAITools() []interface{} {
	tools := make([]interface{}, 0, len(tc.tools))
	for _, t := range tc.tools {
		tool := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        t.Name(),
				"description": t.Description(),
				"parameters":  t.Parameters(),
			},
		}
		tools = append(tools, tool)
	}
	return tools
}

// ParseToolArgs 解析工具参数
func ParseToolArgs(argsJSON string) (map[string]interface{}, error) {
	var args map[string]interface{}
	if argsJSON == "" {
		return make(map[string]interface{}), nil
	}
	err := json.Unmarshal([]byte(argsJSON), &args)
	return args, err
}

