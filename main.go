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
	err = filepath.Walk(*projectPath, func(pathStr string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && extMap[filepath.Ext(pathStr)] {
			code, err := os.ReadFile(pathStr)
			if err != nil {
				return err
			}

			ext := filepath.Ext(pathStr)
			language := strings.TrimPrefix(ext, ".")

			doc, err := client.AnalyzeSource(ctx, string(code), language)
			if err != nil {
				return err
			}

			docMap[pathStr] = doc
		}
		return nil
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
