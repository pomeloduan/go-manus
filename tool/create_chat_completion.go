package tool

import (
	"context"
	"encoding/json"
	"fmt"
)

type CreateChatCompletion struct {
	responseType string // "string" or "object"
}

func NewCreateChatCompletion() *CreateChatCompletion {
	return &CreateChatCompletion{
		responseType: "string",
	}
}

func (c *CreateChatCompletion) Name() string {
	return "create_chat_completion"
}

func (c *CreateChatCompletion) Description() string {
	return "Creates a structured completion with specified output formatting. Use this tool to format the final response to the user in a structured way."
}

func (c *CreateChatCompletion) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"response": map[string]interface{}{
				"type":        "string",
				"description": "The response text that should be delivered to the user.",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "(optional) Output format. Can be 'text', 'json', 'markdown'. Default is 'text'.",
				"enum":        []string{"text", "json", "markdown"},
				"default":     "text",
			},
		},
		"required": []string{"response"},
	}
}

func (c *CreateChatCompletion) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	response, ok := args["response"].(string)
	if !ok {
		return &ToolResult{Error: "response parameter is required"}, nil
	}

	format := "text"
	if f, ok := args["format"].(string); ok && f != "" {
		format = f
	}

	// Format the response based on format type
	var output string
	switch format {
	case "json":
		// Try to parse as JSON and pretty print
		var jsonData interface{}
		if err := json.Unmarshal([]byte(response), &jsonData); err == nil {
			prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
			if err == nil {
				output = string(prettyJSON)
			} else {
				output = response
			}
		} else {
			output = response
		}
	case "markdown":
		output = response // Markdown is already text
	case "text":
		fallthrough
	default:
		output = response
	}

	return &ToolResult{Output: output}, nil
}
