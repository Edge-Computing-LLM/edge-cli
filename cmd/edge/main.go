package main

import (
	"fmt"
	"os"

	"github.com/Edge-Computing-LLM/edge-cli/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
