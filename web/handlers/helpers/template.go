// autodoc/web/handlers/helpers/template.go

package helpers

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
)

// RenderTemplate renders a template with the given data
func RenderTemplate(outputPath string, templateName string, data interface{}, templates embed.FS) error {
	// Log template loading attempt
	log.Printf("Attempting to load embedded template: templates/%s.html", templateName)

	// Create template from embedded files
	tmpl, err := template.New("layout.html").ParseFS(templates,
		"templates/layout.html",
		fmt.Sprintf("templates/%s.html", templateName))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file
	log.Printf("Creating output file at: %s", outputPath)
	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Execute template
	log.Printf("Executing template for: %s", outputPath)
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Printf("Successfully generated: %s", outputPath)
	return nil
}
