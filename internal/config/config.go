package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Max  int    `yaml:"max_items"`
}

type FileConfig struct {
	Sources []Source `yaml:"sources"`
}

type EnvConfig struct {
	OpenAIAPIKey   string
	OpenAIModel    string
	MaxConcurrency int
	OutputPath     string
}

func LoadFile(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if len(cfg.Sources) == 0 {
		return nil, fmt.Errorf("no sources defined in config")
	}

	for i := range cfg.Sources {
		if cfg.Sources[i].Max <= 0 {
			cfg.Sources[i].Max = 5
		}
	}

	return &cfg, nil
}

func LoadEnv() EnvConfig {
	return EnvConfig{
		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:    envOrDefault("OPENAI_MODEL", "gpt-4o-mini"),
		MaxConcurrency: envIntOrDefault("MAX_CONCURRENCY", 5),
		OutputPath:     expandHome(envOrDefault("OUTPUT_PATH", "~/TechDigest/latest.html")),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
