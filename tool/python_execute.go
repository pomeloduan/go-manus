package tool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type PythonExecute struct{}

func NewPythonExecute() *PythonExecute {
	return &PythonExecute{}
}

func (p *PythonExecute) Name() string {
	return "python_execute"
}

func (p *PythonExecute) Description() string {
	return "Executes Python code string. Note: Only print outputs are visible, function return values are not captured. Use print statements to see results. Requires Python 3 to be installed on the system."
}

func (p *PythonExecute) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":        "string",
				"description": "The Python code to execute.",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Execution timeout in seconds. Default is 5.",
				"default":     5,
			},
		},
		"required": []string{"code"},
	}
}

func (p *PythonExecute) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	code, ok := args["code"].(string)
	if !ok {
		return &ToolResult{Error: "code parameter is required"}, nil
	}

	timeout := 5
	if t, ok := args["timeout"].(float64); ok {
		timeout = int(t)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "python_*.py")
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to create temp file: %v", err)}, nil
	}
	defer os.Remove(tmpFile.Name())

	// Write code to file
	if _, err := tmpFile.WriteString(code); err != nil {
		tmpFile.Close()
		return &ToolResult{Error: fmt.Sprintf("Failed to write code: %v", err)}, nil
	}
	tmpFile.Close()

	// Try to find Python executable
	pythonCmd := p.findPython()
	if pythonCmd == "" {
		return &ToolResult{Error: "Python 3 is not installed or not found in PATH. Please install Python 3 to use this tool."}, nil
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Execute Python code
	cmd := exec.CommandContext(execCtx, pythonCmd, tmpFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return &ToolResult{
				Error:  fmt.Sprintf("Execution timeout after %d seconds", timeout),
				Output: string(output),
			}, nil
		}
		return &ToolResult{
			Error:  err.Error(),
			Output: string(output),
		}, nil
	}

	return &ToolResult{Output: string(output)}, nil
}

// findPython tries to find Python 3 executable in PATH
func (p *PythonExecute) findPython() string {
	// Try common Python executable names
	candidates := []string{"python3", "python", "py"}
	
	for _, cmd := range candidates {
		if path, err := exec.LookPath(cmd); err == nil {
			// Verify it's Python 3
			verCmd := exec.Command(path, "--version")
			if output, err := verCmd.Output(); err == nil {
				version := string(output)
				// Check if it's Python 3.x
				if len(version) >= 7 && version[:7] == "Python 3" {
					return path
				}
			}
		}
	}
	
	return ""
}
