// autodoc/internal/templateutil/rendering.go

package templateutil

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
	log.Printf("Attempting to load embedded template: %s.html", templateName)

	// Create template from embedded files
	log.Printf("Attempting to parse template: %s", templateName)
	tmpl, err := template.New("base.html").ParseFS(templates,
		"templates/layouts/base.html",
		"templates/partials/navigation.html",
		"templates/partials/breadcrumb.html",
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

	// Log template rendering attempt
	log.Printf("Attempting to render template: %s", templateName)

	// Execute template
	log.Printf("Executing template for: %s", outputPath)
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Printf("Successfully generated: %s", outputPath)
	return nil
}
