package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/YOUR_USERNAME/AutoDoc/pkg/openai"
)

func main() {
	projectPath := flag.String("path", ".", "Path to the project to document")
	flag.Parse()

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	client := openai.NewClient(config.OpenAIAPIKey)
	ctx := context.Background()

	err = filepath.Walk(*projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			code, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			doc, err := client.AnalyzeCode(ctx, string(code))
			if err != nil {
				return err
			}

			fmt.Printf("Documentation for %s:\n%s\n", path, doc)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error processing files: %v", err)
	}
}
