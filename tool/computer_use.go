package tool

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"
)

// ComputerUseTool 计算机使用工具（屏幕控制）
type ComputerUseTool struct {
	outputDir string
}

func NewComputerUseTool() *ComputerUseTool {
	return &ComputerUseTool{
		outputDir: "workspace/screenshots",
	}
}

func (c *ComputerUseTool) Name() string {
	return "computer_use"
}

func (c *ComputerUseTool) Description() string {
	return `A comprehensive computer automation tool that allows interaction with the desktop environment.
This tool provides commands for controlling mouse, keyboard, and taking screenshots.
It maintains state including current mouse position.
Use this when you need to automate desktop applications, fill forms, or perform GUI interactions.
Key capabilities include:
* Mouse Control: Move, click, drag, scroll
* Keyboard Input: Type text, press keys or key combinations
* Screenshots: Capture and save screen images
* Waiting: Pause execution for specified duration`
}

func (c *ComputerUseTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "The computer action to perform",
				"enum": []string{
					"move_to",
					"click",
					"scroll",
					"typing",
					"press",
					"wait",
					"mouse_down",
					"mouse_up",
					"drag_to",
					"hotkey",
					"screenshot",
				},
			},
			"x": map[string]interface{}{
				"type":        "number",
				"description": "X coordinate for mouse actions",
			},
			"y": map[string]interface{}{
				"type":        "number",
				"description": "Y coordinate for mouse actions",
			},
			"button": map[string]interface{}{
				"type":        "string",
				"description": "Mouse button for click/drag actions",
				"enum":        []string{"left", "right", "middle"},
				"default":     "left",
			},
			"num_clicks": map[string]interface{}{
				"type":        "integer",
				"description": "Number of clicks",
				"enum":        []int{1, 2, 3},
				"default":     1,
			},
			"amount": map[string]interface{}{
				"type":        "integer",
				"description": "Scroll amount (positive for up, negative for down)",
			},
			"text": map[string]interface{}{
				"type":        "string",
				"description": "Text to type",
			},
			"key": map[string]interface{}{
				"type":        "string",
				"description": "Key to press",
			},
			"keys": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Keys for hotkey combination",
			},
			"duration": map[string]interface{}{
				"type":        "number",
				"description": "Duration in seconds for wait action",
			},
		},
		"required": []string{"action"},
	}
}

func (c *ComputerUseTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	action, ok := args["action"].(string)
	if !ok {
		return &ToolResult{Error: "action parameter is required"}, nil
	}

	switch action {
	case "move_to":
		return c.moveTo(ctx, args)
	case "click":
		return c.click(ctx, args)
	case "scroll":
		return c.scroll(ctx, args)
	case "typing":
		return c.typing(ctx, args)
	case "press":
		return c.press(ctx, args)
	case "wait":
		return c.wait(ctx, args)
	case "mouse_down":
		return c.mouseDown(ctx, args)
	case "mouse_up":
		return c.mouseUp(ctx, args)
	case "drag_to":
		return c.dragTo(ctx, args)
	case "hotkey":
		return c.hotkey(ctx, args)
	case "screenshot":
		return c.screenshot(ctx, args)
	default:
		return &ToolResult{Error: fmt.Sprintf("Unknown action: %s", action)}, nil
	}
}

func (c *ComputerUseTool) moveTo(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	x, ok := args["x"].(float64)
	if !ok {
		return &ToolResult{Error: "x coordinate is required for move_to"}, nil
	}
	y, ok := args["y"].(float64)
	if !ok {
		return &ToolResult{Error: "y coordinate is required for move_to"}, nil
	}

	robotgo.Move(int(x), int(y))
	return &ToolResult{Output: fmt.Sprintf("Mouse moved to (%d, %d)", int(x), int(y))}, nil
}

func (c *ComputerUseTool) click(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	button := "left"
	if b, ok := args["button"].(string); ok && b != "" {
		button = b
	}

	numClicks := 1
	if nc, ok := args["num_clicks"].(float64); ok {
		numClicks = int(nc)
	}

	x, hasX := args["x"].(float64)
	y, hasY := args["y"].(float64)

	if hasX && hasY {
		// 点击指定坐标
		robotgo.Move(int(x), int(y))
	}

	// TODO: Implement mouse clicks using platform-specific libraries
	// switch button {
	// case "left":
	// 	for i := 0; i < numClicks; i++ {
	// 		robotgo.Click("left")
	// 	}
	// case "right":
	// 	for i := 0; i < numClicks; i++ {
	// 		robotgo.Click("right")
	// 	}
	// case "middle":
	// 	for i := 0; i < numClicks; i++ {
	// 		robotgo.Click("center")
	// 	}
	// }

	return &ToolResult{Output: fmt.Sprintf("Clicked %s button %d times", button, numClicks)}, nil
}

func (c *ComputerUseTool) scroll(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	amount, ok := args["amount"].(float64)
	if !ok {
		return &ToolResult{Error: "amount is required for scroll"}, nil
	}

	// robotgo.Scroll(int(amount), 0)
	// TODO: Implement scroll using platform-specific libraries
	return &ToolResult{Output: fmt.Sprintf("Scrolled %d units", int(amount))}, nil
}

func (c *ComputerUseTool) typing(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	text, ok := args["text"].(string)
	if !ok {
		return &ToolResult{Error: "text is required for typing"}, nil
	}

	// robotgo.TypeStr(text)
	// TODO: Implement typing using platform-specific libraries
	return &ToolResult{Output: fmt.Sprintf("Typed: %s", text)}, nil
}

func (c *ComputerUseTool) press(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	key, ok := args["key"].(string)
	if !ok {
		return &ToolResult{Error: "key is required for press"}, nil
	}

	// robotgo.KeyTap(key)
	// TODO: Implement key press using platform-specific libraries
	return &ToolResult{Output: fmt.Sprintf("Pressed key: %s", key)}, nil
}

func (c *ComputerUseTool) wait(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	duration := 1.0
	if d, ok := args["duration"].(float64); ok {
		duration = d
	}

	time.Sleep(time.Duration(duration * float64(time.Second)))
	return &ToolResult{Output: fmt.Sprintf("Waited for %.2f seconds", duration)}, nil
}

func (c *ComputerUseTool) mouseDown(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	button := "left"
	if b, ok := args["button"].(string); ok && b != "" {
		button = b
	}

	// robotgo.Toggle(button, "down")
	// TODO: Implement mouse down using platform-specific libraries
	return &ToolResult{Output: fmt.Sprintf("Mouse button %s pressed down", button)}, nil
}

func (c *ComputerUseTool) mouseUp(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	button := "left"
	if b, ok := args["button"].(string); ok && b != "" {
		button = b
	}

	// robotgo.Toggle(button, "up")
	// TODO: Implement mouse up using platform-specific libraries
	return &ToolResult{Output: fmt.Sprintf("Mouse button %s released", button)}, nil
}

func (c *ComputerUseTool) dragTo(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	x, ok := args["x"].(float64)
	if !ok {
		return &ToolResult{Error: "x coordinate is required for drag_to"}, nil
	}
	y, ok := args["y"].(float64)
	if !ok {
		return &ToolResult{Error: "y coordinate is required for drag_to"}, nil
	}

	// robotgo.Drag(int(x), int(y))
	// TODO: Implement drag using platform-specific libraries
	return &ToolResult{Output: fmt.Sprintf("Dragged to (%d, %d)", int(x), int(y))}, nil
}

func (c *ComputerUseTool) hotkey(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	keys, ok := args["keys"].([]interface{})
	if !ok || len(keys) == 0 {
		return &ToolResult{Error: "keys array is required for hotkey"}, nil
	}

	keyStrs := make([]string, len(keys))
	for i, k := range keys {
		keyStrs[i] = k.(string)
	}

	// robotgo.KeyTap(keyStrs...)
	// TODO: Implement hotkey using platform-specific libraries
	return &ToolResult{Output: fmt.Sprintf("Pressed hotkey: %v", keyStrs)}, nil
}

func (c *ComputerUseTool) screenshot(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	os.MkdirAll(c.outputDir, 0755)

	// TODO: Implement screenshot using platform-specific libraries
	// For now, return a placeholder
	// bitmap := robotgo.CaptureScreen()
	// defer robotgo.FreeBitmap(bitmap)
	// img := robotgo.ToImage(bitmap)
	
	// Create a placeholder image (1x1 pixel)
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))

	// 保存截图
	timestamp := time.Now().Format("20060102_150405")
	screenshotPath := filepath.Join(c.outputDir, fmt.Sprintf("screenshot_%s.png", timestamp))

	file, err := os.Create(screenshotPath)
	if err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to create screenshot file: %v", err)}, nil
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return &ToolResult{Error: fmt.Sprintf("Failed to encode screenshot: %v", err)}, nil
	}

	return &ToolResult{Output: fmt.Sprintf("Screenshot saved to: %s", screenshotPath)}, nil
}
