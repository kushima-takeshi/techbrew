package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kushima-takeshi/techbrew/internal/config"
	"github.com/kushima-takeshi/techbrew/internal/fetcher"
	"github.com/kushima-takeshi/techbrew/internal/model"
	"github.com/kushima-takeshi/techbrew/internal/report"
	"github.com/kushima-takeshi/techbrew/internal/summarizer"
)

func main() {
	configPath := flag.String("config", "config/sites.yaml", "path to sites YAML config")
	flag.Parse()

	if err := run(*configPath); err != nil {
		log.Fatalf("curator: %v", err)
	}
}

func run(configPath string) error {
	cfgPath, err := resolveConfigPath(configPath)
	if err != nil {
		return err
	}

	fileCfg, err := config.LoadFile(cfgPath)
	if err != nil {
		return err
	}
	envCfg := config.LoadEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Printf("fetching from %d sources (concurrency=%d)...", len(fileCfg.Sources), envCfg.MaxConcurrency)
	articles := fetcher.New().FetchAll(ctx, fileCfg.Sources, envCfg.MaxConcurrency)
	if len(articles) == 0 {
		return fmt.Errorf("no articles fetched from any source")
	}
	log.Printf("fetched %d articles", len(articles))

	var digest *model.Digest
	fallback := false

	sum := summarizer.NewOpenAI(envCfg.OpenAIAPIKey, envCfg.OpenAIModel)
	d, err := sum.Summarize(ctx, articles)
	if err != nil {
		log.Printf("summarizer failed, using fallback: %v", err)
		digest = report.FallbackDigest(articles)
		fallback = true
	} else {
		digest = d
	}

	if err := report.WriteHTML(envCfg.OutputPath, digest, fallback); err != nil {
		return err
	}

	log.Printf("wrote digest to %s", envCfg.OutputPath)
	return nil
}

func resolveConfigPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	exe, err := os.Executable()
	if err != nil {
		return path, nil
	}
	candidate := filepath.Join(filepath.Dir(exe), path)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return path, nil
	}
	candidate = filepath.Join(wd, path)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	return path, nil
}
