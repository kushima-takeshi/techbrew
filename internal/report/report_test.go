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
		"今日の技術ダイジェスト",
		"Go 1.23 リリース",
		"新機能が追加されました。",
		"Test Article",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected HTML to contain %q", want)
		}
	}
}

func TestFallbackDigest(t *testing.T) {
	articles := []model.Article{
		{Title: "A", URL: "https://a.example"},
		{Title: "B", URL: "https://b.example"},
	}
	d := report.FallbackDigest(articles)
	if len(d.Topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(d.Topics))
	}
	if len(d.Topics[0].RelatedURLs) != 2 {
		t.Fatalf("expected 2 related urls")
	}
}

func TestWriteHTML(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/latest.html"
	digest := &model.Digest{
		Topics: []model.Topic{{Title: "T", Summary: "S"}},
	}
	if err := report.WriteHTML(path, digest, true); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}
}
