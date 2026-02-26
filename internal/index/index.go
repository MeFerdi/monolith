package index

import (
	"context"
	"monolith/pkg/document"
	"time"
)

// Query is the structured search request passed to the index
type Query struct {
	Text   string
	Limit  int
	Offset int
	Fuzzy  int
}
// SearchResult is what the index returns - minimal, ranking-focused
type SearchResult struct {
	ID      string
	Score   float64
	Snippet string
}

// Index is the full-text search contract
type Index interface {
	Add(ctx context.Context, doc document.Document) error
	Search(ctx context.Context, q Query) ([]SearchResult, error)
	Delete(ctx context.Context, id string) error
}

// MetadataStore is the source-of-truth contract for document metadata
type MetadataStore interface {
	Get(ctx context.Context, id string) (DocumentMeta, error)
	Upsert(ctx context.Context, meta DocumentMeta) error
	Delete(ctx context.Context, id string) error
	ListBySource(ctx context.Context, source string) ([]DocumentMeta, error)
}

// DocumentMeta is the full document record stored in SQLite
type DocumentMeta struct {
	ID           string
	Title        string
	Source       string
	Path         string
	LastModified time.Time
	MimeType     string
	ContentHash  string
	Status       string
}
