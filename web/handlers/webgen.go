// autodoc/web/handlers/webgen.go

package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/rgehrsitz/AutoDoc/internal/storage"
	"github.com/rgehrsitz/AutoDoc/internal/templateutil"
)

//go:embed templates/layouts/*.html templates/components/*.html templates/partials/*.html templates/assets/css/*.css templates/assets/js/*.js
var embeddedTemplates embed.FS

// Generator handles the wiki generation process
type Generator struct {
	store storage.Storage
}

// Config contains configuration for wiki generation
type Config struct {
	OutputDir    string            // Directory where wiki files will be generated
	ProjectName  string            // Name of the project
	ProjectURL   string            // URL of the project repository
	Theme        string            // Theme name (e.g., "light" or "dark")
	CustomStyles map[string]string // Custom CSS styles to apply
}

// NewGenerator creates a new wiki generator
func NewGenerator(store storage.Storage) *Generator {
	return &Generator{
		store: store,
	}
}

// Generate generates the complete wiki
func (g *Generator) Generate(cfg Config) error {
	log.Println("Starting wiki generation with configuration:")
	log.Printf("OutputDir: %s", cfg.OutputDir)
	log.Printf("ProjectName: %s", cfg.ProjectName)
	log.Printf("ProjectURL: %s", cfg.ProjectURL)
	log.Printf("Theme: %s", cfg.Theme)

	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		log.Printf("Error creating output directory: %v", err)
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate index page
	if err := g.generateIndex(cfg); err != nil {
		return fmt.Errorf("failed to generate index: %w", err)
	}

	// Generate architecture documentation
	if err := g.generateArchitecture(cfg); err != nil {
		return fmt.Errorf("failed to generate architecture docs: %w", err)
	}

	// Generate module documentation
	if err := g.generateModules(cfg); err != nil {
		return fmt.Errorf("failed to generate module docs: %w", err)
	}

	// Generate search page
	if err := g.generateSearch(cfg); err != nil {
		return fmt.Errorf("failed to generate search page: %w", err)
	}

	// Copy static assets from embedded files
	if err := g.copyAssets(cfg); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	return nil
}

// PageData represents common data for all pages
type PageData struct {
	Title       string
	ProjectName string
	ProjectURL  string
	NavItems    []templateutil.NavItem // Use NavItem from helpers package
	Content     template.HTML
	LastUpdated time.Time
	Theme       string
}

func (g *Generator) generateIndex(cfg Config) error {
	// Get architecture document for overview
	archDocs, err := g.store.ListDocuments(storage.TypeArchitecture)
	if err != nil {
		return fmt.Errorf("failed to list architecture docs: %w", err)
	}

	var overview string
	if len(archDocs) > 0 {
		overview = archDocs[0].Content
	}

	// Get all modules for navigation
	modules, err := g.store.ListDocuments(storage.TypeModule)
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}

	// Sort modules by path
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})

	// Create navigation structure
	nav := templateutil.BuildNavigation(modules) // Update to use helper

	data := PageData{
		Title:       "Home",
		ProjectName: cfg.ProjectName,
		ProjectURL:  cfg.ProjectURL,
		NavItems:    nav,
		Content:     template.HTML(renderMarkdown(overview)),
		LastUpdated: time.Now(),
		Theme:       cfg.Theme,
	}

	return templateutil.RenderTemplate(filepath.Join(cfg.OutputDir, "index.html"), "index", data, embeddedTemplates)
}

func (g *Generator) generateArchitecture(cfg Config) error {
	archDocs, err := g.store.ListDocuments(storage.TypeArchitecture)
	if err != nil {
		return fmt.Errorf("failed to list architecture docs: %w", err)
	}

	if len(archDocs) == 0 {
		return nil // No architecture docs to generate
	}

	doc := archDocs[0]
	modules, err := g.store.ListDocuments(storage.TypeModule)
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}

	nav := templateutil.BuildNavigation(modules)
	data := PageData{
		Title:       "Architecture",
		ProjectName: cfg.ProjectName,
		ProjectURL:  cfg.ProjectURL,
		NavItems:    nav,
		Content:     template.HTML(renderMarkdown(doc.Content)),
		LastUpdated: doc.UpdatedAt,
		Theme:       cfg.Theme,
	}

	return templateutil.RenderTemplate(filepath.Join(cfg.OutputDir, "architecture.html"), "page", data, embeddedTemplates)
}

func (g *Generator) generateModules(cfg Config) error {
	modules, err := g.store.ListDocuments(storage.TypeModule)
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}

	nav := templateutil.BuildNavigation(modules)

	for _, doc := range modules {
		// Create relative path by removing volume name and normalizing separators
		cleanPath := templateutil.SanitizePath(doc.Path)
		// Create output path relative to the wiki directory
		outPath := filepath.Join(cfg.OutputDir, cleanPath+".html")

		// Get references
		refs, err := g.store.GetReferences(doc.ID)
		if err != nil {
			return fmt.Errorf("failed to get references: %w", err)
		}

		// Get back references
		backRefs, err := g.store.GetBackReferences(doc.ID)
		if err != nil {
			return fmt.Errorf("failed to get back references: %w", err)
		}

		// Build content with references
		content := strings.Builder{}
		content.WriteString(doc.Content)

		if len(refs) > 0 {
			content.WriteString("\n\n## Dependencies\n\n")
			for _, ref := range refs {
				target, err := g.store.GetDocument(ref.TargetID)
				if err != nil {
					continue
				}
				relativeURL := templateutil.GetRelativeURL(cleanPath, templateutil.SanitizePath(target.Path)+".html")
				content.WriteString(fmt.Sprintf("- [%s](%s)\n", target.Path, relativeURL))
			}
		}

		if len(backRefs) > 0 {
			content.WriteString("\n\n## Used By\n\n")
			for _, ref := range backRefs {
				source, err := g.store.GetDocument(ref.SourceID)
				if err != nil {
					continue
				}
				relativeURL := templateutil.GetRelativeURL(cleanPath, templateutil.SanitizePath(source.Path)+".html")
				content.WriteString(fmt.Sprintf("- [%s](%s)\n", source.Path, relativeURL))
			}
		}

		// Create the page data
		data := PageData{
			Title:       cleanPath,
			ProjectName: cfg.ProjectName,
			ProjectURL:  cfg.ProjectURL,
			NavItems:    nav,
			Content:     template.HTML(renderMarkdown(content.String())),
			LastUpdated: doc.UpdatedAt,
			Theme:       cfg.Theme,
		}

		// Ensure the directory exists
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		if err := templateutil.RenderTemplate(outPath, "page", data, embeddedTemplates); err != nil {
			return fmt.Errorf("failed to render page: %w", err)
		}
	}

	return nil
}

func (g *Generator) generateSearch(cfg Config) error {
	modules, err := g.store.ListDocuments(storage.TypeModule)
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}

	nav := templateutil.BuildNavigation(modules)
	data := PageData{
		Title:       "Search",
		ProjectName: cfg.ProjectName,
		ProjectURL:  cfg.ProjectURL,
		NavItems:    nav,
		LastUpdated: time.Now(),
		Theme:       cfg.Theme,
	}

	return templateutil.RenderTemplate(filepath.Join(cfg.OutputDir, "search.html"), "search", data, embeddedTemplates)
}

func (g *Generator) copyAssets(cfg Config) error {
	// Create assets directory structure
	cssDir := filepath.Join(cfg.OutputDir, "assets", "css")
	jsDir := filepath.Join(cfg.OutputDir, "assets", "js")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		return fmt.Errorf("failed to create css directory: %w", err)
	}
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		return fmt.Errorf("failed to create js directory: %w", err)
	}

	// List of assets to copy from embedded files
	assetFiles := []string{
		"assets/css/dark.css",
		"assets/css/light.css",
		"assets/css/style.css",
		"assets/js/search.js",
	}

	for _, asset := range assetFiles {
		// Read asset from embedded files
		data, err := embeddedTemplates.ReadFile(asset)
		if err != nil {
			log.Printf("Failed to read embedded asset %s: %v", asset, err)
			return fmt.Errorf("failed to read embedded asset %s: %w", asset, err)
		}

		// Determine the destination path, maintaining directory structure
		relPath := strings.TrimPrefix(asset, "assets/")
		destPath := filepath.Join(cfg.OutputDir, "assets", relPath)

		// Write the asset to the destination
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			log.Printf("Failed to write asset to %s: %v", destPath, err)
			return fmt.Errorf("failed to write asset to %s: %w", destPath, err)
		}

		log.Printf("Successfully copied asset to: %s", destPath)
	}

	return nil
}

// renderMarkdown converts markdown text to HTML
func renderMarkdown(input string) template.HTML {
	// Create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	// Parse markdown
	doc := p.Parse([]byte(input))

	// Create HTML renderer
	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}
	renderer := html.NewRenderer(opts)

	// Render HTML
	html := markdown.Render(doc, renderer)
	return template.HTML(html)
}
