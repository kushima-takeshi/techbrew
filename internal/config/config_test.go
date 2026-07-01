package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kushima-takeshi/techbrew/internal/config"
)

func TestLoadFile(t *testing.T) {
	path := filepath.Join("..", "..", "config", "sites.yaml")
	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if len(cfg.Sources) < 3 {
		t.Fatalf("expected at least 3 sources, got %d", len(cfg.Sources))
	}
	if cfg.Sources[0].Max != 5 {
		t.Fatalf("expected default max_items 5, got %d", cfg.Sources[0].Max)
	}
}

func TestLoadFileEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(path, []byte("sources: []"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := config.LoadFile(path)
	if err == nil {
		t.Fatal("expected error for empty sources")
	}
}

func TestExpandHome(t *testing.T) {
	t.Setenv("HOME", "/tmp/testhome")
	cfg := config.LoadEnv()
	if cfg.OutputPath == "" {
		t.Fatal("expected output path")
	}
}
