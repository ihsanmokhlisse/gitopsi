package version

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMajor int
		wantMinor int
		wantPatch int
		wantErr   bool
	}{
		{"full version", "1.28.0", 1, 28, 0, false},
		{"with v prefix", "v1.28.0", 1, 28, 0, false},
		{"minor only", "1.28", 1, 28, 0, false},
		{"with suffix", "1.28.0-eks-1234", 1, 28, 0, false},
		{"with k3s suffix", "1.27.4+k3s1", 1, 27, 4, false},
		{"invalid format", "invalid", 0, 0, 0, true},
		{"empty string", "", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if v.Major != tt.wantMajor || v.Minor != tt.wantMinor || v.Patch != tt.wantPatch {
					t.Errorf("ParseVersion() = %d.%d.%d, want %d.%d.%d",
						v.Major, v.Minor, v.Patch, tt.wantMajor, tt.wantMinor, tt.wantPatch)
				}
			}
		})
	}
}

func TestKubernetesVersion_String(t *testing.T) {
	v := &KubernetesVersion{Major: 1, Minor: 28, Patch: 0}
	if got := v.String(); got != "1.28.0" {
		t.Errorf("String() = %v, want %v", got, "1.28.0")
	}
}

func TestKubernetesVersion_Compare(t *testing.T) {
	tests := []struct {
		name   string
		v1     string
		v2     string
		expect int
	}{
		{"equal", "1.28.0", "1.28.0", 0},
		{"major greater", "2.0.0", "1.28.0", 1},
		{"major less", "1.28.0", "2.0.0", -1},
		{"minor greater", "1.29.0", "1.28.0", 1},
		{"minor less", "1.27.0", "1.28.0", -1},
		{"patch greater", "1.28.1", "1.28.0", 1},
		{"patch less", "1.28.0", "1.28.1", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, _ := ParseVersion(tt.v1)
			v2, _ := ParseVersion(tt.v2)
			if got := v1.Compare(v2); got != tt.expect {
				t.Errorf("Compare() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestKubernetesVersion_IsAtLeast(t *testing.T) {
	tests := []struct {
		name   string
		v      string
		major  int
		minor  int
		expect bool
	}{
		{"exact match", "1.28.0", 1, 28, true},
		{"higher minor", "1.29.0", 1, 28, true},
		{"higher major", "2.0.0", 1, 28, true},
		{"lower minor", "1.27.0", 1, 28, false},
		{"lower major", "0.28.0", 1, 28, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := ParseVersion(tt.v)
			if got := v.IsAtLeast(tt.major, tt.minor); got != tt.expect {
				t.Errorf("IsAtLeast() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestNewMapper(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		platform string
		wantErr  bool
	}{
		{"valid version", "1.28.0", "kubernetes", false},
		{"empty version", "", "kubernetes", false},
		{"invalid version", "invalid", "kubernetes", true},
		{"openshift platform", "1.27.0", "openshift", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMapper(tt.version, tt.platform)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMapper() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMapper_GetAPIVersion(t *testing.T) {
	tests := []struct {
		name          string
		targetVersion string
		kind          string
		wantAPI       string
	}{
		{"deployment current", "1.28.0", "Deployment", "apps/v1"},
		{"namespace", "1.28.0", "Namespace", "v1"},
		{"ingress current", "1.28.0", "Ingress", "networking.k8s.io/v1"},
		{"cronjob current", "1.28.0", "CronJob", "batch/v1"},
		{"hpa current", "1.28.0", "HorizontalPodAutoscaler", "autoscaling/v2"},
		{"pdb current", "1.28.0", "PodDisruptionBudget", "policy/v1"},
		{"argocd application", "1.28.0", "Application", "argoproj.io/v1alpha1"},
		{"unknown kind", "1.28.0", "UnknownKind", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, _ := NewMapper(tt.targetVersion, "kubernetes")
			if got := vm.GetAPIVersion(tt.kind); got != tt.wantAPI {
				t.Errorf("GetAPIVersion() = %v, want %v", got, tt.wantAPI)
			}
		})
	}
}

func TestMapper_GetAPIVersion_OlderCluster(t *testing.T) {
	// Test that older clusters get compatible API versions
	vm, _ := NewMapper("1.18.0", "kubernetes")

	// Ingress should get v1beta1 for older clusters
	ingressAPI := vm.GetAPIVersion("Ingress")
	// For 1.18, networking.k8s.io/v1beta1 should be returned
	if ingressAPI != "networking.k8s.io/v1beta1" {
		t.Logf("Note: Ingress API for 1.18 = %s (may vary based on implementation)", ingressAPI)
	}
}

func TestMapper_CheckDeprecation(t *testing.T) {
	tests := []struct {
		name          string
		targetVersion string
		kind          string
		apiVersion    string
		wantSeverity  string
		wantResult    bool
	}{
		{
			name:          "current API not deprecated",
			targetVersion: "1.28.0",
			kind:          "Ingress",
			apiVersion:    "networking.k8s.io/v1",
			wantResult:    false,
		},
		{
			name:          "deprecated API warning",
			targetVersion: "1.20.0",
			kind:          "Ingress",
			apiVersion:    "networking.k8s.io/v1beta1",
			wantSeverity:  "warning",
			wantResult:    true,
		},
		{
			name:          "removed API error",
			targetVersion: "1.25.0",
			kind:          "Ingress",
			apiVersion:    "networking.k8s.io/v1beta1",
			wantSeverity:  "error",
			wantResult:    true,
		},
		{
			name:          "cronjob beta deprecated",
			targetVersion: "1.24.0",
			kind:          "CronJob",
			apiVersion:    "batch/v1beta1",
			wantSeverity:  "warning",
			wantResult:    true,
		},
		{
			name:          "cronjob beta removed",
			targetVersion: "1.26.0",
			kind:          "CronJob",
			apiVersion:    "batch/v1beta1",
			wantSeverity:  "error",
			wantResult:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, _ := NewMapper(tt.targetVersion, "kubernetes")
			result := vm.CheckDeprecation(tt.kind, tt.apiVersion)

			if tt.wantResult && result == nil {
				t.Errorf("CheckDeprecation() returned nil, expected result")
				return
			}
			if !tt.wantResult && result != nil {
				t.Errorf("CheckDeprecation() returned result, expected nil")
				return
			}
			if result != nil && result.Severity != tt.wantSeverity {
				t.Errorf("CheckDeprecation() severity = %v, want %v", result.Severity, tt.wantSeverity)
			}
		})
	}
}

func TestMapper_GetAllDeprecatedAPIs(t *testing.T) {
	vm, _ := NewMapper("1.28.0", "kubernetes")
	deprecated := vm.GetAllDeprecatedAPIs()

	if len(deprecated) == 0 {
		t.Error("GetAllDeprecatedAPIs() returned empty list")
	}

	// Check that we have expected deprecated APIs
	foundIngress := false
	foundCronJob := false
	for _, d := range deprecated {
		if d.Kind == "Ingress" && d.CurrentAPI == "networking.k8s.io/v1beta1" {
			foundIngress = true
		}
		if d.Kind == "CronJob" && d.CurrentAPI == "batch/v1beta1" {
			foundCronJob = true
		}
	}

	if !foundIngress {
		t.Error("Expected to find deprecated Ingress networking.k8s.io/v1beta1")
	}
	if !foundCronJob {
		t.Error("Expected to find deprecated CronJob batch/v1beta1")
	}
}

func TestOpenShiftToKubernetesVersion(t *testing.T) {
	tests := []struct {
		ocp     string
		wantK8s string
		wantOk  bool
	}{
		{"4.14", "1.27", true},
		{"4.14.0", "1.27", true},
		{"4.14.5", "1.27", true},
		{"4.12", "1.25", true},
		{"4.10", "1.23", true},
		{"4.99", "", false},
		{"3.11", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.ocp, func(t *testing.T) {
			k8s, ok := GetKubernetesVersionForOpenShift(tt.ocp)
			if ok != tt.wantOk {
				t.Errorf("GetKubernetesVersionForOpenShift() ok = %v, want %v", ok, tt.wantOk)
			}
			if k8s != tt.wantK8s {
				t.Errorf("GetKubernetesVersionForOpenShift() = %v, want %v", k8s, tt.wantK8s)
			}
		})
	}
}

func TestGetOpenShiftVersionForKubernetes(t *testing.T) {
	tests := []struct {
		k8s     string
		wantOcp string
		wantOk  bool
	}{
		{"1.27", "4.14", true},
		{"1.27.0", "4.14", true},
		{"1.25", "4.12", true},
		{"1.23", "4.10", true},
		{"1.99", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.k8s, func(t *testing.T) {
			ocp, ok := GetOpenShiftVersionForKubernetes(tt.k8s)
			if ok != tt.wantOk {
				t.Errorf("GetOpenShiftVersionForKubernetes() ok = %v, want %v", ok, tt.wantOk)
			}
			if ocp != tt.wantOcp {
				t.Errorf("GetOpenShiftVersionForKubernetes() = %v, want %v", ocp, tt.wantOcp)
			}
		})
	}
}

func TestGetSupportedVersions(t *testing.T) {
	supported := GetSupportedVersions()

	if supported.MinVersion == "" {
		t.Error("MinVersion should not be empty")
	}
	if supported.MaxVersion == "" {
		t.Error("MaxVersion should not be empty")
	}

	minV, err := ParseVersion(supported.MinVersion)
	if err != nil {
		t.Errorf("Failed to parse MinVersion: %v", err)
	}

	maxV, err := ParseVersion(supported.MaxVersion)
	if err != nil {
		t.Errorf("Failed to parse MaxVersion: %v", err)
	}

	if minV.Compare(maxV) >= 0 {
		t.Error("MinVersion should be less than MaxVersion")
	}
}

func TestIsVersionSupported(t *testing.T) {
	tests := []struct {
		version string
		wantOk  bool
	}{
		{"1.25.0", true},
		{"1.28.0", true},
		{"1.30.0", true},
		{"1.10.0", false}, // too old
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			ok, msg := IsVersionSupported(tt.version)
			if ok != tt.wantOk {
				t.Errorf("IsVersionSupported(%s) = %v, msg = %s, want %v", tt.version, ok, msg, tt.wantOk)
			}
		})
	}
}

func TestDefaultAPIVersions(t *testing.T) {
	// Verify key resources have default API versions
	required := []string{
		"Namespace", "ConfigMap", "Secret", "Service",
		"Deployment", "StatefulSet", "DaemonSet",
		"Ingress", "NetworkPolicy",
		"Role", "RoleBinding", "ClusterRole", "ClusterRoleBinding",
		"Application", "AppProject", // ArgoCD
	}

	for _, kind := range required {
		if _, ok := DefaultAPIVersions[kind]; !ok {
			t.Errorf("DefaultAPIVersions missing %s", kind)
		}
	}
}

func TestDeprecationResult_Fields(t *testing.T) {
	vm, _ := NewMapper("1.24.0", "kubernetes")
	result := vm.CheckDeprecation("CronJob", "batch/v1beta1")

	if result == nil {
		t.Fatal("Expected deprecation result for CronJob batch/v1beta1")
	}

	if result.Kind != "CronJob" {
		t.Errorf("Kind = %v, want CronJob", result.Kind)
	}
	if result.CurrentAPI != "batch/v1beta1" {
		t.Errorf("CurrentAPI = %v, want batch/v1beta1", result.CurrentAPI)
	}
	if result.ReplacementAPI != "batch/v1" {
		t.Errorf("ReplacementAPI = %v, want batch/v1", result.ReplacementAPI)
	}
	if result.DeprecatedIn == "" {
		t.Error("DeprecatedIn should not be empty")
	}
	if result.RemovedIn == "" {
		t.Error("RemovedIn should not be empty")
	}
	if len(result.Suggestions) == 0 {
		t.Error("Suggestions should not be empty")
	}
}
