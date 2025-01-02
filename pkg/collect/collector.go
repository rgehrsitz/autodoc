// autodoc/pkg/collect/collector.go
package collect

import (
	"context"
)

// Collector handles repository cloning and file enumeration
type Collector interface {
	Clone(ctx context.Context, repoURL string) (string, error)
	ListFiles(ctx context.Context, path string) ([]string, error)
}

// FileInfo represents information about a source file
type FileInfo struct {
	Path     string
	Language string
	Content  string
}
