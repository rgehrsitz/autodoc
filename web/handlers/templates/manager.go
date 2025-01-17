//autodoc/web/handlers/templates/manager.go

package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	analyzer "github.com/rgehrsitz/AutoDoc/internal/analysis"
)

//go:embed layouts/* partials/* components/* index.html assets/css/* assets/js/*
var templateFS embed.FS

// TemplateData represents the data passed to documentation templates
type TemplateData struct {
	Title       string
	ProjectName string
	Description string
	Version     string
	LastUpdated time.Time
	Components  []ComponentData
	Analysis    *analyzer.CodeAnalysisSchema
	Navigation  []NavigationItem
	CurrentPath string
	Theme       string
	Content     string
}

// ComponentData represents a component in the documentation
type ComponentData struct {
	Name          string
	Path          string
	Type          string
	Description   string
	Analysis      *analyzer.CodeAnalysisSchema
	SubComponents []ComponentData
	References    []ReferenceData
}

// ReferenceData represents a relationship between components
type ReferenceData struct {
	Source      string
	Target      string
	Type        string
	Description string
}

// NavigationItem represents a navigation menu item
type NavigationItem struct {
	Title    string
	URL      string
	Active   bool
	Children []NavigationItem
}

// TemplateEngine handles template rendering
type TemplateEngine struct {
	templates  *template.Template
	funcMap    template.FuncMap
	projectDir string
}

// NewTemplateEngine creates a new template engine instance
func NewTemplateEngine(projectDir string) (*TemplateEngine, error) {
	funcMap := template.FuncMap{
		"formatDate":     formatDate,
		"formatType":     formatType,
		"markdownToHTML": markdownToHTML,
		"highlightCode":  highlightCode,
		"relPath":        relPath,
		"isActive":       isActive,
		"hasChildren":    hasChildren,
		"impact":         formatImpact,
		"componentLink":  componentLink,
		"diagram":        generateDiagram,
	}

	log.Printf("Parsing templates from embedded filesystem")
	// Parse all templates
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS,
		"layouts/*.html",
		"partials/*.html",
		"partials/navigation.html",
		"components/*.html",
		"index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	log.Printf("Parsed templates successfully")

	return &TemplateEngine{
		templates:  tmpl,
		funcMap:    funcMap,
		projectDir: projectDir,
	}, nil
}

// Templates returns the underlying template set
func (e *TemplateEngine) Templates() *template.Template {
	return e.templates
}

// RenderPage renders a complete page using the layout template
func (e *TemplateEngine) RenderPage(data *TemplateData, layout, page string) (string, error) {
	var buf bytes.Buffer

	// First render the page content
	pageContent, err := e.RenderTemplate(page, data)
	if err != nil {
		return "", fmt.Errorf("failed to render page content: %w", err)
	}

	// Add rendered page content to data
	data.Content = pageContent

	// Render the complete layout
	if err := e.templates.ExecuteTemplate(&buf, layout, data); err != nil {
		return "", fmt.Errorf("failed to render layout: %w", err)
	}

	return buf.String(), nil
}

// RenderTemplate renders a specific template with data
func (e *TemplateEngine) RenderTemplate(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", name, err)
	}
	return buf.String(), nil
}

// CopyAssets copies static assets to the output directory
func (e *TemplateEngine) CopyAssets() error {
	// Create assets directory structure
	cssDir := filepath.Join(e.projectDir, "assets", "css")
	jsDir := filepath.Join(e.projectDir, "assets", "js")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		return fmt.Errorf("failed to create css directory: %w", err)
	}
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		return fmt.Errorf("failed to create js directory: %w", err)
	}

	// List of assets to copy
	assetFiles := []string{
		"assets/css/dark.css",
		"assets/css/light.css",
		"assets/css/style.css",
		"assets/js/search.js",
	}

	for _, asset := range assetFiles {
		// Read asset from embedded files
		data, err := templateFS.ReadFile(asset)
		if err != nil {
			log.Printf("Failed to read embedded asset %s: %v", asset, err)
			return fmt.Errorf("failed to read embedded asset %s: %w", asset, err)
		}

		// Determine the destination path
		destPath := filepath.Join(e.projectDir, asset)

		// Write the asset to the destination
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			log.Printf("Failed to write asset to %s: %v", destPath, err)
			return fmt.Errorf("failed to write asset to %s: %w", destPath, err)
		}

		log.Printf("Successfully copied asset to: %s", destPath)
	}

	return nil
}

// Template helper functions
func formatDate(t time.Time) string {
	return t.Format("January 2, 2006")
}

func formatType(t string) string {
	return strings.Title(strings.Replace(t, "_", " ", -1))
}

func markdownToHTML(md string) template.HTML {
	// Convert markdown to HTML - implementation depends on chosen markdown library
	// For this example, we'll just wrap in a paragraph
	return template.HTML(fmt.Sprintf("<p>%s</p>", md))
}

func highlightCode(code, language string) template.HTML {
	// Syntax highlighting implementation - could use a library like chroma
	// For this example, just wrap in a pre block
	return template.HTML(fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>",
		language, template.HTMLEscapeString(code)))
}

func relPath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

func isActive(currentPath, targetPath string) bool {
	return currentPath == targetPath
}

func hasChildren(item NavigationItem) bool {
	return len(item.Children) > 0
}

func formatImpact(impact string) template.HTML {
	var class string
	switch strings.ToLower(impact) {
	case "high":
		class = "badge-danger"
	case "medium":
		class = "badge-warning"
	case "low":
		class = "badge-info"
	default:
		class = "badge-secondary"
	}
	return template.HTML(fmt.Sprintf("<span class=\"badge %s\">%s</span>", class, impact))
}

func componentLink(path string) string {
	// Convert component path to documentation URL
	return fmt.Sprintf("/components/%s.html",
		strings.ReplaceAll(path, string(filepath.Separator), "-"))
}

func generateDiagram(components []ComponentData) template.HTML {
	// Generate mermaid diagram markup
	var b strings.Builder
	b.WriteString("```mermaid\ngraph TD\n")

	// Add nodes
	for _, comp := range components {
		b.WriteString(fmt.Sprintf("  %s[%s]\n",
			strings.Replace(comp.Path, "/", "_", -1),
			comp.Name))
	}

	// Add relationships
	for _, comp := range components {
		for _, ref := range comp.References {
			b.WriteString(fmt.Sprintf("  %s --> |%s| %s\n",
				strings.Replace(ref.Source, "/", "_", -1),
				ref.Type,
				strings.Replace(ref.Target, "/", "_", -1)))
		}
	}

	b.WriteString("```")
	return template.HTML(b.String())
}
