// autodoc/internal/analysis/enhanced.go

package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/rgehrsitz/AutoDoc/internal/collector"
)

// ArchitecturalInsight represents a high-level insight about the code
type ArchitecturalInsight struct {
	Type        string   // Pattern, Concern, Recommendation
	Description string   // Detailed description
	Impact      string   // High, Medium, Low
	Components  []string // Affected components
}

// EnhancedAnalysis extends the basic Analysis with additional insights
type EnhancedAnalysis struct {
	*Analysis
	ArchitecturalPatterns []string
	CodeQualityMetrics    map[string]float64
	Insights              []ArchitecturalInsight
	CrossReferences       map[string][]string
}

// EnhancedAnalyzer provides advanced code analysis capabilities
type EnhancedAnalyzer struct {
	*Analyzer
	patternDatabase map[string]string
}

// NewEnhancedAnalyzer creates a new instance of EnhancedAnalyzer
func NewEnhancedAnalyzer(openAIKey string) *EnhancedAnalyzer {
	return &EnhancedAnalyzer{
		Analyzer:        NewAnalyzer(openAIKey),
		patternDatabase: loadPatternDatabase(),
	}
}

// AnalyzeWithInsights performs enhanced analysis of source code
func (ea *EnhancedAnalyzer) AnalyzeWithInsights(ctx context.Context, code string, language string) (*EnhancedAnalysis, error) {
	// Get base analysis first
	baseAnalysis, err := ea.Analyzer.AnalyzeFile(ctx, collector.FileInfo{
		Path:     "",
		Content:  code,
		Language: language,
		Type:     "source",
	})
	if err != nil {
		return nil, fmt.Errorf("base analysis failed: %w", err)
	}

	// Prepare enhanced analysis prompt
	prompt := buildEnhancedPrompt(code, language)

	// Get enhanced insights from OpenAI
	resp, err := ea.openAI.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(enhancedSystemPrompt),
			openai.UserMessage(prompt),
		}),
		Model: openai.F(openai.ChatModelChatgpt4oLatest),
	})
	if err != nil {
		return nil, fmt.Errorf("enhanced analysis failed: %w", err)
	}

	// Parse enhanced insights
	insights, err := ea.parseArchitecturalInsights(resp.Choices[0].Message.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse insights: %w", err)
	}

	// Build cross-references
	crossRefs := buildCrossReferences(baseAnalysis)

	// Detect architectural patterns
	patterns := detectArchitecturalPatterns(code, ea.patternDatabase)

	// Create enhanced analysis
	enhanced := &EnhancedAnalysis{
		Analysis:              baseAnalysis,
		ArchitecturalPatterns: patterns,
		CodeQualityMetrics:    calculateMetrics(code),
		Insights:              insights,
		CrossReferences:       crossRefs,
	}

	return enhanced, nil
}

const enhancedSystemPrompt = `You are an expert software architect analyzing code for architectural patterns,
design principles, and potential improvements. Focus on:
- Architectural patterns and anti-patterns
- SOLID principles adherence
- Code maintainability and scalability
- Potential technical debt
- Security considerations
- Performance implications`

func buildEnhancedPrompt(code string, language string) string {
	return fmt.Sprintf(`Analyze this %s code for architectural insights:

%s

Provide your analysis in the following JSON format:
{
    "architectural_patterns": ["pattern1", "pattern2"],
    "design_principles": {
        "solid_principles": ["principle1", "principle2"],
        "violations": ["violation1", "violation2"]
    },
    "insights": [
        {
            "type": "pattern|concern|recommendation",
            "description": "detailed description",
            "impact": "high|medium|low",
            "affected_components": ["component1", "component2"]
        }
    ]
}`, language, code)
}

func detectArchitecturalPatterns(code string, patterns map[string]string) []string {
	var detected []string
	for pattern, signature := range patterns {
		if strings.Contains(code, signature) {
			detected = append(detected, pattern)
		}
	}
	return detected
}

func calculateMetrics(code string) map[string]float64 {
	return map[string]float64{
		"cyclomatic_complexity": calculateCyclomaticComplexity(code),
		"maintainability_index": calculateMaintainabilityIndex(code),
		"code_coverage":         estimateCodeCoverage(code),
	}
}

func buildCrossReferences(analysis *Analysis) map[string][]string {
	refs := make(map[string][]string)
	for _, comp := range analysis.Components {
		for _, dep := range comp.Dependencies {
			refs[comp.Name] = append(refs[comp.Name], dep)
		}
	}
	return refs
}

// Helper functions for calculating various metrics
func calculateCyclomaticComplexity(code string) float64 {
	// Count decision points (if, for, while, case, &&, ||)
	decisionPoints := strings.Count(code, "if ") +
		strings.Count(code, "for ") +
		strings.Count(code, "while ") +
		strings.Count(code, "case ") +
		strings.Count(code, "&&") +
		strings.Count(code, "||")
	return float64(decisionPoints + 1)
}

func calculateMaintainabilityIndex(code string) float64 {
	// Simplified MI calculation
	loc := float64(len(strings.Split(code, "\n")))
	cc := calculateCyclomaticComplexity(code)
	return 171 - 5.2*cc - 0.23*loc
}

func estimateCodeCoverage(code string) float64 {
	// Simple estimation based on presence of test-related keywords
	if strings.Contains(code, "_test.go") ||
		strings.Contains(code, "func Test") ||
		strings.Contains(code, "t.Run(") {
		return 80.0 // Assume good coverage if tests exist
	}
	return 0.0
}

// AnalysisResponse represents the JSON response from OpenAI
type AnalysisResponse struct {
	ArchitecturalPatterns []string `json:"architectural_patterns"`
	DesignPrinciples      struct {
		SolidPrinciples []string `json:"solid_principles"`
		Violations      []string `json:"violations"`
	} `json:"design_principles"`
	Insights []struct {
		Type               string   `json:"type"`
		Description        string   `json:"description"`
		Impact             string   `json:"impact"`
		AffectedComponents []string `json:"affected_components"`
	} `json:"insights"`
}

// parseArchitecturalInsights parses the LLM response into architectural insights
func (ea *EnhancedAnalyzer) parseArchitecturalInsights(content string) ([]ArchitecturalInsight, error) {
	var response AnalysisResponse
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	insights := make([]ArchitecturalInsight, 0, len(response.Insights))
	for _, insight := range response.Insights {
		insights = append(insights, ArchitecturalInsight{
			Type:        insight.Type,
			Description: insight.Description,
			Impact:      insight.Impact,
			Components:  insight.AffectedComponents,
		})
	}

	return insights, nil
}

func loadPatternDatabase() map[string]string {
	return map[string]string{
		"Singleton":            "sync.Once",
		"Factory":              "New[A-Z]",
		"Observer":             "func.*Subscribe|Notify",
		"Strategy":             "interface.*Execute",
		"Repository":           "interface.*Repository",
		"Dependency Injection": "func New.*\\(.*\\*.*\\)",
	}
}
