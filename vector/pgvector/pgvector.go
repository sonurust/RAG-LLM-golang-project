package pgvector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
	"github.com/sonurust/rag-llm/vector"
)

type Options struct {
	DSN          string
	EmbeddingDim int
}

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, opts Options) (*Store, error) {
	if opts.DSN == "" {
		return nil, fmt.Errorf("DSN is required")
	}
	if opts.EmbeddingDim <= 0 {
		return nil, fmt.Errorf("Embedding dimension is required")
	}
	cfg, err := pgxpool.ParseConfig(opts.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}
	if err := ensureExtension(ctx, opts.DSN); err != nil {
		return nil, fmt.Errorf("failed to ensure pgvector extension: %w", err)
	}
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	s := &Store{
		pool: pool,
	}
	if err := s.migrate(ctx, opts.EmbeddingDim); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return s, nil
}

func ensureExtension(ctx context.Context, dsn string) error {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	return err
}

func (s *Store) migrate(ctx context.Context, dim int) error {
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
			embedding VECTOR(%d) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`, dim),
		`CREATE INDEX IF NOT EXISTS documents_embedding_idx 
		ON documents USING hnsw (embedding vector_cosine_ops)`,
	}

	for _, q := range stmts {
		if _, err := s.pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("exec %q: %w", firstLine(q), err)
		}
	}
	return nil
}

func firstLine(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return s[:i]
		}
	}
	return s
}

func (s *Store) Upsert(ctx context.Context, docs []vector.Document) error {
	if len(docs) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	stmt := `INSERT INTO documents (id, content, metadata, embedding)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE SET content = EXCLUDED.content, metadata = EXCLUDED.metadata, embedding = EXCLUDED.embedding`

	for _, doc := range docs {
		meta, err := marshalMetadata(doc.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
		if _, err := tx.Exec(ctx, stmt, doc.ID, doc.Content, meta, pgvector.NewVector(doc.Embedding)); err != nil {
			return fmt.Errorf("exec upsert: %w", err)
		}

	}
	return tx.Commit(ctx)
}

func unmarshalMetadata(data []byte, dst *map[string]string) error {
	if len(data) == 0 {
		*dst = nil
		return nil
	}
	return json.Unmarshal(data, dst)
}

func (s *Store) Query(ctx context.Context, embedding []float32, topK int) ([]vector.Result, error) {
	if topK <= 0 {
		return nil, nil
	}
	const stmt = `SELECT id, content, metadata, embedding <=> $1 as distance
	FROM documents
	ORDER BY embedding <=> $1
	LIMIT $2`
	rows, err := s.pool.Query(ctx, stmt, pgvector.NewVector(embedding), topK)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var results []vector.Result
	for rows.Next() {
		var (
			r        vector.Result
			metaRaw  []byte
			distance float64
		)
		if err := rows.Scan(&r.Document.ID, &r.Document.Content, &metaRaw, &distance); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		if err := unmarshalMetadata(metaRaw, &r.Document.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
		r.Score = float32(1 - distance)
		results = append(results, r)
	}

	return results, rows.Err()
}

func (s *Store) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	stmt := `DELETE FROM documents WHERE id = ANY($1)`
	_, err := s.pool.Exec(ctx, stmt, ids)
	return err
}

func (s *Store) DeleteBySource(ctx context.Context, source string) error {
	if source == "" {
		return nil
	}
	stmt := `DELETE FROM documents WHERE metadata->>'source' = $1`
	_, err := s.pool.Exec(ctx, stmt, source)
	return err
}

func (s *Store) Close() error {
	if s.pool != nil {
		s.pool.Close()
	}
	return nil
}

func marshalMetadata(metadata map[string]string) ([]byte, error) {
	if len(metadata) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(metadata)
}
