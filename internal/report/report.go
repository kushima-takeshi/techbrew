package report

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/kushima-takeshi/techbrew/internal/model"
)

const pageTemplate = `<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Tech Digest — {{.GeneratedAt}}</title>
  <style>
    :root {
      color-scheme: light dark;
      --bg: #0f1419;
      --card: #1a2332;
      --text: #e7ecf3;
      --muted: #8b9cb3;
      --accent: #5b9fd4;
      --border: #2a3544;
    }
    @media (prefers-color-scheme: light) {
      :root {
        --bg: #f4f7fb;
        --card: #ffffff;
        --text: #1a2332;
        --muted: #5a6b82;
        --accent: #2563eb;
        --border: #d8e0ec;
      }
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: var(--bg);
      color: var(--text);
      line-height: 1.6;
      padding: 2rem 1rem 4rem;
    }
    .container { max-width: 760px; margin: 0 auto; }
    h1 { font-size: 1.75rem; margin: 0 0 0.25rem; }
    .meta { color: var(--muted); font-size: 0.9rem; margin-bottom: 2rem; }
    .topic {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 1.25rem 1.5rem;
      margin-bottom: 1rem;
    }
    .topic h2 { margin: 0 0 0.5rem; font-size: 1.1rem; }
    .topic p { margin: 0 0 0.75rem; }
    .topic ul { margin: 0; padding-left: 1.25rem; }
    .topic a { color: var(--accent); text-decoration: none; }
    .topic a:hover { text-decoration: underline; }
    details {
      margin-top: 2rem;
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 1rem 1.25rem;
    }
    summary { cursor: pointer; font-weight: 600; }
    .article-list { list-style: none; padding: 0; margin: 1rem 0 0; }
    .article-list li {
      padding: 0.5rem 0;
      border-top: 1px solid var(--border);
      font-size: 0.9rem;
    }
    .article-list .source { color: var(--muted); font-size: 0.8rem; }
    .fallback-note {
      background: var(--card);
      border-left: 4px solid var(--accent);
      padding: 1rem 1.25rem;
      margin-bottom: 1.5rem;
      border-radius: 0 8px 8px 0;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>今日の技術ダイジェスト</h1>
    <p class="meta">生成日時: {{.GeneratedAt}}</p>

    {{if .Fallback}}
    <div class="fallback-note">AI要約を取得できなかったため、取得した記事一覧を表示しています。</div>
    {{end}}

    {{range .Topics}}
    <section class="topic">
      <h2>{{.Title}}</h2>
      <p>{{.Summary}}</p>
      {{if .RelatedURLs}}
      <ul>
        {{range .RelatedURLs}}
        <li><a href="{{.}}" target="_blank" rel="noopener">{{.}}</a></li>
        {{end}}
      </ul>
      {{end}}
    </section>
    {{end}}

    <details>
      <summary>取得した記事一覧（{{len .Articles}}件）</summary>
      <ul class="article-list">
        {{range .Articles}}
        <li>
          <span class="source">{{.Source}}</span><br>
          <a href="{{.URL}}" target="_blank" rel="noopener">{{.Title}}</a>
        </li>
        {{end}}
      </ul>
    </details>
  </div>
</body>
</html>`

type pageData struct {
	GeneratedAt string
	Topics      []model.Topic
	Articles    []model.Article
	Fallback    bool
}

func WriteHTML(outputPath string, digest *model.Digest, fallback bool) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	tmpl, err := template.New("digest").Parse(pageTemplate)
	if err != nil {
		return err
	}

	now := time.Now()
	data := pageData{
		GeneratedAt: now.Format("2006-01-02 15:04"),
		Topics:      digest.Topics,
		Articles:    digest.Articles,
		Fallback:    fallback,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write latest: %w", err)
	}

	datedPath := filepath.Join(filepath.Dir(outputPath), now.Format("2006-01-02")+".html")
	if datedPath != outputPath {
		if err := os.WriteFile(datedPath, buf.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write dated copy: %w", err)
		}
	}

	return nil
}

func RenderHTML(digest *model.Digest, fallback bool) (string, error) {
	tmpl, err := template.New("digest").Parse(pageTemplate)
	if err != nil {
		return "", err
	}

	data := pageData{
		GeneratedAt: time.Now().Format("2006-01-02 15:04"),
		Topics:      digest.Topics,
		Articles:    digest.Articles,
		Fallback:    fallback,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func FallbackDigest(articles []model.Article) *model.Digest {
	topics := []model.Topic{
		{
			Title:   "取得記事のハイライト",
			Summary: fmt.Sprintf("%d件の記事を取得しました。詳細は下部の一覧をご確認ください。", len(articles)),
		},
	}
	urls := make([]string, 0, 3)
	for _, a := range articles {
		if a.URL != "" {
			urls = append(urls, a.URL)
		}
		if len(urls) >= 3 {
			break
		}
	}
	if len(urls) > 0 {
		topics[0].RelatedURLs = urls
	}
	return &model.Digest{Topics: topics, Articles: articles}
}
