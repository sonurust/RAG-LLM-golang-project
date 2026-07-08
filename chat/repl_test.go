package chat

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSeedHistoryCreatesMissingSystemPromptFile(t *testing.T) {
	tempDir := t.TempDir()
	promptPath := filepath.Join(tempDir, "documents", "system_prompt.md")

	messages, err := seedHistory(promptPath)
	if err != nil {
		t.Fatalf("seedHistory returned error: %v", err)
	}
	if messages != nil {
		t.Fatalf("expected no messages for a newly created prompt file, got %v", messages)
	}

	data, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("expected the prompt file to be created: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected an empty prompt file, got %q", string(data))
	}
}
