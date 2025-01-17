// autodoc/internal/analysis/converter.go

package analyzer

// ConvertToCodeAnalysis converts an Analysis to a CodeAnalysisSchema
func ConvertToCodeAnalysis(analysis *Analysis) *CodeAnalysisSchema {
	if analysis == nil {
		return nil
	}

	// Convert insights to ArchitecturalInsights
	insights := make([]ArchitecturalInsight, len(analysis.Insights))
	for i, insight := range analysis.Insights {
		insights[i] = ArchitecturalInsight{
			Type:        "general",
			Description: insight,
			Impact:      "medium", // Default impact
		}
	}

	// Convert relationships to cross-references
	crossRefs := make(map[string][]string)
	for _, rel := range analysis.Relations {
		if refs, exists := crossRefs[rel.From]; exists {
			crossRefs[rel.From] = append(refs, rel.To)
		} else {
			crossRefs[rel.From] = []string{rel.To}
		}
	}

	return &CodeAnalysisSchema{
		Insights:        insights,
		CrossReferences: crossRefs,
		// Initialize empty but non-nil slices/maps for other fields
		ArchitecturalPatterns: []string{},
		CodeQualityMetrics:    make(map[string]any),
	}
}
