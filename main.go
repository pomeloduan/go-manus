package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"go-manus/agent"
	"go-manus/logger"
)

func main() {
	// 初始化日志
	logger.Setup("INFO", "DEBUG", "go-manus")

	// 创建 Agent
	manusAgent := agent.NewManus()

	// 创建上下文
	ctx := context.Background()

	// 交互式循环
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Go-Manus - Enter your prompt (or 'exit' to quit):")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		prompt := strings.TrimSpace(scanner.Text())
		if prompt == "" {
			continue
		}

		if strings.ToLower(prompt) == "exit" {
			logger.Info("Goodbye!")
			break
		}

		logger.Warn("Processing your request...")

		// 执行 Agent
		result, err := manusAgent.Run(ctx, prompt)
		if err != nil {
			logger.Errorf("Error: %v", err)
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println(result)
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		logger.Errorf("Error reading input: %v", err)
	}
}

