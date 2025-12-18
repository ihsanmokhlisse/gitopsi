// Package operator provides Kubernetes Operator management for GitOps.
package operator

import (
	"fmt"
	"strings"
)

// Operator represents a Kubernetes Operator configuration.
type Operator struct {
	// Name is the operator name (e.g., "prometheus-operator")
	Name string `yaml:"name" json:"name"`
	// Namespace where the operator will be installed
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	// Channel is the OLM channel to subscribe to
	Channel string `yaml:"channel,omitempty" json:"channel,omitempty"`
	// Source is the CatalogSource (e.g., "community-operators", "redhat-operators")
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
	// SourceNamespace is the namespace of the CatalogSource
	SourceNamespace string `yaml:"source_namespace,omitempty" json:"source_namespace,omitempty"`
	// Version specifies a specific version to install
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	// InstallPlanApproval is Automatic or Manual
	InstallPlanApproval string `yaml:"install_plan_approval,omitempty" json:"install_plan_approval,omitempty"`
	// Config allows passing custom configuration
	Config map[string]any `yaml:"config,omitempty" json:"config,omitempty"`
	// Enabled controls whether this operator should be deployed
	Enabled bool `yaml:"enabled" json:"enabled"`
	// TargetNamespaces for AllNamespaces/OwnNamespace/SingleNamespace/MultiNamespace install modes
	TargetNamespaces []string `yaml:"target_namespaces,omitempty" json:"target_namespaces,omitempty"`
	// InstallMode is the OLM install mode
	InstallMode string `yaml:"install_mode,omitempty" json:"install_mode,omitempty"`
}

// InstallMode represents OLM operator install modes.
type InstallMode string

const (
	InstallModeOwnNamespace    InstallMode = "OwnNamespace"
	InstallModeSingleNamespace InstallMode = "SingleNamespace"
	InstallModeMultiNamespace  InstallMode = "MultiNamespace"
	InstallModeAllNamespaces   InstallMode = "AllNamespaces"
)

// CatalogSource represents an OLM CatalogSource.
type CatalogSource string

const (
	CatalogSourceCommunity   CatalogSource = "community-operators"
	CatalogSourceRedHat      CatalogSource = "redhat-operators"
	CatalogSourceCertified   CatalogSource = "certified-operators"
	CatalogSourceMarketplace CatalogSource = "redhat-marketplace"
	CatalogSourceOperatorHub CatalogSource = "operatorhubio-catalog"
)

// CommonOperators provides presets for commonly used operators.
var CommonOperators = map[string]Operator{
	"prometheus": {
		Name:            "prometheus-operator",
		Channel:         "beta",
		Source:          string(CatalogSourceCommunity),
		SourceNamespace: "openshift-marketplace",
		Namespace:       "openshift-monitoring",
		InstallMode:     string(InstallModeAllNamespaces),
		Enabled:         true,
	},
	"grafana": {
		Name:            "grafana-operator",
		Channel:         "v5",
		Source:          string(CatalogSourceCommunity),
		SourceNamespace: "openshift-marketplace",
		Namespace:       "grafana",
		InstallMode:     string(InstallModeOwnNamespace),
		Enabled:         true,
	},
	"cert-manager": {
		Name:            "cert-manager",
		Channel:         "stable",
		Source:          string(CatalogSourceCommunity),
		SourceNamespace: "openshift-marketplace",
		Namespace:       "cert-manager",
		InstallMode:     string(InstallModeAllNamespaces),
		Enabled:         true,
	},
	"external-secrets": {
		Name:            "external-secrets-operator",
		Channel:         "stable",
		Source:          string(CatalogSourceCommunity),
		SourceNamespace: "openshift-marketplace",
		Namespace:       "external-secrets",
		InstallMode:     string(InstallModeAllNamespaces),
		Enabled:         true,
	},
	"sealed-secrets": {
		Name:            "sealed-secrets-operator-helm",
		Channel:         "alpha",
		Source:          string(CatalogSourceCommunity),
		SourceNamespace: "openshift-marketplace",
		Namespace:       "sealed-secrets",
		InstallMode:     string(InstallModeOwnNamespace),
		Enabled:         true,
	},
	"kyverno": {
		Name:            "kyverno-operator",
		Channel:         "stable",
		Source:          string(CatalogSourceCommunity),
		SourceNamespace: "openshift-marketplace",
		Namespace:       "kyverno",
		InstallMode:     string(InstallModeAllNamespaces),
		Enabled:         true,
	},
	"opa-gatekeeper": {
		Name:            "gatekeeper-operator-product",
		Channel:         "stable",
		Source:          string(CatalogSourceRedHat),
		SourceNamespace: "openshift-marketplace",
		Namespace:       "gatekeeper-system",
		InstallMode:     string(InstallModeAllNamespaces),
		Enabled:         true,
	},
	"elasticsearch": {
		Name:            "elasticsearch-operator",
		Channel:         "stable-5.8",
		Source:          string(CatalogSourceRedHat),
		SourceNamespace: "openshift-marketplace",
		Namespace:       "openshift-operators-redhat",
		InstallMode:     string(InstallModeAllNamespaces),
		Enabled:         true,
	},
	"jaeger": {
		Name:            "jaeger-product",
		Channel:         "stable",
		Source:          string(CatalogSourceRedHat),
		SourceNamespace: "openshift-marketplace",
		Namespace:       "openshift-distributed-tracing",
		InstallMode:     string(InstallModeAllNamespaces),
		Enabled:         true,
	},
	"servicemesh": {
		Name:            "servicemeshoperator",
		Channel:         "stable",
		Source:          string(CatalogSourceRedHat),
		SourceNamespace: "openshift-marketplace",
		Namespace:       "openshift-operators",
		InstallMode:     string(InstallModeAllNamespaces),
		Enabled:         true,
	},
}

// GetOperatorPreset returns a preset operator configuration by name.
func GetOperatorPreset(name string) (*Operator, bool) {
	op, ok := CommonOperators[strings.ToLower(name)]
	if ok {
		return &op, true
	}
	return nil, false
}

// ListOperatorPresets returns all available operator presets.
func ListOperatorPresets() []string {
	names := make([]string, 0, len(CommonOperators))
	for name := range CommonOperators {
		names = append(names, name)
	}
	return names
}

// Config holds the operator management configuration.
type Config struct {
	// Enabled controls whether operator management is enabled
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Operators is the list of operators to manage
	Operators []Operator `yaml:"operators,omitempty" json:"operators,omitempty"`
	// CreateOperatorGroup controls whether to create OperatorGroups
	CreateOperatorGroup bool `yaml:"create_operator_group" json:"create_operator_group"`
	// DefaultSource is the default CatalogSource
	DefaultSource string `yaml:"default_source,omitempty" json:"default_source,omitempty"`
	// DefaultSourceNamespace is the default CatalogSource namespace
	DefaultSourceNamespace string `yaml:"default_source_namespace,omitempty" json:"default_source_namespace,omitempty"`
}

// NewDefaultConfig creates a default operator configuration.
func NewDefaultConfig() Config {
	return Config{
		Enabled:                false,
		Operators:              []Operator{},
		CreateOperatorGroup:    true,
		DefaultSource:          string(CatalogSourceCommunity),
		DefaultSourceNamespace: "openshift-marketplace",
	}
}

// Validate validates the operator configuration.
func (o *Operator) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("operator name is required")
	}
	if o.Namespace == "" {
		return fmt.Errorf("operator namespace is required for %s", o.Name)
	}
	if o.InstallMode != "" {
		validModes := []string{
			string(InstallModeOwnNamespace),
			string(InstallModeSingleNamespace),
			string(InstallModeMultiNamespace),
			string(InstallModeAllNamespaces),
		}
		valid := false
		for _, m := range validModes {
			if o.InstallMode == m {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid install mode: %s", o.InstallMode)
		}
	}
	return nil
}

// GetSource returns the operator source with default fallback.
func (o *Operator) GetSource(defaultSource string) string {
	if o.Source != "" {
		return o.Source
	}
	return defaultSource
}

// GetSourceNamespace returns the source namespace with default fallback.
func (o *Operator) GetSourceNamespace(defaultSourceNamespace string) string {
	if o.SourceNamespace != "" {
		return o.SourceNamespace
	}
	return defaultSourceNamespace
}

// GetChannel returns the channel with a default fallback.
func (o *Operator) GetChannel() string {
	if o.Channel != "" {
		return o.Channel
	}
	return "stable"
}

// GetInstallPlanApproval returns the install plan approval with default.
func (o *Operator) GetInstallPlanApproval() string {
	if o.InstallPlanApproval != "" {
		return o.InstallPlanApproval
	}
	return "Automatic"
}

// GetInstallMode returns the install mode with default.
func (o *Operator) GetInstallMode() InstallMode {
	if o.InstallMode != "" {
		return InstallMode(o.InstallMode)
	}
	return InstallModeOwnNamespace
}

// SubscriptionManifest represents the data needed for Subscription template.
type SubscriptionManifest struct {
	Name                string
	Namespace           string
	Channel             string
	Source              string
	SourceNamespace     string
	InstallPlanApproval string
	StartingCSV         string
}

// GroupManifest represents the data needed for OperatorGroup template.
type GroupManifest struct {
	Name             string
	Namespace        string
	TargetNamespaces []string
}

// ToSubscriptionManifest converts an Operator to SubscriptionManifest.
func (o *Operator) ToSubscriptionManifest(defaultSource, defaultSourceNamespace string) SubscriptionManifest {
	return SubscriptionManifest{
		Name:                o.Name,
		Namespace:           o.Namespace,
		Channel:             o.GetChannel(),
		Source:              o.GetSource(defaultSource),
		SourceNamespace:     o.GetSourceNamespace(defaultSourceNamespace),
		InstallPlanApproval: o.GetInstallPlanApproval(),
		StartingCSV:         o.Version,
	}
}

// ToGroupManifest converts an Operator to GroupManifest.
func (o *Operator) ToGroupManifest() GroupManifest {
	targetNamespaces := o.TargetNamespaces
	if len(targetNamespaces) == 0 {
		switch o.GetInstallMode() {
		case InstallModeOwnNamespace, InstallModeSingleNamespace:
			targetNamespaces = []string{o.Namespace}
		case InstallModeAllNamespaces:
			targetNamespaces = []string{} // Empty means all namespaces
		}
	}
	return GroupManifest{
		Name:             o.Name + "-og",
		Namespace:        o.Namespace,
		TargetNamespaces: targetNamespaces,
	}
}

// Manager handles operator management operations.
type Manager struct {
	config Config
}

// NewManager creates a new operator manager.
func NewManager(config Config) *Manager {
	return &Manager{config: config}
}

// GetEnabledOperators returns all enabled operators.
func (m *Manager) GetEnabledOperators() []Operator {
	var enabled []Operator
	for i := range m.config.Operators {
		if m.config.Operators[i].Enabled {
			enabled = append(enabled, m.config.Operators[i])
		}
	}
	return enabled
}

// AddOperator adds an operator to the configuration.
func (m *Manager) AddOperator(op *Operator) error {
	if err := op.Validate(); err != nil {
		return err
	}
	m.config.Operators = append(m.config.Operators, *op)
	return nil
}

// AddPreset adds a preset operator by name.
func (m *Manager) AddPreset(name string) error {
	preset, ok := GetOperatorPreset(name)
	if !ok {
		return fmt.Errorf("unknown operator preset: %s", name)
	}
	return m.AddOperator(preset)
}

// RemoveOperator removes an operator by name.
func (m *Manager) RemoveOperator(name string) bool {
	for i := range m.config.Operators {
		if m.config.Operators[i].Name == name {
			m.config.Operators = append(m.config.Operators[:i], m.config.Operators[i+1:]...)
			return true
		}
	}
	return false
}

// GetOperator returns an operator by name.
func (m *Manager) GetOperator(name string) (*Operator, bool) {
	for i := range m.config.Operators {
		if m.config.Operators[i].Name == name {
			return &m.config.Operators[i], true
		}
	}
	return nil, false
}

// ValidateAll validates all operators in the configuration.
func (m *Manager) ValidateAll() []error {
	var errors []error
	for i := range m.config.Operators {
		if err := m.config.Operators[i].Validate(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// GetConfig returns the operator configuration.
func (m *Manager) GetConfig() Config {
	return m.config
}
