// autodoc/internal/storage/badger_test.go

package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBadgerStorage(t *testing.T) {
	// Create temporary directory for database
	tmpDir, err := os.MkdirTemp("", "badger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize storage
	storage, err := NewBadgerStorage(filepath.Join(tmpDir, "db"))
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Test document operations
	doc := &Document{
		ID:         "test1",
		Path:       "/test/path",
		Type:       TypeModule,
		Content:    "Test content",
		Embedding:  []float64{0.1, 0.2, 0.3},
		References: []string{"ref1", "ref2"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Test SaveDocument
	if err := storage.SaveDocument(doc); err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	// Test GetDocument
	retrieved, err := storage.GetDocument(doc.ID)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Retrieved document is nil")
	}
	if retrieved.ID != doc.ID {
		t.Errorf("Expected ID %s, got %s", doc.ID, retrieved.ID)
	}

	// Test ListDocuments
	docs, err := storage.ListDocuments(TypeModule)
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("Expected 1 document, got %d", len(docs))
	}

	// Test reference operations
	ref := &Reference{
		SourceID:  "test1",
		TargetID:  "test2",
		Type:      "imports",
		CreatedAt: time.Now(),
	}

	// Test SaveReference
	if err := storage.SaveReference(ref); err != nil {
		t.Fatalf("Failed to save reference: %v", err)
	}

	// Test GetReferences
	refs, err := storage.GetReferences(ref.SourceID)
	if err != nil {
		t.Fatalf("Failed to get references: %v", err)
	}
	if len(refs) != 1 {
		t.Errorf("Expected 1 reference, got %d", len(refs))
	}

	// Test GetBackReferences
	backRefs, err := storage.GetBackReferences(ref.TargetID)
	if err != nil {
		t.Fatalf("Failed to get back references: %v", err)
	}
	if len(backRefs) != 1 {
		t.Errorf("Expected 1 back reference, got %d", len(backRefs))
	}

	// Test SearchSimilar
	similar, err := storage.SearchSimilar([]float64{0.1, 0.2, 0.3}, 10)
	if err != nil {
		t.Fatalf("Failed to search similar documents: %v", err)
	}
	if len(similar) != 1 {
		t.Errorf("Expected 1 similar document, got %d", len(similar))
	}
}
