package helpers

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/rgehrsitz/AutoDoc/pkg/storage"
)

// NavItem represents a navigation menu item
type NavItem struct {
	Title    string
	URL      string
	Active   bool
	Children []NavItem
}

// BuildNavigation creates the navigation structure
func BuildNavigation(modules []*storage.Document) []NavItem {
	nav := []NavItem{
		{
			Title: "Home",
			URL:   "index.html",
		},
		{
			Title: "Architecture",
			URL:   "architecture.html",
		},
	}

	// Create module navigation items
	moduleNav := make(map[string]*NavItem)

	// Sort modules to ensure consistent ordering
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})

	for _, doc := range modules {
		// Create relative URL from doc path
		url := PathToURL(doc.Path)
		parts := strings.Split(strings.Trim(doc.Path, string(filepath.Separator)), string(filepath.Separator))

		if len(parts) == 0 {
			continue
		}

		// Create or get parent nav item
		var parent *NavItem
		if len(parts) > 1 {
			parentPath := parts[0]
			if p, exists := moduleNav[parentPath]; exists {
				parent = p
			} else {
				parent = &NavItem{
					Title: parentPath,
					URL:   "#",
				}
				moduleNav[parentPath] = parent
				nav = append(nav, *parent)
			}
		}

		// Create nav item for the current module
		item := NavItem{
			Title: filepath.Base(doc.Path),
			URL:   url,
		}

		// Add to parent or main nav
		if parent != nil {
			p := moduleNav[parts[0]]
			p.Children = append(p.Children, item)
		} else {
			nav = append(nav, item)
		}
	}

	return nav
}
