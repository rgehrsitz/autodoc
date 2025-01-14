// autodoc/pkg/storage/types.go

package storage

import (
	"time"
)

// DocumentType represents the type of documentation
type DocumentType string

const (
	TypeArchitecture DocumentType = "architecture"
	TypeModule       DocumentType = "module"
	TypeFunction     DocumentType = "function"
	TypeClass        DocumentType = "class"
	TypeAPI         DocumentType = "api"
)

// Document represents a piece of documentation
type Document struct {
	ID          string       // Unique identifier
	Path        string       // File path this document relates to
	Type        DocumentType // Type of documentation
	Content     string       // The actual documentation content
	Embedding   []float64    // Vector embedding for semantic search
	References  []string     // List of other document IDs this references
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Reference represents a relationship between two pieces of code/documentation
type Reference struct {
	SourceID      string    // ID of the source document
	TargetID      string    // ID of the target document
	Type          string    // Type of reference (e.g., "imports", "calls", "implements")
	CreatedAt     time.Time
}

// Storage interface defines the methods required for storing and retrieving documentation
type Storage interface {
	// Document operations
	SaveDocument(doc *Document) error
	GetDocument(id string) (*Document, error)
	ListDocuments(docType DocumentType) ([]*Document, error)
	SearchSimilar(embedding []float64, limit int) ([]*Document, error)
	
	// Reference operations
	SaveReference(ref *Reference) error
	GetReferences(sourceID string) ([]*Reference, error)
	GetBackReferences(targetID string) ([]*Reference, error)
	
	// Batch operations
	BatchSaveDocuments(docs []*Document) error
	BatchSaveReferences(refs []*Reference) error
}
