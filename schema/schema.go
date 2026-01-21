package schema

// AgentState 表示 Agent 的执行状态
type AgentState string

const (
	AgentStateIDLE     AgentState = "IDLE"
	AgentStateRUNNING  AgentState = "RUNNING"
	AgentStateFINISHED AgentState = "FINISHED"
	AgentStateERROR    AgentState = "ERROR"
)

// MessageRole 消息角色
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// Function 表示函数调用
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCall 表示工具调用
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Message 表示对话消息
type Message struct {
	Role         MessageRole `json:"role"`
	Content      *string     `json:"content,omitempty"`
	ToolCalls    []ToolCall  `json:"tool_calls,omitempty"`
	Name         *string     `json:"name,omitempty"`
	ToolCallID   *string     `json:"tool_call_id,omitempty"`
}

// NewUserMessage 创建用户消息
func NewUserMessage(content string) Message {
	return Message{
		Role:    RoleUser,
		Content: &content,
	}
}

// NewSystemMessage 创建系统消息
func NewSystemMessage(content string) Message {
	return Message{
		Role:    RoleSystem,
		Content: &content,
	}
}

// NewAssistantMessage 创建助手消息
func NewAssistantMessage(content string) Message {
	return Message{
		Role:    RoleAssistant,
		Content: &content,
	}
}

// NewToolMessage 创建工具消息
func NewToolMessage(content string, name string, toolCallID string) Message {
	return Message{
		Role:       RoleTool,
		Content:    &content,
		Name:       &name,
		ToolCallID: &toolCallID,
	}
}

// NewMessageFromToolCalls 从工具调用创建消息
func NewMessageFromToolCalls(content string, toolCalls []ToolCall) Message {
	msg := Message{
		Role:      RoleAssistant,
		ToolCalls: toolCalls,
	}
	if content != "" {
		msg.Content = &content
	}
	return msg
}

// Memory 表示 Agent 的记忆存储
type Memory struct {
	Messages   []Message `json:"messages"`
	MaxMessages int      `json:"max_messages"`
}

// NewMemory 创建新的记忆
func NewMemory() *Memory {
	return &Memory{
		Messages:    make([]Message, 0),
		MaxMessages: 100,
	}
}

// AddMessage 添加消息
func (m *Memory) AddMessage(msg Message) {
	m.Messages = append(m.Messages, msg)
	if len(m.Messages) > m.MaxMessages {
		m.Messages = m.Messages[len(m.Messages)-m.MaxMessages:]
	}
}

// AddMessages 添加多条消息
func (m *Memory) AddMessages(msgs []Message) {
	m.Messages = append(m.Messages, msgs...)
	if len(m.Messages) > m.MaxMessages {
		m.Messages = m.Messages[len(m.Messages)-m.MaxMessages:]
	}
}

// Clear 清空消息
func (m *Memory) Clear() {
	m.Messages = make([]Message, 0)
}

// GetRecentMessages 获取最近 N 条消息
func (m *Memory) GetRecentMessages(n int) []Message {
	if n > len(m.Messages) {
		n = len(m.Messages)
	}
	start := len(m.Messages) - n
	if start < 0 {
		start = 0
	}
	return m.Messages[start:]
}
