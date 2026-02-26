package search

import (
	"context"
	"fmt"
	"monolith/internal/index"
	"strings"
	"time"
)

type Query struct {
	Text    string
	Sources []string // nil = all sources
	Limit   int      // default 20
	Offset  int      // for pagination
	Fuzzy   int      // enable typo tolerance
}

type Result struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Source       string    `json:"source"`
	Path         string    `json:"path"`
	Snippet      string    `json:"snippet"`
	LastModified time.Time `json:"last_modified"`
	Score        float64   `json:"score"`
	MimeType     string    `json:"mime_type"`
}

type Response struct {
	Query   string   `json:"query"`
	Total   int      `json:"total"`
	Results []Result `json:"results"`
}

type Searcher struct {
	idx  index.Index
	meta index.MetadataStore // SQLite - source of truth for document metadata
}

func New(idx index.Index, meta index.MetadataStore) *Searcher {
	return &Searcher{idx: idx, meta: meta}
}

func (s *Searcher) Search(ctx context.Context, q Query) (Response, error) {
	if q.Limit == 0 {
		q.Limit = 20
	}

	// Execute against Bleve
	hits, err := s.idx.Search(ctx, s.buildQuery(q))
	if err != nil {
		return Response{}, fmt.Errorf("search: %w", err)
	}

	// Hydrate with SQLite metadata and also apply source filter
	results := make([]Result, 0, len(hits))
	for _, hit := range hits {
		meta, err := s.meta.Get(ctx, hit.ID)
		if err != nil {
			continue // doc indexed but metadata missing - skip gracefully
		}
		if !Contains(q.Sources, meta.Source) {
			continue
		}
		results = append(results, Result{
			ID:           hit.ID,
			Title:        meta.Title,
			Source:       meta.Source,
			Path:         meta.Path,
			Snippet:      hit.Snippet,
			LastModified: meta.LastModified,
			Score:        hit.Score,
			MimeType:     meta.MimeType,
		})
	}

	// Paginate after source filtering
	start := q.Offset
	if start > len(results) {
		start = len(results)
	}
	end := start + q.Limit
	if end > len(results) {
		end = len(results)
	}
	return Response{
		Query:   q.Text,
		Total:   len(results),
		Results: results[start:end],
	}, nil
}

func (s *Searcher) buildQuery(q Query) index.Query {
	return index.Query{
		Text:   strings.TrimSpace(q.Text),
		Limit:  q.Limit + q.Offset, // fetch enough for pagination slice
		Offset: 0,                  // always fetch from top; slice after source filter
		Fuzzy:  q.Fuzzy,
	}
}

func Contains(sources []string, target string) bool {
	if len(sources) == 0 {
		return true // no filter = all sources pass
	}
	for _, s := range sources {
		if strings.EqualFold(s, target) {
			return true
		}
	}
	return false
}
