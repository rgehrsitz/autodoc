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
	"github.com/rgehrsitz/AutoDoc/pkg/docs"
	"github.com/rgehrsitz/AutoDoc/pkg/openai"
)

func main() {
	// Print startup message
	fmt.Println("Starting AutoDoc...")

	// Define CLI flags
	projectPath := flag.String("path", ".", "Path to the project to document")
	extensions := flag.String("extensions", ".js,.ts,.go,.rs,.py,.java", "Comma-separated list of file extensions to include")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize OpenAI client
	client := openai.NewClient(config.OpenAIKey)
	fmt.Println("OpenAI client initialized.")

	ctx := context.Background()

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
	err = filepath.Walk(*projectPath, func(pathStr string, info os.FileInfo, err error) error {
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

			// Analyze references
			fileRefs := analysis.ExtractReferences(string(code), filepath.Ext(pathStr))
			references[pathStr] = fileRefs

			chunks := client.Chunker.Split(string(code))
			fmt.Printf("File %s split into %d chunks\n", pathStr, len(chunks))

			ext := filepath.Ext(pathStr)
			language := strings.TrimPrefix(ext, ".")

			doc, err := client.AnalyzeSource(ctx, string(code), language)
			if err != nil {
				return err
			}

			docMap[pathStr] = doc

			fmt.Println("Documentation generated for:", pathStr)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error processing files: %v", err)
	}

	// Generate documentation
	err = docs.GenerateDocumentation(*projectPath, docMap, references)
	if err != nil {
		log.Fatalf("Failed to generate documentation: %v", err)
	}

}
