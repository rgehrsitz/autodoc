// internal/testutil/template_test_helper.go

package testutil

import (
	"bytes"
	"testing"

	"github.com/rgehrsitz/AutoDoc/web/handlers/templates"
)

// TemplateTestHelper provides utilities for testing templates
type TemplateTestHelper struct {
	t      *testing.T
	engine *templates.TemplateEngine
}

// NewTemplateTestHelper creates a new template test helper
func NewTemplateTestHelper(t *testing.T) *TemplateTestHelper {
	// Create template engine with in-memory templates
	engine, err := templates.NewTemplateEngine("")
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	return &TemplateTestHelper{
		t:      t,
		engine: engine,
	}
}

// RenderTemplate renders a template with test data
func (h *TemplateTestHelper) RenderTemplate(name string, data interface{}) string {
	var buf bytes.Buffer
	err := h.engine.RenderTemplate(&buf, name, data)
	if err != nil {
		h.t.Fatalf("Failed to render template %s: %v", name, err)
	}
	return buf.String()
}

// AssertTemplateContains checks if rendered template contains expected content
func (h *TemplateTestHelper) AssertTemplateContains(rendered, expected string) {
	if !bytes.Contains([]byte(rendered), []byte(expected)) {
		h.t.Errorf("Template output did not contain expected content.\nExpected substring: %s\nGot: %s",
			expected, rendered)
	}
}

// AssertTemplateNotContains checks if rendered template doesn't contain unexpected content
func (h *TemplateTestHelper) AssertTemplateNotContains(rendered, unexpected string) {
	if bytes.Contains([]byte(rendered), []byte(unexpected)) {
		h.t.Errorf("Template output contained unexpected content.\nUnexpected substring: %s\nGot: %s",
			unexpected, rendered)
	}
}

// CreateTestData creates common test data structures
func (h *TemplateTestHelper) CreateTestData() *templates.TemplateData {
	return &templates.TemplateData{
		Title:       "Test Page",
		ProjectName: "Test Project",
		Description: "Test Description",
		Version:     "1.0.0",
		Components: []templates.ComponentData{
			{
				Name:        "TestComponent",
				Path:        "test/path",
				Type:        "test",
				Description: "Test component description",
			},
		},
		Navigation: []templates.NavigationItem{
			{
				Title: "Home",
				URL:   "/",
			},
		},
		Theme: "light",
	}
}
