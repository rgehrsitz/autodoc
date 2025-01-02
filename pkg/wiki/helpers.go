package wiki

import (
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rgehrsitz/AutoDoc/pkg/storage"
	"github.com/russross/blackfriday/v2"
)

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

// renderTemplate renders a template with the given data
func renderTemplate(path string, name string, data interface{}) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Parse templates
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/"+name+".html",
	)
	if err != nil {
		return err
	}

	// Create output file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Execute template
	return tmpl.Execute(f, data)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open source file
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Create destination file
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy contents
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// Sync to ensure write
	return out.Sync()
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
