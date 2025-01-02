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

	// Walk through the project files
	err = filepath.Walk(*projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && extMap[filepath.Ext(path)] {
			code, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			doc, err := client.AnalyzeCode(ctx, string(code))
			if err != nil {
				return err
			}

			docMap[path] = doc
		}
		log.Fatalf("Failed to generate documentation: %v", err)
		return err
	})

	if err != nil {
		log.Fatalf("Error processing files: %v", err)
	}

	// Generate documentation
	err = docs.GenerateDocumentation(*projectPath, docMap)
	if err != nil {
		log.Fatalf("Failed to generate documentation: %v", err)
	}

}
