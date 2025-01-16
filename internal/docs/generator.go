// autodoc/internal/docs/generator.go

package docs

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	analyzer "github.com/rgehrsitz/AutoDoc/internal/analysis"
	"github.com/rgehrsitz/AutoDoc/web/handlers/templates"
	"encoding/json"
	"log"
)

// DocumentationGenerator handles the generation of documentation
type DocumentationGenerator struct {
	outDir      string
	projectName string
	templates   *template.Template
}

// Config holds configuration for documentation generation
type Config struct {
	OutputDir    string            // Directory where docs will be generated
	ProjectName  string            // Name of the project
	TemplatePath string            // Path to HTML templates
	CustomStyles map[string]string // Optional custom CSS styles
	Theme        string            // Theme name (e.g., "light", "dark")
}

// PageData represents the data passed to documentation templates
type PageData struct {
	Title       string
	ProjectName string
	Content     template.HTML
	Navigation  []NavItem
	LastUpdated time.Time
	Theme       string
}

// NavItem represents a navigation menu item
type NavItem struct {
	Title    string
	URL      string
	Active   bool
	Children []NavItem
}

// NewDocumentationGenerator creates a new documentation generator
func NewDocumentationGenerator(config Config) (*DocumentationGenerator, error) {
	// Create base template from layout
	tmpl, err := template.New("layout").ParseFiles(
		filepath.Join(config.TemplatePath, "layout.html"),
		filepath.Join(config.TemplatePath, "index.html"),
		filepath.Join(config.TemplatePath, "page.html"),
		filepath.Join(config.TemplatePath, "search.html"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &DocumentationGenerator{
		outDir:      config.OutputDir,
		projectName: config.ProjectName,
		templates:   tmpl,
	}, nil
}

// Generate generates the complete documentation site
func (g *DocumentationGenerator) Generate(structure *analyzer.ProjectStructure) error {
	if err := os.MkdirAll(g.outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	nav := g.buildNavigation(structure)

	if err := g.generatePages(structure, nav); err != nil {
		return fmt.Errorf("failed to generate pages: %w", err)
	}

	if err := g.copyAssets(); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	return nil
}

// generatePages generates all documentation pages
func (g *DocumentationGenerator) generatePages(structure *analyzer.ProjectStructure, nav []NavItem) error {
	if err := g.generateIndexPage(structure, nav); err != nil {
		return fmt.Errorf("failed to generate index page: %w", err)
	}

	if err := g.generateArchitecturePage(structure, nav); err != nil {
		return fmt.Errorf("failed to generate architecture page: %w", err)
	}

	for _, comp := range structure.Components {
		if err := g.generateComponentPage(comp, structure, nav); err != nil {
			return fmt.Errorf("failed to generate component page %s: %w", comp.Path, err)
		}
	}

	return nil
}

// generateIndexPage creates the main index.html
func (g *DocumentationGenerator) generateIndexPage(structure *analyzer.ProjectStructure, nav []NavItem) error {
	content := strings.Builder{}
	content.WriteString(fmt.Sprintf("# %s Documentation\n\n", g.projectName))
	content.WriteString(fmt.Sprintf("**Project Type:** %s  \n", structure.Type))
	content.WriteString(fmt.Sprintf("**Primary Language:** %s\n\n", structure.Language))

	content.WriteString("## Project Structure\n\n")
	for _, comp := range structure.Components {
		content.WriteString(fmt.Sprintf("### [%s](components/%s.html)\n", comp.Name, g.sanitizePath(comp.Path)))
		if comp.Description != "" {
			content.WriteString(comp.Description + "\n\n")
		}
	}

	return g.renderPage("index.html", "Home", content.String(), nav)
}

// generateArchitecturePage creates the architecture overview page
func (g *DocumentationGenerator) generateArchitecturePage(structure *analyzer.ProjectStructure, nav []NavItem) error {
	content := strings.Builder{}
	content.WriteString("# Architecture Overview\n\n")
	content.WriteString(fmt.Sprintf("This is a %s project primarily written in %s.\n\n", structure.Type, structure.Language))

	content.WriteString("## Component Diagram\n\n")
	content.WriteString("```mermaid\ngraph TD\n")
	for _, ref := range structure.References {
		content.WriteString(fmt.Sprintf("    %s-->|%s|%s\n",
			g.sanitizePath(ref.Source),
			ref.Type,
			g.sanitizePath(ref.Target)))
	}
	content.WriteString("```\n\n")

	content.WriteString("## Components\n\n")
	for _, comp := range structure.Components {
		content.WriteString(fmt.Sprintf("### %s\n", comp.Name))
		content.WriteString(fmt.Sprintf("**Type:** %s\n\n", comp.Type))
		if comp.Description != "" {
			content.WriteString(comp.Description + "\n\n")
		}
		if len(comp.References) > 0 {
			content.WriteString("**Dependencies:**\n\n")
			for _, ref := range comp.References {
				content.WriteString(fmt.Sprintf("- %s\n", ref))
			}
			content.WriteString("\n")
		}
	}

	return g.renderPage("architecture.html", "Architecture", content.String(), nav)
}

// generateComponentPage creates documentation for a component
func (g *DocumentationGenerator) generateComponentPage(comp analyzer.ProjectComponent, structure *analyzer.ProjectStructure, nav []NavItem) error {
	content := strings.Builder{}
	content.WriteString(fmt.Sprintf("# %s\n\n", comp.Name))
	content.WriteString(fmt.Sprintf("**Type:** %s\n\n", comp.Type))

	if comp.Description != "" {
		content.WriteString("## Overview\n\n")
		content.WriteString(comp.Description + "\n\n")
	}

	content.WriteString("## Files\n\n")
	for _, file := range comp.Files {
		content.WriteString(fmt.Sprintf("- `%s`\n", file))
	}
	content.WriteString("\n")

	if len(comp.References) > 0 {
		content.WriteString("## Dependencies\n\n")
		for _, ref := range comp.References {
			content.WriteString(fmt.Sprintf("- [%s](components/%s.html)\n",
				ref, g.sanitizePath(ref)))
		}
		content.WriteString("\n")
	}

	outPath := filepath.Join("components", g.sanitizePath(comp.Path)+".html")
	return g.renderPage(outPath, comp.Name, content.String(), nav)
}

// renderPage renders a markdown page through the HTML template
func (g *DocumentationGenerator) renderPage(outPath, title, markdown string, nav []NavItem) error {
	html := g.markdownToHTML(markdown)

	data := PageData{
		Title:       title,
		ProjectName: g.projectName,
		Content:     template.HTML(html),
		Navigation:  nav,
		LastUpdated: time.Now(),
	}

	outPath = filepath.Join(g.outDir, outPath)
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Use the appropriate template based on content type
	templateName := "page"
	if outPath == "index.html" {
		templateName = "index"
	} else if outPath == "search.html" {
		templateName = "search"
	}

	if err := g.templates.ExecuteTemplate(f, templateName, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// buildNavigation creates the navigation structure
func (g *DocumentationGenerator) buildNavigation(structure *analyzer.ProjectStructure) []NavItem {
	nav := []NavItem{
		{Title: "Home", URL: "index.html"},
		{Title: "Architecture", URL: "architecture.html"},
	}

	if len(structure.Components) > 0 {
		components := NavItem{
			Title:    "Components",
			Children: make([]NavItem, 0, len(structure.Components)),
		}

		for _, comp := range structure.Components {
			components.Children = append(components.Children, NavItem{
				Title: comp.Name,
				URL:   fmt.Sprintf("components/%s.html", g.sanitizePath(comp.Path)),
			})
		}

		nav = append(nav, components)
	}

	return nav
}

// copyAssets copies static assets to the output directory
func (g *DocumentationGenerator) copyAssets() error {
	assetsDir := filepath.Join(g.outDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return fmt.Errorf("failed to create assets directory: %w", err)
	}

	assets := map[string]string{
		"style.css": defaultStyles,
		"script.js": defaultScript,
	}

	for name, content := range assets {
		if err := os.WriteFile(filepath.Join(assetsDir, name), []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write asset %s: %w", name, err)
		}
	}

	return nil
}

// markdownToHTML converts markdown to HTML with our preferred settings
func (g *DocumentationGenerator) markdownToHTML(input string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)

	doc := p.Parse([]byte(input))

	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}

// sanitizePath creates a safe filename from a path
func (g *DocumentationGenerator) sanitizePath(path string) string {
	name := strings.ReplaceAll(path, string(filepath.Separator), "_")
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	return name
}

// GenerateDocumentation creates documentation from the analyses map and references.
// This is a package-level function for backward compatibility with existing code.
func GenerateDocumentation(outputDir string, analyses map[string]string, references map[string][]string) error {
	language := determineLanguage(analyses)
	projectType := determineProjectType(analyses)

	// Create template engine
	templateEngine, err := templates.NewTemplateEngine(outputDir)
	if err != nil {
		return fmt.Errorf("failed to create template engine: %w", err)
	}

	// Copy static assets
	if err := templateEngine.CopyAssets(); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	components := convertToComponents(analyses, references)

	// Create components directory
	componentsDir := filepath.Join(outputDir, "components")
	if err := os.MkdirAll(componentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create components directory: %w", err)
	}

	// Create template data
	data := &templates.TemplateData{
		Title:       "Project Documentation",
		ProjectName: "sample",
		Description: fmt.Sprintf("Generated documentation for sample (%s %s)", language, projectType),
		LastUpdated: time.Now(),
		Components:  make([]templates.ComponentData, 0, len(components)),
		Theme:       "light",
		Navigation: []templates.NavigationItem{
			{
				Title: "Overview",
				URL:   "../index.html",
			},
			{
				Title: "Components",
				Children: []templates.NavigationItem{
					{
						Title: "Calculator",
						URL:   "calculator.html",
					},
					{
						Title: "Operations",
						URL:   "operations.html",
					},
				},
			},
		},
	}

	// Convert components to template data and generate component pages
	for _, file := range components {
		// Parse the analysis JSON
		var analysis analyzer.Analysis
		if analysisJSON, ok := analyses[file.Path]; ok {
			if err := json.Unmarshal([]byte(analysisJSON), &analysis); err != nil {
				log.Printf("Warning: failed to parse analysis for %s: %v", file.Path, err)
			}
		}

		// Create component data for the file
		fileData := templates.ComponentData{
			Name:        filepath.Base(file.Path),
			Path:        strings.TrimPrefix(file.Path, "/"),
			Type:        "file",
			Description: analysis.Purpose,
			Analysis:    &analysis,
		}

		// Create sub-components for each component in the file
		var subComponents []templates.ComponentData
		for _, comp := range analysis.Components {
			subComponents = append(subComponents, templates.ComponentData{
				Name:        comp.Name,
				Path:        strings.ToLower(strings.ReplaceAll(comp.Name, ".", "/")),
				Type:        comp.Type,
				Description: comp.Description,
				Analysis:    &analysis,
			})
		}
		fileData.SubComponents = subComponents

		data.Components = append(data.Components, fileData)

		// Generate component page
		componentHTML, err := templateEngine.RenderPage(&templates.TemplateData{
			Title:       fileData.Name,
			ProjectName: "sample",
			Description: fileData.Description,
			LastUpdated: time.Now(),
			Components:  []templates.ComponentData{fileData},
			Theme:      "light",
			Navigation: []templates.NavigationItem{
				{
					Title: "Overview",
					URL:   "../index.html",
				},
				{
					Title: "Components",
					Children: []templates.NavigationItem{
						{
							Title: "Calculator",
							URL:   "calculator.html",
						},
						{
							Title: "Operations",
							URL:   "operations.html",
						},
					},
				},
			},
			CurrentPath: fmt.Sprintf("components/%s.html", strings.ToLower(fileData.Name)),
		}, "component", "content")
		if err != nil {
			return fmt.Errorf("failed to render component page for %s: %w", fileData.Name, err)
		}

		// Write component page
		componentPath := filepath.Join(componentsDir, strings.ToLower(fileData.Name)+".html")
		if err := os.WriteFile(componentPath, []byte(componentHTML), 0644); err != nil {
			return fmt.Errorf("failed to write component page for %s: %w", fileData.Name, err)
		}
	}

	// Generate documentation
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Render index page
	indexHTML, err := templateEngine.RenderPage(data, "layout", "index")
	if err != nil {
		return fmt.Errorf("failed to render index page: %w", err)
	}

	// Write index page
	indexPath := filepath.Join(outputDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(indexHTML), 0644); err != nil {
		return fmt.Errorf("failed to write index page: %w", err)
	}

	return nil
}

// Helper functions for GenerateDocumentation
func determineLanguage(analyses map[string]string) string {
	for path := range analyses {
		ext := filepath.Ext(path)
		switch ext {
		case ".go":
			return "go"
		case ".cs":
			return "csharp"
		}
	}
	return "unknown"
}

func determineProjectType(analyses map[string]string) string {
	for path := range analyses {
		switch {
		case strings.HasSuffix(path, "go.mod"):
			return "go-module"
		case strings.HasSuffix(path, ".sln"):
			return "dotnet-solution"
		}
	}
	return "unknown"
}

func convertToComponents(analyses map[string]string, references map[string][]string) []analyzer.ProjectComponent {
	components := make([]analyzer.ProjectComponent, 0)
	componentMap := make(map[string]*analyzer.ProjectComponent)

	for path, content := range analyses {
		dir := filepath.Dir(path)
		comp, exists := componentMap[dir]
		if !exists {
			comp = &analyzer.ProjectComponent{
				Path:        dir,
				Name:        filepath.Base(dir),
				Type:        "package",
				Description: "",
				Files:       []string{},
				References:  references[path],
			}
			componentMap[dir] = comp
		}
		comp.Files = append(comp.Files, path)

		if comp.Description == "" {
			parts := strings.SplitN(content, "\n\n", 2)
			if len(parts) > 0 {
				comp.Description = strings.TrimSpace(parts[0])
			}
		}
	}

	for _, comp := range componentMap {
		components = append(components, *comp)
	}

	return components
}

// Default styles and scripts
const (
	defaultStyles = `
body {
    margin: 0;
    padding: 0;
    font-family: -apple-system, system-ui, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    line-height: 1.5;
}

.nav {
    position: fixed;
    width: 250px;
    height: 100vh;
    overflow-y: auto;
    padding: 20px;
    background: #f8f9fa;
    border-right: 1px solid #dee2e6;
}

.content {
    margin-left: 290px;
    padding: 20px 40px;
    max-width: 900px;
}

pre {
    background: #f8f9fa;
    padding: 15px;
    border-radius: 4px;
    overflow-x: auto;
}

code {
    background: #f8f9fa;
    padding: 2px 4px;
    border-radius: 4px;
}

.nav-item {
    margin: 10px 0;
}

.nav-group {
    margin: 20px 0;
}

.nav-children {
    margin-left: 20px;
}

a {
    color: #0366d6;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}
`

	defaultScript = `
document.addEventListener('DOMContentLoaded', function() {
    if (typeof mermaid !== 'undefined') {
        mermaid.initialize({ startOnLoad: true });
    }
});
`
)
