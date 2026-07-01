package summarizer

import (
	"context"

	"github.com/kushima-takeshi/techbrew/internal/model"
)

type Summarizer interface {
	Summarize(ctx context.Context, articles []model.Article) (*model.Digest, error)
}
