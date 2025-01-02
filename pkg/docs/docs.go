package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateDocumentation creates Markdown documentation from the analyses map.
func GenerateDocumentation(projectPath string, analyses map[string]string) error {
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

	// Write INDEX.md
	indexPath := filepath.Join(outputDir, "INDEX.md")
	f, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("failed to create INDEX.md: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(indexContent.String())
	if err != nil {
		return fmt.Errorf("failed to write INDEX.md: %v", err)
	}

	fmt.Printf("Documentation written to ./%s/\n", outputDir)
	return nil
}
