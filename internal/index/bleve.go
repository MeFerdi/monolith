package index

import (
	"context"
	"fmt"
	"monolith/pkg/document"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

type BleveIndex struct {
	b bleve.Index
}

func NewBleveIndex(path string) (*BleveIndex, error) {
	b, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		mapping := bleve.NewIndexMapping()
		b, err = bleve.New(path, mapping)
	}
	if err != nil {
		return nil, fmt.Errorf("bleve open: %w", err)

	}
	return &BleveIndex{b: b}, nil
}
func (i *BleveIndex) Add(_ context.Context, doc document.Document) error {
	return i.b.Index(doc.ID, doc)
}
func (i *BleveIndex) Delete(_ context.Context, id string) error {
	return i.b.Delete(id)
}

func (i *BleveIndex) Search(_ context.Context, q Query) ([]SearchResult, error) {
	var bleveQuery query.Query
	if q.Fuzzy > 0 {
		fq := bleve.NewFuzzyQuery(q.Text)
		fq.Fuzziness = q.Fuzzy
		bleveQuery = fq
	} else {
		bleveQuery = bleve.NewMatchQuery(q.Text)
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}

	req := bleve.NewSearchRequestOptions(bleveQuery, limit, q.Offset, false)
	req.Highlight = bleve.NewHighlight()
	req.Fields = []string{"Title", "Source", "Path"}

	res, err := i.b.Search(req)
	if err != nil {
		return nil, fmt.Errorf("bleve search: %w", err)
	}
	out := make([]SearchResult, 0, len(res.Hits))
	for _, hit := range res.Hits {
		sr := SearchResult{
			ID:    hit.ID,
			Score: hit.Score,
		}
		if frags, ok := hit.Fragments["Content"]; ok && len(frags) > 0 {
			sr.Snippet = frags[0]
		}
		out = append(out, sr)
	}
	return out, nil
}
