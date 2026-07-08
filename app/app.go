package app

import (
	"context"

	"github.com/sonurust/rag-llm/chat"
	"github.com/sonurust/rag-llm/config"
	"github.com/sonurust/rag-llm/llm"
)

func Run(ctx context.Context, cfg config.Config) error {
	// Set up the LLM client
	client := llm.New(cfg)
	return chat.RunREPL(ctx, client, chat.Options{
		// Set up the REPL options here
		SystemPromptFile: cfg.SystemPromptFile,
	})
}
