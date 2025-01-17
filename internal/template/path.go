// autodoc/internal/template/path.go

package helpers

import (
	"path/filepath"
	"strings"
)

// PathToURL converts a file path to a URL-friendly path from wiki root
func PathToURL(path string) string {
	// Normalize the path separators
	path = filepath.ToSlash(path)
	// Remove volume name, leading slashes, and clean the path
	path = strings.TrimPrefix(path, filepath.VolumeName(path))
	path = strings.TrimPrefix(path, "/")
	// For index file detection
	base := filepath.Base(path)
	dir := filepath.Dir(path)
	if base == "index.html" {
		return dir
	}
	// Keep original extension and add .html
	return path + ".html"
}

// SanitizePath creates a safe file path from the input path
func SanitizePath(path string) string {
	// Remove volume name if present
	path = strings.TrimPrefix(path, filepath.VolumeName(path))

	// Convert to forward slashes
	path = filepath.ToSlash(path)

	// Remove leading separator
	path = strings.TrimPrefix(path, "/")

	// Remove any "." or ".." components
	parts := strings.Split(path, "/")
	var cleaned []string
	for _, part := range parts {
		if part == "." || part == ".." {
			continue
		}
		cleaned = append(cleaned, part)
	}

	return strings.Join(cleaned, "/")
}

// GetRelativeURL returns a relative URL from one page to another
func GetRelativeURL(fromPath, toPath string) string {
	from := filepath.Dir(SanitizePath(fromPath))
	to := SanitizePath(toPath)

	if from == "." {
		return PathToURL(to)
	}

	// Calculate the number of directory levels to go up
	levels := len(strings.Split(from, "/"))
	prefix := strings.Repeat("../", levels)

	return prefix + PathToURL(to)
}
