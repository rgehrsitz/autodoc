// autodoc/pkg/collect/collector.go
package collect

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// Collector handles repository cloning and file enumeration
type Collector interface {
	Clone(ctx context.Context, repoURL string) (string, error)
	ListFiles(ctx context.Context, path string) ([]string, error)
	ReadFile(path string) ([]byte, error)
}

// FileInfo represents information about a source file
type FileInfo struct {
	Path     string
	Language string
	Content  string
}

// YourCollectorStruct is the concrete implementation of the Collector interface.
type YourCollectorStruct struct{}

// NewCollector initializes and returns a new Collector.
func NewCollector() Collector {
	return &YourCollectorStruct{}
}

// Clone clones the repository from the given URL to a temporary directory.
func (c *YourCollectorStruct) Clone(ctx context.Context, repoURL string) (string, error) {
	tempDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, tempDir)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return tempDir, nil
}

// ListFiles lists all files within the specified directory.
func (c *YourCollectorStruct) ListFiles(ctx context.Context, path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// ReadFile reads the content of a file.
func (c *YourCollectorStruct) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
