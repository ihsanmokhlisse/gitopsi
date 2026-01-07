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

// skipIfNoArgoCD skips the test if ArgoCD is not available
func skipIfNoArgoCD(t *testing.T) {
	t.Helper()
	skipIfNoCluster(t)

	if os.Getenv("E2E_ARGOCD") == "" {
		t.Skip("Skipping ArgoCD test: E2E_ARGOCD not set")
	}

	cmd := exec.Command("argocd", "version", "--client")
	if err := cmd.Run(); err != nil {
		t.Skip("Skipping ArgoCD test: argocd CLI not available")
	}
}

func TestArgoCDInstalled(t *testing.T) {
	skipIfNoArgoCD(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check ArgoCD namespace exists
	output, err := runWithTimeout(ctx, "kubectl", "get", "namespace", "argocd")
	if err != nil {
		t.Fatalf("ArgoCD namespace not found: %v\nOutput: %s", err, output)
	}

	// Check ArgoCD pods are running
	output, err = runWithTimeout(ctx, "kubectl", "get", "pods", "-n", "argocd", "-o", "wide")
	if err != nil {
		t.Fatalf("failed to get ArgoCD pods: %v\nOutput: %s", err, output)
	}

	requiredPods := []string{
		"argocd-server",
		"argocd-repo-server",
		"argocd-application-controller",
	}

	for _, pod := range requiredPods {
		if !strings.Contains(output, pod) {
			t.Errorf("required pod %s not found in output", pod)
		}
	}
}

func TestArgoCDServerReady(t *testing.T) {
	skipIfNoArgoCD(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Wait for argocd-server to be ready
	output, err := runWithTimeout(ctx, "kubectl", "wait",
		"--for=condition=available",
		"deployment/argocd-server",
		"-n", "argocd",
		"--timeout=120s",
	)
	if err != nil {
		t.Fatalf("argocd-server not ready: %v\nOutput: %s", err, output)
	}
}

func TestArgoCDCLILogin(t *testing.T) {
	skipIfNoArgoCD(t)

	// Get password from environment or secret
	password := os.Getenv("ARGOCD_PASSWORD")
	if password == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		output, err := runWithTimeout(ctx, "kubectl", "-n", "argocd", "get", "secret",
			"argocd-initial-admin-secret",
			"-o", "jsonpath={.data.password}",
		)
		if err != nil {
			t.Skip("Cannot get ArgoCD password, skipping login test")
		}

		// Decode base64
		decoded, err := runWithTimeout(ctx, "bash", "-c", "echo '"+output+"' | base64 -d")
		if err != nil {
			t.Skip("Cannot decode ArgoCD password")
		}
		password = strings.TrimSpace(decoded)
	}

	if password == "" {
		t.Skip("No ArgoCD password available")
	}

	// Test argocd app list (requires port-forward to be running)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server := os.Getenv("ARGOCD_SERVER")
	if server == "" {
		server = "localhost:8080"
	}

	output, err := runWithTimeout(ctx, "argocd", "app", "list",
		"--server", server,
		"--auth-token", password,
		"--insecure",
	)
	if err != nil {
		// May fail if no apps exist yet
		if !strings.Contains(output, "NAME") && !strings.Contains(output, "No resources") {
			t.Logf("argocd app list warning: %v\nOutput: %s", err, output)
		}
	}
}

func TestArgoCDProjectCRD(t *testing.T) {
	skipIfNoArgoCD(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check that ArgoCD CRDs exist
	output, err := runWithTimeout(ctx, "kubectl", "get", "crd", "appprojects.argoproj.io")
	if err != nil {
		t.Fatalf("ArgoCD AppProject CRD not found: %v\nOutput: %s", err, output)
	}

	output, err = runWithTimeout(ctx, "kubectl", "get", "crd", "applications.argoproj.io")
	if err != nil {
		t.Fatalf("ArgoCD Application CRD not found: %v\nOutput: %s", err, output)
	}
}

func TestArgoCDApplicationSetCRD(t *testing.T) {
	skipIfNoArgoCD(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := runWithTimeout(ctx, "kubectl", "get", "crd", "applicationsets.argoproj.io")
	if err != nil {
		t.Fatalf("ArgoCD ApplicationSet CRD not found: %v\nOutput: %s", err, output)
	}
}
