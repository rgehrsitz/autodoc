// autodoc/internal/analysis/project.go

package analyzer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rgehrsitz/AutoDoc/internal/collector"
	"github.com/rgehrsitz/AutoDoc/internal/storage"
)

// ProjectStructure represents the overall structure of the analyzed project
type ProjectStructure struct {
	Language   string             // Primary language (go, csharp)
	Type       string             // Project type (go-module, dotnet-solution)
	Root       string             // Root directory path
	Components []ProjectComponent // List of project components
	References []ProjectReference // Cross-component references
}

// ProjectComponent represents a major component in the project
type ProjectComponent struct {
	Path        string   // Relative path from project root
	Type        string   // Component type (package, project, assembly)
	Name        string   // Component name
	Description string   // Component description
	References  []string // Dependencies
	Files       []string // Source files in this component
}

// ProjectReference represents a relationship between components
type ProjectReference struct {
	Source      string // Source component path
	Target      string // Target component path
	Type        string // Reference type (imports, implements, etc.)
	Description string // Description of the relationship
}

// ProjectAnalyzer coordinates the analysis of an entire project
type ProjectAnalyzer struct {
	collector collector.Collector
	analyzer  *Analyzer
	storage   storage.Storage
}

// NewProjectAnalyzer creates a new ProjectAnalyzer instance
func NewProjectAnalyzer(collector collector.Collector, analyzer *Analyzer, storage storage.Storage) *ProjectAnalyzer {
	return &ProjectAnalyzer{
		collector: collector,
		analyzer:  analyzer,
		storage:   storage,
	}
}

// AnalyzeProject analyzes an entire project
func (p *ProjectAnalyzer) AnalyzeProject(ctx context.Context, path string) (*ProjectStructure, error) {
	// Collect all project files
	files, err := p.collector.CollectFiles(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to collect files: %w", err)
	}

	// Create initial project structure
	structure := &ProjectStructure{
		Root:       path,
		Language:   p.determineMainLanguage(files),
		Type:       p.determineProjectType(files),
		Components: []ProjectComponent{},
		References: []ProjectReference{},
	}

	// Group files into components
	components := p.groupFilesByComponent(path, files)

	// Analyze each component
	for i := range components {
		comp := &components[i]
		if err := p.analyzeComponent(ctx, comp, files); err != nil {
			return nil, fmt.Errorf("failed to analyze component %s: %w", comp.Path, err)
		}
	}

	// Store the analyzed components
	structure.Components = components

	return structure, nil
}

// analyzeComponent analyzes a single component and its files
func (p *ProjectAnalyzer) analyzeComponent(ctx context.Context, comp *ProjectComponent, files []collector.FileInfo) error {
	for _, filePath := range comp.Files {
		fileInfo := p.findFileInfo(files, filePath)
		if fileInfo == nil {
			continue
		}

		analysis, rawResponse, err := p.analyzer.AnalyzeFile(ctx, *fileInfo)
		if err != nil {
			// Log the raw response if there was an error
			if rawResponse != "" {
				return fmt.Errorf("failed to analyze %s (raw response: %s): %w", filePath, rawResponse, err)
			}
			return fmt.Errorf("failed to analyze %s: %w", filePath, err)
		}

		// Update component information based on analysis
		if comp.Description == "" {
			comp.Description = analysis.Purpose
		}

		// Add any references found in the file
		for _, ref := range analysis.Relations {
			if !contains(comp.References, ref.To) {
				comp.References = append(comp.References, ref.To)
			}
		}
	}

	return nil
}

// determineProjectType identifies the project type based on files
func (p *ProjectAnalyzer) determineProjectType(files []collector.FileInfo) string {
	for _, file := range files {
		switch {
		case strings.HasSuffix(file.Path, "go.mod"):
			return "go-module"
		case strings.HasSuffix(file.Path, ".sln"):
			return "dotnet-solution"
		}
	}
	return "unknown"
}

// determineMainLanguage identifies the primary language
func (p *ProjectAnalyzer) determineMainLanguage(files []collector.FileInfo) string {
	counts := make(map[string]int)
	for _, file := range files {
		counts[file.Language]++
	}

	maxCount := 0
	mainLang := "unknown"
	for lang, count := range counts {
		if count > maxCount {
			maxCount = count
			mainLang = lang
		}
	}
	return mainLang
}

// groupFilesByComponent organizes files into logical components
func (p *ProjectAnalyzer) groupFilesByComponent(root string, files []collector.FileInfo) []ProjectComponent {
	components := make(map[string]*ProjectComponent)

	for _, file := range files {
		// Get relative path from root
		relPath, err := filepath.Rel(root, file.Path)
		if err != nil {
			continue
		}

		// Determine component path (directory for Go, project file for C#)
		compPath := filepath.Dir(relPath)
		if strings.HasSuffix(file.Path, ".csproj") {
			compPath = relPath
		}

		// Create or update component
		comp, exists := components[compPath]
		if !exists {
			comp = &ProjectComponent{
				Path:  compPath,
				Type:  p.determineComponentType(file),
				Name:  filepath.Base(compPath),
				Files: []string{},
			}
			components[compPath] = comp
		}
		comp.Files = append(comp.Files, file.Path)
	}

	// Convert map to slice
	result := make([]ProjectComponent, 0, len(components))
	for _, comp := range components {
		result = append(result, *comp)
	}

	return result
}

// determineComponentType identifies the type of a component
func (p *ProjectAnalyzer) determineComponentType(file collector.FileInfo) string {
	switch {
	case file.Language == "go":
		return "package"
	case strings.HasSuffix(file.Path, ".csproj"):
		return "project"
	default:
		return "unknown"
	}
}

// findFileInfo finds the FileInfo for a given path
func (p *ProjectAnalyzer) findFileInfo(files []collector.FileInfo, path string) *collector.FileInfo {
	for _, file := range files {
		if file.Path == path {
			return &file
		}
	}
	return nil
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
