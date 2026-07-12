// hello shows the normal Go path: load a character YAML and run one turn.
//
//	OPENROUTER_API_KEY=... go run ./examples/hello
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/snowztech/vikusha"
)

func main() {
	_ = godotenv.Load()
	a, err := vikusha.LoadAgent("examples/character.yaml", vikusha.BuildOptions{})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reply, err := a.Chat(ctx, "lucas", "Say hello and tell me what model you are.")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}
