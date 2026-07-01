package summarizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kushima-takeshi/techbrew/internal/model"
)

const openAIEndpoint = "https://api.openai.com/v1/chat/completions"

type OpenAISummarizer struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewOpenAI(apiKey, model string) *OpenAISummarizer {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAISummarizer{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type llmResponse struct {
	Topics []model.Topic `json:"topics"`
}

func (s *OpenAISummarizer) Summarize(ctx context.Context, articles []model.Article) (*model.Digest, error) {
	if s.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	prompt := buildPrompt(articles)

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		digest, err := s.call(ctx, prompt)
		if err == nil {
			digest.Articles = articles
			return digest, nil
		}
		lastErr = err
	}

	return nil, lastErr
}

func (s *OpenAISummarizer) call(ctx context.Context, userPrompt string) (*model.Digest, error) {
	body := map[string]any{
		"model": s.Model,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "You are a tech news curator. Respond only with valid JSON matching the requested schema. " +
					"Write all topic titles and summaries in Japanese.",
			},
			{"role": "user", "content": userPrompt},
		},
		"temperature":     0.3,
		"response_format": map[string]string{"type": "json_object"},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("decode openai response: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	var parsed llmResponse
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &parsed); err != nil {
		return nil, fmt.Errorf("decode llm json: %w", err)
	}
	if len(parsed.Topics) == 0 {
		return nil, fmt.Errorf("llm returned no topics")
	}
	if len(parsed.Topics) > 3 {
		parsed.Topics = parsed.Topics[:3]
	}

	return &model.Digest{Topics: parsed.Topics}, nil
}

func buildPrompt(articles []model.Article) string {
	var b strings.Builder
	b.WriteString("以下の技術記事一覧から、今日最も重要なトピックを3つ選び、日本語で要約してください。\n")
	b.WriteString("JSON形式: {\"topics\":[{\"title\":\"...\",\"summary\":\"1行要約\",\"related_urls\":[\"...\"]}]}\n\n")
	b.WriteString("記事一覧:\n")
	for i, a := range articles {
		fmt.Fprintf(&b, "%d. [%s] %s\n   URL: %s\n   概要: %s\n\n", i+1, a.Source, a.Title, a.URL, a.Summary)
	}
	return b.String()
}
