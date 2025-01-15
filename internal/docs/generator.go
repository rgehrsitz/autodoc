// autodoc/internal/docs/generator.go

package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateDocumentation creates Markdown documentation from the analyses map and references.
func GenerateDocumentation(projectPath string, analyses map[string]string, references map[string][]string) error {
	outputDir := "docs_out"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	var indexContent strings.Builder
	indexContent.WriteString("# Project Documentation\n\n")

	for path, doc := range analyses {
		relativePath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		ext := filepath.Ext(relativePath)
		language := strings.TrimPrefix(ext, ".")
		languageDir := filepath.Join(outputDir, language)
		if err := os.MkdirAll(languageDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create language directory %s: %v", languageDir, err)
		}

		sanitizedPath := strings.ReplaceAll(relativePath, string(filepath.Separator), "_")
		mdFileName := sanitizedPath + ".md"
		mdFilePath := filepath.Join(languageDir, mdFileName)

		// Write individual Markdown file
		f, err := os.Create(mdFilePath)
		if err != nil {
			return fmt.Errorf("failed to create Markdown file for %s: %v", path, err)
		}
		defer f.Close()

		_, err = f.WriteString(fmt.Sprintf("# Documentation for %s\n\n%s", relativePath, doc))
		if err != nil {
			return fmt.Errorf("failed to write to Markdown file for %s: %v", path, err)
		}

		// Add entry to INDEX.md
		indexContent.WriteString(fmt.Sprintf("- [%s](%s/%s)\n", relativePath, language, mdFileName))
	}

	// Generate Mermaid diagram
	diagram := buildMermaidDiagram(references)
	indexContent.WriteString("\n## Architecture Overview\n\n")
	indexContent.WriteString("```mermaid\n")
	indexContent.WriteString(diagram)
	indexContent.WriteString("\n```\n\n")

	// Write INDEX.md
	indexPath := filepath.Join(outputDir, "INDEX.md")
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("failed to create INDEX.md: %v", err)
	}
	defer indexFile.Close()

	_, err = indexFile.WriteString(indexContent.String())
	if err != nil {
		return fmt.Errorf("failed to write to INDEX.md: %v", err)
	}

	fmt.Printf("Documentation written to %s/\n", outputDir)
	return nil
}

// buildMermaidDiagram creates a Mermaid diagram representing file-to-file references.
func buildMermaidDiagram(references map[string][]string) string {
	// This is a simple placeholder. Feel free to enrich the logic.
	var b strings.Builder

	b.WriteString("graph LR\n")

	// references is expected to be map[filePath][]importedFilePaths
	for file, refs := range references {
		sanitizedFile := sanitizeMermaidNode(file)
		// If a file has no references, make sure it still appears as a node
		if len(refs) == 0 {
			b.WriteString(fmt.Sprintf("    %s\n", sanitizedFile))
		}
		for _, ref := range refs {
			sanitizedRef := sanitizeMermaidNode(ref)
			b.WriteString(fmt.Sprintf("    %s --> %s\n", sanitizedFile, sanitizedRef))
		}
	}

	return b.String()
}

// sanitizeMermaidNode replaces characters that might break Mermaid diagrams
func sanitizeMermaidNode(path string) string {
	// Replace path separators, etc. You can refine this logic as needed.
	return strings.ReplaceAll(path, string(filepath.Separator), "_")
}
