// cmd/test/main.go

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	analyzer "github.com/rgehrsitz/AutoDoc/internal/analysis"
	"github.com/rgehrsitz/AutoDoc/internal/collector"
	"github.com/rgehrsitz/AutoDoc/internal/storage"
	"github.com/rgehrsitz/AutoDoc/pkg/config"
)

// getProjectRoot returns the absolute path to the project root directory
func getProjectRoot() (string, error) {
	// Get the directory of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	// Navigate up two levels from cmd/test to reach project root
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absProjectRoot, nil
}

func main() {
	// Get project root path
	projectRoot, err := getProjectRoot()
	if err != nil {
		log.Fatalf("Failed to get project root: %v", err)
	}
	log.Printf("Project root: %s", projectRoot)

	// Create directory paths
	testDataDir := filepath.Join(projectRoot, "testdata")
	dbDir := filepath.Join(testDataDir, "db")
	sampleDir := filepath.Join(testDataDir, "sample")

	// Ensure directories exist
	dirs := []string{testDataDir, dbDir, sampleDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize storage with absolute path
	// Ensure complete database cleanup
	dbDir = filepath.Join(testDataDir, "db")
	if err := os.RemoveAll(dbDir); err != nil {
		log.Fatalf("Failed to remove existing database: %v", err)
	}

	// Recreate the directory to ensure a clean slate
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Initialize storage with the clean directory
	store, err := storage.NewBadgerStorage(dbDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// 3. Initialize collector
	collector := collector.NewCollector()

	// 4. Initialize analyzer
	openAIAnalyzer := analyzer.NewAnalyzer(cfg.OpenAIKey)

	// 5. Create sample files for testing
	if err := createSampleFiles(sampleDir); err != nil {
		log.Fatalf("Failed to create sample files: %v", err)
	}

	// 6. Collect files
	files, err := collector.CollectFiles(context.Background(), sampleDir)
	if err != nil {
		log.Fatalf("Failed to collect files: %v", err)
	}

	log.Printf("Found %d files to analyze", len(files))

	// 7. Analyze each file and store results
	for _, file := range files {
		log.Printf("Analyzing file: %s", file.Path)

		// Perform analysis
		analysis, rawResponse, err := openAIAnalyzer.AnalyzeFile(context.Background(), file)
		if err != nil {
			log.Printf("Error analyzing file %s: %v", file.Path, err)
			log.Printf("Raw response: %s", rawResponse)
			continue
		}

		// Debug: Print raw response
		log.Printf("Raw OpenAI response for %s:\n%s\n", file.Path, rawResponse)

		// Create document from analysis
		doc := &storage.Document{
			ID:         generateID(file.Path),
			Path:       file.Path,
			Type:       storage.TypeModule,
			Content:    analysis.Purpose,
			Purpose:    analysis.Purpose,
			Components: make([]storage.ComponentInfo, len(analysis.Components)),
			Relations:  make([]storage.RelationInfo, len(analysis.Relations)),
			Insights:   analysis.Insights,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// Convert components
		for i, comp := range analysis.Components {
			doc.Components[i] = storage.ComponentInfo{
				Name:            comp.Name,
				Type:            comp.Type,
				Description:     comp.Description,
				Visibility:      comp.Visibility,
				Dependencies:    comp.Dependencies,
				NotableFeatures: comp.NotableFeatures,
			}
		}

		// Convert relations
		for i, rel := range analysis.Relations {
			doc.Relations[i] = storage.RelationInfo{
				From: rel.From,
				To:   rel.To,
				Type: rel.Type,
			}
		}

		// Store document
		if err := store.SaveDocument(doc); err != nil {
			log.Printf("Error storing document %s: %v", file.Path, err)
			continue
		}

		// Initialize reference processor
		refProcessor := analyzer.NewReferenceProcessor(store)

		// Process and store references
		if err := refProcessor.ProcessReferences(doc, analysis); err != nil {
			log.Printf("Error processing references for %s: %v", file.Path, err)
			continue
		}

		log.Printf("Processed references for: %s", file.Path)

		// Print analysis results
		printAnalysis(file.Path, analysis)
	}

	// 8. Test cross-file reference retrieval
	log.Println("\nTesting cross-file references:")
	for _, file := range files {
		refs, err := store.GetReferences(generateID(file.Path))
		if err != nil {
			log.Printf("Error getting references for %s: %v", file.Path, err)
			continue
		}

		backRefs, err := store.GetBackReferences(generateID(file.Path))
		if err != nil {
			log.Printf("Error getting back references for %s: %v", file.Path, err)
			continue
		}

		log.Printf("\nFile: %s", file.Path)
		log.Printf("Dependencies (%d):", len(refs))
		for _, ref := range refs {
			refDoc, err := store.GetDocument(ref.TargetID)
			if err != nil {
				log.Printf("  - [Error getting target document: %v]", err)
				continue
			}
			log.Printf("  - %s (%s) -> %s", ref.Type, refDoc.Path, ref.TargetID)
		}

		log.Printf("Used by (%d):", len(backRefs))
		for _, ref := range backRefs {
			refDoc, err := store.GetDocument(ref.SourceID)
			if err != nil {
				log.Printf("  - [Error getting source document: %v]", err)
				continue
			}
			log.Printf("  - %s (%s) -> %s", ref.Type, refDoc.Path, ref.SourceID)
		}
	}
}

func createSampleFiles(sampleDir string) error {
	samples := map[string]string{
		"math\\calculator.go": `package math

type Calculator struct {}

func (c *Calculator) Add(a, b int) int {
    return a + b
}`,
		"math\\operations.go": `package math

type Operations interface {
    Add(a, b int) int
}`,
		"main.go": `package main

import "./math"

func main() {
    calc := &math.Calculator{}
    result := calc.Add(1, 2)
    println(result)
}`,
	}

	for path, content := range samples {
		fullPath := filepath.Join(sampleDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	return nil
}

func printAnalysis(path string, analysis *analyzer.Analysis) {
	log.Printf("\nAnalysis for: %s", path)
	log.Printf("Purpose: %s", analysis.Purpose)

	log.Println("Components:")
	for _, comp := range analysis.Components {
		log.Printf("  - %s (%s)", comp.Name, comp.Type)
		log.Printf("    Description: %s", comp.Description)
		if len(comp.Dependencies) > 0 {
			log.Printf("    Dependencies: %v", comp.Dependencies)
		}
	}

	if len(analysis.Relations) > 0 {
		log.Println("Relations:")
		for _, rel := range analysis.Relations {
			log.Printf("  - %s -> %s (%s)", rel.From, rel.To, rel.Type)
		}
	}

	if len(analysis.Insights) > 0 {
		log.Println("Insights:")
		for _, insight := range analysis.Insights {
			log.Printf("  - %s", insight)
		}
	}
}

func generateID(path string) string {
	// Get package name from path
	pkgName := filepath.Base(filepath.Dir(path))
	if pkgName == "." {
		pkgName = "main"
	}

	// Create stable ID string
	idStr := pkgName + ":" + filepath.Base(path)

	// Hash the string for a stable ID
	hash := sha256.Sum256([]byte(idStr))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for readable length
}
