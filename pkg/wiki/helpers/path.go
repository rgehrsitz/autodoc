package helpers

import (
	"path/filepath"
	"strings"
)

// PathToURL converts a file path to a URL-friendly path
func PathToURL(path string) string {
	// Convert path separators to forward slashes
	path = filepath.ToSlash(path)

	// Remove volume name (e.g. "C:")
	path = strings.TrimPrefix(path, filepath.VolumeName(path))

	// Strip any leading slashes
	path = strings.TrimPrefix(path, "/")

	// Create relative path from root
	return "code/" + path + ".html"
}

// SanitizePath creates a safe file path from the input path
func SanitizePath(path string) string {
	// Remove volume name and normalize separators
	path = strings.TrimPrefix(filepath.ToSlash(path), filepath.ToSlash(filepath.VolumeName(path)))
	// Remove leading/trailing separators
	return strings.Trim(path, "/")
}

// GetRelativePath returns a path relative to the documentation root
func GetRelativePath(path, basePath string) string {
	// Convert both paths to forward slashes
	path = filepath.ToSlash(path)
	basePath = filepath.ToSlash(basePath)

	// Remove volume names
	path = strings.TrimPrefix(path, filepath.VolumeName(path))
	basePath = strings.TrimPrefix(basePath, filepath.VolumeName(basePath))

	// Calculate relative path
	rel, err := filepath.Rel(basePath, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
