// autodoc/internal/analysis/chunker.go

package analyzer

import "strings"

type Chunk struct {
	Content   string
	StartLine int
	EndLine   int
}

// Chunker splits code into manageable pieces for analysis
type Chunker struct {
	MaxChunkSize int
}

func NewChunker(maxChunkSize int) *Chunker {
	return &Chunker{
		MaxChunkSize: maxChunkSize,
	}
}

func (c *Chunker) Split(content string) []Chunk {
	lines := strings.Split(content, "\n")
	chunks := make([]Chunk, 0)
	currentChunk := ""
	startLine := 1
	lineCount := 0

	for i, line := range lines {
		currentChunk += line + "\n"
		lineCount++

		if len(currentChunk) >= c.MaxChunkSize || i == len(lines)-1 {
			chunks = append(chunks, Chunk{
				Content:   currentChunk,
				StartLine: startLine,
				EndLine:   startLine + lineCount - 1,
			})
			currentChunk = ""
			startLine += lineCount
			lineCount = 0
		}
	}

	return chunks
}
