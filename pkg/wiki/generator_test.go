package wiki

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rgehrsitz/AutoDoc/pkg/storage"
)

var ErrNotFound = errors.New("not found")

// MockStorage implements storage.Storage for testing
type MockStorage struct {
	docs      map[string]*storage.Document
	refs      map[string][]*storage.Reference
	backRefs  map[string][]*storage.Reference
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		docs:      make(map[string]*storage.Document),
		refs:      make(map[string][]*storage.Reference),
		backRefs:  make(map[string][]*storage.Reference),
	}
}

func (m *MockStorage) SaveDocument(doc *storage.Document) error {
	m.docs[doc.ID] = doc
	return nil
}

func (m *MockStorage) BatchSaveDocuments(docs []*storage.Document) error {
	for _, doc := range docs {
		m.docs[doc.ID] = doc
	}
	return nil
}

func (m *MockStorage) GetDocument(id string) (*storage.Document, error) {
	if doc, ok := m.docs[id]; ok {
		return doc, nil
	}
	return nil, ErrNotFound
}

func (m *MockStorage) ListDocuments(docType storage.DocumentType) ([]*storage.Document, error) {
	var docs []*storage.Document
	for _, doc := range m.docs {
		if doc.Type == docType {
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

func (m *MockStorage) SaveReference(ref *storage.Reference) error {
	m.refs[ref.SourceID] = append(m.refs[ref.SourceID], ref)
	m.backRefs[ref.TargetID] = append(m.backRefs[ref.TargetID], ref)
	return nil
}

func (m *MockStorage) BatchSaveReferences(refs []*storage.Reference) error {
	for _, ref := range refs {
		if err := m.SaveReference(ref); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockStorage) GetReferences(sourceID string) ([]*storage.Reference, error) {
	if refs, ok := m.refs[sourceID]; ok {
		return refs, nil
	}
	return nil, nil
}

func (m *MockStorage) GetBackReferences(targetID string) ([]*storage.Reference, error) {
	if refs, ok := m.backRefs[targetID]; ok {
		return refs, nil
	}
	return nil, nil
}

func (m *MockStorage) SearchSimilar(embedding []float64, limit int) ([]*storage.Document, error) {
	// For testing, just return all documents up to the limit
	var docs []*storage.Document
	for _, d := range m.docs {
		docs = append(docs, d)
		if len(docs) >= limit {
			break
		}
	}
	return docs, nil
}

func (m *MockStorage) Close() error {
	return nil
}

func TestGenerator(t *testing.T) {
	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "wiki-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock storage
	store := NewMockStorage()

	// Add test documents
	archDoc := &storage.Document{
		ID:        "arch1",
		Type:      storage.TypeArchitecture,
		Path:      "architecture.md",
		Content:   "# Architecture\n\nThis is the architecture overview.",
		UpdatedAt: time.Now(),
	}
	store.SaveDocument(archDoc)

	moduleDoc := &storage.Document{
		ID:        "mod1",
		Type:      storage.TypeModule,
		Path:      "pkg/example/example.go",
		Content:   "# Example Package\n\nThis is an example package.",
		UpdatedAt: time.Now(),
	}
	store.SaveDocument(moduleDoc)

	// Add test reference
	ref := &storage.Reference{
		SourceID: moduleDoc.ID,
		TargetID: archDoc.ID,
		Type:     "import", // Use string literal since we don't have access to the constant
	}
	store.SaveReference(ref)

	// Create generator
	gen := NewGenerator(store)

	// Generate wiki
	cfg := Config{
		OutputDir:    tmpDir,
		ProjectName:  "Test Project",
		ProjectURL:   "https://example.com/test",
		Theme:        "light",
	}

	if err := gen.Generate(cfg); err != nil {
		t.Fatalf("Failed to generate wiki: %v", err)
	}

	// Check if files were generated
	files := []string{
		"index.html",
		"architecture.html",
		filepath.Join("pkg", "example", "example.go.html"),
		"search.html",
		filepath.Join("assets", "style.css"),
		filepath.Join("assets", "search.js"),
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not generated", file)
		}
	}
}
