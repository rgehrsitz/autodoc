package helpers

import (
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
		{
			Title: "Search",
			URL:   "search.html",
		},
	}

	// Create module navigation items
	moduleNav := make(map[string]*NavItem)

	// Sort modules to ensure consistent ordering
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})

	for _, doc := range modules {
		// Sanitize and split the path
		cleanPath := SanitizePath(doc.Path)
		if cleanPath == "" {
			continue
		}

		parts := strings.Split(cleanPath, "/")
		if len(parts) == 0 {
			continue
		}

		// Create the URL for this document
		url := PathToURL(doc.Path)

		// Handle root level files
		if len(parts) == 1 {
			nav = append(nav, NavItem{
				Title: parts[0],
				URL:   url,
			})
			continue
		}

		// Handle nested files
		currentPath := ""
		var currentNav *NavItem

		// Create or update the navigation hierarchy
		for _, part := range parts[:len(parts)-1] {
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = currentPath + "/" + part
			}

			if existing, exists := moduleNav[currentPath]; exists {
				currentNav = existing
			} else {
				newNav := &NavItem{
					Title: part,
					URL:   "#",
				}
				if currentNav == nil {
					nav = append(nav, *newNav)
					moduleNav[currentPath] = &nav[len(nav)-1]
				} else {
					currentNav.Children = append(currentNav.Children, *newNav)
					moduleNav[currentPath] = &currentNav.Children[len(currentNav.Children)-1]
				}
				currentNav = moduleNav[currentPath]
			}
		}

		// Add the leaf node (file)
		fileNav := NavItem{
			Title: parts[len(parts)-1],
			URL:   url,
		}

		if currentNav != nil {
			currentNav.Children = append(currentNav.Children, fileNav)
		} else {
			nav = append(nav, fileNav)
		}
	}

	return nav
}
