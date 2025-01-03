package helpers

import (
	"path/filepath"
	"strings"
)

// PathToURL converts a file path to a URL-friendly path
func PathToURL(path string) string {
	// Convert path separators
	path = filepath.ToSlash(path)

	// Remove volume name (e.g. "C:")
	vol := filepath.VolumeName(path)
	if vol != "" {
		path = strings.TrimPrefix(path, vol)
		// Remove any leftover leading slash
		path = strings.TrimLeft(path, "/")
	}

	// Strip any leading slashes
	path = strings.TrimLeft(path, "/")

	// Append .html
	return path + ".html"
}

// SanitizePath creates a safe file path from the input path
func SanitizePath(path string) string {
	// Remove volume name (drive letter)
	path = strings.TrimPrefix(strings.TrimPrefix(path, filepath.VolumeName(path)), string(filepath.Separator))
	return filepath.FromSlash(path)
}
