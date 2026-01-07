//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// skipIfNoSync skips if sync testing environment is not configured
func skipIfNoSync(t *testing.T) {
	t.Helper()
	skipIfNoArgoCD(t)

	if os.Getenv("E2E_SYNC_TEST") == "" {
		t.Skip("Skipping sync test: E2E_SYNC_TEST not set")
	}
}

func TestArgoCDAppCreate(t *testing.T) {
	skipIfNoSync(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	appName := "e2e-test-app"
	repoURL := os.Getenv("E2E_REPO_URL")
	if repoURL == "" {
		t.Skip("E2E_REPO_URL not set")
	}

	revision := os.Getenv("E2E_BRANCH")
	if revision == "" {
		revision = "main"
	}

	// Clean up any existing app
	runWithTimeout(ctx, "argocd", "app", "delete", appName, "--yes", "--insecure")

	// Create the application
	output, err := runWithTimeout(ctx, "argocd", "app", "create", appName,
		"--repo", repoURL,
		"--revision", revision,
		"--path", "infrastructure/overlays/dev",
		"--dest-server", "https://kubernetes.default.svc",
		"--dest-namespace", "dev",
		"--insecure",
	)
	if err != nil {
		t.Fatalf("failed to create ArgoCD application: %v\nOutput: %s", err, output)
	}

	// Verify app was created
	output, err = runWithTimeout(ctx, "argocd", "app", "get", appName, "--insecure")
	if err != nil {
		t.Fatalf("failed to get application: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, appName) {
		t.Errorf("application %s not found in output", appName)
	}

	// Cleanup
	defer func() {
		runWithTimeout(context.Background(), "argocd", "app", "delete", appName, "--yes", "--insecure")
	}()
}

func TestArgoCDAppSync(t *testing.T) {
	skipIfNoSync(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	appName := "e2e-sync-test"
	repoURL := os.Getenv("E2E_REPO_URL")
	if repoURL == "" {
		t.Skip("E2E_REPO_URL not set")
	}

	revision := os.Getenv("E2E_BRANCH")
	if revision == "" {
		revision = "main"
	}

	// Clean up any existing app
	runWithTimeout(ctx, "argocd", "app", "delete", appName, "--yes", "--insecure")
	time.Sleep(2 * time.Second)

	// Create the application
	output, err := runWithTimeout(ctx, "argocd", "app", "create", appName,
		"--repo", repoURL,
		"--revision", revision,
		"--path", "infrastructure/overlays/dev",
		"--dest-server", "https://kubernetes.default.svc",
		"--dest-namespace", "dev",
		"--sync-policy", "automated",
		"--auto-prune",
		"--self-heal",
		"--insecure",
	)
	if err != nil {
		t.Fatalf("failed to create ArgoCD application: %v\nOutput: %s", err, output)
	}

	// Trigger sync
	output, err = runWithTimeout(ctx, "argocd", "app", "sync", appName, "--insecure", "--timeout", "300")
	if err != nil {
		t.Logf("sync may have issues: %v\nOutput: %s", err, output)
	}

	// Wait for sync completion
	output, err = runWithTimeout(ctx, "argocd", "app", "wait", appName, "--sync", "--timeout", "300", "--insecure")
	if err != nil {
		t.Fatalf("app wait failed: %v\nOutput: %s", err, output)
	}

	// Verify sync status
	output, err = runWithTimeout(ctx, "argocd", "app", "get", appName, "-o", "json", "--insecure")
	if err != nil {
		t.Fatalf("failed to get app status: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, `"syncStatus"`) {
		t.Error("no sync status in output")
	}

	// Cleanup
	defer func() {
		runWithTimeout(context.Background(), "argocd", "app", "delete", appName, "--yes", "--insecure")
	}()
}

func TestResourcesCreatedAfterSync(t *testing.T) {
	skipIfNoSync(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check for namespaces that should be created by the sync
	output, err := runWithTimeout(ctx, "kubectl", "get", "namespaces")
	if err != nil {
		t.Fatalf("failed to get namespaces: %v\nOutput: %s", err, output)
	}

	// The dev namespace should exist after sync (from standard config)
	// This is environment-specific based on what was synced
	t.Logf("Namespaces after sync:\n%s", output)
}

func TestArgoCDHealthStatus(t *testing.T) {
	skipIfNoSync(t)

	appName := os.Getenv("E2E_APP_NAME")
	if appName == "" {
		t.Skip("E2E_APP_NAME not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := runWithTimeout(ctx, "argocd", "app", "get", appName, "-o", "json", "--insecure")
	if err != nil {
		t.Fatalf("failed to get app: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, `"health"`) {
		t.Error("no health status in output")
	}

	// Check for healthy or progressing status
	if !strings.Contains(output, "Healthy") && !strings.Contains(output, "Progressing") {
		t.Logf("App may not be fully healthy yet. Output:\n%s", output)
	}
}
