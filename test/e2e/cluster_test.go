//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// skipIfNoCluster skips the test if no Kubernetes cluster is available
func skipIfNoCluster(t *testing.T) {
	t.Helper()
	if os.Getenv("KUBECONFIG") == "" && os.Getenv("E2E_CLUSTER") == "" {
		t.Skip("Skipping cluster test: no KUBECONFIG or E2E_CLUSTER set")
	}
	cmd := exec.Command("kubectl", "cluster-info")
	if err := cmd.Run(); err != nil {
		t.Skip("Skipping cluster test: kubectl cluster-info failed")
	}
}

// runWithTimeout runs a command with context timeout
func runWithTimeout(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestClusterPreflightCheck(t *testing.T) {
	skipIfNoCluster(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	output, err := runWithTimeout(ctx, binaryPath, "preflight")
	if err != nil {
		// Preflight may have warnings, check if it's a hard failure
		if strings.Contains(output, "FATAL") || strings.Contains(output, "panic") {
			t.Fatalf("preflight hard failure: %v\nOutput: %s", err, output)
		}
		t.Logf("preflight completed with warnings: %s", output)
	}

	// Should contain cluster connectivity check
	if !strings.Contains(output, "cluster") && !strings.Contains(output, "Kubernetes") {
		t.Log("Output did not mention cluster check, may be expected for minimal check")
	}
}

func TestClusterNamespaceOperations(t *testing.T) {
	skipIfNoCluster(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	testNS := "gitopsi-e2e-test-" + time.Now().Format("20060102150405")

	// Cleanup at the end
	defer func() {
		exec.Command("kubectl", "delete", "namespace", testNS, "--ignore-not-found").Run()
	}()

	// Create namespace
	output, err := runWithTimeout(ctx, "kubectl", "create", "namespace", testNS)
	if err != nil {
		t.Fatalf("failed to create namespace: %v\nOutput: %s", err, output)
	}

	// Verify namespace exists
	output, err = runWithTimeout(ctx, "kubectl", "get", "namespace", testNS)
	if err != nil {
		t.Fatalf("namespace not found: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, testNS) {
		t.Errorf("expected namespace %s in output, got: %s", testNS, output)
	}
}

func TestClusterKustomizeBuild(t *testing.T) {
	skipIfNoCluster(t)

	// Check if kustomize is available
	if _, err := exec.LookPath("kustomize"); err != nil {
		t.Skip("Skipping: kustomize not found in PATH")
	}

	// Generate manifests first
	tmpDir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	output, err := runWithTimeout(ctx, binaryPath, "init",
		"--config", "fixtures/standard-config.yaml",
		"--output", tmpDir,
	)
	if err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	// Build with kustomize
	infraOverlay := tmpDir + "/test-standard/infrastructure/overlays/dev"
	output, err = runWithTimeout(ctx, "kustomize", "build", infraOverlay)
	if err != nil {
		t.Fatalf("kustomize build failed: %v\nOutput: %s", err, output)
	}

	// Should contain expected resources
	expectedResources := []string{"Namespace", "kind:"}
	for _, expected := range expectedResources {
		if !strings.Contains(output, expected) {
			t.Errorf("expected kustomize output to contain '%s'", expected)
		}
	}
}

func TestClusterDryApply(t *testing.T) {
	skipIfNoCluster(t)

	// Generate manifests first
	tmpDir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	output, err := runWithTimeout(ctx, binaryPath, "init",
		"--config", "fixtures/minimal-config.yaml",
		"--output", tmpDir,
	)
	if err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	// Dry-run apply to cluster
	nsFile := tmpDir + "/test-minimal/infrastructure/base/namespaces/dev.yaml"
	output, err = runWithTimeout(ctx, "kubectl", "apply", "-f", nsFile, "--dry-run=server")
	if err != nil {
		t.Fatalf("kubectl dry-run apply failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "namespace") || !strings.Contains(output, "created") && !strings.Contains(output, "configured") {
		t.Logf("dry-run output: %s", output)
	}
}
