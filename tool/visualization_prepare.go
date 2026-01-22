package tool

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// VisualizationPrepare 可视化准备工具
type VisualizationPrepare struct {
	outputDir string
}

func NewVisualizationPrepare() *VisualizationPrepare {
	return &VisualizationPrepare{
		outputDir: "workspace/charts",
	}
}

func (v *VisualizationPrepare) Name() string {
	return "visualization_prepare"
}

func (v *VisualizationPrepare) Description() string {
	return `Prepare data for visualization. Generates CSV and JSON metadata files for data visualization.
This tool processes data and creates the necessary files for chart generation.`
}

func (v *VisualizationPrepare) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"data": map[string]interface{}{
				"type":        "string",
				"description": "The data to prepare (can be CSV content or file path)",
			},
			"chart_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of chart to create",
				"enum":        []string{"line", "bar", "pie", "scatter"},
				"default":     "line",
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Title for the chart",
			},
			"x_label": map[string]interface{}{
				"type":        "string",
				"description": "Label for X axis",
			},
			"y_label": map[string]interface{}{
				"type":        "string",
				"description": "Label for Y axis",
			},
		},
		"required": []string{"data"},
	}
}

func (v *VisualizationPrepare) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	data, ok := args["data"].(string)
	if !ok {
		return &ToolResult{Error: "data parameter is required"}, nil
	}

	chartType := "line"
	if ct, ok := args["chart_type"].(string); ok && ct != "" {
		chartType = ct
	}

	title := "Chart"
	if t, ok := args["title"].(string); ok && t != "" {
		title = t
	}

	xLabel := "X"
	if xl, ok := args["x_label"].(string); ok && xl != "" {
		xLabel = xl
	}

	yLabel := "Y"
	if yl, ok := args["y_label"].(string); ok && yl != "" {
		yLabel = yl
	}

	// 确保输出目录存在
	os.MkdirAll(v.outputDir, 0755)

	// 处理数据（可能是文件路径或 CSV 内容）
	var csvPath string
	if strings.HasSuffix(data, ".csv") || strings.Contains(data, "\n") {
		// 如果是 CSV 内容或文件路径
		if strings.Contains(data, "\n") {
			// 是 CSV 内容，保存到文件
			csvPath = filepath.Join(v.outputDir, fmt.Sprintf("%s.csv", strings.ReplaceAll(title, " ", "_")))
			if err := os.WriteFile(csvPath, []byte(data), 0644); err != nil {
				return &ToolResult{Error: fmt.Sprintf("Failed to write CSV: %v", err)}, nil
			}
		} else {
			// 是文件路径
			csvPath = data
			if !filepath.IsAbs(csvPath) {
				csvPath = filepath.Join("workspace", csvPath)
			}
		}
	} else {
		// 尝试解析为 CSV
		csvPath = filepath.Join(v.outputDir, fmt.Sprintf("%s.csv", strings.ReplaceAll(title, " ", "_")))
		if err := os.WriteFile(csvPath, []byte(data), 0644); err != nil {
			return &ToolResult{Error: fmt.Sprintf("Failed to write CSV: %v", err)}, nil
		}
	}

	// 生成 JSON 元数据
	jsonPath := filepath.Join(v.outputDir, fmt.Sprintf("%s.json", strings.ReplaceAll(title, " ", "_")))
	metadata := map[string]interface{}{
		"csvFilePath": csvPath,
		"chartType":   chartType,
		"title":        title,
		"xLabel":       xLabel,
		"yLabel":       yLabel,
	}

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to marshal JSON: %v", err)}, nil
	}

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to write JSON: %v", err)}, nil
	}

	// 验证 CSV 文件
	if err := v.validateCSV(csvPath); err != nil {
		return &ToolResult{Error: fmt.Sprintf("CSV validation failed: %v", err)}, nil
	}

	output := fmt.Sprintf("Data prepared successfully!\nCSV: %s\nJSON: %s\n\nUse data_visualization tool with json_path='%s' to generate the chart.", csvPath, jsonPath, jsonPath)
	return &ToolResult{Output: output}, nil
}

func (v *VisualizationPrepare) validateCSV(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.ReadAll()
	return err
}
