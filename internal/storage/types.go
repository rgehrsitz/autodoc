// internal/storage/types.go

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
	TypeAPI          DocumentType = "api"
)

// ComponentInfo represents a code component within a document
type ComponentInfo struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Description     string   `json:"description"`
	Visibility      string   `json:"visibility"`
	Dependencies    []string `json:"dependencies"`
	NotableFeatures []string `json:"notable_features"`
}

// RelationInfo represents a relationship between components
type RelationInfo struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// Document represents a piece of documentation
type Document struct {
	ID         string          `json:"id"`         // Unique identifier
	Path       string          `json:"path"`       // File path this document relates to
	Type       DocumentType    `json:"type"`       // Type of documentation
	Content    string          `json:"content"`    // The actual documentation content
	Purpose    string          `json:"purpose"`    // Brief description of the code's purpose
	Components []ComponentInfo `json:"components"` // List of components in this document
	Relations  []RelationInfo  `json:"relations"`  // List of relationships
	Insights   []string        `json:"insights"`   // Important observations
	Embedding  []float64       `json:"embedding"`  // Vector embedding for semantic search
	References []string        `json:"references"` // List of other document IDs this references
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// Reference represents a relationship between two pieces of code/documentation
type Reference struct {
	SourceID  string    `json:"source_id"` // ID of the source document
	TargetID  string    `json:"target_id"` // ID of the target document
	Type      string    `json:"type"`      // Type of reference (e.g., "imports", "calls", "implements")
	CreatedAt time.Time `json:"created_at"`
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
