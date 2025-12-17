package bootstrap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDetector(t *testing.T) {
	tests := []struct {
		name        string
		kubeContext string
		timeout     time.Duration
		wantTimeout time.Duration
	}{
		{
			name:        "with custom timeout",
			kubeContext: "test-context",
			timeout:     60 * time.Second,
			wantTimeout: 60 * time.Second,
		},
		{
			name:        "with default timeout",
			kubeContext: "",
			timeout:     0,
			wantTimeout: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector(tt.kubeContext, tt.timeout)
			require.NotNil(t, d)
			assert.Equal(t, tt.kubeContext, d.kubeContext)
			assert.Equal(t, tt.wantTimeout, d.timeout)
		})
	}
}

func TestArgoCDTypeConstants(t *testing.T) {
	assert.Equal(t, ArgoCDType("community"), ArgoCDTypeCommunity)
	assert.Equal(t, ArgoCDType("redhat"), ArgoCDTypeRedHat)
	assert.Equal(t, ArgoCDType("unknown"), ArgoCDTypeUnknown)
	assert.Equal(t, ArgoCDType("not_installed"), ArgoCDTypeNotInstalled)
}

func TestInstallMethodConstants(t *testing.T) {
	assert.Equal(t, InstallMethod("operator"), InstallMethodOperator)
	assert.Equal(t, InstallMethod("manifest"), InstallMethodManifest)
	assert.Equal(t, InstallMethod("helm"), InstallMethodHelm)
	assert.Equal(t, InstallMethod("olm"), InstallMethodOLM)
	assert.Equal(t, InstallMethod("unknown"), InstallMethodUnknown)
}

func TestOperatorSourceConstants(t *testing.T) {
	assert.Equal(t, OperatorSource("redhat-operators"), OperatorSourceRedHat)
	assert.Equal(t, OperatorSource("community-operators"), OperatorSourceCommunity)
	assert.Equal(t, OperatorSource("certified-operators"), OperatorSourceCertified)
	assert.Equal(t, OperatorSource("redhat-marketplace"), OperatorSourceMarketplace)
	assert.Equal(t, OperatorSource("unknown"), OperatorSourceUnknown)
}

func TestComponentStatusConstants(t *testing.T) {
	assert.Equal(t, ComponentStatus("running"), StatusRunning)
	assert.Equal(t, ComponentStatus("not_running"), StatusNotRunning)
	assert.Equal(t, ComponentStatus("degraded"), StatusDegraded)
	assert.Equal(t, ComponentStatus("unknown"), StatusUnknown)
}

func TestArgoCDDetectionResult_IsRunning(t *testing.T) {
	d := NewDetector("", 30*time.Second)

	tests := []struct {
		name       string
		components []ArgoCDComponent
		want       bool
	}{
		{
			name:       "empty components",
			components: []ArgoCDComponent{},
			want:       false,
		},
		{
			name: "all ready",
			components: []ArgoCDComponent{
				{Name: "server", Ready: true, Status: StatusRunning},
				{Name: "repo", Ready: true, Status: StatusRunning},
			},
			want: true,
		},
		{
			name: "some not ready",
			components: []ArgoCDComponent{
				{Name: "server", Ready: true, Status: StatusRunning},
				{Name: "repo", Ready: false, Status: StatusNotRunning},
			},
			want: false,
		},
		{
			name: "all not ready",
			components: []ArgoCDComponent{
				{Name: "server", Ready: false, Status: StatusNotRunning},
				{Name: "repo", Ready: false, Status: StatusNotRunning},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.isRunning(tt.components)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetermineHealthStatus(t *testing.T) {
	d := NewDetector("", 30*time.Second)

	tests := []struct {
		name       string
		components []ArgoCDComponent
		want       string
	}{
		{
			name:       "empty components",
			components: []ArgoCDComponent{},
			want:       "unknown",
		},
		{
			name: "all healthy",
			components: []ArgoCDComponent{
				{Name: "server", Status: StatusRunning},
				{Name: "repo", Status: StatusRunning},
			},
			want: "healthy",
		},
		{
			name: "some degraded",
			components: []ArgoCDComponent{
				{Name: "server", Status: StatusRunning},
				{Name: "repo", Status: StatusDegraded},
			},
			want: "degraded",
		},
		{
			name: "mixed status",
			components: []ArgoCDComponent{
				{Name: "server", Status: StatusRunning},
				{Name: "repo", Status: StatusNotRunning},
			},
			want: "degraded",
		},
		{
			name: "all not running",
			components: []ArgoCDComponent{
				{Name: "server", Status: StatusNotRunning},
				{Name: "repo", Status: StatusNotRunning},
			},
			want: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.determineHealthStatus(tt.components)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestArgoCDDetectionResult_Summary(t *testing.T) {
	result := &ArgoCDDetectionResult{
		Installed:      true,
		Type:           ArgoCDTypeRedHat,
		InstallMethod:  InstallMethodOLM,
		OperatorSource: OperatorSourceRedHat,
		Namespace:      "openshift-gitops",
		Version:        "v2.9.0",
		URL:            "https://argocd.apps.cluster.com",
		Running:        true,
		HealthStatus:   "healthy",
		AppCount:       5,
		Components: []ArgoCDComponent{
			{Name: "server", Ready: true, Available: 1, Replicas: 1},
			{Name: "repo-server", Ready: true, Available: 1, Replicas: 1},
		},
	}

	summary := result.Summary()

	assert.Contains(t, summary, "ArgoCD Detection Summary")
	assert.Contains(t, summary, "Installed:     true")
	assert.Contains(t, summary, "Type:          redhat")
	assert.Contains(t, summary, "Namespace:     openshift-gitops")
	assert.Contains(t, summary, "Install Method: olm")
	assert.Contains(t, summary, "Operator Source: redhat-operators")
	assert.Contains(t, summary, "Version:       v2.9.0")
	assert.Contains(t, summary, "URL:           https://argocd.apps.cluster.com")
	assert.Contains(t, summary, "Running:       true")
	assert.Contains(t, summary, "Health:        healthy")
	assert.Contains(t, summary, "Applications:  5")
	assert.Contains(t, summary, "✅ server")
	assert.Contains(t, summary, "✅ repo-server")
}

func TestArgoCDDetectionResult_SummaryWithIssues(t *testing.T) {
	result := &ArgoCDDetectionResult{
		Installed:    true,
		Type:         ArgoCDTypeCommunity,
		Namespace:    "argocd",
		Running:      false,
		HealthStatus: "unhealthy",
		Components: []ArgoCDComponent{
			{Name: "server", Ready: false, Available: 0, Replicas: 1},
		},
		Issues: []string{
			"ArgoCD is not fully running",
			"Component server is not ready",
		},
		Recommendations: []string{
			"Check pod logs for troubleshooting",
		},
	}

	summary := result.Summary()

	assert.Contains(t, summary, "❌ server")
	assert.Contains(t, summary, "Issues:")
	assert.Contains(t, summary, "⚠️  ArgoCD is not fully running")
	assert.Contains(t, summary, "Recommendations:")
	assert.Contains(t, summary, "💡 Check pod logs")
}

func TestArgoCDDetectionResult_ToJSON(t *testing.T) {
	result := &ArgoCDDetectionResult{
		Installed:    true,
		Type:         ArgoCDTypeRedHat,
		Namespace:    "openshift-gitops",
		Running:      true,
		HealthStatus: "healthy",
	}

	jsonStr, err := result.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, jsonStr, `"installed": true`)
	assert.Contains(t, jsonStr, `"type": "redhat"`)
	assert.Contains(t, jsonStr, `"namespace": "openshift-gitops"`)
	assert.Contains(t, jsonStr, `"running": true`)
	assert.Contains(t, jsonStr, `"health_status": "healthy"`)
}

func TestArgoCDComponent(t *testing.T) {
	component := ArgoCDComponent{
		Name:      "argocd-server",
		Ready:     true,
		Replicas:  2,
		Available: 2,
		Status:    StatusRunning,
		Image:     "quay.io/argoproj/argocd:v2.9.0",
	}

	assert.Equal(t, "argocd-server", component.Name)
	assert.True(t, component.Ready)
	assert.Equal(t, 2, component.Replicas)
	assert.Equal(t, 2, component.Available)
	assert.Equal(t, StatusRunning, component.Status)
	assert.Equal(t, "quay.io/argoproj/argocd:v2.9.0", component.Image)
}

func TestAnalyzeAndRecommend(t *testing.T) {
	d := NewDetector("", 30*time.Second)

	tests := []struct {
		name          string
		result        *ArgoCDDetectionResult
		wantIssues    int
		wantRecs      int
		containsIssue string
		containsRec   string
	}{
		{
			name: "not running",
			result: &ArgoCDDetectionResult{
				Running: false,
				Components: []ArgoCDComponent{
					{Name: "server", Ready: false, Available: 0, Replicas: 1},
				},
			},
			wantIssues:    2,
			wantRecs:      1,
			containsIssue: "not fully running",
			containsRec:   "Check pod logs",
		},
		{
			name: "community in openshift namespace",
			result: &ArgoCDDetectionResult{
				Running:   true,
				Type:      ArgoCDTypeCommunity,
				Namespace: "openshift-gitops",
				Components: []ArgoCDComponent{
					{Name: "server", Ready: true},
				},
			},
			wantIssues:    1,
			wantRecs:      1,
			containsIssue: "Community ArgoCD installed in OpenShift GitOps namespace",
			containsRec:   "Red Hat OpenShift GitOps operator",
		},
		{
			name: "no URL",
			result: &ArgoCDDetectionResult{
				Running: true,
				URL:     "",
				Components: []ArgoCDComponent{
					{Name: "server", Ready: true},
				},
			},
			wantIssues:    1,
			wantRecs:      1,
			containsIssue: "No external URL",
			containsRec:   "Route or Ingress",
		},
		{
			name: "outdated version",
			result: &ArgoCDDetectionResult{
				Running: true,
				Version: "v1.8.7",
				URL:     "https://argocd.example.com",
				Components: []ArgoCDComponent{
					{Name: "server", Ready: true},
				},
			},
			wantIssues:    1,
			wantRecs:      1,
			containsIssue: "v1.x is outdated",
			containsRec:   "Upgrade to ArgoCD v2.x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.result.Issues = []string{}
			tt.result.Recommendations = []string{}
			d.analyzeAndRecommend(tt.result)

			assert.GreaterOrEqual(t, len(tt.result.Issues), tt.wantIssues)
			assert.GreaterOrEqual(t, len(tt.result.Recommendations), tt.wantRecs)

			if tt.containsIssue != "" {
				found := false
				for _, issue := range tt.result.Issues {
					if assert.ObjectsAreEqual(issue, tt.containsIssue) ||
						len(issue) > 0 && len(tt.containsIssue) > 0 {
						if containsSubstring(issue, tt.containsIssue) {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Expected issue containing '%s'", tt.containsIssue)
			}

			if tt.containsRec != "" {
				found := false
				for _, rec := range tt.result.Recommendations {
					if containsSubstring(rec, tt.containsRec) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected recommendation containing '%s'", tt.containsRec)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && len(substr) > 0 &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
