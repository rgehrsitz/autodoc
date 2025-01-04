package helpers

import (
	"path/filepath"
	"strings"
)

// PathToURL converts a file path to a URL-friendly path from wiki root
func PathToURL(path string) string {
	// Normalize the path separators
	path = filepath.ToSlash(path)
	// Remove volume name and any leading slashes
	path = strings.TrimPrefix(path, filepath.VolumeName(path))
	path = strings.TrimPrefix(path, "/")
	// Replace file extension with .html
	if ext := filepath.Ext(path); ext != "" {
		path = strings.TrimSuffix(path, ext)
	}
	return path + ".html"
}

// SanitizePath creates a safe file path from the input path
func SanitizePath(path string) string {
	// Normalize path separators
	path = filepath.ToSlash(path)
	// Remove volume name and any leading/trailing slashes
	path = strings.TrimPrefix(path, filepath.VolumeName(path))
	path = strings.Trim(path, "/")
	return path
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
