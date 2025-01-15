//autodoc/internal/templates/templates.go

package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	analyzer "github.com/rgehrsitz/AutoDoc/internal/analysis"
)

//go:embed layouts/* partials/* components/*
var templateFS embed.FS

// TemplateData represents the data passed to documentation templates
type TemplateData struct {
	Title       string
	ProjectName string
	Description string
	Version     string
	LastUpdated time.Time
	Components  []ComponentData
	Analysis    *analyzer.EnhancedAnalysis
	Navigation  []NavigationItem
	CurrentPath string
	Theme       string
}

// ComponentData represents a single component in the documentation
type ComponentData struct {
	Name        string
	Path        string
	Type        string
	Description string
	Content     string
	Analysis    *analyzer.EnhancedAnalysis
	References  []ReferenceData
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

	// Parse all templates
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS,
		"layouts/*.html",
		"partials/*.html",
		"components/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &TemplateEngine{
		templates:  tmpl,
		funcMap:    funcMap,
		projectDir: projectDir,
	}, nil
}

// RenderPage renders a complete page using the layout template
func (e *TemplateEngine) RenderPage(data *TemplateData, layout, page string) (string, error) {
	var buf bytes.Buffer

	// First render the page content
	pageContent, err := e.renderTemplate(page, data)
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
func (e *TemplateEngine) renderTemplate(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", name, err)
	}
	return buf.String(), nil
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
