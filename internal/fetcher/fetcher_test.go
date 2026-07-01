package fetcher_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kushima-takeshi/techbrew/internal/config"
	"github.com/kushima-takeshi/techbrew/internal/fetcher"
)

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <item>
      <title>Hello Go</title>
      <link>https://example.com/go</link>
      <description>Go is great</description>
      <pubDate>Mon, 30 Jun 2026 06:00:00 GMT</pubDate>
    </item>
    <item>
      <title>Hello RSS</title>
      <link>https://example.com/rss</link>
      <description>RSS parsing</description>
    </item>
  </channel>
</rss>`

func TestFetchSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(sampleRSS))
	}))
	defer srv.Close()

	f := fetcher.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	articles := f.FetchAll(ctx, []config.Source{
		{Name: "Test", URL: srv.URL, Max: 2},
	}, 1)

	if len(articles) != 2 {
		t.Fatalf("expected 2 articles, got %d", len(articles))
	}
	if articles[0].Title != "Hello Go" {
		t.Fatalf("unexpected title: %s", articles[0].Title)
	}
	if articles[0].Source != "Test" {
		t.Fatalf("unexpected source: %s", articles[0].Source)
	}
}

func TestStripHTMLViaFetcher(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := `<?xml version="1.0"?><rss version="2.0"><channel><item>` +
			`<title>T</title><link>https://example.com</link>` +
			`<description><![CDATA[<p>bold <b>text</b></p>]]></description>` +
			`</item></channel></rss>`
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	f := fetcher.New()
	ctx := context.Background()
	articles := f.FetchAll(ctx, []config.Source{{Name: "X", URL: srv.URL, Max: 1}}, 1)
	if len(articles) != 1 {
		t.Fatal("expected 1 article")
	}
	if articles[0].Summary == "" {
		t.Fatal("expected non-empty summary")
	}
}
