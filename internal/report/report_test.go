package report_test

import (
	"strings"
	"testing"
	"time"

	"github.com/kushima-takeshi/techbrew/internal/model"
	"github.com/kushima-takeshi/techbrew/internal/report"
)

func TestRenderHTML(t *testing.T) {
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

	html, err := report.RenderHTML(digest, false)
	if err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}
	for _, want := range []string{
		"TechBrew ダイジェスト",
		"AI ピックアップ",
		"Go 1.23 リリース",
		"ソース別の記事一覧",
		"Test Article",
		"RSS excerpt text",
		"元記事を読む",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected HTML to contain %q", want)
		}
	}
}

func TestRenderHTMLArticlesOnly(t *testing.T) {
	digest := report.ArticlesDigest([]model.Article{
		{Title: "A", URL: "https://a.example", Source: "Qiita", Summary: "summary a"},
		{Title: "B", URL: "https://b.example", Source: "Zenn", Summary: "summary b"},
	})
	html, err := report.RenderHTML(digest, true)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "AI ピックアップ") {
		t.Fatal("articles-only mode should not show AI section")
	}
	if !strings.Contains(html, "Qiita") || !strings.Contains(html, "summary a") {
		t.Fatal("expected source group with summary")
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
	if err := report.WriteHTML(path, digest, true); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}
}
