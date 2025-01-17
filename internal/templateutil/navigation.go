// autodoc/internal/templateutil/navigation.go

package templateutil

import (
	"sort"
	"strings"

	"github.com/rgehrsitz/AutoDoc/internal/storage"
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

		// Keep the original filename for display
		fileName := parts[len(parts)-1]
		url := PathToURL(doc.Path)

		// Handle root level files
		if len(parts) == 1 {
			nav = append(nav, NavItem{
				Title: fileName, // Keep original filename
				URL:   url,
			})
			continue
		}

		// Handle nested files
		currentPath := ""
		var currentNav *NavItem

		// Create directory structure
		for i, part := range parts[:len(parts)-1] {
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = currentPath + "/" + part
			}

			if existing, exists := moduleNav[currentPath]; exists {
				currentNav = existing
			} else {
				// Find first file in directory for the URL
				firstFileURL := ""
				currentDir := strings.Join(parts[:i+1], "/") + "/"
				for _, m := range modules {
					if strings.HasPrefix(SanitizePath(m.Path), currentDir) {
						firstFileURL = PathToURL(m.Path)
						break
					}
				}

				newNav := &NavItem{
					Title:    part,
					URL:      firstFileURL,
					Children: []NavItem{},
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

		// Add file to current directory
		fileNav := NavItem{
			Title: fileName, // Keep original filename with extension
			URL:   url,      // Use full path with .html extension
		}

		if currentNav != nil {
			currentNav.Children = append(currentNav.Children, fileNav)
			// Sort children by title
			sort.Slice(currentNav.Children, func(i, j int) bool {
				return currentNav.Children[i].Title < currentNav.Children[j].Title
			})
		}
	}

	return nav
}
