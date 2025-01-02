// autodoc/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rgehrsitz/AutoDoc/pkg/analysis"
	"github.com/rgehrsitz/AutoDoc/pkg/collect"
	"github.com/rgehrsitz/AutoDoc/pkg/docs"
	"github.com/rgehrsitz/AutoDoc/pkg/openai"
)

func main() {
	// Print startup message
	fmt.Println("Starting AutoDoc...")

	// Define CLI flags
	repoURL := flag.String("repo", "", "URL of the repository to document")
	extensions := flag.String("extensions", ".js,.ts,.go,.rs,.py,.java", "Comma-separated list of file extensions to include")
	flag.Parse()

	if *repoURL == "" {
		log.Fatalf("Repository URL must be provided using the -repo flag.")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize OpenAI client
	client := openai.NewClient(config.OpenAIKey)
	fmt.Println("OpenAI client initialized.")

	ctx := context.Background()

	// Initialize Collector
	collector := collect.NewCollector()

	// Clone repository
	repoPath, err := collector.Clone(ctx, *repoURL)
	if err != nil {
		log.Fatalf("Failed to clone repository: %v", err)
	}
	fmt.Printf("Repository cloned to %s\n", repoPath)

	// Parse extensions
	extList := strings.Split(*extensions, ",")
	extMap := make(map[string]bool)
	for _, ext := range extList {
		trimmedExt := strings.TrimSpace(ext)
		if !strings.HasPrefix(trimmedExt, ".") {
			trimmedExt = "." + trimmedExt
		}
		extMap[trimmedExt] = true
	}

	// Map to hold documentation for each file
	docMap := make(map[string]string)

	// Initialize reference map
	references := make(map[string][]string)

	// Walk through the project files
	err = filepath.Walk(repoPath, func(pathStr string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v", pathStr, err)
			return err
		}
		if !info.IsDir() && extMap[filepath.Ext(pathStr)] {
			fmt.Println("Analyzing file:", pathStr)
			code, err := os.ReadFile(pathStr)
			if err != nil {
				log.Printf("Failed to read file %s: %v", pathStr, err)
				return err
			}

			// Extract metadata
			metadata := analysis.ExtractMetadata(string(code), filepath.Ext(pathStr))
			references[pathStr] = metadata.Imports

			// Split code into chunks
			chunks := client.Chunker.Split(string(code))
			fmt.Printf("File %s split into %d chunks\n", pathStr, len(chunks))

			// Generate documentation
			doc, err := client.AnalyzeSource(ctx, string(code), strings.TrimPrefix(filepath.Ext(pathStr), "."))
			if err != nil {
				return err
			}

			// Generate embedding
			embedding, err := client.GenerateEmbedding(ctx, doc)
			if err != nil {
				log.Printf("Failed to generate embedding for %s: %v", pathStr, err)
			}
			// TODO: Store embedding in a vector database
			_ = embedding

			docMap[pathStr] = doc

			fmt.Println("Documentation generated for:", pathStr)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error processing files: %v", err)
	}

	// Generate wiki documentation
	err = docs.GenerateDocumentation(repoPath, docMap, references)
	if err != nil {
		log.Fatalf("Failed to generate wiki documentation: %v", err)
	}

	fmt.Println("AutoDoc completed successfully.")
}
