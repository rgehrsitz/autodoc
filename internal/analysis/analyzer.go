// autodoc/internal/analysis/analyzer.go

package analyzer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/rgehrsitz/AutoDoc/internal/collector"
)

// Analysis represents the LLM's understanding of a code file
type Analysis struct {
	Purpose    string      `json:"purpose"`
	Components []Component `json:"components"`
	Relations  []Relation  `json:"relationships"`
	Insights   []string    `json:"insights"`
}

// Component represents a code component identified by the LLM
type Component struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Description     string   `json:"description"`
	Visibility      string   `json:"visibility"`
	Dependencies    []string `json:"dependencies"`
	NotableFeatures []string `json:"notable_features"`
}

// Relation represents a relationship between components
type Relation struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// Analyzer handles code analysis using LLM
type Analyzer struct {
	openAI  *openai.Client
	prompts map[string]string
}

// NewAnalyzer creates a new Analyzer instance
func NewAnalyzer(openAIKey string) *Analyzer {
	return &Analyzer{
		openAI: openai.NewClient(
			option.WithAPIKey(openAIKey),
		),
		prompts: loadPrompts(),
	}
}

// AnalyzeFile analyzes a single file using the LLM
func (a *Analyzer) AnalyzeFile(ctx context.Context, file collector.FileInfo) (*Analysis, error) {
	// Get appropriate prompt for file type
	prompt := a.getPrompt(file.Language, file.Type)

	// Create chat completion request
	resp, err := a.openAI.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(a.prompts["system"]),
			openai.UserMessage(fmt.Sprintf(prompt, file.Content)),
		}),
		Model: openai.F(openai.ChatModelChatgpt4oLatest),
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	// Parse the response into our Analysis struct
	var analysis Analysis
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis: %w", err)
	}

	return &analysis, nil
}

// loadPrompts loads the analysis prompts
func loadPrompts() map[string]string {
	return map[string]string{
		"system": `You are an expert code analyzer. Analyze the provided code and extract key information in JSON format.
Focus on understanding:
- Purpose and functionality
- Public interfaces/types
- Dependencies and relationships
- Key patterns and practices
- Notable features or concerns`,

		"go_source": `Analyze this Go code and provide structured information about its contents:

%s

Provide your analysis in the following JSON format.`,

		"csharp_source": `Analyze this C# code and provide structured information about its contents:

%s

Provide your analysis in the following JSON format.`,

		"csharp_project": `Analyze this C# project file and provide structured information about its contents:

%s

Provide your analysis in the following JSON format.`,

		"csharp_solution": `Analyze this C# solution file and provide structured information about its contents:

%s

Provide your analysis in the following JSON format.`,
	}
}

// getPrompt returns the appropriate prompt for the file type
func (a *Analyzer) getPrompt(language, fileType string) string {
	key := fmt.Sprintf("%s_%s", language, fileType)
	if prompt, ok := a.prompts[key]; ok {
		return prompt
	}
	// Default to source prompt if specific one not found
	return a.prompts[fmt.Sprintf("%s_source", language)]
}
