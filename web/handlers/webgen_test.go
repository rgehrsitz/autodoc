// web/handlers/webgen_test.go

package handlers

import (
	"testing"
	"time"

	"github.com/rgehrsitz/AutoDoc/internal/storage"
	"github.com/rgehrsitz/AutoDoc/internal/testutil"
)

func TestGenerator(t *testing.T) {
	// Initialize test helper
	helper := testutil.NewTemplateTestHelper(t)

	// Create mock storage
	store := NewMockStorage()

	// Add test documents
	archDoc := &storage.Document{
		ID:        "arch1",
		Type:      storage.TypeArchitecture,
		Path:      "architecture.md",
		Content:   "# Architecture\n\nThis is the architecture overview.",
		UpdatedAt: time.Now(),
	}
	if err := store.SaveDocument(archDoc); err != nil {
		t.Fatalf("Failed to save architecture document: %v", err)
	}

	moduleDoc := &storage.Document{
		ID:        "mod1",
		Type:      storage.TypeModule,
		Path:      "pkg/example/example.go",
		Content:   "# Example Package\n\nThis is an example package.",
		UpdatedAt: time.Now(),
	}
	if err := store.SaveDocument(moduleDoc); err != nil {
		t.Fatalf("Failed to save module document: %v", err)
	}

	// Add test reference
	ref := &storage.Reference{
		SourceID: moduleDoc.ID,
		TargetID: archDoc.ID,
		Type:     "import",
	}
	if err := store.SaveReference(ref); err != nil {
		t.Fatalf("Failed to save reference: %v", err)
	}

	// Test index page generation
	indexData := helper.CreateTestData()
	rendered := helper.RenderTemplate("index", indexData)
	helper.AssertTemplateContains(rendered, "Test Project")
	helper.AssertTemplateContains(rendered, "Test Description")

	// Test architecture page generation
	archData := helper.CreateTestData()
	archData.Title = "Architecture"
	rendered = helper.RenderTemplate("architecture", archData)
	helper.AssertTemplateContains(rendered, "Architecture Overview")

	// Test component page generation
	compData := helper.CreateTestData()
	compData.Title = "Example Package"
	rendered = helper.RenderTemplate("component", compData)
	helper.AssertTemplateContains(rendered, "Example Package")
	helper.AssertTemplateContains(rendered, "Test component description")
}
