package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type StrReplaceEditor struct {
	fileHistory map[string][]string
}

func NewStrReplaceEditor() *StrReplaceEditor {
	return &StrReplaceEditor{
		fileHistory: make(map[string][]string),
	}
}

func (s *StrReplaceEditor) Name() string {
	return "str_replace_editor"
}

func (s *StrReplaceEditor) Description() string {
	return `Custom editing tool for viewing, creating and editing files
* State is persistent across command calls and discussions with the user
* If path is a file, view displays the result of applying cat -n. If path is a directory, view lists non-hidden files and directories up to 2 levels deep
* The create command cannot be used if the specified path already exists as a file
* If a command generates a long output, it will be truncated and marked with <response clipped>
* The undo_edit command will revert the last edit made to the file at path

Notes for using the str_replace command:
* The old_str parameter should match EXACTLY one or more consecutive lines from the original file. Be mindful of whitespaces!
* If the old_str parameter is not unique in the file, the replacement will not be performed. Make sure to include enough context in old_str to make it unique
* The new_str parameter should contain the edited lines that should replace the old_str`
}

func (s *StrReplaceEditor) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"description": "The commands to run. Allowed options are: view, create, str_replace, insert, undo_edit.",
				"enum":        []string{"view", "create", "str_replace", "insert", "undo_edit"},
				"type":        "string",
			},
			"path": map[string]interface{}{
				"description": "Absolute path to file or directory.",
				"type":        "string",
			},
			"file_text": map[string]interface{}{
				"description": "Required parameter of create command, with the content of the file to be created.",
				"type":        "string",
			},
			"old_str": map[string]interface{}{
				"description": "Required parameter of str_replace command containing the string in path to replace.",
				"type":        "string",
			},
			"new_str": map[string]interface{}{
				"description": "Optional parameter of str_replace command containing the new string (if not given, no string will be added). Required parameter of insert command containing the string to insert.",
				"type":        "string",
			},
			"insert_line": map[string]interface{}{
				"description": "Required parameter of insert command. The new_str will be inserted AFTER the line insert_line of path.",
				"type":        "integer",
			},
			"view_range": map[string]interface{}{
				"description": "Optional parameter of view command when path points to a file. If none is given, the full file is shown. If provided, the file will be shown in the indicated line number range, e.g. [11, 12] will show lines 11 and 12. Indexing at 1 to start. Setting [start_line, -1] shows all lines from start_line to the end of the file.",
				"type":        "array",
				"items": map[string]interface{}{
					"type": "integer",
				},
			},
		},
		"required": []string{"command", "path"},
	}
}

func (s *StrReplaceEditor) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	command, ok := args["command"].(string)
	if !ok {
		return &ToolResult{Error: "command parameter is required"}, nil
	}

	path, ok := args["path"].(string)
	if !ok {
		return &ToolResult{Error: "path parameter is required"}, nil
	}

	// Validate path is absolute
	if !filepath.IsAbs(path) {
		return &ToolResult{Error: fmt.Sprintf("The path %s is not an absolute path", path)}, nil
	}

	switch command {
	case "view":
		return s.view(ctx, path, args)
	case "create":
		return s.create(ctx, path, args)
	case "str_replace":
		return s.strReplace(ctx, path, args)
	case "insert":
		return s.insert(ctx, path, args)
	case "undo_edit":
		return s.undoEdit(ctx, path)
	default:
		return &ToolResult{Error: fmt.Sprintf("Unrecognized command: %s", command)}, nil
	}
}

func (s *StrReplaceEditor) view(ctx context.Context, path string, args map[string]interface{}) (*ToolResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("The path %s does not exist", path)}, nil
	}

	if info.IsDir() {
		return s.viewDirectory(ctx, path)
	}

	// View file
	var viewRange []int
	if vr, ok := args["view_range"].([]interface{}); ok && len(vr) > 0 {
		viewRange = make([]int, len(vr))
		for i, v := range vr {
			if f, ok := v.(float64); ok {
				viewRange[i] = int(f)
			}
		}
	}

	return s.viewFile(ctx, path, viewRange)
}

func (s *StrReplaceEditor) viewDirectory(ctx context.Context, path string) (*ToolResult, error) {
	// List directory contents up to 2 levels deep
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Here's the files and directories up to 2 levels deep in %s, excluding hidden items:\n", path))

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(path, p)
		if err != nil {
			return err
		}

		// Skip hidden files
		if strings.HasPrefix(rel, ".") && rel != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Limit depth to 2 levels
		depth := strings.Count(rel, string(filepath.Separator))
		if depth > 2 {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if rel != "." {
			if info.IsDir() {
				result.WriteString(rel + "/\n")
			} else {
				result.WriteString(rel + "\n")
			}
		}

		return nil
	})

	if err != nil {
		return &ToolResult{Error: err.Error()}, nil
	}

	return &ToolResult{Output: result.String()}, nil
}

func (s *StrReplaceEditor) viewFile(ctx context.Context, path string, viewRange []int) (*ToolResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to read file: %v", err)}, nil
	}

	lines := strings.Split(string(content), "\n")
	initLine := 1

	if len(viewRange) == 2 {
		initLine = viewRange[0]
		finalLine := viewRange[1]

		if initLine < 1 || initLine > len(lines) {
			return &ToolResult{Error: fmt.Sprintf("Invalid view_range: [%d, %d]. First element should be within [1, %d]", initLine, finalLine, len(lines))}, nil
		}

		if finalLine == -1 {
			lines = lines[initLine-1:]
		} else {
			if finalLine > len(lines) {
				return &ToolResult{Error: fmt.Sprintf("Invalid view_range: [%d, %d]. Second element should be <= %d", initLine, finalLine, len(lines))}, nil
			}
			if finalLine < initLine {
				return &ToolResult{Error: fmt.Sprintf("Invalid view_range: [%d, %d]. Second element should be >= first element", initLine, finalLine)}, nil
			}
			lines = lines[initLine-1 : finalLine]
		}
	}

	// Format with line numbers
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Here's the result of running `cat -n` on %s:\n", path))
	for i, line := range lines {
		result.WriteString(fmt.Sprintf("%6d\t%s\n", i+initLine, line))
	}

	output := result.String()
	// Truncate if too long
	const maxLength = 16000
	if len(output) > maxLength {
		output = output[:maxLength] + "<response clipped><NOTE>To save on context only part of this file has been shown to you. You should retry this tool after you have searched inside the file with `grep -n` in order to find the line numbers of what you are looking for.</NOTE>"
	}

	return &ToolResult{Output: output}, nil
}

func (s *StrReplaceEditor) create(ctx context.Context, path string, args map[string]interface{}) (*ToolResult, error) {
	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		return &ToolResult{Error: fmt.Sprintf("File already exists at: %s. Cannot overwrite files using command create.", path)}, nil
	}

	fileText, ok := args["file_text"].(string)
	if !ok {
		return &ToolResult{Error: "file_text parameter is required for create command"}, nil
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to create directory: %v", err)}, nil
	}

	// Write file
	if err := os.WriteFile(path, []byte(fileText), 0644); err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to write file: %v", err)}, nil
	}

	// Save to history
	s.fileHistory[path] = append(s.fileHistory[path], fileText)

	return &ToolResult{Output: fmt.Sprintf("File created successfully at: %s", path)}, nil
}

func (s *StrReplaceEditor) strReplace(ctx context.Context, path string, args map[string]interface{}) (*ToolResult, error) {
	oldStr, ok := args["old_str"].(string)
	if !ok {
		return &ToolResult{Error: "old_str parameter is required for str_replace command"}, nil
	}

	newStr := ""
	if ns, ok := args["new_str"].(string); ok {
		newStr = ns
	}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to read file: %v", err)}, nil
	}

	fileContent := strings.ReplaceAll(string(content), "\t", "    ") // Expand tabs
	oldStr = strings.ReplaceAll(oldStr, "\t", "    ")
	newStr = strings.ReplaceAll(newStr, "\t", "    ")

	// Check occurrences
	occurrences := strings.Count(fileContent, oldStr)
	if occurrences == 0 {
		return &ToolResult{Error: fmt.Sprintf("No replacement was performed, old_str did not appear verbatim in %s.", path)}, nil
	} else if occurrences > 1 {
		// Find line numbers
		lines := make([]int, 0)
		fileLines := strings.Split(fileContent, "\n")
		for i, line := range fileLines {
			if strings.Contains(line, oldStr) {
				lines = append(lines, i+1)
			}
		}
		return &ToolResult{Error: fmt.Sprintf("No replacement was performed. Multiple occurrences of old_str in lines %v. Please ensure it is unique", lines)}, nil
	}

	// Replace
	newFileContent := strings.Replace(fileContent, oldStr, newStr, 1)

	// Write file
	if err := os.WriteFile(path, []byte(newFileContent), 0644); err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to write file: %v", err)}, nil
	}

	// Save to history
	s.fileHistory[path] = append(s.fileHistory[path], fileContent)

	// Create snippet
	replacementLine := strings.Count(strings.Split(fileContent, oldStr)[0], "\n")
	startLine := replacementLine - 4
	if startLine < 0 {
		startLine = 0
	}
	endLine := replacementLine + 4 + strings.Count(newStr, "\n")
	newLines := strings.Split(newFileContent, "\n")
	if endLine >= len(newLines) {
		endLine = len(newLines) - 1
	}
	snippet := strings.Join(newLines[startLine:endLine+1], "\n")

	// Format output with line numbers
	var result strings.Builder
	result.WriteString(fmt.Sprintf("The file %s has been edited. ", path))
	result.WriteString(fmt.Sprintf("Here's the result of running `cat -n` on a snippet of %s:\n", path))
	snippetLines := strings.Split(snippet, "\n")
	for i, line := range snippetLines {
		result.WriteString(fmt.Sprintf("%6d\t%s\n", startLine+i+1, line))
	}
	result.WriteString("Review the changes and make sure they are as expected. Edit the file again if necessary.")

	return &ToolResult{Output: result.String()}, nil
}

func (s *StrReplaceEditor) insert(ctx context.Context, path string, args map[string]interface{}) (*ToolResult, error) {
	insertLine, ok := args["insert_line"].(float64)
	if !ok {
		return &ToolResult{Error: "insert_line parameter is required for insert command"}, nil
	}

	newStr, ok := args["new_str"].(string)
	if !ok {
		return &ToolResult{Error: "new_str parameter is required for insert command"}, nil
	}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to read file: %v", err)}, nil
	}

	fileText := strings.ReplaceAll(string(content), "\t", "    ")
	newStr = strings.ReplaceAll(newStr, "\t", "    ")
	fileLines := strings.Split(fileText, "\n")
	nLines := len(fileLines)

	lineNum := int(insertLine)
	if lineNum < 0 || lineNum > nLines {
		return &ToolResult{Error: fmt.Sprintf("Invalid insert_line parameter: %d. It should be within [0, %d]", lineNum, nLines)}, nil
	}

	// Insert
	newStrLines := strings.Split(newStr, "\n")
	newFileLines := append(fileLines[:lineNum], append(newStrLines, fileLines[lineNum:]...)...)

	// Create snippet
	startLine := lineNum - 4
	if startLine < 0 {
		startLine = 0
	}
	endLine := lineNum + len(newStrLines) + 4
	if endLine >= len(newFileLines) {
		endLine = len(newFileLines) - 1
	}
	snippetLines := newFileLines[startLine : endLine+1]

	// Write file
	newFileText := strings.Join(newFileLines, "\n")
	if err := os.WriteFile(path, []byte(newFileText), 0644); err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to write file: %v", err)}, nil
	}

	// Save to history
	s.fileHistory[path] = append(s.fileHistory[path], fileText)

	// Format output
	var result strings.Builder
	result.WriteString(fmt.Sprintf("The file %s has been edited. ", path))
	result.WriteString("Here's the result of running `cat -n` on a snippet of the edited file:\n")
	for i, line := range snippetLines {
		result.WriteString(fmt.Sprintf("%6d\t%s\n", startLine+i+1, line))
	}
	result.WriteString("Review the changes and make sure they are as expected (correct indentation, no duplicate lines, etc). Edit the file again if necessary.")

	return &ToolResult{Output: result.String()}, nil
}

func (s *StrReplaceEditor) undoEdit(ctx context.Context, path string) (*ToolResult, error) {
	history, exists := s.fileHistory[path]
	if !exists || len(history) == 0 {
		return &ToolResult{Error: fmt.Sprintf("No edit history found for %s.", path)}, nil
	}

	// Get last version
	oldText := history[len(history)-1]
	s.fileHistory[path] = history[:len(history)-1]

	// Write old content
	if err := os.WriteFile(path, []byte(oldText), 0644); err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to write file: %v", err)}, nil
	}

	// Format output
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Last edit to %s undone successfully. ", path))
	result.WriteString(fmt.Sprintf("Here's the result of running `cat -n` on %s:\n", path))
	lines := strings.Split(oldText, "\n")
	for i, line := range lines {
		result.WriteString(fmt.Sprintf("%6d\t%s\n", i+1, line))
	}

	return &ToolResult{Output: result.String()}, nil
}
