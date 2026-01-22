package tool

import (
	"bufio"
	"context"
	"fmt"
	"os"
)

type AskHuman struct{}

func NewAskHuman() *AskHuman {
	return &AskHuman{}
}

func (a *AskHuman) Name() string {
	return "ask_human"
}

func (a *AskHuman) Description() string {
	return "Use this tool to ask human for help. When you need clarification, additional information, or user confirmation, use this tool to interact with the user."
}

func (a *AskHuman) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"inquire": map[string]interface{}{
				"type":        "string",
				"description": "The question you want to ask human.",
			},
		},
		"required": []string{"inquire"},
	}
}

func (a *AskHuman) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	inquire, ok := args["inquire"].(string)
	if !ok {
		return &ToolResult{Error: "inquire parameter is required"}, nil
	}

	// Print question and wait for user input
	fmt.Printf("Bot: %s\n\nYou: ", inquire)

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return &ToolResult{Error: "Failed to read user input"}, nil
	}

	response := scanner.Text()
	return &ToolResult{Output: response}, nil
}
