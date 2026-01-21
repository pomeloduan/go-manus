package tool

import (
	"context"
)

type Terminate struct{}

func NewTerminate() *Terminate {
	return &Terminate{}
}

func (t *Terminate) Name() string {
	return "terminate"
}

func (t *Terminate) Description() string {
	return "Terminate the interaction when the request is met OR if the assistant cannot proceed further with the task."
}

func (t *Terminate) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "The finish status of the interaction.",
				"enum":        []string{"success", "failure"},
			},
		},
		"required": []string{"status"},
	}
}

func (t *Terminate) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	status, _ := args["status"].(string)
	if status == "" {
		status = "success"
	}
	return &ToolResult{
		Output: "The interaction has been completed with status: " + status,
	}, nil
}

