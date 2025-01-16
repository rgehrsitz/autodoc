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
func (a *Analyzer) AnalyzeFile(ctx context.Context, file collector.FileInfo) (*Analysis, string, error) {
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
		return nil, "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, "", fmt.Errorf("no response from LLM")
	}

	rawResponse := resp.Choices[0].Message.Content

	// Parse the response into our Analysis struct
	var analysis Analysis
	if err := json.Unmarshal([]byte(rawResponse), &analysis); err != nil {
		return nil, rawResponse, fmt.Errorf("failed to parse analysis: %w", err)
	}

	return &analysis, rawResponse, nil
}

// loadPrompts loads the analysis prompts
func loadPrompts() map[string]string {
	return map[string]string{
		"system": `You are an expert code analyzer. Analyze the provided code and return ONLY a JSON object with the following structure:
{
    "purpose": "Brief description of the code's purpose",
    "components": [
        {
            "name": "Component name",
            "type": "Type of component",
            "description": "Component description",
            "visibility": "public/private/etc",
            "dependencies": ["List of dependencies"],
            "notable_features": ["List of notable features"]
        }
    ],
    "relationships": [
        {
            "from": "Source component",
            "to": "Target component",
            "type": "Relationship type"
        }
    ],
    "insights": ["List of important observations"]
}
Do not include any text before or after the JSON. Do not use markdown formatting.`,

		"go_source": `Analyze this Go code and provide your analysis as a JSON object matching the specified structure:

%s`,

		"csharp_source": `Analyze this C# code and provide your analysis as a JSON object matching the specified structure:

%s`,

		"csharp_project": `Analyze this C# project file and provide your analysis as a JSON object matching the specified structure:

%s`,

		"csharp_solution": `Analyze this C# solution file and provide your analysis as a JSON object matching the specified structure:

%s`,
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
