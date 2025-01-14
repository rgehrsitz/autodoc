// autodoc/pkg/generator/openai.go

package generator

import "context"

// OpenAIClient defines the interface for OpenAI operations
type OpenAIClient interface {
	// AnalyzeSource analyzes and documents the provided source code
	AnalyzeSource(ctx context.Context, code string, language string) (string, error)
	// GenerateEmbedding generates an embedding for the given text
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)
}
