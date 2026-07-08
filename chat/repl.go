package chat

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sonurust/rag-llm/llm"
)

type Options struct {
	SystemPromptFile string
}

func RunREPL(ctx context.Context, client *llm.Client, opts Options) error {
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 1024*1024) // Set a larger buffer size for input

	history, err := seedHistory(opts.SystemPromptFile)
	if err != nil {
		return err
	}

	fmt.Println("Welcome to the RAG-LLM REPL. Type your messages below. Press Q to exit.")
	for {
		fmt.Print("\n> ")
		if !in.Scan() {
			if err := in.Err(); err != nil {
				return err
			}
			return nil // EOF reached
		}

		input := strings.TrimSpace(in.Text())
		if input == "" {
			continue
		}
		if strings.EqualFold(input, "q") || strings.EqualFold(input, "/exit") || strings.EqualFold(input, "exit") || strings.EqualFold(input, "quit") {
			fmt.Println("Goodbye!")
			return nil
		}

		history = append(history, llm.Message{
			Role:    "user",
			Content: input,
		})
		spin := startSpinner("thinking")
		var stopOnce sync.Once
		reply, err := client.ChatStream(ctx, history, func(s string) {
			stopOnce.Do(spin.Stop)
			fmt.Print(s)
		})
		stopOnce.Do(spin.Stop)
		fmt.Println()

		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			history = history[:len(history)-1] // Remove the last user message on error
			continue
		}

		history = append(history, reply)
	}
}

type spinner struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once
}

func startSpinner(message string) *spinner {
	s := &spinner{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go func() {
		defer close(s.done)
		frames := []string{"|", "/", "-", "\\"}
		t := time.NewTicker(80 * time.Millisecond)
		defer t.Stop()
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K") // Clear the line
				return
			case <-t.C:
				fmt.Printf("\r%s %s", frames[i%len(frames)], message)
				i++
			}
		}
	}()

	return s
}

func (s *spinner) Stop() {
	s.once.Do(func() {
		close(s.stop)
		<-s.done
	})
}

func seedHistory(systemPromptFile string) ([]llm.Message, error) {
	if systemPromptFile == "" {
		return nil, nil
	}

	if err := ensureSystemPromptFile(systemPromptFile); err != nil {
		return nil, fmt.Errorf("prepare system prompt: %w", err)
	}

	data, err := os.ReadFile(systemPromptFile)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read system prompt: %w", err)
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, nil
	}
	return []llm.Message{
		{
			Role:    "system",
			Content: content,
		},
	}, nil
}

func ensureSystemPromptFile(path string) error {
	if path == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0o644)
	if err != nil {
		return err
	}
	return file.Close()
}
