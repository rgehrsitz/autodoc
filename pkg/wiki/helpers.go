package wiki

import (
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/russross/blackfriday/v2"
)

// NavItem defines a single navigation entry
type NavItem struct {
	Title    string
	URL      string
	Active   bool
	Children []NavItem
}

// renderMarkdown converts markdown content to HTML
func renderMarkdown(content string) string {
	md := blackfriday.Run([]byte(content),
		blackfriday.WithExtensions(
			blackfriday.CommonExtensions|
				blackfriday.AutoHeadingIDs|
				blackfriday.NoEmptyLineBeforeBlock,
		),
	)
	return string(md)
}

// RenderTemplate renders a template with the provided data to the specified output path.
func RenderTemplate(outputPath, templateName string, data PageData) error {
	// Load the template from embedded files
	templatePath := "templates/layout.html"
	log.Printf("Attempting to load embedded template: %s", templatePath)

	tmpl, err := template.ParseFS(embeddedTemplates, templatePath)
	if err != nil {
		log.Printf("Error parsing template %s: %v", templatePath, err)
		return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Log output path
	log.Printf("Creating output file at: %s", outputPath)
	file, err := os.Create(outputPath)
	if err != nil {
		log.Printf("Error creating file %s: %v", outputPath, err)
		return fmt.Errorf("failed to create file %s: %w", outputPath, err)
	}
	defer file.Close()

	// Execute the template with the provided data
	log.Printf("Executing template for: %s", outputPath)
	if err := tmpl.Execute(file, data); err != nil {
		log.Printf("Error executing template for %s: %v", outputPath, err)
		return fmt.Errorf("failed to execute template for %s: %w", outputPath, err)
	}

	log.Printf("Successfully generated: %s", outputPath)
	return nil
}
