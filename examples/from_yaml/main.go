// from_yaml shows the high-level path: load a character YAML and chat with
// the resulting agent.
//
//	OPENAI_API_KEY=... go run ./examples/from_yaml
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	reply, err := a.Chat(ctx, "lucas", "Read go.mod and tell me the module name.")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}
