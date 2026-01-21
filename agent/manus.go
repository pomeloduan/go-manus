package agent

import (
	"go-manus/tool"
)

// Manus 通用 Agent，包含多种工具
type Manus struct {
	*ToolCallAgent
}

// NewManus 创建 Manus Agent
func NewManus() *Manus {
	manus := &Manus{
		ToolCallAgent: NewToolCallAgent("Manus"),
	}

	// 设置提示词
	manus.SystemPrompt = "You are OpenManus, an all-capable AI assistant, aimed at solving any task presented by the user. You have various tools at your disposal that you can call upon to efficiently complete complex requests. Whether it's programming, information retrieval, file processing, or web browsing, you can handle it all."

	manus.NextStepPrompt = `You can interact with the computer by saving important content and information files through FileSaver, opening browsers with BrowserUseTool, and retrieving information using GoogleSearch.

FileSaver: Save files locally, such as txt, html, json, etc.

BrowserUseTool: Open, browse, and use web browsers. If you open a local HTML file, you must provide the absolute path to the file.

GoogleSearch: Perform web information retrieval

Based on user needs, proactively select the most appropriate tool or combination of tools. For complex tasks, you can break down the problem and use different tools step by step to solve it. After using each tool, clearly explain the execution results and suggest the next steps.`

	// 添加工具（不包含 PythonExecute，避免依赖 Python 环境）
	manus.AvailableTools = tool.NewToolCollection(
		tool.NewGoogleSearch(),
		tool.NewBrowserUse(),
		tool.NewFileSaver(),
		tool.NewTerminate(),
	)

	manus.Description = "A versatile agent that can solve various tasks using multiple tools"

	return manus
}

