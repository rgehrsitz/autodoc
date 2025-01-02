package wiki

import (
	"embed" // Added embed package
	"fmt"
	"html/template"
	"log" // Ensure log is imported
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rgehrsitz/AutoDoc/pkg/storage"
	// "github.com/rgehrsitz/AutoDoc/pkg/wiki/helpers" // Removed incorrect import
)

//go:embed templates/*.html templates/assets/*.css templates/assets/*.js
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
	Theme        string            // Theme name (default, dark, light)
	CustomStyles map[string]string // Custom CSS styles
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
	NavItems    []NavItem
	Content     template.HTML
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
	nav := buildNavigation(modules)

	data := PageData{
		Title:       "Home",
		ProjectName: cfg.ProjectName,
		ProjectURL:  cfg.ProjectURL,
		NavItems:    nav,
		Content:     template.HTML(renderMarkdown(overview)),
		LastUpdated: time.Now(),
		Theme:       cfg.Theme,
	}

	return RenderTemplate(filepath.Join(cfg.OutputDir, "index.html"), "index", data) // Updated call
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

	nav := buildNavigation(modules)
	data := PageData{
		Title:       "Architecture",
		ProjectName: cfg.ProjectName,
		ProjectURL:  cfg.ProjectURL,
		NavItems:    nav,
		Content:     template.HTML(renderMarkdown(doc.Content)),
		LastUpdated: doc.UpdatedAt,
		Theme:       cfg.Theme,
	}

	return RenderTemplate(filepath.Join(cfg.OutputDir, "architecture.html"), "page", data) // Updated call
}

func (g *Generator) generateModules(cfg Config) error {
	modules, err := g.store.ListDocuments(storage.TypeModule)
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}

	nav := buildNavigation(modules)

	for _, doc := range modules {
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
				content.WriteString(fmt.Sprintf("- [%s](%s)\n", target.Path, pathToURL(target.Path)))
			}
		}

		if len(backRefs) > 0 {
			content.WriteString("\n\n## Used By\n\n")
			for _, ref := range backRefs {
				source, err := g.store.GetDocument(ref.SourceID)
				if err != nil {
					continue
				}
				content.WriteString(fmt.Sprintf("- [%s](%s)\n", source.Path, pathToURL(source.Path)))
			}
		}

		data := PageData{
			Title:       doc.Path,
			ProjectName: cfg.ProjectName,
			ProjectURL:  cfg.ProjectURL,
			NavItems:    nav,
			Content:     template.HTML(renderMarkdown(content.String())),
			LastUpdated: doc.UpdatedAt,
			Theme:       cfg.Theme,
		}

		outPath := filepath.Join(cfg.OutputDir, pathToURL(doc.Path))
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		if err := RenderTemplate(outPath, "page", data); err != nil { // Updated call
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

	nav := buildNavigation(modules)
	data := PageData{
		Title:       "Search",
		ProjectName: cfg.ProjectName,
		ProjectURL:  cfg.ProjectURL,
		NavItems:    nav,
		LastUpdated: time.Now(),
		Theme:       cfg.Theme,
	}

	return RenderTemplate(filepath.Join(cfg.OutputDir, "search.html"), "search", data) // Updated call
}

func (g *Generator) copyAssets(cfg Config) error {
	assetsDir := filepath.Join(cfg.OutputDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return fmt.Errorf("failed to create assets directory: %w", err)
	}

	// List of assets to copy from embedded files
	assetFiles := []string{
		"templates/assets/style.css",
		"templates/assets/search.js",
	}

	for _, asset := range assetFiles {
		// Read asset from embedded files
		data, err := embeddedTemplates.ReadFile(asset)
		if err != nil {
			log.Printf("Failed to read embedded asset %s: %v", asset, err)
			return fmt.Errorf("failed to read embedded asset %s: %w", asset, err)
		}

		// Determine the destination path
		destPath := filepath.Join(assetsDir, filepath.Base(asset))

		// Write the asset to the destination
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			log.Printf("Failed to write asset to %s: %v", destPath, err)
			return fmt.Errorf("failed to write asset to %s: %w", destPath, err)
		}

		log.Printf("Successfully copied asset to: %s", destPath)
	}

	return nil
}
