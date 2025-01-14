// autodoc/pkg/generator/generator_test.go

package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rgehrsitz/AutoDoc/pkg/storage"
)

// MockCollector implements the collect.Collector interface for testing
type MockCollector struct{}

func (m *MockCollector) Clone(ctx context.Context, repoURL string) (string, error) {
	return "test_repo", nil
}

func (m *MockCollector) ListFiles(ctx context.Context, path string) ([]string, error) {
	return []string{"test.go", "main.go"}, nil
}

func (m *MockCollector) ReadFile(path string) ([]byte, error) {
	return []byte(`package main

func main() {
    println("Hello, World!")
}`), nil
}

// MockOpenAIClient implements the necessary OpenAI client methods for testing
type MockOpenAIClient struct {
	Chunker MockChunker
}

type MockChunker struct{}

func (m MockChunker) Split(content string) []string {
	return []string{content}
}

func NewMockOpenAIClient() *MockOpenAIClient {
	return &MockOpenAIClient{
		Chunker: MockChunker{},
	}
}

func (m *MockOpenAIClient) AnalyzeSource(ctx context.Context, code string, language string) (string, error) {
	return "Test documentation for " + language + " code", nil
}

func (m *MockOpenAIClient) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	return []float64{0.1, 0.2, 0.3}, nil
}

func TestGenerator(t *testing.T) {
	// Create temporary directory for database
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize components
	collector := &MockCollector{}
	client := NewMockOpenAIClient()
	store, err := storage.NewBadgerStorage(filepath.Join(tmpDir, "db"))
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Initialize generator
	gen := NewGenerator(collector, client, store)

	// Test documentation generation
	ctx := context.Background()
	err = gen.GenerateDocumentation(ctx, "test_repo")
	if err != nil {
		t.Fatalf("Failed to generate documentation: %v", err)
	}

	// Verify documents were created
	docs, err := store.ListDocuments(storage.TypeModule)
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}

	if len(docs) != 2 { // We expect 2 files from our mock
		t.Errorf("Expected 2 documents, got %d", len(docs))
	}

	// Verify architecture document was created
	archDocs, err := store.ListDocuments(storage.TypeArchitecture)
	if err != nil {
		t.Fatalf("Failed to list architecture documents: %v", err)
	}

	if len(archDocs) != 1 {
		t.Errorf("Expected 1 architecture document, got %d", len(archDocs))
	}

	// Verify document content
	if docs[0].Content == "" {
		t.Error("Document content is empty")
	}

	// Verify embeddings
	if len(docs[0].Embedding) != 3 {
		t.Errorf("Expected embedding of length 3, got %d", len(docs[0].Embedding))
	}
}
