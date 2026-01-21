package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

type LLMSettings struct {
	Model       string  `toml:"model"`
	BaseURL     string  `toml:"base_url"`
	APIKey      string  `toml:"api_key"`
	MaxTokens   int     `toml:"max_tokens"`
	Temperature float64 `toml:"temperature"`
}

type AppConfig struct {
	LLM map[string]LLMSettings `toml:"llm"`
}

type Config struct {
	config *AppConfig
	mu     sync.RWMutex
}

var (
	instance *Config
	once     sync.Once
)

// GetInstance 获取配置单例
func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{}
		instance.loadConfig()
	})
	return instance
}

// getConfigPath 获取配置文件路径
func (c *Config) getConfigPath() (string, error) {
	// 尝试获取项目根目录
	workDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 查找 config.toml
	configPath := filepath.Join(workDir, "config", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// 回退到 example
	examplePath := filepath.Join(workDir, "config", "config.example.toml")
	if _, err := os.Stat(examplePath); err == nil {
		return examplePath, nil
	}

	return "", fmt.Errorf("no configuration file found in config directory")
}

// loadConfig 加载配置
func (c *Config) loadConfig() {
	c.mu.Lock()
	defer c.mu.Unlock()

	configPath, err := c.getConfigPath()
	if err != nil {
		panic(fmt.Sprintf("Failed to find config file: %v", err))
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config file: %v", err))
	}

	var rawConfig map[string]interface{}
	if err := toml.Unmarshal(data, &rawConfig); err != nil {
		panic(fmt.Sprintf("Failed to parse config file: %v", err))
	}

	// 解析 LLM 配置
	llmConfig := make(map[string]LLMSettings)
	llmRaw, ok := rawConfig["llm"].(map[string]interface{})
	if !ok {
		panic("llm configuration not found")
	}

	// 获取基础配置
	baseLLM := LLMSettings{
		Model:       getString(llmRaw, "model", ""),
		BaseURL:     getString(llmRaw, "base_url", ""),
		APIKey:      getString(llmRaw, "api_key", ""),
		MaxTokens:   getInt(llmRaw, "max_tokens", 4096),
		Temperature: getFloat(llmRaw, "temperature", 0.0),
	}

	llmConfig["default"] = baseLLM

	// 处理覆盖配置（如 llm.vision）
	for k, v := range llmRaw {
		if k == "model" || k == "base_url" || k == "api_key" || k == "max_tokens" || k == "temperature" {
			continue
		}
		if override, ok := v.(map[string]interface{}); ok {
			overrideSettings := baseLLM
			if model := getString(override, "model", ""); model != "" {
				overrideSettings.Model = model
			}
			if baseURL := getString(override, "base_url", ""); baseURL != "" {
				overrideSettings.BaseURL = baseURL
			}
			if apiKey := getString(override, "api_key", ""); apiKey != "" {
				overrideSettings.APIKey = apiKey
			}
			if maxTokens := getInt(override, "max_tokens", 0); maxTokens > 0 {
				overrideSettings.MaxTokens = maxTokens
			}
			if temp := getFloat(override, "temperature", -1); temp >= 0 {
				overrideSettings.Temperature = temp
			}
			llmConfig[k] = overrideSettings
		}
	}

	c.config = &AppConfig{LLM: llmConfig}
}

// GetLLM 获取 LLM 配置
func (c *Config) GetLLM(name string) LLMSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if settings, ok := c.config.LLM[name]; ok {
		return settings
	}
	return c.config.LLM["default"]
}

// 辅助函数
func getString(m map[string]interface{}, key string, defaultValue string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultValue
}

func getInt(m map[string]interface{}, key string, defaultValue int) int {
	if v, ok := m[key].(int64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return defaultValue
}

func getFloat(m map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return defaultValue
}

