package main

import (
	"fmt"
	"os"

	"github.com/Low-Stack-Technologies/temp/cli/client"
	"github.com/Low-Stack-Technologies/temp/cli/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: temp <file1> [file2] ...")
		os.Exit(1)
	}

	filePaths := os.Args[1:]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Start upload
	if err := client.Upload(cfg, filePaths); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
