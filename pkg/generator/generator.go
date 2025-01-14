// autodoc/pkg/generator/generator.go

package generator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rgehrsitz/AutoDoc/pkg/analysis"
	"github.com/rgehrsitz/AutoDoc/pkg/collect"
	"github.com/rgehrsitz/AutoDoc/pkg/storage"
)

// Generator handles the documentation generation process
type Generator struct {
	collector collect.Collector
	client    OpenAIClient
	store     storage.Storage
}

// NewGenerator creates a new documentation generator
func NewGenerator(collector collect.Collector, client OpenAIClient, store storage.Storage) *Generator {
	return &Generator{
		collector: collector,
		client:    client,
		store:     store,
	}
}

// GenerateDocumentation generates documentation for a repository
func (g *Generator) GenerateDocumentation(ctx context.Context, repoURL string) error {
	// Clone repository
	repoPath, err := g.collector.Clone(ctx, repoURL)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// List all files
	files, err := g.collector.ListFiles(ctx, repoPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Warning: Directory %s does not exist, skipping", repoPath)
			return nil
		}
		return fmt.Errorf("failed to list files: %w", err)
	}

	// Process each file
	for _, file := range files {
		if err := g.processFile(ctx, file); err != nil {
			log.Printf("Warning: Error processing file %s: %v", file, err)
			continue // Continue with other files instead of stopping
		}
	}

	// Generate high-level documentation
	if err := g.generateHighLevelDocs(ctx); err != nil {
		return fmt.Errorf("failed to generate high-level documentation: %w", err)
	}

	return nil
}

// processFile processes a single file and generates its documentation
func (g *Generator) processFile(ctx context.Context, path string) error {
	// Generate a unique ID for the file
	fileID := generateID(path)

	// Extract metadata
	code, err := g.collector.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	ext := filepath.Ext(path)
	metadata := analysis.ExtractMetadata(string(code), ext)

	// Generate documentation using OpenAI
	doc, err := g.client.AnalyzeSource(ctx, string(code), strings.TrimPrefix(ext, "."))
	if err != nil {
		return fmt.Errorf("failed to analyze source: %w", err)
	}

	// Generate embedding for semantic search
	embedding, err := g.client.GenerateEmbedding(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create document
	document := &storage.Document{
		ID:         fileID,
		Path:       path,
		Type:       storage.TypeModule,
		Content:    doc,
		Embedding:  embedding,
		References: metadata.Imports,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save document
	if err := g.store.SaveDocument(document); err != nil {
		return fmt.Errorf("failed to save document: %w", err)
	}

	// Process references
	for _, imp := range metadata.Imports {
		ref := &storage.Reference{
			SourceID:  fileID,
			TargetID:  generateID(imp),
			Type:      "imports",
			CreatedAt: time.Now(),
		}
		if err := g.store.SaveReference(ref); err != nil {
			return fmt.Errorf("failed to save reference: %w", err)
		}
	}

	return nil
}

// generateHighLevelDocs generates high-level documentation for the entire project
func (g *Generator) generateHighLevelDocs(ctx context.Context) error {
	// Get all module documents
	docs, err := g.store.ListDocuments(storage.TypeModule)
	if err != nil {
		return fmt.Errorf("failed to list documents: %w", err)
	}

	// Combine all documentation for high-level analysis
	var combinedDocs strings.Builder
	for _, doc := range docs {
		combinedDocs.WriteString(fmt.Sprintf("File: %s\n", doc.Path))
		combinedDocs.WriteString(doc.Content)
		combinedDocs.WriteString("\n\n")
	}

	// Generate architecture documentation
	archDoc, err := g.client.AnalyzeSource(ctx, combinedDocs.String(), "architecture")
	if err != nil {
		return fmt.Errorf("failed to generate architecture documentation: %w", err)
	}

	// Generate embedding for the architecture document
	archEmbedding, err := g.client.GenerateEmbedding(ctx, archDoc)
	if err != nil {
		return fmt.Errorf("failed to generate architecture embedding: %w", err)
	}

	// Create architecture document
	archDocument := &storage.Document{
		ID:         generateID("architecture"),
		Path:       "architecture",
		Type:       storage.TypeArchitecture,
		Content:    archDoc,
		Embedding:  archEmbedding,
		References: []string{}, // Will be populated with file IDs
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Add references to all analyzed files
	for _, doc := range docs {
		archDocument.References = append(archDocument.References, doc.ID)
	}

	// Save architecture document
	if err := g.store.SaveDocument(archDocument); err != nil {
		return fmt.Errorf("failed to save architecture document: %w", err)
	}

	return nil
}

// generateID generates a unique ID for a path
func generateID(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars of hash
}
