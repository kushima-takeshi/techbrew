package report_test

import (
	"strings"
	"testing"
	"time"

	"github.com/kushima-takeshi/techbrew/internal/model"
	"github.com/kushima-takeshi/techbrew/internal/report"
)

func TestRenderHTMLAIEnabled(t *testing.T) {
	digest := &model.Digest{
		Topics: []model.Topic{
			{
				Title:       "Go 1.23 リリース",
				Summary:     "新機能が追加されました。",
				RelatedURLs: []string{"https://example.com/go"},
			},
		},
		Articles: []model.Article{
			{
				Title:     "Test Article",
				URL:       "https://example.com/article",
				Summary:   "RSS excerpt text",
				Source:    "Test",
				Published: time.Now(),
			},
		},
	}

	html, err := report.RenderHTML(digest, report.ModeAIEnabled)
	if err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}
	for _, want := range []string{
		"TechBrew ダイジェスト",
		"AI ピックアップ",
		"AI 要約は参考情報",
		"Go 1.23 リリース",
		"Test Article",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected HTML to contain %q", want)
		}
	}
}

func TestRenderHTMLRSSOnly(t *testing.T) {
	digest := report.ArticlesDigest([]model.Article{
		{Title: "A", URL: "https://a.example", Source: "Qiita", Summary: "summary a"},
	})
	html, err := report.RenderHTML(digest, report.ModeRSSOnly)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "AI ピックアップ") {
		t.Fatal("RSS-only mode should not show AI section")
	}
	if !strings.Contains(html, "RSS から取得") {
		t.Fatal("expected RSS note")
	}
}

func TestRenderHTMLAIFailed(t *testing.T) {
	digest := report.ArticlesDigest([]model.Article{{Title: "A", URL: "https://a.example", Source: "Test"}})
	html, err := report.RenderHTML(digest, report.ModeAIFailed)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "AI 要約に失敗") {
		t.Fatal("expected AI failed note")
	}
}

func TestArticlesDigest(t *testing.T) {
	articles := []model.Article{
		{Title: "A", URL: "https://a.example"},
	}
	d := report.ArticlesDigest(articles)
	if len(d.Articles) != 1 || len(d.Topics) != 0 {
		t.Fatalf("unexpected digest: %+v", d)
	}
}

func TestWriteHTML(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/latest.html"
	digest := report.ArticlesDigest([]model.Article{
		{Title: "T", URL: "https://example.com", Source: "Test"},
	})
	if err := report.WriteHTML(path, digest, report.ModeRSSOnly); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}
}
