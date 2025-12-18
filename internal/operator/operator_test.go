package operator

import (
	"testing"
)

func TestOperator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		op      Operator
		wantErr bool
	}{
		{
			name: "valid operator",
			op: Operator{
				Name:      "test-operator",
				Namespace: "test-ns",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			op: Operator{
				Namespace: "test-ns",
			},
			wantErr: true,
		},
		{
			name: "missing namespace",
			op: Operator{
				Name: "test-operator",
			},
			wantErr: true,
		},
		{
			name: "valid install mode",
			op: Operator{
				Name:        "test-operator",
				Namespace:   "test-ns",
				InstallMode: "AllNamespaces",
			},
			wantErr: false,
		},
		{
			name: "invalid install mode",
			op: Operator{
				Name:        "test-operator",
				Namespace:   "test-ns",
				InstallMode: "InvalidMode",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.op.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOperator_GetSource(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		defaultSource string
		want          string
	}{
		{"custom source", "custom-operators", "community-operators", "custom-operators"},
		{"default source", "", "community-operators", "community-operators"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := Operator{Source: tt.source}
			if got := op.GetSource(tt.defaultSource); got != tt.want {
				t.Errorf("GetSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperator_GetChannel(t *testing.T) {
	tests := []struct {
		name    string
		channel string
		want    string
	}{
		{"custom channel", "beta", "beta"},
		{"default channel", "", "stable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := Operator{Channel: tt.channel}
			if got := op.GetChannel(); got != tt.want {
				t.Errorf("GetChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperator_GetInstallPlanApproval(t *testing.T) {
	tests := []struct {
		name     string
		approval string
		want     string
	}{
		{"manual approval", "Manual", "Manual"},
		{"default approval", "", "Automatic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := Operator{InstallPlanApproval: tt.approval}
			if got := op.GetInstallPlanApproval(); got != tt.want {
				t.Errorf("GetInstallPlanApproval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperator_GetInstallMode(t *testing.T) {
	tests := []struct {
		name        string
		installMode string
		want        InstallMode
	}{
		{"all namespaces", "AllNamespaces", InstallModeAllNamespaces},
		{"own namespace", "OwnNamespace", InstallModeOwnNamespace},
		{"default", "", InstallModeOwnNamespace},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := Operator{InstallMode: tt.installMode}
			if got := op.GetInstallMode(); got != tt.want {
				t.Errorf("GetInstallMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperator_ToSubscriptionManifest(t *testing.T) {
	op := Operator{
		Name:                "test-operator",
		Namespace:           "test-ns",
		Channel:             "beta",
		Source:              "custom-operators",
		SourceNamespace:     "custom-marketplace",
		InstallPlanApproval: "Manual",
		Version:             "1.0.0",
	}

	manifest := op.ToSubscriptionManifest("default-source", "default-ns")

	if manifest.Name != "test-operator" {
		t.Errorf("Name = %v, want test-operator", manifest.Name)
	}
	if manifest.Namespace != "test-ns" {
		t.Errorf("Namespace = %v, want test-ns", manifest.Namespace)
	}
	if manifest.Channel != "beta" {
		t.Errorf("Channel = %v, want beta", manifest.Channel)
	}
	if manifest.Source != "custom-operators" {
		t.Errorf("Source = %v, want custom-operators", manifest.Source)
	}
	if manifest.SourceNamespace != "custom-marketplace" {
		t.Errorf("SourceNamespace = %v, want custom-marketplace", manifest.SourceNamespace)
	}
	if manifest.InstallPlanApproval != "Manual" {
		t.Errorf("InstallPlanApproval = %v, want Manual", manifest.InstallPlanApproval)
	}
	if manifest.StartingCSV != "1.0.0" {
		t.Errorf("StartingCSV = %v, want 1.0.0", manifest.StartingCSV)
	}
}

func TestOperator_ToGroupManifest(t *testing.T) {
	tests := []struct {
		name                 string
		op                   Operator
		wantTargetNamespaces []string
	}{
		{
			name: "own namespace",
			op: Operator{
				Name:        "test-operator",
				Namespace:   "test-ns",
				InstallMode: "OwnNamespace",
			},
			wantTargetNamespaces: []string{"test-ns"},
		},
		{
			name: "all namespaces",
			op: Operator{
				Name:        "test-operator",
				Namespace:   "test-ns",
				InstallMode: "AllNamespaces",
			},
			wantTargetNamespaces: []string{},
		},
		{
			name: "custom target namespaces",
			op: Operator{
				Name:             "test-operator",
				Namespace:        "test-ns",
				TargetNamespaces: []string{"ns1", "ns2"},
			},
			wantTargetNamespaces: []string{"ns1", "ns2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := tt.op.ToGroupManifest()
			if len(manifest.TargetNamespaces) != len(tt.wantTargetNamespaces) {
				t.Errorf("TargetNamespaces length = %v, want %v", len(manifest.TargetNamespaces), len(tt.wantTargetNamespaces))
			}
		})
	}
}

func TestGetOperatorPreset(t *testing.T) {
	tests := []struct {
		name   string
		preset string
		want   bool
	}{
		{"prometheus", "prometheus", true},
		{"grafana", "grafana", true},
		{"cert-manager", "cert-manager", true},
		{"unknown", "unknown-operator", false},
		{"case insensitive", "PROMETHEUS", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := GetOperatorPreset(tt.preset)
			if ok != tt.want {
				t.Errorf("GetOperatorPreset(%s) ok = %v, want %v", tt.preset, ok, tt.want)
			}
		})
	}
}

func TestListOperatorPresets(t *testing.T) {
	presets := ListOperatorPresets()
	if len(presets) == 0 {
		t.Error("ListOperatorPresets() should return non-empty list")
	}

	// Check that expected presets are in the list
	expected := []string{"prometheus", "grafana", "cert-manager"}
	for _, e := range expected {
		found := false
		for _, p := range presets {
			if p == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected preset %s not found in list", e)
		}
	}
}

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	if config.Enabled {
		t.Error("Default config should have Enabled = false")
	}
	if len(config.Operators) != 0 {
		t.Error("Default config should have empty Operators list")
	}
	if !config.CreateOperatorGroup {
		t.Error("Default config should have CreateOperatorGroup = true")
	}
	if config.DefaultSource != string(CatalogSourceCommunity) {
		t.Errorf("DefaultSource = %v, want %v", config.DefaultSource, CatalogSourceCommunity)
	}
}

func TestManager_AddOperator(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	op := &Operator{
		Name:      "test-operator",
		Namespace: "test-ns",
		Enabled:   true,
	}

	err := manager.AddOperator(op)
	if err != nil {
		t.Errorf("AddOperator() error = %v", err)
	}

	if len(manager.GetConfig().Operators) != 1 {
		t.Error("Operator should be added")
	}
}

func TestManager_AddOperator_Invalid(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	op := &Operator{
		Name: "test-operator",
		// Missing namespace
	}

	err := manager.AddOperator(op)
	if err == nil {
		t.Error("AddOperator() should fail for invalid operator")
	}
}

func TestManager_AddPreset(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	err := manager.AddPreset("prometheus")
	if err != nil {
		t.Errorf("AddPreset() error = %v", err)
	}

	if len(manager.GetConfig().Operators) != 1 {
		t.Error("Preset operator should be added")
	}

	op, ok := manager.GetOperator("prometheus-operator")
	if !ok {
		t.Error("Should find prometheus-operator")
	}
	if op.Channel == "" {
		t.Error("Preset should have channel set")
	}
}

func TestManager_AddPreset_Unknown(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	err := manager.AddPreset("unknown-preset")
	if err == nil {
		t.Error("AddPreset() should fail for unknown preset")
	}
}

func TestManager_RemoveOperator(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	_ = manager.AddOperator(&Operator{Name: "op1", Namespace: "ns1"})
	_ = manager.AddOperator(&Operator{Name: "op2", Namespace: "ns2"})

	removed := manager.RemoveOperator("op1")
	if !removed {
		t.Error("RemoveOperator() should return true")
	}

	if len(manager.GetConfig().Operators) != 1 {
		t.Error("Should have 1 operator after removal")
	}

	_, ok := manager.GetOperator("op1")
	if ok {
		t.Error("op1 should be removed")
	}
}

func TestManager_RemoveOperator_NotFound(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	removed := manager.RemoveOperator("nonexistent")
	if removed {
		t.Error("RemoveOperator() should return false for nonexistent operator")
	}
}

func TestManager_GetEnabledOperators(t *testing.T) {
	config := NewDefaultConfig()
	config.Operators = []Operator{
		{Name: "op1", Namespace: "ns1", Enabled: true},
		{Name: "op2", Namespace: "ns2", Enabled: false},
		{Name: "op3", Namespace: "ns3", Enabled: true},
	}
	manager := NewManager(config)

	enabled := manager.GetEnabledOperators()
	if len(enabled) != 2 {
		t.Errorf("GetEnabledOperators() count = %d, want 2", len(enabled))
	}
}

func TestManager_ValidateAll(t *testing.T) {
	config := NewDefaultConfig()
	config.Operators = []Operator{
		{Name: "op1", Namespace: "ns1"}, // Valid
		{Name: "op2"},                   // Invalid - no namespace
		{Name: "", Namespace: "ns3"},    // Invalid - no name
	}
	manager := NewManager(config)

	errors := manager.ValidateAll()
	if len(errors) != 2 {
		t.Errorf("ValidateAll() errors count = %d, want 2", len(errors))
	}
}

func TestInstallModeConstants(t *testing.T) {
	if InstallModeOwnNamespace != "OwnNamespace" {
		t.Error("InstallModeOwnNamespace constant wrong")
	}
	if InstallModeSingleNamespace != "SingleNamespace" {
		t.Error("InstallModeSingleNamespace constant wrong")
	}
	if InstallModeMultiNamespace != "MultiNamespace" {
		t.Error("InstallModeMultiNamespace constant wrong")
	}
	if InstallModeAllNamespaces != "AllNamespaces" {
		t.Error("InstallModeAllNamespaces constant wrong")
	}
}

func TestCatalogSourceConstants(t *testing.T) {
	if CatalogSourceCommunity != "community-operators" {
		t.Error("CatalogSourceCommunity constant wrong")
	}
	if CatalogSourceRedHat != "redhat-operators" {
		t.Error("CatalogSourceRedHat constant wrong")
	}
	if CatalogSourceCertified != "certified-operators" {
		t.Error("CatalogSourceCertified constant wrong")
	}
}
