package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Environment     string          `yaml:"environment"`
	Tenant          string          `yaml:"tenant"`
	Region          string          `yaml:"region"`
	RefreshInterval string          `yaml:"refresh_interval"`
	Server          ServerConfig    `yaml:"server"`
	Services        []ServiceConfig `yaml:"services"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name        string   `yaml:"name"`
	DisplayName string   `yaml:"display_name"`
	BaseURL     string   `yaml:"base_url"`
	DocPath     string   `yaml:"doc_path"`
	HealthCheck string   `yaml:"health_check"`
	Status      string   `yaml:"status"`
	Owner       string   `yaml:"owner"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	if cfg.Server.Port == "" {
		cfg.Server.Port = "9000"
	}
	if cfg.RefreshInterval == "" {
		cfg.RefreshInterval = "5m"
	}

	return &cfg, nil
}

// GetRefreshDuration 获取刷新间隔
func (c *Config) GetRefreshDuration() time.Duration {
	d, err := time.ParseDuration(c.RefreshInterval)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}

