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
		for i, part := range parts[:len(parts)-1] {
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = currentPath + "/" + part
			}

			if existing, exists := moduleNav[currentPath]; exists {
				currentNav = existing
			} else {
				// Create new nav item for directory
				newNav := &NavItem{
					Title:    part,
					URL:      "",          // Will be updated if index file is found
					Children: []NavItem{}, // Initialize empty children slice
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

		// Add the file as a child of the current navigation item
		fileName := parts[len(parts)-1]
		fileNav := NavItem{
			Title: fileName,
			URL:   PathToURL(doc.Path), // Use the full path for the URL
		}

		if currentNav != nil {
			// Sort children alphabetically when adding new item
			currentNav.Children = append(currentNav.Children, fileNav)
			sort.Slice(currentNav.Children, func(i, j int) bool {
				return currentNav.Children[i].Title < currentNav.Children[j].Title
			})
		} else {
			nav = append(nav, fileNav)
		}
	}

	return nav
}
