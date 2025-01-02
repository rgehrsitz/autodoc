package helpers

import (
	"path/filepath"
	"strings"
)

// PathToURL converts a file path to a URL-friendly path
func PathToURL(path string) string {
	// Convert path separators to forward slashes
	path = filepath.ToSlash(path)
	// Remove any drive letter prefix (e.g., C:)
	path = strings.TrimPrefix(strings.TrimPrefix(path, filepath.VolumeName(path)), "/")
	// Add .html extension
	return path + ".html"
}

// SanitizePath creates a safe file path from the input path
func SanitizePath(path string) string {
	// Remove volume name (drive letter)
	path = strings.TrimPrefix(strings.TrimPrefix(path, filepath.VolumeName(path)), string(filepath.Separator))
	return filepath.FromSlash(path)
}
