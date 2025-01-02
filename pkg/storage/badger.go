package storage

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/dgraph-io/badger/v4"
)

type BadgerStorage struct {
	db *badger.DB
}

// NewBadgerStorage creates a new BadgerDB storage instance
func NewBadgerStorage(path string) (*BadgerStorage, error) {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &BadgerStorage{db: db}, nil
}

// Close closes the database connection
func (s *BadgerStorage) Close() error {
	return s.db.Close()
}

// SaveDocument saves a document to the database
func (s *BadgerStorage) SaveDocument(doc *Document) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("doc:"+doc.ID), data)
	})
}

// GetDocument retrieves a document from the database
func (s *BadgerStorage) GetDocument(id string) (*Document, error) {
	var doc Document
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("doc:" + id))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &doc)
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &doc, nil
}

// ListDocuments lists all documents of a specific type
func (s *BadgerStorage) ListDocuments(docType DocumentType) ([]*Document, error) {
	var docs []*Document

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("doc:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var doc Document
				if err := json.Unmarshal(val, &doc); err != nil {
					return err
				}
				if doc.Type == docType {
					docs = append(docs, &doc)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	return docs, nil
}

// SaveReference saves a reference to the database
func (s *BadgerStorage) SaveReference(ref *Reference) error {
	data, err := json.Marshal(ref)
	if err != nil {
		return fmt.Errorf("failed to marshal reference: %w", err)
	}

	key := fmt.Sprintf("ref:%s:%s:%s", ref.SourceID, ref.TargetID, ref.Type)
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// GetReferences gets all references from a source document
func (s *BadgerStorage) GetReferences(sourceID string) ([]*Reference, error) {
	var refs []*Reference

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(fmt.Sprintf("ref:%s:", sourceID))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var ref Reference
				if err := json.Unmarshal(val, &ref); err != nil {
					return err
				}
				refs = append(refs, &ref)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}

	return refs, nil
}

// GetBackReferences gets all references to a target document
func (s *BadgerStorage) GetBackReferences(targetID string) ([]*Reference, error) {
	var refs []*Reference

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("ref:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var ref Reference
				if err := json.Unmarshal(val, &ref); err != nil {
					return err
				}
				if ref.TargetID == targetID {
					refs = append(refs, &ref)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get back references: %w", err)
	}

	return refs, nil
}

// BatchSaveDocuments saves multiple documents in a batch
func (s *BadgerStorage) BatchSaveDocuments(docs []*Document) error {
	wb := s.db.NewWriteBatch()
	defer wb.Cancel()

	for _, doc := range docs {
		data, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %w", err)
		}

		err = wb.Set([]byte("doc:"+doc.ID), data)
		if err != nil {
			return fmt.Errorf("failed to batch set document: %w", err)
		}
	}

	return wb.Flush()
}

// BatchSaveReferences saves multiple references in a batch
func (s *BadgerStorage) BatchSaveReferences(refs []*Reference) error {
	wb := s.db.NewWriteBatch()
	defer wb.Cancel()

	for _, ref := range refs {
		data, err := json.Marshal(ref)
		if err != nil {
			return fmt.Errorf("failed to marshal reference: %w", err)
		}

		key := fmt.Sprintf("ref:%s:%s:%s", ref.SourceID, ref.TargetID, ref.Type)
		err = wb.Set([]byte(key), data)
		if err != nil {
			return fmt.Errorf("failed to batch set reference: %w", err)
		}
	}

	return wb.Flush()
}

// SearchSimilar finds documents with similar embeddings using cosine similarity
func (s *BadgerStorage) SearchSimilar(embedding []float64, limit int) ([]*Document, error) {
	var docs []*Document

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		type docWithSimilarity struct {
			doc        *Document
			similarity float64
		}

		var results []docWithSimilarity

		prefix := []byte("doc:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var doc Document
				if err := json.Unmarshal(val, &doc); err != nil {
					return err
				}
				if len(doc.Embedding) > 0 {
					similarity := cosineSimilarity(embedding, doc.Embedding)
					results = append(results, docWithSimilarity{&doc, similarity})
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		// Sort by similarity (descending)
		sort.Slice(results, func(i, j int) bool {
			return results[i].similarity > results[j].similarity
		})

		// Take top k results
		if len(results) > limit {
			results = results[:limit]
		}

		// Extract documents
		docs = make([]*Document, len(results))
		for i, r := range results {
			docs[i] = r.doc
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search similar documents: %w", err)
	}

	return docs, nil
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
