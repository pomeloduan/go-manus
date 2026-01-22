package flow

import (
	"fmt"
	"go-manus/agent"
)

// FlowType Flow 类型
type FlowType string

const (
	FlowTypePlanning FlowType = "planning"
)

// FlowFactory Flow 工厂
type FlowFactory struct{}

// NewFlowFactory 创建 Flow 工厂
func NewFlowFactory() *FlowFactory {
	return &FlowFactory{}
}

// CreateFlow 创建 Flow
func (f *FlowFactory) CreateFlow(flowType FlowType, agents map[string]*agent.BaseAgent, primaryKey string) (BaseFlow, error) {
	switch flowType {
	case FlowTypePlanning:
		return NewPlanningFlow(agents, primaryKey), nil
	default:
		return nil, fmt.Errorf("unknown flow type: %s", flowType)
	}
}

// CreateFlowFromAgents 从 Agent 列表创建 Flow
func (f *FlowFactory) CreateFlowFromAgents(flowType FlowType, agentsList []*agent.BaseAgent, primaryKey string) (BaseFlow, error) {
	agents := make(map[string]*agent.BaseAgent)
	for i, ag := range agentsList {
		key := fmt.Sprintf("agent_%d", i)
		if primaryKey == "" && i == 0 {
			primaryKey = key
		}
		agents[key] = ag
	}

	if primaryKey == "" && len(agents) > 0 {
		// 使用第一个作为主 Agent
		for k := range agents {
			primaryKey = k
			break
		}
	}

	return f.CreateFlow(flowType, agents, primaryKey)
}
