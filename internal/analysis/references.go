// internal/analysis/references.go

package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rgehrsitz/AutoDoc/internal/storage"
)

// generateReferenceKey creates a stable, unique identifier for a reference
func generateReferenceKey(sourceID, targetID, refType string) string {
	// Create a consistent hash that captures all meaningful properties
	hashInput := fmt.Sprintf("%s:%s:%s",
		sourceID,
		targetID,
		normalizeRelationType(refType),
	)

	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes of hash
}

// normalizeRelationType standardizes relationship types to reduce duplication
func normalizeRelationType(relType string) string {
	// Comprehensive mapping to consolidate similar relationship types
	relationMap := map[string]string{
		// Instantiation variations
		"instantiation": "creates",
		"instantiates":  "creates",
		"usage":         "uses",
		"uses":          "uses",

		// Method-related variations
		"method call":       "calls",
		"method invocation": "calls",
		"invokes":           "calls",
		"calls":             "calls",
		"containment":       "contains",

		// Structural relationships
		"method-of":                "belongs-to",
		"method-to-struct":         "belongs-to",
		"method-belongs-to":        "belongs-to",
		"method-belongs-to-struct": "belongs-to",
		"association":              "relates",
		"has-method":               "defines",
		"defines method":           "defines",

		// Implementation relationships
		"implements": "implements",
	}

	// Convert to lowercase for case-insensitive matching
	normalized := strings.ToLower(strings.TrimSpace(relType))

	// Check for exact match first
	if mappedType, exists := relationMap[normalized]; exists {
		return mappedType
	}

	// If no exact match, try partial match
	for key, value := range relationMap {
		if strings.Contains(normalized, key) {
			return value
		}
	}

	// If no mapping found, return original in lowercase
	return normalized
}

// ReferenceProcessor handles the extraction and storage of cross-file references
type ReferenceProcessor struct {
	store storage.Storage
	// Removed sync.Map as it wasn't being used effectively
}

// NewReferenceProcessor creates a new ReferenceProcessor instance
func NewReferenceProcessor(store storage.Storage) *ReferenceProcessor {
	return &ReferenceProcessor{
		store: store,
	}
}

// ProcessReferences coordinates the processing of different types of references
func (r *ReferenceProcessor) ProcessReferences(doc *storage.Document, analysis *Analysis) error {
	// Track unique references more comprehensively
	processedRefs := make(map[string]bool)

	// Process imports and relationships
	if err := r.processImports(doc, analysis, processedRefs); err != nil {
		return fmt.Errorf("error processing imports: %w", err)
	}

	if err := r.processRelationships(doc, analysis, processedRefs); err != nil {
		return fmt.Errorf("error processing relationships: %w", err)
	}

	// Optional: Process interface implementations if needed
	if err := r.processInterfaceImplementations(doc, analysis, processedRefs); err != nil {
		return fmt.Errorf("error processing interface implementations: %w", err)
	}

	return nil
}

// processImports handles package-level import references
func (r *ReferenceProcessor) processImports(doc *storage.Document, analysis *Analysis, processedRefs map[string]bool) error {
	importedPackages := make(map[string]bool)

	for _, comp := range analysis.Components {
		for _, dep := range comp.Dependencies {
			// Skip empty or internal dependencies
			if dep == "" || !strings.Contains(dep, ".") {
				continue
			}

			// Extract package name
			parts := strings.Split(dep, ".")
			pkgName := parts[0]

			// Avoid duplicate package imports
			if importedPackages[pkgName] {
				continue
			}
			importedPackages[pkgName] = true

			// Attempt to resolve package path
			basePath := filepath.Dir(doc.Path)
			possiblePaths := []string{
				filepath.Join(basePath, "..", pkgName),
				filepath.Join(basePath, pkgName),
			}

			var targetPath string
			for _, path := range possiblePaths {
				if _, err := os.Stat(path); err == nil {
					targetPath = path
					break
				}
			}

			if targetPath == "" {
				log.Printf("Warning: Could not resolve import path for package %s", pkgName)
				continue
			}

			// Generate a unique reference key
			refKey := generateReferenceKey(doc.ID, targetPath, "imports")

			// Prevent duplicate references
			if processedRefs[refKey] {
				continue
			}

			// Create and save the reference
			ref := &storage.Reference{
				SourceID:  doc.ID,
				TargetID:  targetPath,
				Type:      "imports",
				CreatedAt: time.Now(),
			}

			if err := r.store.SaveReference(ref); err != nil {
				return err
			}

			processedRefs[refKey] = true
			log.Printf("Added import reference: %s -> %s", doc.Path, targetPath)
		}
	}
	return nil
}

// processRelationships handles relationships between components
func (r *ReferenceProcessor) processRelationships(doc *storage.Document, analysis *Analysis, processedRefs map[string]bool) error {
	for _, rel := range analysis.Relations {
		if rel.From == "" || rel.To == "" {
			continue
		}

		// Normalize relationship type
		normalizedType := normalizeRelationType(rel.Type)

		// Resolve target path
		basePath := filepath.Dir(doc.Path)
		targetPath := filepath.Join(basePath, rel.To)

		// Fallback to simple path if file doesn't exist
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			targetPath = rel.To
		}

		// Generate unique reference key
		refKey := generateReferenceKey(doc.ID, targetPath, normalizedType)

		// Prevent duplicate references
		if processedRefs[refKey] {
			continue
		}

		// Create and save reference
		ref := &storage.Reference{
			SourceID:  doc.ID,
			TargetID:  targetPath,
			Type:      normalizedType,
			CreatedAt: time.Now(),
		}

		if err := r.store.SaveReference(ref); err != nil {
			return err
		}

		processedRefs[refKey] = true
		log.Printf("Added relationship reference: %s -%s-> %s",
			filepath.Base(doc.Path), normalizedType, filepath.Base(targetPath))
	}
	return nil
}

// processInterfaceImplementations checks for interface implementations
func (r *ReferenceProcessor) processInterfaceImplementations(doc *storage.Document, analysis *Analysis, processedRefs map[string]bool) error {
	for _, comp := range analysis.Components {
		if strings.EqualFold(comp.Type, "struct") {
			// Find interfaces in the same package
			pkgPath := filepath.Dir(doc.Path)
			interfaces, err := r.findInterfaces(pkgPath)
			if err != nil {
				return fmt.Errorf("error finding interfaces: %w", err)
			}

			// Check each interface for potential implementation
			for _, iface := range interfaces {
				if r.implementsInterface(doc, iface) {
					// Generate unique reference key
					refKey := generateReferenceKey(doc.ID, iface.ID, "implements")

					// Prevent duplicate references
					if processedRefs[refKey] {
						continue
					}

					// Create and save reference
					ref := &storage.Reference{
						SourceID:  doc.ID,
						TargetID:  iface.ID,
						Type:      "implements",
						CreatedAt: time.Now(),
					}

					if err := r.store.SaveReference(ref); err != nil {
						return err
					}

					processedRefs[refKey] = true
					log.Printf("Added interface implementation: %s implements %s",
						comp.Name, filepath.Base(iface.Path))
				}
			}
		}
	}
	return nil
}

// findInterfaces retrieves interface definitions from a package
func (r *ReferenceProcessor) findInterfaces(pkgPath string) ([]*storage.Document, error) {
	docs, err := r.store.ListDocuments(storage.TypeModule)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	var interfaces []*storage.Document
	for _, doc := range docs {
		if filepath.Dir(doc.Path) == pkgPath {
			for _, comp := range doc.Components {
				if strings.EqualFold(comp.Type, "interface") {
					interfaces = append(interfaces, doc)
					break
				}
			}
		}
	}

	return interfaces, nil
}

// implementsInterface checks if a document represents a type that implements an interface
func (r *ReferenceProcessor) implementsInterface(doc *storage.Document, iface *storage.Document) bool {
	for _, ifaceMethod := range iface.Components {
		if strings.EqualFold(ifaceMethod.Type, "method") {
			methodImplemented := false

			for _, docMethod := range doc.Components {
				if strings.EqualFold(docMethod.Type, "method") {
					if strings.EqualFold(ifaceMethod.Name, docMethod.Name) {
						methodImplemented = true
						break
					}
				}
			}

			// If any required method is not implemented, return false
			if !methodImplemented {
				return false
			}
		}
	}

	return true
}
