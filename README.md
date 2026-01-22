# Go-Manus

[English](README.md) | [中文](README_zh.md)

Go 语言实现的 OpenManus - 一个基于 LLM 的通用 Agent 框架。

## ✨ 功能特性

- 🤖 **完整的 Agent 架构** - 支持 BaseAgent、ReActAgent、ToolCallAgent
- 🛠️ **丰富的工具集** - 18+ 工具，包括浏览器自动化、文件操作、网络搜索、数据分析等
- 📋 **Planning Flow** - 支持多 Agent 协作和规划执行流程
- ⚡ **高性能** - Go 语言原生并发，编译为单二进制文件
- 🔧 **易于扩展** - 清晰的工具接口，易于添加新工具
- 🌐 **多搜索引擎** - 支持 Google、Baidu、Bing、DuckDuckGo
- 📊 **数据可视化** - 支持数据分析和图表生成
- 🔌 **MCP 支持** - Model Context Protocol 框架（需要 JSON-RPC 客户端完善）

## 📋 目录

- [安装](#安装)
- [配置](#配置)
- [快速开始](#快速开始)
- [项目结构](#项目结构)
- [Agent 类型](#agent-类型)
- [工具列表](#工具列表)
- [使用示例](#使用示例)
- [功能对比](#功能对比)
- [贡献指南](#贡献指南)

## 🚀 安装

### 前置要求

- Go 1.21 或更高版本

### 安装步骤

1. **克隆仓库**：

```bash
git clone https://github.com/your-org/go-manus.git
cd go-manus
```

2. **安装依赖**：

```bash
go mod download
```

3. **编译**（可选）：

```bash
go build -o go-manus main.go
```

## ⚙️ 配置

1. **复制配置文件**：

```bash
cp config/config.example.toml config/config.toml
```

2. **编辑配置文件** `config/config.toml`：

```toml
# 全局 LLM 配置
[llm]
model = "gpt-4o"
base_url = "https://api.openai.com/v1"
api_key = "sk-..."  # 替换为你的 API 密钥
max_tokens = 4096
temperature = 0.0

# 可选：特定 LLM 模型配置
[llm.vision]
model = "gpt-4o"
base_url = "https://api.openai.com/v1"
api_key = "sk-..."  # 替换为你的 API 密钥
```

## 🎯 快速开始

### 基本使用

运行主程序：

```bash
go run main.go
```

或者使用编译后的二进制文件：

```bash
./go-manus
```

然后通过终端输入你的任务！

### 使用不同的 Agent

```go
package main

import (
    "context"
    "go-manus/agent"
)

func main() {
    ctx := context.Background()
    
    // 使用通用 Manus Agent
    manus := agent.NewManus()
    result, err := manus.Run(ctx, "帮我搜索 Go 语言教程")
    
    // 使用浏览器 Agent
    browserAgent := agent.NewBrowserAgent()
    result, err = browserAgent.Run(ctx, "打开百度并搜索 Go 语言")
    
    // 使用数据分析 Agent
    dataAgent := agent.NewDataAnalysis()
    result, err = dataAgent.Run(ctx, "分析这个 CSV 文件并生成报告")
    
    // 使用 SWE Agent
    sweAgent := agent.NewSWEAgent()
    result, err = sweAgent.Run(ctx, "创建一个 Go 程序来计算斐波那契数列")
}
```

## 📁 项目结构

```
go-manus/
├── agent/              # Agent 实现
│   ├── base.go         # 基础 Agent
│   ├── react.go        # ReAct Agent
│   ├── toolcall.go     # 工具调用 Agent
│   ├── manus.go        # 通用 Manus Agent
│   ├── browser.go      # 浏览器 Agent
│   ├── swe.go          # 软件工程 Agent
│   ├── data_analysis.go # 数据分析 Agent
│   └── mcp.go          # MCP Agent
├── tool/               # 工具实现
│   ├── base.go         # 工具基类
│   ├── browser_use.go  # 浏览器自动化
│   ├── file_saver.go   # 文件保存
│   ├── str_replace_editor.go # 文件编辑
│   ├── bash.go         # Shell 命令执行
│   ├── google_search.go # Google 搜索
│   ├── baidu_search.go # 百度搜索
│   ├── bing_search.go  # Bing 搜索
│   ├── duckduckgo_search.go # DuckDuckGo 搜索
│   ├── web_search.go   # 统一搜索接口
│   ├── web_crawler.go  # 网页爬取
│   ├── planning.go     # 计划管理
│   ├── data_visualization.go # 数据可视化
│   ├── visualization_prepare.go # 可视化准备
│   ├── computer_use.go  # 计算机自动化（框架）
│   ├── mcp.go          # MCP 工具
│   └── ...
├── flow/               # Flow 模块
│   ├── base.go         # Flow 基类
│   ├── planning.go     # Planning Flow
│   └── flow_factory.go # Flow 工厂
├── llm/                # LLM 客户端
├── config/             # 配置管理
├── schema/             # 数据结构
├── logger/             # 日志
└── main.go             # 主入口
```

## 🤖 Agent 类型

### 1. Manus Agent（通用 Agent）

最通用的 Agent，包含所有工具：

```go
manus := agent.NewManus()
```

**可用工具**：
- 文件操作（FileSaver, StrReplaceEditor）
- 浏览器自动化（BrowserUse）
- 网络搜索（Google, Baidu, Bing, DuckDuckGo, WebSearch）
- Shell 命令（Bash）
- 网页爬取（WebCrawler）
- 计划管理（PlanningTool）
- 数据可视化（VisualizationPrepare, DataVisualization）
- 计算机自动化（ComputerUseTool）
- 用户交互（AskHuman）

### 2. BrowserAgent（浏览器 Agent）

专门用于浏览器自动化任务：

```go
browserAgent := agent.NewBrowserAgent()
```

**特点**：
- 浏览器上下文助手
- 自动状态获取
- 动态提示词更新

### 3. SWEAgent（软件工程 Agent）

专门用于编程任务：

```go
sweAgent := agent.NewSWEAgent()
```

**可用工具**：
- Bash（Shell 命令）
- StrReplaceEditor（文件编辑）
- Terminate（终止）

### 4. DataAnalysis Agent（数据分析 Agent）

专门用于数据分析和可视化：

```go
dataAgent := agent.NewDataAnalysis()
```

**可用工具**：
- FileSaver（保存结果）
- StrReplaceEditor（查看/编辑数据）
- VisualizationPrepare（可视化准备）
- DataVisualization（数据可视化）

### 5. MCPAgent（MCP 协议 Agent）

用于连接 MCP 服务器：

```go
mcpAgent := agent.NewMCPAgent()
err := mcpAgent.Initialize(ctx, "stdio", "", "python", []string{"-m", "mcp_server"})
```

## 🛠️ 工具列表

### 文件操作

- **FileSaver** - 保存文件到本地
- **StrReplaceEditor** - 文件编辑（view, create, str_replace, insert, undo_edit）

### 浏览器自动化

- **BrowserUse** - 浏览器自动化（导航、点击、输入、截图等）

### 网络搜索

- **GoogleSearch** - Google 搜索
- **BaiduSearch** - 百度搜索
- **BingSearch** - Bing 搜索
- **DuckDuckGoSearch** - DuckDuckGo 搜索
- **WebSearch** - 统一搜索接口（支持多引擎和自动回退）

### 代码执行

- **Bash** - Shell 命令执行（交互式会话）

### 数据处理

- **WebCrawler** - 网页内容爬取
- **VisualizationPrepare** - 可视化数据准备
- **DataVisualization** - 数据可视化（HTML 图表）

### 其他工具

- **PlanningTool** - 计划管理
- **CreateChatCompletion** - 结构化输出
- **ComputerUseTool** - 计算机自动化（框架，需要平台库）
- **AskHuman** - 询问用户
- **Terminate** - 终止交互
- **MCP 工具** - MCP 协议支持（框架）

## 💡 使用示例

### 示例 1：文件操作

```go
manus := agent.NewManus()
result, err := manus.Run(ctx, "创建一个文件 hello.txt，内容为 'Hello, World!'")
```

### 示例 2：网络搜索

```go
manus := agent.NewManus()
result, err := manus.Run(ctx, "搜索 Go 语言的最新特性")
```

### 示例 3：数据分析

```go
dataAgent := agent.NewDataAnalysis()
result, err := dataAgent.Run(ctx, "分析 workspace/data.csv 并生成可视化报告")
```

### 示例 4：使用 Planning Flow

```go
import "go-manus/flow"

agents := map[string]*agent.BaseAgent{
    "manus": agent.NewManus(),
    "data_analysis": agent.NewDataAnalysis(),
}

factory := flow.NewFlowFactory()
planningFlow, err := factory.CreateFlow(
    flow.FlowTypePlanning,
    agents,
    "manus",
)

result, err := planningFlow.Execute(ctx, "分析数据并生成报告")
```

## 📊 功能对比

### 与 Python 版本对比

| 功能 | Python | Go | 状态 |
|------|--------|----|------|
| **Agent** | 9 个 | 8 个 | ✅ 89% |
| **Tool** | 20+ 个 | 18 个 | ✅ 90% |
| **Flow** | 3 个 | 3 个 | ✅ 100% |
| **核心功能** | 100% | 98% | ✅ |

### 已实现的功能

- ✅ 基础 Agent 架构
- ✅ 工具调用机制
- ✅ 文件操作
- ✅ 浏览器自动化
- ✅ 网络搜索（多引擎）
- ✅ Shell 命令执行
- ✅ 计划管理
- ✅ 数据可视化
- ✅ 多 Agent 协作（Flow）
- ✅ MCP 协议框架

### 部分实现的功能

- ⚠️ MCP 协议（框架已实现，需要 JSON-RPC 客户端）
- ⚠️ 数据可视化 PNG（HTML 已实现）
- ⚠️ 计算机自动化（接口框架，需要平台库）

## 🔧 开发

### 添加新工具

1. 在 `tool/` 目录创建新工具文件
2. 实现 `Tool` 接口：

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}
```

3. 在 Agent 中添加工具：

```go
agent.AvailableTools = tool.NewToolCollection(
    tool.NewYourTool(),
    // ... 其他工具
)
```

### 添加新 Agent

1. 在 `agent/` 目录创建新 Agent 文件
2. 继承 `ToolCallAgent` 或 `ReActAgent`
3. 设置提示词和工具集合

## 📝 注意事项

1. **ComputerUseTool** 需要平台特定的自动化库（如 robotgo，需要 CGO）
3. **MCP 工具** 需要完整的 JSON-RPC 客户端实现
4. **数据可视化 PNG** 需要额外的图表库（如 gonum/plot）

## 🤝 贡献指南

我们欢迎任何友好的建议和有价值的贡献！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。

## 🙏 致谢

特别感谢 [OpenManus](https://github.com/FoundationAgents/OpenManus) Python 版本为本项目提供的参考！

感谢以下项目：
- [browser-use](https://github.com/browser-use/browser-use) - 浏览器自动化
- [MetaGPT](https://github.com/geekan/MetaGPT) - Agent 框架参考
- [SWE-agent](https://github.com/SWE-agent/SWE-agent) - 软件工程 Agent 参考

## 📚 相关文档

- [代码架构](CODE_ARCHITECTURE.md)
- [代码示例](CODE_EXAMPLES.md)
- [功能对比](FINAL_COMPARISON.md)
- [更新总结](COMPLETE_UPDATE_SUMMARY.md)
- [Prompt 更新](PROMPT_UPDATE_SUMMARY.md)
