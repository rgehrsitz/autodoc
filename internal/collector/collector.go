// autodoc/internal/collector/collector.go

package collector

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileInfo represents information about a source file
type FileInfo struct {
	Path     string
	Language string
	Type     string
	Content  string
}

// Collector handles repository cloning and file enumeration
type Collector interface {
	Clone(ctx context.Context, repoURL string) (string, error)
	CollectFiles(ctx context.Context, path string) ([]FileInfo, error)
	ReadFile(path string) ([]byte, error)
}

// FSCollector implements the Collector interface for filesystem operations
type FSCollector struct{}

// NewCollector initializes and returns a new Collector
func NewCollector() Collector {
	return &FSCollector{}
}

// CollectFiles walks through the directory and collects relevant files
func (c *FSCollector) CollectFiles(ctx context.Context, path string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			if os.IsNotExist(err) {
				return nil // Skip non-existent paths
			}
			if os.IsPermission(err) {
				return nil // Skip inaccessible paths
			}
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(filePath))
			language, fileType := classifyFile(ext)

			// Only collect files we're interested in
			if language != "" {
				content, err := c.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file %s: %w", filePath, err)
				}

				files = append(files, FileInfo{
					Path:     filePath,
					Language: language,
					Type:     fileType,
					Content:  string(content),
				})
			}
		}
		return nil
	})

	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist: %s", path)
		}
		return nil, fmt.Errorf("error walking path %s: %w", path, err)
	}

	return files, nil
}

// ReadFile reads the content of the specified file
func (c *FSCollector) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// Clone clones the repository from the given URL to a temporary directory
func (c *FSCollector) Clone(ctx context.Context, repoURL string) (string, error) {
	tempDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, tempDir)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir) // Clean up on error
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return tempDir, nil
}

// classifyFile determines the language and type of a file based on its extension
func classifyFile(ext string) (language, fileType string) {
	switch ext {
	case ".go":
		return "go", "source"
	case ".cs":
		return "csharp", "source"
	case ".csproj":
		return "csharp", "project"
	case ".sln":
		return "csharp", "solution"
	case ".mod":
		return "go", "module"
	default:
		return "", ""
	}
}
