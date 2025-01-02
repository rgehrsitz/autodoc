// autodoc/main.go
package main

import (
	"context"
	"crypto/sha256" // Added import for generateID
	"encoding/hex"  // Added import for generateID
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rgehrsitz/AutoDoc/pkg/analysis" // Ensure cache is imported if used
	"github.com/rgehrsitz/AutoDoc/pkg/collect"
	"github.com/rgehrsitz/AutoDoc/pkg/docs"
	"github.com/rgehrsitz/AutoDoc/pkg/openai"
	"github.com/rgehrsitz/AutoDoc/pkg/storage" // Imported the storage package
	"github.com/rgehrsitz/AutoDoc/pkg/wiki"    // Existing import
)

// generateID creates a unique ID based on the provided path
func generateID(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:])[:16] // Uses first 16 characters of the hash
}

func main() {
	// Print startup message
	fmt.Println("Starting AutoDoc...")

	// Define CLI flags
	repoURL := flag.String("repo", "", "URL of the repository to document")
	path := flag.String("path", "", "Path to the local repository to document") // Added flag
	extensions := flag.String("extensions", ".js,.ts,.go,.rs,.py,.java", "Comma-separated list of file extensions to include")
	flag.Parse()

	// Validate flags
	if *repoURL == "" && *path == "" {
		log.Fatalf("Either repository URL (-repo) or repository path (-path) must be provided.")
	}
	if *repoURL != "" && *path != "" {
		log.Fatalf("Please provide only one of -repo or -path flags, not both.")
	}

	// Load configuration using the LoadConfig from config.go
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

	var repoPath string
	if *path != "" {
		// Use the provided local path
		repoPath = *path
		fmt.Printf("Using local repository path: %s\n", repoPath)
	} else {
		// Clone repository
		repoPath, err = collector.Clone(ctx, *repoURL)
		if err != nil {
			log.Fatalf("Failed to clone repository: %v", err)
		}
		fmt.Printf("Repository cloned to %s\n", repoPath)
	}

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

	// Generate Markdown documentation
	err = docs.GenerateDocumentation(repoPath, docMap, references)
	if err != nil {
		log.Fatalf("Failed to generate Markdown documentation: %v", err)
	}

	fmt.Println("Markdown documentation generated successfully.")

	// Initialize Storage using NewBadgerStorage
	store, err := storage.NewBadgerStorage(filepath.Join(repoPath, "storage")) // Changed from "path/to/storage"
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Save documents and references to storage
	for path, doc := range docMap {
		document := &storage.Document{
			ID:         generateID(path),
			Path:       path,
			Type:       storage.TypeModule,
			Content:    doc,
			Embedding:  nil, // Add embedding if necessary
			References: references[path],
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := store.SaveDocument(document); err != nil {
			log.Printf("Failed to save document %s: %v", path, err)
		}
		// Save references
		for _, ref := range references[path] {
			reference := &storage.Reference{
				SourceID:  document.ID,
				TargetID:  generateID(ref),
				Type:      "import",
				CreatedAt: time.Now(),
			}
			if err := store.SaveReference(reference); err != nil {
				log.Printf("Failed to save reference from %s to %s: %v", path, ref, err)
			}
		}
	}

	// Generate Wiki documentation
	log.Println("Starting wiki documentation generation...")
	wikiGen := wiki.NewGenerator(store)
	wikiCfg := wiki.Config{
		OutputDir:    filepath.Join(repoPath, "wiki"), // Desired output directory
		ProjectName:  config.ProjectName,
		ProjectURL:   config.ProjectURL,
		Theme:        config.Theme,
		CustomStyles: config.CustomStyles, // Add if you have custom styles
		// No need to specify TemplatesDir if generator.go uses the correct path
	}

	if err := wikiGen.Generate(wikiCfg); err != nil {
		log.Fatalf("Failed to generate wiki documentation: %v", err)
	}

	log.Println("Wiki documentation generated successfully.")
}
