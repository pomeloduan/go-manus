package flow

import (
	"context"
	"go-manus/agent"
)

// BaseFlow Flow 基类，支持多 Agent 协作
type BaseFlow interface {
	// Execute 执行 Flow
	Execute(ctx context.Context, inputText string) (string, error)

	// GetAgent 获取指定 Agent
	GetAgent(key string) *agent.BaseAgent

	// AddAgent 添加 Agent
	AddAgent(key string, ag *agent.BaseAgent)

	// GetPrimaryAgent 获取主 Agent
	GetPrimaryAgent() *agent.BaseAgent
}

// FlowBase Flow 基础实现
type FlowBase struct {
	agents         map[string]*agent.BaseAgent
	primaryAgentKey string
}

// NewFlowBase 创建 Flow 基础实例
func NewFlowBase(agents map[string]*agent.BaseAgent, primaryKey string) *FlowBase {
	if primaryKey == "" && len(agents) > 0 {
		// 如果没有指定主 Agent，使用第一个
		for k := range agents {
			primaryKey = k
			break
		}
	}

	return &FlowBase{
		agents:         agents,
		primaryAgentKey: primaryKey,
	}
}

// GetAgent 获取指定 Agent
func (f *FlowBase) GetAgent(key string) *agent.BaseAgent {
	return f.agents[key]
}

// AddAgent 添加 Agent
func (f *FlowBase) AddAgent(key string, ag *agent.BaseAgent) {
	if f.agents == nil {
		f.agents = make(map[string]*agent.BaseAgent)
	}
	f.agents[key] = ag
}

// GetPrimaryAgent 获取主 Agent
func (f *FlowBase) GetPrimaryAgent() *agent.BaseAgent {
	if f.primaryAgentKey == "" {
		return nil
	}
	return f.agents[f.primaryAgentKey]
}
