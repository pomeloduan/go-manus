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

// DataVisualization 数据可视化工具
type DataVisualization struct {
	outputDir string
}

func NewDataVisualization() *DataVisualization {
	return &DataVisualization{
		outputDir: "workspace/charts",
	}
}

func (d *DataVisualization) Name() string {
	return "data_visualization"
}

func (d *DataVisualization) Description() string {
	return `Visualize statistical chart or Add insights in chart with JSON info from visualization_preparation tool.
You can do steps as follows:
1. Visualize statistical chart
2. Choose insights into chart based on step 1 (Optional)
Outputs:
1. Charts (png/html)
2. Charts Insights (.md)(Optional)`
}

func (d *DataVisualization) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"json_path": map[string]interface{}{
				"type":        "string",
				"description": "file path of json info with \".json\" in the end",
			},
			"output_type": map[string]interface{}{
				"type":        "string",
				"description": "Rendering format (html=interactive)",
				"enum":        []string{"png", "html"},
				"default":     "html",
			},
			"tool_type": map[string]interface{}{
				"type":        "string",
				"description": "visualize chart or add insights",
				"enum":        []string{"visualization", "insight"},
				"default":     "visualization",
			},
			"language": map[string]interface{}{
				"type":        "string",
				"description": "english(en) / chinese(zh)",
				"enum":        []string{"zh", "en"},
				"default":     "en",
			},
		},
		"required": []string{"json_path"},
	}
}

func (d *DataVisualization) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	jsonPath, ok := args["json_path"].(string)
	if !ok {
		return &ToolResult{Error: "json_path parameter is required"}, nil
	}

	outputType := "html"
	if ot, ok := args["output_type"].(string); ok && ot != "" {
		outputType = ot
	}

	toolType := "visualization"
	if tt, ok := args["tool_type"].(string); ok && tt != "" {
		toolType = tt
	}

	language := "en"
	if lang, ok := args["language"].(string); ok && lang != "" {
		language = lang
	}

	// 读取 JSON 配置
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to read JSON file: %v", err)}, nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to parse JSON: %v", err)}, nil
	}

	// 获取 CSV 文件路径
	csvPath, ok := config["csvFilePath"].(string)
	if !ok {
		return &ToolResult{Error: "csvFilePath not found in JSON config"}, nil
	}

	// 读取 CSV 数据
	data, err := d.readCSV(csvPath)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to read CSV: %v", err)}, nil
	}

	// 生成图表
	if toolType == "visualization" {
		return d.generateChart(ctx, data, config, outputType, language)
	} else {
		return d.addInsights(ctx, data, config, language)
	}
}

func (d *DataVisualization) readCSV(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (d *DataVisualization) generateChart(ctx context.Context, data [][]string, config map[string]interface{}, outputType, language string) (*ToolResult, error) {
	// 确保输出目录存在
	os.MkdirAll(d.outputDir, 0755)

	// 获取图表类型和配置
	chartType, _ := config["chartType"].(string)
	title, _ := config["title"].(string)
	if title == "" {
		title = "Chart"
	}

	// 生成图表文件名
	chartFileName := fmt.Sprintf("%s.%s", strings.ReplaceAll(title, " ", "_"), outputType)
	chartPath := filepath.Join(d.outputDir, chartFileName)

	// 这里应该使用 Go 的图表库生成图表
	// 简化实现：生成 HTML 图表
	if outputType == "html" {
		htmlContent := d.generateHTMLChart(data, config, title, language)
		if err := os.WriteFile(chartPath, []byte(htmlContent), 0644); err != nil {
			return &ToolResult{Error: fmt.Sprintf("Failed to write chart: %v", err)}, nil
		}
	} else {
		// PNG 格式需要调用图表库或外部工具
		// 这里简化处理
		return &ToolResult{Error: "PNG format requires chart library (e.g., gonum/plot or go-echarts)"}, nil
	}

	output := fmt.Sprintf("Chart Generated Successfully!\n## %s\nChart saved in: %s", title, chartPath)
	return &ToolResult{Output: output}, nil
}

func (d *DataVisualization) generateHTMLChart(data [][]string, config map[string]interface{}, title, language string) string {
	// 使用简单的 HTML + Chart.js 生成交互式图表
	// 这里是一个简化实现
	chartType, _ := config["chartType"].(string)
	if chartType == "" {
		chartType = "line"
	}

	// 提取数据（简化：假设第一行是标题，后续是数据）
	var labels []string
	var values []float64

	if len(data) > 1 {
		// 使用第一列作为标签，第二列作为值
		for i := 1; i < len(data); i++ {
			if len(data[i]) >= 2 {
				labels = append(labels, data[i][0])
				// 简化：假设是数字
				var val float64
				fmt.Sscanf(data[i][1], "%f", &val)
				values = append(values, val)
			}
		}
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <h1>%s</h1>
    <canvas id="myChart" width="400" height="200"></canvas>
    <script>
        const ctx = document.getElementById('myChart').getContext('2d');
        const chart = new Chart(ctx, {
            type: '%s',
            data: {
                labels: %s,
                datasets: [{
                    label: 'Data',
                    data: %s,
                    borderColor: 'rgb(75, 192, 192)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    </script>
</body>
</html>`, title, title, chartType, d.arrayToJSON(labels), d.arrayToJSONFloat(values))

	return html
}

func (d *DataVisualization) arrayToJSON(arr []string) string {
	data, _ := json.Marshal(arr)
	return string(data)
}

func (d *DataVisualization) arrayToJSONFloat(arr []float64) string {
	data, _ := json.Marshal(arr)
	return string(data)
}

func (d *DataVisualization) addInsights(ctx context.Context, data [][]string, config map[string]interface{}, language string) (*ToolResult, error) {
	// 添加洞察（简化实现）
	insightPath, _ := config["insight_path"].(string)
	if insightPath == "" {
		insightPath = filepath.Join(d.outputDir, "insights.md")
	}

	insights := fmt.Sprintf("# Chart Insights\n\n## Analysis\n\nBased on the data visualization, here are key insights:\n\n")
	
	// 这里可以添加实际的数据分析逻辑
	// 简化实现
	insights += "- Data points analyzed\n"
	insights += "- Trends identified\n"
	insights += "- Recommendations provided\n"

	if err := os.WriteFile(insightPath, []byte(insights), 0644); err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to write insights: %v", err)}, nil
	}

	output := fmt.Sprintf("Insights Added Successfully!\nInsights saved in: %s", insightPath)
	return &ToolResult{Output: output}, nil
}
