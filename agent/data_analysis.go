package agent

import (
	"fmt"
	"go-manus/tool"
)

// DataAnalysis 数据分析 Agent
type DataAnalysis struct {
	*ToolCallAgent
}

// NewDataAnalysis 创建数据分析 Agent
func NewDataAnalysis() *DataAnalysis {
	agent := &DataAnalysis{
		ToolCallAgent: NewToolCallAgent("Data_Analysis"),
	}

	// 设置提示词（来自 Python 版本的 app/prompt/visualization.py）
	// 注意：需要从 config 获取 workspace_root，这里先使用默认值
	workspaceRoot := "workspace" // TODO: 从 config 获取
	agent.SystemPrompt = fmt.Sprintf(`You are an AI agent designed to data analysis / visualization task. You have various tools at your disposal that you can call upon to efficiently complete complex requests.
# Note:
1. The workspace directory is: %s; Read / write file in workspace
2. Generate analysis conclusion report in the end
3. Use FileSaver to save analysis results, StrReplaceEditor to view/edit data files, VisualizationPrepare and DataVisualization for creating charts`, workspaceRoot)

	agent.NextStepPrompt = `Based on user needs, break down the problem and use different tools step by step to solve it.
# Note
1. Each step select the most appropriate tool proactively (ONLY ONE).
2. After using each tool, clearly explain the execution results and suggest the next steps.
3. When observation with Error, review and fix it.

Available tools:
- FileSaver: Save analysis results, reports, and processed data
- StrReplaceEditor: View and edit data files
- VisualizationPrepare: Prepare data for visualization
- DataVisualization: Generate charts and visualizations`

	// 配置工具（数据分析 Agent 使用 FileSaver, StrReplaceEditor, VisualizationPrepare, DataVisualization）
	agent.AvailableTools = tool.NewToolCollection(
		tool.NewFileSaver(),
		tool.NewStrReplaceEditor(),
		tool.NewVisualizationPrepare(),
		tool.NewDataVisualization(),
		tool.NewTerminate(),
	)

	agent.SpecialToolNames = []string{"terminate"}
	agent.Description = "An analytical agent that utilizes data visualization tools to solve diverse data analysis tasks"
	agent.MaxSteps = 20
	agent.MaxObserve = 15000

	return agent
}
