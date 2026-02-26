package index

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SQLiteMetaDataStore struct {
	db *sql.DB
}

func NewSQLiteMetaDataStore(path string) (*SQLiteMetaDataStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("sqlite migrate: %w", err)
	}
	return &SQLiteMetaDataStore{db: db}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS documents(
	id   TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	source TEXT NOT NULL,
	path TEXT NOT NULL,
	last_modified DATETIME NOT NULL,
	mime_type TEXT,
	content_hash TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'indexed'
	);
	CREATE INDEX IF NOT EXISTS idx_source ON documents(source);
	CREATE INDEX IF NOT EXISTS idx_hash ON documents(content_hash);
	`)
	return err
}
func (s *SQLiteMetaDataStore) Get(ctx context.Context, id string) (DocumentMeta, error) {
	row := s.db.QueryRowContext(ctx, `
	SELECT id, title, source, path, last_modified, mime_type, content_hash, status
	FROM documents WHERE id = ?`, id)

	var m DocumentMeta
	var lastMod string
	err := row.Scan(&m.ID, &m.Title, &m.Source, &m.Path,
		&lastMod, &m.MimeType, &m.ContentHash, &m.Status)
	if err == sql.ErrNoRows {
		return DocumentMeta{}, fmt.Errorf("metadata: %s not found", id)
	}
	if err != nil {
		return DocumentMeta{}, fmt.Errorf("metadata get: %w", err)
	}
	m.LastModified, _ = time.Parse(time.RFC3339, lastMod)
	return m, nil
}

func (s *SQLiteMetaDataStore) Upsert(ctx context.Context, m DocumentMeta) error {
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO documents (id, title, source, path, last_modified, mime_type, content_hash, status)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET
            title         = excluded.title,
            source        = excluded.source,
            path          = excluded.path,
            last_modified = excluded.last_modified,
            mime_type     = excluded.mime_type,
            content_hash  = excluded.content_hash,
            status        = excluded.status`,
		m.ID, m.Title, m.Source, m.Path,
		m.LastModified.Format(time.RFC3339),
		m.MimeType, m.ContentHash, m.Status,
	)
	return err
}

func (s *SQLiteMetaDataStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM documents WHERE id = ?`, id)
	return err
}

func (s *SQLiteMetaDataStore) ListBySource(ctx context.Context, source string) ([]DocumentMeta, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, title, source, path, last_modified, mime_type, content_hash, status
        FROM documents WHERE source = ?`, source)
	if err != nil {
		return nil, fmt.Errorf("metadata list: %w", err)
	}
	defer rows.Close()

	var docs []DocumentMeta
	for rows.Next() {
		var m DocumentMeta
		var lastMod string
		if err := rows.Scan(&m.ID, &m.Title, &m.Source, &m.Path,
			&lastMod, &m.MimeType, &m.ContentHash, &m.Status); err != nil {
			return nil, err
		}
		m.LastModified, _ = time.Parse(time.RFC3339, lastMod)
		docs = append(docs, m)
	}
	return docs, rows.Err()
}
