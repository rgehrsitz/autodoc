// autodoc/internal/generator/generator.go

package generator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rgehrsitz/AutoDoc/internal/storage"
)

// Generator handles the documentation generation process.
type Generator struct {
	store  storage.Storage
	openai *OpenAIClient
}

// NewGenerator creates a new Generator instance.
func NewGenerator(store storage.Storage, openaiKey string) *Generator {
	return &Generator{
		store:  store,
		openai: NewOpenAIClient(openaiKey),
	}
}

// ProcessFile processes a single file and generates its documentation.
func (g *Generator) ProcessFile(ctx context.Context, path string, content []byte) error {
	// Extract file extension
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if ext == "" {
		return fmt.Errorf("file has no extension: %s", path)
	}

	// Generate documentation
	doc, err := g.openai.AnalyzeSource(ctx, string(content), ext)
	if err != nil {
		return fmt.Errorf("failed to analyze source: %w", err)
	}

	// Generate a unique ID for the file
	fileID := generateID(path)

	// Store the documentation
	document := &storage.Document{
		ID:         fileID,
		Path:       path,
		Type:       storage.TypeModule,
		Content:    doc,
		References: []string{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := g.store.SaveDocument(document); err != nil {
		return fmt.Errorf("failed to save document: %w", err)
	}

	return nil
}

// ProcessDirectory processes all files in a directory recursively.
func (g *Generator) ProcessDirectory(ctx context.Context, dir string, extensions []string) error {
	// Create extension map for faster lookup
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extMap[ext] = true
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if !extMap[ext] {
			return nil
		}

		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			content, err := os.ReadFile(filePath)
			if err != nil {
				errChan <- fmt.Errorf("failed to read file %s: %w", filePath, err)
				return
			}

			if err := g.ProcessFile(ctx, filePath, content); err != nil {
				errChan <- fmt.Errorf("failed to process file %s: %w", filePath, err)
				return
			}

			log.Printf("Processed file: %s", filePath)
		}(path)

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// generateID generates a unique ID for a path
func generateID(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars of hash
}
