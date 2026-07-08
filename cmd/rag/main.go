package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sonurust/rag-llm/app"
	"github.com/sonurust/rag-llm/config"
)

func main() {
	// We need to:
	// Set up this app
	// set up config
	// set up an LLM client
	// set up the Readt-Eval-Print Loop (REPL)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx, config.Load()); err != nil {
		fmt.Println(os.Stderr, err)
		panic(err)
	}
}
