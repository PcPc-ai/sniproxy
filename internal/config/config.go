package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Whitelist WhitelistConfig `yaml:"whitelist"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTPSPort string `yaml:"https_port" default:"443"`
	HTTPPort  string `yaml:"http_port" default:"80"`
	BindAddr  string `yaml:"bind_addr" default:"0.0.0.0"`
}

// WhitelistConfig 白名单配置
type WhitelistConfig struct {
	Enabled     bool     `yaml:"enabled" default:"true"`
	DefaultMode string   `yaml:"default_mode" default:"deny"` // allow 或 deny
	Domains     []string `yaml:"domains"`
	DomainsFile string   `yaml:"domains_file"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `yaml:"level" default:"info"`
	Format string `yaml:"format" default:"text"` // text 或 json
}

// LoadConfig 从文件加载配置
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	if config.Server.HTTPSPort == "" {
		config.Server.HTTPSPort = "443"
	}
	if config.Server.HTTPPort == "" {
		config.Server.HTTPPort = "80"
	}
	if config.Server.BindAddr == "" {
		config.Server.BindAddr = "0.0.0.0"
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
	if config.Whitelist.DefaultMode == "" {
		config.Whitelist.DefaultMode = "deny"
	}

	return &config, nil
}
