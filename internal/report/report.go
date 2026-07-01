package report

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/kushima-takeshi/techbrew/internal/model"
)

const pageTemplate = `<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>TechBrew — {{.GeneratedAt}}</title>
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
    .container { max-width: 820px; margin: 0 auto; }
    h1 { font-size: 1.75rem; margin: 0 0 0.25rem; }
    .meta { color: var(--muted); font-size: 0.9rem; margin-bottom: 1.5rem; }
    .stats {
      display: flex;
      gap: 1rem;
      flex-wrap: wrap;
      margin-bottom: 2rem;
      font-size: 0.85rem;
      color: var(--muted);
    }
    .stats span {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 999px;
      padding: 0.35rem 0.85rem;
    }
    .note {
      background: var(--card);
      border-left: 4px solid var(--accent);
      padding: 1rem 1.25rem;
      margin-bottom: 1.5rem;
      border-radius: 0 8px 8px 0;
      font-size: 0.9rem;
    }
    .section-title {
      font-size: 1.15rem;
      margin: 2rem 0 1rem;
      padding-bottom: 0.35rem;
      border-bottom: 2px solid var(--border);
    }
    .ai-topic {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 1.25rem 1.5rem;
      margin-bottom: 1rem;
    }
    .ai-topic h3 { margin: 0 0 0.5rem; font-size: 1.05rem; }
    .ai-topic p { margin: 0 0 0.75rem; }
    .ai-topic ul { margin: 0; padding-left: 1.25rem; }
    .source-group { margin-bottom: 2rem; }
    .source-group h2 {
      font-size: 1rem;
      color: var(--accent);
      margin: 0 0 0.75rem;
      letter-spacing: 0.02em;
    }
    .article-card {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 1.1rem 1.35rem;
      margin-bottom: 0.75rem;
    }
    .article-card h3 {
      margin: 0 0 0.35rem;
      font-size: 1rem;
      line-height: 1.45;
    }
    .article-card h3 a {
      color: var(--text);
      text-decoration: none;
    }
    .article-card h3 a:hover { color: var(--accent); text-decoration: underline; }
    .article-meta {
      font-size: 0.78rem;
      color: var(--muted);
      margin-bottom: 0.5rem;
    }
    .article-summary {
      margin: 0;
      font-size: 0.9rem;
      color: var(--muted);
    }
    .read-link {
      display: inline-block;
      margin-top: 0.65rem;
      font-size: 0.85rem;
      color: var(--accent);
      text-decoration: none;
    }
    .read-link:hover { text-decoration: underline; }
    a { color: var(--accent); }
  </style>
</head>
<body>
  <div class="container">
    <h1>TechBrew ダイジェスト</h1>
    <p class="meta">生成日時: {{.GeneratedAt}}</p>

    <div class="stats">
      <span>{{len .Articles}} 件の記事</span>
      <span>{{len .SourceGroups}} ソース</span>
    </div>

    {{if .Fallback}}
    <div class="note">元記事の RSS テキストをそのまま表示しています。全文は各サイトでご覧ください。</div>
    {{end}}

    {{if .ShowAI}}
    <h2 class="section-title">AI ピックアップ（参考）</h2>
    {{range .Topics}}
    <section class="ai-topic">
      <h3>{{.Title}}</h3>
      <p>{{.Summary}}</p>
      {{if .RelatedURLs}}
      <ul>
        {{range .RelatedURLs}}
        <li><a href="{{.}}" target="_blank" rel="noopener">元記事を読む</a></li>
        {{end}}
      </ul>
      {{end}}
    </section>
    {{end}}
    {{end}}

    <h2 class="section-title">ソース別の記事一覧</h2>
    {{range .SourceGroups}}
    <section class="source-group">
      <h2>{{.Name}}（{{len .Articles}}件）</h2>
      {{range .Articles}}
      <article class="article-card">
        <h3><a href="{{.URL}}" target="_blank" rel="noopener">{{.Title}}</a></h3>
        <p class="article-meta">{{.Published}}</p>
        {{if .Summary}}
        <p class="article-summary">{{.Summary}}</p>
        {{end}}
        <a class="read-link" href="{{.URL}}" target="_blank" rel="noopener">元記事を読む →</a>
      </article>
      {{end}}
    </section>
    {{end}}
  </div>
</body>
</html>`

type sourceGroup struct {
	Name     string
	Articles []articleView
}

type articleView struct {
	Title     string
	URL       string
	Summary   string
	Published string
}

type pageData struct {
	GeneratedAt  string
	Topics       []model.Topic
	Articles     []model.Article
	SourceGroups []sourceGroup
	Fallback     bool
	ShowAI       bool
}

func WriteHTML(outputPath string, digest *model.Digest, fallback bool) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	now := time.Now()
	html, err := render(digest, fallback, now)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, []byte(html), 0o644); err != nil {
		return fmt.Errorf("write latest: %w", err)
	}

	datedPath := filepath.Join(filepath.Dir(outputPath), now.Format("2006-01-02")+".html")
	if datedPath != outputPath {
		if err := os.WriteFile(datedPath, []byte(html), 0o644); err != nil {
			return fmt.Errorf("write dated copy: %w", err)
		}
	}

	return nil
}

func RenderHTML(digest *model.Digest, fallback bool) (string, error) {
	return render(digest, fallback, time.Now())
}

func render(digest *model.Digest, fallback bool, now time.Time) (string, error) {
	tmpl, err := template.New("digest").Parse(pageTemplate)
	if err != nil {
		return "", err
	}

	showAI := !fallback && len(digest.Topics) > 0
	data := pageData{
		GeneratedAt:  now.Format("2006-01-02 15:04"),
		Topics:       digest.Topics,
		Articles:     digest.Articles,
		SourceGroups: groupBySource(digest.Articles),
		Fallback:     fallback,
		ShowAI:       showAI,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}
	return buf.String(), nil
}

func groupBySource(articles []model.Article) []sourceGroup {
	order := make([]string, 0)
	groups := make(map[string][]articleView)

	for _, a := range articles {
		if _, ok := groups[a.Source]; !ok {
			order = append(order, a.Source)
		}
		groups[a.Source] = append(groups[a.Source], articleView{
			Title:     a.Title,
			URL:       a.URL,
			Summary:   a.Summary,
			Published: a.Published.Format("2006-01-02 15:04"),
		})
	}

	sort.Strings(order)
	result := make([]sourceGroup, 0, len(order))
	for _, name := range order {
		result = append(result, sourceGroup{Name: name, Articles: groups[name]})
	}
	return result
}

func ArticlesDigest(articles []model.Article) *model.Digest {
	return &model.Digest{Articles: articles}
}

func FallbackDigest(articles []model.Article) *model.Digest {
	return ArticlesDigest(articles)
}
