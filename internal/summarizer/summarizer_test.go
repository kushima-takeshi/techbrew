package summarizer_test

import (
	"context"
	"testing"

	"github.com/kushima-takeshi/techbrew/internal/model"
	"github.com/kushima-takeshi/techbrew/internal/summarizer"
)

type mockSummarizer struct{}

func (m *mockSummarizer) Summarize(_ context.Context, articles []model.Article) (*model.Digest, error) {
	return &model.Digest{
		Topics: []model.Topic{
			{Title: "モックトピック", Summary: "テスト要約", RelatedURLs: []string{"https://example.com"}},
		},
		Articles: articles,
	}, nil
}

func TestSummarizerInterface(t *testing.T) {
	var s summarizer.Summarizer = &mockSummarizer{}
	d, err := s.Summarize(context.Background(), []model.Article{{Title: "x"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Topics) != 1 {
		t.Fatalf("expected 1 topic")
	}
}

func TestOpenAINoKey(t *testing.T) {
	s := summarizer.NewOpenAI("", "gpt-4o-mini")
	_, err := s.Summarize(context.Background(), []model.Article{{Title: "x"}})
	if err == nil {
		t.Fatal("expected error without API key")
	}
}
