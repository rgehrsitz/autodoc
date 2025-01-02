// autodoc/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rgehrsitz/AutoDoc/pkg/collect"
	"github.com/rgehrsitz/AutoDoc/pkg/generator"
	"github.com/rgehrsitz/AutoDoc/pkg/openai"
	"github.com/rgehrsitz/AutoDoc/pkg/storage"
)

func main() {
	// Print startup message
	fmt.Println("Starting AutoDoc...")

	// Define CLI flags
	repoURL := flag.String("repo", "", "URL of the repository to document")
	dbPath := flag.String("db", "autodoc.db", "Path to the database file")
	flag.Parse()

	if *repoURL == "" {
		log.Fatalf("Repository URL must be provided using the -repo flag.")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize components
	collector := collect.NewCollector()
	client := openai.NewClient(config.OpenAIKey)
	
	// Create storage directory if it doesn't exist
	storageDir := filepath.Dir(*dbPath)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// Initialize storage
	store, err := storage.NewBadgerStorage(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize generator
	gen := generator.NewGenerator(collector, client, store)

	// Generate documentation
	ctx := context.Background()
	if err := gen.GenerateDocumentation(ctx, *repoURL); err != nil {
		log.Fatalf("Failed to generate documentation: %v", err)
	}

	fmt.Println("Documentation generation complete!")
}
