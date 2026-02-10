package cluster

import "testing"

// TestNewClusterManager tests ClusterManager creation
func TestNewClusterManager(t *testing.T) {
	cm := NewClusterManager()

	if cm == nil {
		t.Error("NewClusterManager() returned nil")
	}

	if cm.config == nil {
		t.Error("NewClusterManager() config is nil")
	}
}

// TestGetTemplate tests template retrieval
func TestGetTemplate(t *testing.T) {
	cm := NewClusterManager()

	tests := []struct {
		name      string
		templateName string
		wantErr   bool
	}{
		{"default template", "default", false},
		{"gpu template", "gpu", false},
		{"non-existent template", "non-existent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := cm.GetTemplate(tt.templateName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplate() err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tmpl == nil {
				t.Error("GetTemplate() returned nil template")
			}
		})
	}
}

// TestListTemplates tests template listing
func TestListTemplates(t *testing.T) {
	cm := NewClusterManager()

	templates, err := cm.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() err = %v", err)
	}

	if templates == nil {
		t.Error("ListTemplates() returned nil")
	}

	if len(templates) == 0 {
		t.Error("ListTemplates() returned empty list")
	}
}
