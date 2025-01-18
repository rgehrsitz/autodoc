// autodoc/internal/docs/docgen.go

package docs

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	analyzer "github.com/rgehrsitz/AutoDoc/internal/analysis"
)

// DocumentationGenerator handles the generation of documentation
type DocumentationGenerator struct {
	outDir      string
	projectName string
	templates   *template.Template
}

// Config holds configuration for documentation generation
type Config struct {
	OutputDir    string
	ProjectName  string
	TemplatePath string
	CustomStyles map[string]string
	Theme        string
}

// PageData represents the data passed to documentation templates
type PageData struct {
	Title       string
	ProjectName string
	Content     template.HTML
	Navigation  []NavItem
	LastUpdated time.Time
	Theme       string
	Description string
	CurrentPath string
	Components  []analyzer.ProjectComponent
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
		filepath.Join(config.TemplatePath, "layouts", "base.html"),
		filepath.Join(config.TemplatePath, "index.html"),
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
	// Build navigation first
	nav := g.buildNavigation(structure)

	// Copy static assets
	if err := g.copyAssets(); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	// Generate all pages
	if err := g.generatePages(structure, nav); err != nil {
		return fmt.Errorf("failed to generate pages: %w", err)
	}

	return nil
}

// generatePages generates all documentation pages
func (g *DocumentationGenerator) generatePages(structure *analyzer.ProjectStructure, nav []NavItem) error {
	if err := g.generateIndexPage(structure, nav); err != nil {
		return fmt.Errorf("failed to generate index page: %w", err)
	}

	for _, comp := range structure.Components {
		if err := g.generateComponentPage(comp, structure, nav); err != nil {
			return fmt.Errorf("failed to generate component page %s: %w", comp.Path, err)
		}
	}

	return nil
}

// GenerateDocumentation creates documentation from the analyses map and references
func GenerateDocumentation(outputDir string, analyses map[string]string, references map[string][]string) error {
	// Convert analyses to components
	components := convertToComponents(analyses, references)

	// Create project structure
	structure := &analyzer.ProjectStructure{
		Type:       determineProjectType(analyses),
		Language:   determineLanguage(analyses),
		Components: components,
	}

	// Create documentation generator
	config := Config{
		OutputDir:    outputDir,
		ProjectName:  filepath.Base(filepath.Dir(outputDir)),
		TemplatePath: filepath.Join(filepath.Dir(filepath.Dir(outputDir)), "web", "handlers", "templates"),
	}

	generator, err := NewDocumentationGenerator(config)
	if err != nil {
		return fmt.Errorf("failed to create documentation generator: %w", err)
	}

	// Generate documentation
	if err := generator.Generate(structure); err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	return nil
}

// buildNavigation creates the navigation structure
func (g *DocumentationGenerator) buildNavigation(structure *analyzer.ProjectStructure) []NavItem {
	nav := []NavItem{
		{
			Title: "Overview",
			URL:   "index.html",
		},
	}

	// Group components by package
	packageGroups := make(map[string][]analyzer.ProjectComponent)
	for _, comp := range structure.Components {
		pkgPath := filepath.Dir(comp.Path)
		if pkgPath == "." {
			pkgPath = "root"
		}
		packageGroups[pkgPath] = append(packageGroups[pkgPath], comp)
	}

	if len(structure.Components) > 0 {
		packages := NavItem{
			Title:    "Packages",
			Children: make([]NavItem, 0),
		}

		// Sort package names for consistent ordering
		pkgNames := make([]string, 0, len(packageGroups))
		for pkg := range packageGroups {
			pkgNames = append(pkgNames, pkg)
		}
		sort.Strings(pkgNames)

		// Create navigation structure for each package
		for _, pkgName := range pkgNames {
			components := packageGroups[pkgName]
			pkgNav := NavItem{
				Title:    filepath.Base(pkgName),
				Children: make([]NavItem, 0, len(components)),
			}

			// Sort components within package
			sort.Slice(components, func(i, j int) bool {
				return components[i].Name < components[j].Name
			})

			for _, comp := range components {
				pkgNav.Children = append(pkgNav.Children, NavItem{
					Title: comp.Name,
					URL:   g.getComponentURL(comp),
				})
			}

			packages.Children = append(packages.Children, pkgNav)
		}

		nav = append(nav, packages)
	}

	return nav
}

// getComponentURL generates the URL for a component
func (g *DocumentationGenerator) getComponentURL(comp analyzer.ProjectComponent) string {
	pkgPath := filepath.Dir(comp.Path)
	baseName := g.sanitizePath(comp.Name)
	
	if pkgPath == "." {
		return fmt.Sprintf("components/%s.html", baseName)
	}
	
	// Create package-specific directory
	pkgDir := g.sanitizePath(pkgPath)
	return fmt.Sprintf("components/%s/%s.html", pkgDir, baseName)
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

	content.WriteString("## Package\n\n")
	pkgPath := filepath.Dir(comp.Path)
	if pkgPath == "." {
		content.WriteString("Root package\n\n")
	} else {
		content.WriteString(fmt.Sprintf("Package `%s`\n\n", pkgPath))
	}

	content.WriteString("## Files\n\n")
	for _, file := range comp.Files {
		content.WriteString(fmt.Sprintf("- `%s`\n", file))
	}
	content.WriteString("\n")

	if len(comp.References) > 0 {
		content.WriteString("## Dependencies\n\n")
		for _, ref := range comp.References {
			// Find the component being referenced to get its proper URL
			var refURL string
			for _, c := range structure.Components {
				if c.Name == ref {
					refURL = g.getComponentURL(c)
					break
				}
			}
			if refURL == "" {
				refURL = fmt.Sprintf("components/%s.html", g.sanitizePath(ref))
			}
			content.WriteString(fmt.Sprintf("- [%s](%s)\n", ref, refURL))
		}
		content.WriteString("\n")
	}

	// Create subdirectories based on package structure
	pkgPath = filepath.Dir(comp.Path)
	var outPath string
	if pkgPath == "." {
		outPath = filepath.Join("components", g.sanitizePath(comp.Name)+".html")
	} else {
		// Create package-specific directory
		pkgDir := g.sanitizePath(pkgPath)
		outPath = filepath.Join("components", pkgDir, g.sanitizePath(comp.Name)+".html")
		
		// Create the package directory if it doesn't exist
		if err := os.MkdirAll(filepath.Join(g.outDir, "components", pkgDir), 0755); err != nil {
			return fmt.Errorf("failed to create package directory: %w", err)
		}
	}

	return g.renderPage(outPath, comp.Name, content.String(), nav, structure)
}

// renderPage renders a markdown page through the HTML template
func (g *DocumentationGenerator) renderPage(outPath, title, markdown string, nav []NavItem, structure *analyzer.ProjectStructure) error {
	html := g.markdownToHTML(markdown)

	data := PageData{
		Title:       title,
		ProjectName: g.projectName,
		Content:     template.HTML(html),
		Navigation:  nav,
		LastUpdated: time.Now(),
		Theme:       "light",
		Description: "", // Add empty description for now
		CurrentPath: outPath,
		Components:  structure.Components, // Include components for templates
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
	templateName := "layout"
	if filepath.Base(outPath) == "index.html" {
		templateName = "index"
	}

	if err := g.templates.ExecuteTemplate(f, templateName, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// generateIndexPage creates the main index.html
func (g *DocumentationGenerator) generateIndexPage(structure *analyzer.ProjectStructure, nav []NavItem) error {
	content := strings.Builder{}
	content.WriteString("# Project Documentation\n\n")

	if len(structure.Components) > 0 {
		content.WriteString("## Components\n\n")

		// Group components by package
		packageGroups := make(map[string][]analyzer.ProjectComponent)
		for _, comp := range structure.Components {
			pkgPath := filepath.Dir(comp.Path)
			if pkgPath == "." {
				pkgPath = "root"
			}
			packageGroups[pkgPath] = append(packageGroups[pkgPath], comp)
		}

		// Sort package names for consistent ordering
		pkgNames := make([]string, 0, len(packageGroups))
		for pkg := range packageGroups {
			pkgNames = append(pkgNames, pkg)
		}
		sort.Strings(pkgNames)

		// List components by package
		for _, pkg := range pkgNames {
			if pkg == "root" {
				content.WriteString("### Root Package\n\n")
			} else {
				content.WriteString(fmt.Sprintf("### Package %s\n\n", pkg))
			}

			components := packageGroups[pkg]
			sort.Slice(components, func(i, j int) bool {
				return components[i].Name < components[j].Name
			})

			for _, comp := range components {
				url := g.getComponentURL(comp)
				content.WriteString(fmt.Sprintf("- [%s](%s) - %s\n", 
					comp.Name, 
					url, 
					comp.Description))
			}
			content.WriteString("\n")
		}
	}

	return g.renderPage("index.html", "Overview", content.String(), nav, structure)
}

// Helper functions for GenerateDocumentation
func determineLanguage(analyses map[string]string) string {
	for path := range analyses {
		switch {
		case strings.HasSuffix(path, ".go"):
			return "Go"
		case strings.HasSuffix(path, ".cs"):
			return "C#"
		case strings.HasSuffix(path, ".js"):
			return "JavaScript"
		case strings.HasSuffix(path, ".py"):
			return "Python"
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
	components := make([]analyzer.ProjectComponent, 0, len(analyses))
	for path, content := range analyses {
		component := analyzer.ProjectComponent{
			Name:        filepath.Base(path),
			Path:        path,
			Type:        "file",
			Description: content,
			References:  references[path],
		}
		components = append(components, component)
	}
	return components
}

// sanitizePath creates a safe filename from a path
func (g *DocumentationGenerator) sanitizePath(path string) string {
	// Only sanitize the filename portion
	name := filepath.Base(path)
	
	// Remove .go extension if present
	name = strings.TrimSuffix(name, ".go")
	
	// Replace problematic characters while preserving meaningful ones
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_':
			return r
		case r == ' ':
			return '-'
		default:
			return '_'
		}
	}, name)
	
	return strings.ToLower(name)
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

// copyAssets copies static assets to the output directory
func (g *DocumentationGenerator) copyAssets() error {
	assetsDir := filepath.Join(g.outDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return fmt.Errorf("failed to create assets directory: %w", err)
	}

	// Create CSS subdirectory
	cssDir := filepath.Join(assetsDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		return fmt.Errorf("failed to create css directory: %w", err)
	}

	// Create JS subdirectory
	jsDir := filepath.Join(assetsDir, "js")
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		return fmt.Errorf("failed to create js directory: %w", err)
	}

	// Copy CSS files
	cssFiles := map[string]string{
		"style.css": defaultStyles,
		"light.css": lightTheme,
		"dark.css":  darkTheme,
	}

	for name, content := range cssFiles {
		if err := os.WriteFile(filepath.Join(cssDir, name), []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write CSS file %s: %w", name, err)
		}
	}

	// Copy JS files
	jsFiles := map[string]string{
		"search.js": searchScript,
	}

	for name, content := range jsFiles {
		if err := os.WriteFile(filepath.Join(jsDir, name), []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write JS file %s: %w", name, err)
		}
	}

	return nil
}

// Default styles and scripts
const (
	defaultStyles = `
/* Base styles */
body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    line-height: 1.6;
    margin: 0;
    padding: 0;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
}

/* Navigation */
nav {
    padding: 1rem;
}

.nav-item {
    margin-bottom: 1rem;
}

.nav-group {
    margin-bottom: 1.5rem;
}

/* Content */
.content {
    margin-left: 16rem;
    padding: 2rem;
}

/* Code blocks */
pre {
    background-color: #f5f5f5;
    padding: 1rem;
    border-radius: 4px;
    overflow-x: auto;
}

code {
    font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
}

/* Links */
a {
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}
`

	lightTheme = `
/* Light theme */
body {
    background-color: #ffffff;
    color: #1a1a1a;
}

a {
    color: #0066cc;
}

pre {
    background-color: #f5f5f5;
    color: #1a1a1a;
}

nav {
    background-color: #f8f9fa;
    border-right: 1px solid #dee2e6;
}
`

	darkTheme = `
/* Dark theme */
body {
    background-color: #1a1a1a;
    color: #ffffff;
}

a {
    color: #66b3ff;
}

pre {
    background-color: #2d2d2d;
    color: #ffffff;
}

nav {
    background-color: #2d2d2d;
    border-right: 1px solid #404040;
}
`

	searchScript = `
document.addEventListener('DOMContentLoaded', function() {
    const themeToggle = document.getElementById('theme-toggle');
    const html = document.documentElement;

    // Check for saved theme preference
    const savedTheme = localStorage.getItem('theme') || 'light';
    html.classList.toggle('dark', savedTheme === 'dark');

    // Theme toggle functionality
    themeToggle.addEventListener('click', function() {
        html.classList.toggle('dark');
        localStorage.setItem('theme', html.classList.contains('dark') ? 'dark' : 'light');
    });
});
`
)
