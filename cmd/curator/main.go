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
	noAI := flag.Bool("no-ai", false, "skip AI summarization and show RSS articles only")
	flag.Parse()

	if err := run(*configPath, *noAI); err != nil {
		log.Fatalf("curator: %v", err)
	}
}

func run(configPath string, noAI bool) error {
	log.Println("[1/4] 設定を読み込み中...")
	cfgPath, err := resolveConfigPath(configPath)
	if err != nil {
		return err
	}

	fileCfg, err := config.LoadFile(cfgPath)
	if err != nil {
		return err
	}
	envCfg := config.LoadEnv()
	log.Printf("      ソース %d 件 / 同時接続 %d / 出力先 %s", len(fileCfg.Sources), envCfg.MaxConcurrency, envCfg.OutputPath)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Println("[2/4] RSS を並行取得中...")
	articles := fetcher.New().FetchAll(ctx, fileCfg.Sources, envCfg.MaxConcurrency)
	if len(articles) == 0 {
		return fmt.Errorf("no articles fetched from any source")
	}
	log.Printf("      合計 %d 件取得完了", len(articles))

	var digest *model.Digest
	fallback := false

	if noAI {
		log.Println("[3/4] AI 要約をスキップ（--no-ai）")
		digest = report.ArticlesDigest(articles)
		fallback = true
	} else {
		log.Println("[3/4] AI 要約を生成中...")
		sum := summarizer.NewOpenAI(envCfg.OpenAIAPIKey, envCfg.OpenAIModel)
		d, err := sum.Summarize(ctx, articles)
		if err != nil {
			log.Printf("      要約失敗 → RSS 原文表示に切替: %v", err)
			digest = report.ArticlesDigest(articles)
			fallback = true
		} else {
			log.Printf("      トピック %d 件を生成", len(d.Topics))
			digest = d
		}
	}

	log.Println("[4/4] HTML を書き出し中...")
	if err := report.WriteHTML(envCfg.OutputPath, digest, fallback); err != nil {
		return err
	}

	log.Printf("      完了 → %s", envCfg.OutputPath)
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
	if err == nil {
		exeDir := filepath.Dir(exe)
		for _, candidate := range []string{
			filepath.Join(exeDir, path),
			filepath.Join(exeDir, "..", path),
		} {
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}

	wd, err := os.Getwd()
	if err == nil {
		candidate := filepath.Join(wd, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return path, nil
}
