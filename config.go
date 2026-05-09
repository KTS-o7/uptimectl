package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ServiceConfig defines a service to monitor.
type ServiceConfig struct {
	Name           string        `yaml:"name"`
	URL            string        `yaml:"url"`
	Method         string        `yaml:"method"` // GET or HEAD
	Timeout        time.Duration `yaml:"timeout"`
	ExpectedStatus int           `yaml:"expected_status"`
}

func (s ServiceConfig) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if s.URL == "" {
		return fmt.Errorf("service %q: URL is required", s.Name)
	}
	if s.Method != "GET" && s.Method != "HEAD" {
		return fmt.Errorf("service %q: method must be GET or HEAD, got %q", s.Name, s.Method)
	}
	if s.Timeout <= 0 {
		return fmt.Errorf("service %q: timeout must be positive", s.Name)
	}
	if s.ExpectedStatus <= 0 {
		return fmt.Errorf("service %q: expected_status must be positive", s.Name)
	}
	return nil
}

// Config is the application configuration.
type Config struct {
	Services    []ServiceConfig `yaml:"services"`
	Interval    time.Duration   `yaml:"interval"`
	HistoryDays int             `yaml:"history_days"`
	OutputPath  string          `yaml:"output_path"`
	DataPath    string          `yaml:"data_path"`
}

func (c Config) Validate() error {
	if len(c.Services) == 0 {
		return fmt.Errorf("at least one service is required")
	}
	for _, s := range c.Services {
		if err := s.Validate(); err != nil {
			return err
		}
	}
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	if c.HistoryDays <= 0 {
		return fmt.Errorf("history_days must be positive")
	}
	if c.OutputPath == "" {
		return fmt.Errorf("output_path is required")
	}
	if c.DataPath == "" {
		return fmt.Errorf("data_path is required")
	}
	return nil
}

// LoadConfig reads and parses the YAML config file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}
