package app

import (
	"context"
	"log"
	"os"

	"github.com/sonurust/rag-llm/chat"
	"github.com/sonurust/rag-llm/config"
	"github.com/sonurust/rag-llm/llm"
	"github.com/sonurust/rag-llm/vector"
	"github.com/sonurust/rag-llm/vector/pgvector"
)

func Run(ctx context.Context, cfg config.Config) error {
	// Set up the LLM client
	logger := log.New(os.Stderr, "[rag]: ", log.LstdFlags)
	client := llm.New(cfg)

	store, err := openStore(ctx, cfg)
	if err != nil {
		return err
	}
	if store != nil {
		defer store.Close()
		logger.Println("Connected to vector store")
	}

	return chat.RunREPL(ctx, client, chat.Options{
		// Set up the REPL options here
		SystemPromptFile: cfg.SystemPromptFile,
	})
}

func openStore(ctx context.Context, cfg config.Config) (vector.Store, error) {
	if cfg.DatabaseURL == "" {
		return nil, nil // No database configured, return nil store
	}
	s, err := pgvector.New(ctx, pgvector.Options{
		DSN:          cfg.DatabaseURL,
		EmbeddingDim: cfg.EmbeddingDim,
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}
