package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// MCPClientTool MCP 客户端工具
type MCPClientTool struct {
	name         string
	description  string
	parameters   map[string]interface{}
	serverID     string
	originalName string
	// session 用于与 MCP 服务器通信
	// 这里简化实现，实际需要 JSON-RPC 客户端
}

func NewMCPClientTool(name, description string, parameters map[string]interface{}, serverID, originalName string) *MCPClientTool {
	return &MCPClientTool{
		name:         name,
		description:  description,
		parameters:   parameters,
		serverID:     serverID,
		originalName: originalName,
	}
}

func (m *MCPClientTool) Name() string {
	return m.name
}

func (m *MCPClientTool) Description() string {
	return m.description
}

func (m *MCPClientTool) Parameters() map[string]interface{} {
	return m.parameters
}

func (m *MCPClientTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// 这里应该通过 JSON-RPC 调用 MCP 服务器
	// 简化实现：返回错误提示需要实现 JSON-RPC 客户端
	return &ToolResult{
		Error: fmt.Sprintf("MCP tool execution requires JSON-RPC client implementation. Tool: %s (original: %s) on server: %s", m.name, m.originalName, m.serverID),
	}, nil
}

// MCPClients MCP 客户端集合
type MCPClients struct {
	sessions map[string]interface{} // MCP session，实际应该是 JSON-RPC 客户端
	toolMap  map[string]*MCPClientTool
	tools    []*MCPClientTool
	mu       sync.RWMutex
}

func NewMCPClients() *MCPClients {
	return &MCPClients{
		sessions: make(map[string]interface{}),
		toolMap:  make(map[string]*MCPClientTool),
		tools:    make([]*MCPClientTool, 0),
	}
}

// ConnectSSE 通过 SSE 连接 MCP 服务器
func (m *MCPClients) ConnectSSE(ctx context.Context, serverURL, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 这里应该实现 SSE 连接和 JSON-RPC 客户端
	// 简化实现：只记录连接信息
	m.sessions[serverID] = map[string]string{
		"type": "sse",
		"url":  serverURL,
	}

	// 模拟工具发现（实际应该通过 list_tools 调用）
	// 这里返回一个示例工具
	tool := NewMCPClientTool(
		fmt.Sprintf("mcp_%s_example", serverID),
		"Example MCP tool",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
		serverID,
		"example",
	)

	m.toolMap[tool.Name()] = tool
	m.tools = append(m.tools, tool)

	return nil
}

// ConnectStdio 通过 stdio 连接 MCP 服务器
func (m *MCPClients) ConnectStdio(ctx context.Context, command string, args []string, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 这里应该启动子进程并通过 stdio 进行 JSON-RPC 通信
	// 简化实现：只记录连接信息
	m.sessions[serverID] = map[string]interface{}{
		"type":    "stdio",
		"command": command,
		"args":    args,
	}

	// 模拟工具发现
	tool := NewMCPClientTool(
		fmt.Sprintf("mcp_%s_example", serverID),
		"Example MCP tool",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
		serverID,
		"example",
	)

	m.toolMap[tool.Name()] = tool
	m.tools = append(m.tools, tool)

	return nil
}

// ListTools 列出所有可用工具
func (m *MCPClients) ListTools(ctx context.Context) ([]*MCPClientTool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.tools, nil
}

// Disconnect 断开连接
func (m *MCPClients) Disconnect(serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, serverID)

	// 移除该服务器的工具
	newTools := make([]*MCPClientTool, 0)
	for _, tool := range m.tools {
		if tool.serverID != serverID {
			newTools = append(newTools, tool)
		} else {
			delete(m.toolMap, tool.Name())
		}
	}
	m.tools = newTools

	return nil
}

// GetTool 获取工具
func (m *MCPClients) GetTool(name string) (Tool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tool, ok := m.toolMap[name]
	return tool, ok
}

// AddTool 添加工具（实现 ToolCollection 接口）
func (m *MCPClients) AddTool(t Tool) {
	if mcpTool, ok := t.(*MCPClientTool); ok {
		m.mu.Lock()
		defer m.mu.Unlock()

		m.toolMap[mcpTool.Name()] = mcpTool
		m.tools = append(m.tools, mcpTool)
	}
}

// Execute 执行工具（实现 ToolCollection 接口）
func (m *MCPClients) Execute(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	tool, ok := m.GetTool(name)
	if !ok {
		return &ToolResult{Error: fmt.Sprintf("Tool %s not found", name)}, nil
	}
	return tool.Execute(ctx, args)
}

// ToOpenAITools 转换为 OpenAI 工具格式
func (m *MCPClients) ToOpenAITools() []interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]interface{}, 0, len(m.tools))
	for _, t := range m.tools {
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
