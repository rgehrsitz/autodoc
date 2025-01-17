// web/handlers/mock_storage.go

package handlers

import (
	"fmt"
	"sync"

	"github.com/rgehrsitz/AutoDoc/internal/storage"
)

// MockStorage implements the Storage interface for testing
type MockStorage struct {
	docs     map[string]*storage.Document
	refs     map[string][]*storage.Reference
	backRefs map[string][]*storage.Reference
	mu       sync.RWMutex
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		docs:     make(map[string]*storage.Document),
		refs:     make(map[string][]*storage.Reference),
		backRefs: make(map[string][]*storage.Reference),
	}
}

func (m *MockStorage) SaveDocument(doc *storage.Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.docs[doc.ID] = doc
	return nil
}

func (m *MockStorage) GetDocument(id string) (*storage.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if doc, exists := m.docs[id]; exists {
		return doc, nil
	}
	return nil, fmt.Errorf("document not found: %s", id)
}

func (m *MockStorage) ListDocuments(docType storage.DocumentType) ([]*storage.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var docs []*storage.Document
	for _, doc := range m.docs {
		if doc.Type == docType {
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

func (m *MockStorage) SaveReference(ref *storage.Reference) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refs[ref.SourceID] = append(m.refs[ref.SourceID], ref)
	m.backRefs[ref.TargetID] = append(m.backRefs[ref.TargetID], ref)
	return nil
}

func (m *MockStorage) GetReferences(sourceID string) ([]*storage.Reference, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.refs[sourceID], nil
}

func (m *MockStorage) GetBackReferences(targetID string) ([]*storage.Reference, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.backRefs[targetID], nil
}

func (m *MockStorage) SearchSimilar(embedding []float64, limit int) ([]*storage.Document, error) {
	return nil, nil
}

func (m *MockStorage) BatchSaveDocuments(docs []*storage.Document) error {
	for _, doc := range docs {
		if err := m.SaveDocument(doc); err != nil {
			return err
		}
	}
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
