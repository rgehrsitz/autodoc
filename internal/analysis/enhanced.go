// autodoc/internal/analysis/enhanced.go

package analyzer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/rgehrsitz/AutoDoc/internal/collector"
)

// CodeAnalysisSchema defines the structure for our code analysis
type CodeAnalysisSchema struct {
	ArchitecturalPatterns []string               `json:"architectural_patterns,omitempty"`
	CodeQualityMetrics    map[string]any         `json:"code_quality_metrics,omitempty"`
	Insights              []ArchitecturalInsight `json:"insights,omitempty"`
	CrossReferences       map[string][]string    `json:"cross_references,omitempty"`
}

// ArchitecturalInsight represents a high-level insight about the code
type ArchitecturalInsight struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Components  []string `json:"components"`
}

// EnhancedAnalyzer provides advanced code analysis capabilities
type EnhancedAnalyzer struct {
	client *openai.Client
}

// NewEnhancedAnalyzer creates a new instance of EnhancedAnalyzer
func NewEnhancedAnalyzer(openAIKey string) *EnhancedAnalyzer {
	return &EnhancedAnalyzer{
		client: openai.NewClient(
			option.WithAPIKey(openAIKey),
		),
	}
}

// generateSystemPrompt creates a dynamic system prompt for code analysis
func (ea *EnhancedAnalyzer) generateSystemPrompt(file collector.FileInfo) string {
	return fmt.Sprintf(`You are an expert software architect analyzing %s code.
Provide a comprehensive analysis of the code, focusing on architectural insights, code quality, and system interactions.
Respond with a structured JSON output that captures the nuanced understanding of an experienced architect.

Include the following in your analysis:
- Architectural patterns and design principles
- Code complexity and maintainability metrics
- Potential improvements or refactoring opportunities
- System and component interactions`, file.Language)
}

// generateAnalysisPrompt creates a dynamic analysis prompt
func (ea *EnhancedAnalyzer) generateAnalysisPrompt(file collector.FileInfo) string {
	return fmt.Sprintf(`Analyze this %s code comprehensively:

%s

Please provide a detailed JSON analysis covering:
- Architectural patterns discovered
- Code quality metrics
- Architectural insights and potential improvements
- Cross-component references`, file.Language, file.Content)
}

// AnalyzeWithInsights performs enhanced analysis of source code
func (ea *EnhancedAnalyzer) AnalyzeWithInsights(ctx context.Context, file collector.FileInfo) (*CodeAnalysisSchema, error) {
	// Validate input
	if file.Content == "" {
		return nil, fmt.Errorf("empty file content")
	}

	// Prepare OpenAI request
	resp, err := ea.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(ea.generateSystemPrompt(file)),
			openai.UserMessage(ea.generateAnalysisPrompt(file)),
		}),
		Model:       openai.F(openai.ChatModelGPT4o),
		Temperature: openai.F(0.3), // Lower temperature for more consistent results
	}, option.WithHeader("Accept", "application/json"))

	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Validate response
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no analysis response received")
	}

	// Parse the response
	var analysis CodeAnalysisSchema
	content := resp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &analysis, nil
}
