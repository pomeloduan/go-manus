# Go-Manus

Go 语言实现的 OpenManus - 一个基于 LLM 的通用 Agent 框架。

## 功能特性

- 🤖 基于 Tool Calling 的 Agent 架构
- 🛠️ 丰富的工具集（浏览器自动化、文件操作、网络搜索等）
- 📋 支持 Planning Flow 模式
- ⚡ 高性能并发执行
- 🔧 易于扩展的工具系统

## 安装

```bash
go mod download
```

## 配置

复制 `config/config.example.toml` 为 `config/config.toml` 并配置你的 API 密钥：

```toml
[llm]
model = "gpt-4o"
base_url = "https://api.openai.com/v1"
api_key = "sk-..."
max_tokens = 4096
temperature = 0.0
```

## 运行

```bash
go run main.go
```

## 项目结构

```
go-manus/
├── agent/           # Agent 实现
├── config/          # 配置管理（代码和配置文件）
├── llm/             # LLM 客户端
├── logger/          # 日志
├── schema/          # 数据结构
├── tool/            # 工具实现
└── main.go          # 主入口

```

