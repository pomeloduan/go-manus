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

	// 设置提示词（来自 Python 版本的 app/prompt/manus.py）
	// 注意：需要从 config 获取 workspace_root，这里先使用默认值
	workspaceRoot := "workspace" // TODO: 从 config 获取
	manus.SystemPrompt = fmt.Sprintf("You are OpenManus, an all-capable AI assistant, aimed at solving any task presented by the user. You have various tools at your disposal that you can call upon to efficiently complete complex requests. Whether it's programming, information retrieval, file processing, web browsing, or human interaction (only for extreme cases), you can handle it all.\nThe initial directory is: %s", workspaceRoot)

	manus.NextStepPrompt = `You can interact with the computer using various tools:

FileSaver: Save files locally, such as txt, html, json, etc.

StrReplaceEditor: View, create, and edit files. Supports commands: view, create, str_replace, insert, undo_edit.

Bash: Execute bash commands in the terminal. Supports interactive sessions, background tasks, and process management.

BrowserUseTool: Open, browse, and use web browsers. If you open a local HTML file, you must provide the absolute path to the file.

WebSearch: Unified web search supporting multiple engines (google, baidu, bing, duckduckgo). Automatically falls back to other engines if one fails.

WebCrawler: Extract clean, AI-ready content from web pages. Perfect for content analysis and research.

PythonExecute: Execute Python code. Note: Requires Python 3 to be installed.

Planning: Create and manage plans for complex tasks. Track progress and manage multi-step workflows.

CreateChatCompletion: Format the final response in a structured way (text, json, markdown).

ComputerUseTool: Control mouse, keyboard, and take screenshots. Automate desktop applications.

VisualizationPrepare: Prepare data for visualization. Generates CSV and JSON metadata files.

DataVisualization: Visualize statistical charts with JSON info. Generate charts in PNG or HTML format.

AskHuman: Ask the user for clarification, additional information, or confirmation when needed.

Based on user needs, proactively select the most appropriate tool or combination of tools. For complex tasks, you can break down the problem and use different tools step by step to solve it. After using each tool, clearly explain the execution results and suggest the next steps.

If you want to stop the interaction at any point, use the terminate tool/function call.`

	// 添加工具集合
	manus.AvailableTools = tool.NewToolCollection(
		tool.NewGoogleSearch(),
		tool.NewBaiduSearch(),
		tool.NewBingSearch(),
		tool.NewDuckDuckGoSearch(),
		tool.NewWebSearch(),
		tool.NewBrowserUse(),
		tool.NewFileSaver(),
		tool.NewStrReplaceEditor(),
		tool.NewBash(),
		tool.NewAskHuman(),
		tool.NewPythonExecute(),
		tool.NewWebCrawler(),
		tool.NewPlanningTool(),
		tool.NewCreateChatCompletion(),
		tool.NewComputerUseTool(),
		tool.NewVisualizationPrepare(),
		tool.NewDataVisualization(),
		tool.NewTerminate(),
	)

	manus.Description = "A versatile agent that can solve various tasks using multiple tools"

	return manus
}

