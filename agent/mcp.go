package agent

import (
	"context"
	"fmt"

	"go-manus/logger"
	"go-manus/schema"
	"go-manus/tool"
)

// MCPAgent MCP 协议 Agent
type MCPAgent struct {
	*ToolCallAgent
	mcpClients        *tool.MCPClients
	connectionType    string // "stdio" or "sse"
	toolSchemas       map[string]map[string]interface{}
	refreshInterval   int
	connectedServers  map[string]string
}

// NewMCPAgent 创建 MCP Agent
func NewMCPAgent() *MCPAgent {
	agent := &MCPAgent{
		ToolCallAgent:     NewToolCallAgent("mcp_agent"),
		mcpClients:        tool.NewMCPClients(),
		connectionType:    "stdio",
		toolSchemas:       make(map[string]map[string]interface{}),
		refreshInterval:   5,
		connectedServers: make(map[string]string),
	}

	// 设置提示词（来自 Python 版本的 app/prompt/mcp.py）
	agent.SystemPrompt = `You are an AI assistant with access to a Model Context Protocol (MCP) server.
You can use the tools provided by the MCP server to complete tasks.
The MCP server will dynamically expose tools that you can use - always check the available tools first.

When using an MCP tool:
1. Choose the appropriate tool based on your task requirements
2. Provide properly formatted arguments as required by the tool
3. Observe the results and use them to determine next steps
4. Tools may change during operation - new tools might appear or existing ones might disappear

Follow these guidelines:
- Call tools with valid parameters as documented in their schemas
- Handle errors gracefully by understanding what went wrong and trying again with corrected parameters
- For multimedia responses (like images), you'll receive a description of the content
- Complete user requests step by step, using the most appropriate tools
- If multiple tools need to be called in sequence, make one call at a time and wait for results

Remember to clearly explain your reasoning and actions to the user.`

	agent.NextStepPrompt = `Based on the current state and available tools, what should be done next?
Think step by step about the problem and identify which MCP tool would be most helpful for the current stage.
If you've already made progress, consider what additional information you need or what actions would move you closer to completing the task.`
	agent.Description = "An agent that connects to an MCP server and uses its tools"
	agent.MaxSteps = 20
	agent.SpecialToolNames = []string{"terminate"}

	return agent
}

// Initialize 初始化 MCP 连接
func (m *MCPAgent) Initialize(ctx context.Context, connectionType string, serverURL string, command string, args []string) error {
	if connectionType != "" {
		m.connectionType = connectionType
	}

	serverID := fmt.Sprintf("server_%d", len(m.connectedServers))

	var err error
	if m.connectionType == "sse" {
		if serverURL == "" {
			return fmt.Errorf("server URL is required for SSE connection")
		}
		err = m.mcpClients.ConnectSSE(ctx, serverURL, serverID)
	} else if m.connectionType == "stdio" {
		if command == "" {
			return fmt.Errorf("command is required for stdio connection")
		}
		err = m.mcpClients.ConnectStdio(ctx, command, args, serverID)
	} else {
		return fmt.Errorf("unsupported connection type: %s", m.connectionType)
	}

	if err != nil {
		return err
	}

	m.connectedServers[serverID] = serverURL
	if serverURL == "" {
		m.connectedServers[serverID] = command
	}

	// 更新可用工具
	m.AvailableTools = tool.NewToolCollection()
	// 添加 MCP 工具
	tools, err := m.mcpClients.ListTools(ctx)
	if err == nil {
		for _, t := range tools {
			m.AvailableTools.AddTool(t)
		}
	}

	// 添加 Terminate 工具
	m.AvailableTools.AddTool(tool.NewTerminate())

	// 存储工具模式
	m.refreshTools(ctx)

	// 添加系统消息
	toolNames := make([]string, 0, len(tools))
	for _, t := range tools {
		toolNames = append(toolNames, t.Name())
	}
	toolsInfo := fmt.Sprintf("Available MCP tools: %v", toolNames)

	agentMessage := schema.NewSystemMessage(fmt.Sprintf("%s\n\n%s", m.SystemPrompt, toolsInfo))
	m.Memory.AddMessage(agentMessage)

	return nil
}

// refreshTools 刷新工具列表
func (m *MCPAgent) refreshTools(ctx context.Context) {
	tools, err := m.mcpClients.ListTools(ctx)
	if err != nil {
		logger.Warningf("Failed to refresh MCP tools: %v", err)
		return
	}

	// 更新工具模式
	for _, t := range tools {
		m.toolSchemas[t.Name()] = t.Parameters()
	}
}

// Think 思考下一步行动
func (m *MCPAgent) Think(ctx context.Context) (bool, error) {
	// 检查 MCP 会话和工具可用性
	if len(m.mcpClients.Sessions()) == 0 || len(m.mcpClients.Tools()) == 0 {
		logger.Info("MCP service is no longer available, ending interaction")
		m.State = schema.AgentStateFINISHED
		return false, nil
	}

	// 定期刷新工具
	if m.CurrentStep%m.refreshInterval == 0 {
		m.refreshTools(ctx)
		// 如果所有工具都被移除，表示服务器关闭
		if len(m.mcpClients.Tools()) == 0 {
			logger.Info("MCP service has shut down, ending interaction")
			m.State = schema.AgentStateFINISHED
			return false, nil
		}
	}

	// 使用父类的 Think 方法
	return m.ToolCallAgent.Think(ctx)
}

// Cleanup 清理 MCP 连接
func (m *MCPAgent) Cleanup(ctx context.Context) error {
	for serverID := range m.connectedServers {
		if err := m.mcpClients.Disconnect(serverID); err != nil {
			logger.Warningf("Error disconnecting from MCP server %s: %v", serverID, err)
		}
	}
	logger.Info("MCP connection closed")
	return nil
}
