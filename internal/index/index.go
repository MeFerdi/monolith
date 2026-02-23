package index

import (
	"context"
	"monolith/pkg/document"
)

type Index interface {
	Add(ctx context.Context, doc document.Document) error
	Search(ctx context.Context, query string) ([]SearchResult, error)
	Delete(ctx context.Context, id string) error
}
