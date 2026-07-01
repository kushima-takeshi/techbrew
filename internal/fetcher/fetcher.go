package fetcher

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"golang.org/x/sync/semaphore"

	"github.com/kushima-takeshi/techbrew/internal/config"
	"github.com/kushima-takeshi/techbrew/internal/model"
)

const (
	sourceTimeout = 10 * time.Second
	userAgent     = "TechBrew/1.0 (+https://github.com/kushima-takeshi/techbrew)"
)

type Fetcher struct {
	parser *gofeed.Parser
}

func New() *Fetcher {
	parser := gofeed.NewParser()
	parser.UserAgent = userAgent
	return &Fetcher{parser: parser}
}

func (f *Fetcher) FetchAll(ctx context.Context, sources []config.Source, maxConcurrency int) []model.Article {
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	sem := semaphore.NewWeighted(int64(maxConcurrency))
	var (
		mu       sync.Mutex
		articles []model.Article
		wg       sync.WaitGroup
	)

	for _, src := range sources {
		wg.Add(1)
		go func(s config.Source) {
			defer wg.Done()

			if err := sem.Acquire(ctx, 1); err != nil {
				log.Printf("fetcher: skip %s: %v", s.Name, err)
				return
			}
			defer sem.Release(1)

			fetched, err := f.fetchSource(ctx, s)
			if err != nil {
				log.Printf("      ✗ %s: %v", s.Name, err)
				return
			}

			log.Printf("      ✓ %s: %d 件", s.Name, len(fetched))

			mu.Lock()
			articles = append(articles, fetched...)
			mu.Unlock()
		}(src)
	}

	wg.Wait()
	return articles
}

func (f *Fetcher) fetchSource(ctx context.Context, src config.Source) ([]model.Article, error) {
	srcCtx, cancel := context.WithTimeout(ctx, sourceTimeout)
	defer cancel()

	feed, err := f.parser.ParseURLWithContext(src.URL, srcCtx)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}

	limit := src.Max
	if limit <= 0 {
		limit = 5
	}
	if len(feed.Items) < limit {
		limit = len(feed.Items)
	}

	articles := make([]model.Article, 0, limit)
	for _, item := range feed.Items[:limit] {
		published := time.Now()
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			published = *item.UpdatedParsed
		}

		summary := strings.TrimSpace(item.Description)
		if summary == "" {
			summary = strings.TrimSpace(item.Content)
		}
		summary = stripHTML(summary)
		if len(summary) > 300 {
			summary = summary[:300] + "..."
		}

		articles = append(articles, model.Article{
			Title:     strings.TrimSpace(item.Title),
			URL:       strings.TrimSpace(item.Link),
			Summary:   summary,
			Published: published,
			Source:    src.Name,
		})
	}

	return articles, nil
}

func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
