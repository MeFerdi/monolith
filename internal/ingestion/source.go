package ingestion

import (
	"context"
	"monolith/pkg/document"
	"time"
)

type Source interface {
	Name() string
	FetchAll(ctx context.Context) ([]document.Document, error)
	FetchSince(ctx context.Context, since time.Time) ([]document.Document, error)
}
