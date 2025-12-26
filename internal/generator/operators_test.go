package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ihsanmokhlisse/gitopsi/internal/config"
	"github.com/ihsanmokhlisse/gitopsi/internal/operator"
	"github.com/ihsanmokhlisse/gitopsi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_GenerateOperators_CreatesSubscription(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "operator-test"},
		Platform:   "openshift",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Operators: operator.Config{
			Enabled:             true,
			CreateOperatorGroup: true,
			Operators: []operator.Operator{
				{
					Name:        "prometheus-operator",
					Namespace:   "openshift-monitoring",
					Channel:     "beta",
					Source:      "community-operators",
					InstallMode: "AllNamespaces",
					Enabled:     true,
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.generateOperators()
	require.NoError(t, err, "generateOperators should not error")

	subscriptionPath := filepath.Join(tmpDir, "operator-test/infrastructure/operators/prometheus-operator/subscription.yaml")
	_, err = os.Stat(subscriptionPath)
	assert.False(t, os.IsNotExist(err), "Subscription YAML should exist at %s", subscriptionPath)
}

func TestGenerator_GenerateOperators_CreatesOperatorGroup(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "operator-test"},
		Platform:   "openshift",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Operators: operator.Config{
			Enabled:             true,
			CreateOperatorGroup: true,
			Operators: []operator.Operator{
				{
					Name:        "grafana-operator",
					Namespace:   "grafana",
					Channel:     "v5",
					Source:      "community-operators",
					InstallMode: "OwnNamespace",
					Enabled:     true,
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.generateOperators()
	require.NoError(t, err, "generateOperators should not error")

	operatorGroupPath := filepath.Join(tmpDir, "operator-test/infrastructure/operators/grafana-operator/operatorgroup.yaml")
	_, err = os.Stat(operatorGroupPath)
	assert.False(t, os.IsNotExist(err), "OperatorGroup YAML should exist at %s", operatorGroupPath)
}

func TestGenerator_GenerateOperators_SkipsWhenDisabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "operator-test"},
		Platform:   "openshift",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Operators: operator.Config{
			Enabled: false,
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.generateOperators()
	require.NoError(t, err, "generateOperators should not error when disabled")

	operatorsDir := filepath.Join(tmpDir, "operator-test/infrastructure/operators")
	_, err = os.Stat(operatorsDir)
	assert.True(t, os.IsNotExist(err), "Operators directory should not exist when disabled")
}

func TestGenerator_GenerateOperators_CreatesKustomization(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Project:    config.Project{Name: "operator-test"},
		Platform:   "openshift",
		Scope:      "infrastructure",
		GitOpsTool: "argocd",
		Git:        config.GitConfig{URL: "https://github.com/test/repo.git"},
		Environments: []config.Environment{
			{Name: "dev"},
		},
		Operators: operator.Config{
			Enabled:             true,
			CreateOperatorGroup: true,
			Operators: []operator.Operator{
				{
					Name:        "cert-manager",
					Namespace:   "cert-manager",
					Channel:     "stable",
					Source:      "community-operators",
					InstallMode: "AllNamespaces",
					Enabled:     true,
				},
			},
		},
	}

	writer := output.New(tmpDir, false, false)
	gen := New(cfg, writer, false)

	err := gen.generateOperators()
	require.NoError(t, err, "generateOperators should not error")

	kustomizationPath := filepath.Join(tmpDir, "operator-test/infrastructure/operators/kustomization.yaml")
	_, err = os.Stat(kustomizationPath)
	assert.False(t, os.IsNotExist(err), "Kustomization YAML should exist at %s", kustomizationPath)
}

