package wiki

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rgehrsitz/AutoDoc/pkg/storage"
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

// buildNavigation builds the navigation structure from a list of documents
func buildNavigation(docs []*storage.Document) []NavItem {
	// Group documents by directory
	groups := make(map[string][]NavItem)

	for _, doc := range docs {
		dir := filepath.Dir(doc.Path)
		if dir == "." {
			dir = ""
		}

		item := NavItem{
			Title: filepath.Base(doc.Path),
			URL:   pathToURL(doc.Path),
		}

		groups[dir] = append(groups[dir], item)
	}

	// Sort each group
	for _, items := range groups {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Title < items[j].Title
		})
	}

	// Build tree structure
	var nav []NavItem

	// Add architecture as first item
	nav = append(nav, NavItem{
		Title: "Architecture",
		URL:   "architecture.html",
	})

	// Add modules
	var dirs []string
	for dir := range groups {
		if dir != "" {
			dirs = append(dirs, dir)
		}
	}
	sort.Strings(dirs)

	// Add root files first
	if items, ok := groups[""]; ok {
		nav = append(nav, items...)
	}

	// Add directories
	for _, dir := range dirs {
		item := NavItem{
			Title:    filepath.Base(dir),
			Children: groups[dir],
		}
		nav = append(nav, item)
	}

	return nav
}

// pathToURL converts a file path to a URL-friendly format
func pathToURL(path string) string {
	url := strings.ReplaceAll(path, string(filepath.Separator), "/")
	return url + ".html"
}
