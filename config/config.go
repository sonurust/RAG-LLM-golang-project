package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL          string
	APIKey           string
	Model            string
	SystemPromptFile string

	DatabaseURL  string
	EmbeddingDim int
}

func Load() Config {
	_ = godotenv.Load() // Load environment variables from .env file if it exists

	cfg := Config{
		BaseURL:          os.Getenv("OPENAI_BASE_URL"),
		APIKey:           os.Getenv("OPENAI_API_KEY"),
		Model:            os.Getenv("OPENAI_MODEL"),
		SystemPromptFile: os.Getenv("SYSTEM_PROMPT_FILE"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		EmbeddingDim:     atoiOr(os.Getenv("EMBEDDING_DIM"), 0),
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4-mini"
	}

	if cfg.EmbeddingDim == 0 {
		cfg.EmbeddingDim = 768 // Default embedding dimension for OpenAI's text-embedding-3-small
	}

	return cfg
}

func atoiOr(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return n
}
