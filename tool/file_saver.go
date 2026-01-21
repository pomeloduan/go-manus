package tool

import (
	"context"
	"os"
	"path/filepath"
)

type FileSaver struct{}

func NewFileSaver() *FileSaver {
	return &FileSaver{}
}

func (f *FileSaver) Name() string {
	return "file_saver"
}

func (f *FileSaver) Description() string {
	return "Save content to a local file at a specified path. Use this tool when you need to save text, code, or generated content to a file on the local filesystem."
}

func (f *FileSaver) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"content": map[string]interface{}{
				"type":        "string",
				"description": "(required) The content to save to the file.",
			},
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "(required) The path where the file should be saved, including filename and extension.",
			},
			"mode": map[string]interface{}{
				"type":        "string",
				"description": "(optional) The file opening mode. Default is 'w' for write. Use 'a' for append.",
				"enum":        []string{"w", "a"},
				"default":     "w",
			},
		},
		"required": []string{"content", "file_path"},
	}
}

func (f *FileSaver) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	content, ok := args["content"].(string)
	if !ok {
		return &ToolResult{Error: "content parameter is required"}, nil
	}

	filePath, ok := args["file_path"].(string)
	if !ok {
		return &ToolResult{Error: "file_path parameter is required"}, nil
	}

	mode := "w"
	if m, ok := args["mode"].(string); ok && m != "" {
		mode = m
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return &ToolResult{Error: "Failed to create directory: " + err.Error()}, nil
		}
	}

	// 写入文件
	var flag int
	if mode == "a" {
		flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flag = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	}

	file, err := os.OpenFile(filePath, flag, 0644)
	if err != nil {
		return &ToolResult{Error: "Failed to open file: " + err.Error()}, nil
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return &ToolResult{Error: "Failed to write file: " + err.Error()}, nil
	}

	return &ToolResult{
		Output: "Content successfully saved to " + filePath,
	}, nil
}

