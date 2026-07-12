// file_read shows advanced manual construction. The agent is given file tools
// directly and asked a question that requires reading a file to answer.
//
//	OPENROUTER_API_KEY=... go run ./examples/file_read
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/snowztech/vikusha/core/agent"
	"github.com/snowztech/vikusha/core/llm"
	"github.com/snowztech/vikusha/core/tool"
	"github.com/snowztech/vikusha/core/tools/file"
)

func main() {
	_ = godotenv.Load()
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY not set")
	}

	reg := tool.NewRegistry()
	reg.Register(file.NewList())
	reg.Register(file.NewRead())

	a, err := agent.New(agent.Options{
		Name:         "reader",
		Model:        "openai/gpt-4o-mini",
		SystemPrompt: "You answer questions about files. Use file_read when you need the contents.",
		Provider:     llm.NewOpenRouter(apiKey),
		Tools:        reg,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	reply, err := a.Chat(ctx, "lucas", "Read go.mod and tell me the module name.")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}
