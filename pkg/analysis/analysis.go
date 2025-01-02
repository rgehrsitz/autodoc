package analysis

import (
	"regexp"
	"strings"
)

// FileMetadata represents metadata about a source file.
type FileMetadata struct {
	Path      string
	Imports   []string
	Functions []string
}

// ExtractReferences scans code for import statements and function calls.
func ExtractReferences(code string, ext string) []string {
	var refs []string

	importRegex := getImportRegex(ext)
	funcRegex := getFuncCallRegex(ext)

	imports := importRegex.FindAllStringSubmatch(code, -1)
	for _, match := range imports {
		if len(match) > 1 {
			refs = append(refs, strings.Trim(match[1], "\"'"))
		}
	}

	funcCalls := funcRegex.FindAllStringSubmatch(code, -1)
	for _, match := range funcCalls {
		if len(match) > 1 {
			refs = append(refs, match[1])
		}
	}

	return refs
}

// ExtractMetadata extracts metadata from code.
func ExtractMetadata(code string, ext string) FileMetadata {
	var metadata FileMetadata
	metadata.Path = "..."

	// Extract imports
	importRegex := getImportRegex(ext)
	imports := importRegex.FindAllStringSubmatch(code, -1)
	for _, match := range imports {
		if len(match) > 1 {
			metadata.Imports = append(metadata.Imports, strings.Trim(match[1], "\"'"))
		}
	}

	// Extract function names
	funcRegex := regexp.MustCompile(`func\s+(\w+)`)
	funcMatches := funcRegex.FindAllStringSubmatch(code, -1)
	for _, match := range funcMatches {
		if len(match) > 1 {
			metadata.Functions = append(metadata.Functions, match[1])
		}
	}

	return metadata
}

func getImportRegex(ext string) *regexp.Regexp {
	switch ext {
	case ".go":
		return regexp.MustCompile(`import\s+"([^"]+)"`)
	case ".js", ".ts":
		return regexp.MustCompile(`import\s+.*\s+from\s+'([^']+)'`)
	case ".py":
		return regexp.MustCompile(`import\s+(\w+)`)
	case ".java":
		return regexp.MustCompile(`import\s+([\w\.]+);`)
	case ".rs":
		return regexp.MustCompile(`use\s+([\w::]+);`)
	default:
		return regexp.MustCompile(``)
	}
}

func getFuncCallRegex(ext string) *regexp.Regexp {
	switch ext {
	case ".go", ".js", ".ts", ".py", ".java", ".rs":
		return regexp.MustCompile(`(\w+)\s*\(`)
	default:
		return regexp.MustCompile(``)
	}
}
