package cluster

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCluster(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		clname   string
		platform Platform
	}{
		{
			name:     "kubernetes cluster",
			url:      "https://kubernetes.default.svc",
			clname:   "prod-cluster",
			platform: PlatformKubernetes,
		},
		{
			name:     "openshift cluster",
			url:      "https://api.ocp.example.com:6443",
			clname:   "ocp-cluster",
			platform: PlatformOpenShift,
		},
		{
			name:     "aks cluster",
			url:      "https://aks-cluster.azmk8s.io",
			clname:   "aks-prod",
			platform: PlatformAKS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.url, tt.clname, tt.platform)

			if c.GetURL() != tt.url {
				t.Errorf("GetURL() = %v, want %v", c.GetURL(), tt.url)
			}
			if c.GetName() != tt.clname {
				t.Errorf("GetName() = %v, want %v", c.GetName(), tt.clname)
			}
			if c.GetPlatform() != tt.platform {
				t.Errorf("GetPlatform() = %v, want %v", c.GetPlatform(), tt.platform)
			}
			if c.IsAuthenticated() {
				t.Error("IsAuthenticated() should be false before authentication")
			}
		})
	}
}

func TestAuthMethodConstants(t *testing.T) {
	tests := []struct {
		method AuthMethod
		want   string
	}{
		{AuthKubeconfig, "kubeconfig"},
		{AuthToken, "token"},
		{AuthOIDC, "oidc"},
		{AuthServiceAccount, "service-account"},
	}

	for _, tt := range tests {
		if string(tt.method) != tt.want {
			t.Errorf("AuthMethod %v = %s, want %s", tt.method, string(tt.method), tt.want)
		}
	}
}

func TestPlatformConstants(t *testing.T) {
	tests := []struct {
		platform Platform
		want     string
	}{
		{PlatformKubernetes, "kubernetes"},
		{PlatformOpenShift, "openshift"},
		{PlatformAKS, "aks"},
		{PlatformEKS, "eks"},
		{PlatformGKE, "gke"},
	}

	for _, tt := range tests {
		if string(tt.platform) != tt.want {
			t.Errorf("Platform %v = %s, want %s", tt.platform, string(tt.platform), tt.want)
		}
	}
}

func TestAuthenticate_Token(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Authenticate(&AuthOptions{
		Method: AuthToken,
		Token:  "test-token-12345",
	})

	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
	}
	if !c.IsAuthenticated() {
		t.Error("IsAuthenticated() should be true after authentication")
	}
	if c.GetAuthMethod() != AuthToken {
		t.Errorf("GetAuthMethod() = %v, want %v", c.GetAuthMethod(), AuthToken)
	}
}

func TestAuthenticate_Token_Empty(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Authenticate(&AuthOptions{
		Method: AuthToken,
		Token:  "",
	})

	if err == nil {
		t.Error("Authenticate() should fail with empty token")
	}
}

func TestAuthenticate_Token_FromEnv(t *testing.T) {
	os.Setenv("TEST_CLUSTER_TOKEN", "env-token-12345")
	defer os.Unsetenv("TEST_CLUSTER_TOKEN")

	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Authenticate(&AuthOptions{
		Method:   AuthToken,
		TokenEnv: "TEST_CLUSTER_TOKEN",
	})

	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
	}
	if !c.IsAuthenticated() {
		t.Error("IsAuthenticated() should be true after authentication")
	}
}

func TestAuthenticate_Kubeconfig(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")
	err := os.WriteFile(kubeconfigPath, []byte("apiVersion: v1\nkind: Config"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err = c.Authenticate(&AuthOptions{
		Method:     AuthKubeconfig,
		Kubeconfig: kubeconfigPath,
	})

	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
	}
	if !c.IsAuthenticated() {
		t.Error("IsAuthenticated() should be true after authentication")
	}
}

func TestAuthenticate_Kubeconfig_NotFound(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Authenticate(&AuthOptions{
		Method:     AuthKubeconfig,
		Kubeconfig: "/nonexistent/path/to/kubeconfig",
	})

	if err == nil {
		t.Error("Authenticate() should fail with non-existent kubeconfig")
	}
}

func TestAuthenticate_UnsupportedMethod(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Authenticate(&AuthOptions{
		Method: AuthMethod("invalid"),
	})

	if err == nil {
		t.Error("Authenticate() should fail with unsupported method")
	}
}

func TestBuildKubectlArgs(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		authMethod AuthMethod
		token      string
		kubeconfig string
		context    string
		args       []string
		wantLen    int
	}{
		{
			name:       "token auth with server",
			url:        "https://kubernetes.default.svc",
			authMethod: AuthToken,
			token:      "test-token",
			args:       []string{"get", "pods"},
			wantLen:    6, // --server, url, --token, token, get, pods
		},
		{
			name:       "kubeconfig auth",
			authMethod: AuthKubeconfig,
			kubeconfig: "/path/to/kubeconfig",
			args:       []string{"get", "pods"},
			wantLen:    4, // --kubeconfig, path, get, pods
		},
		{
			name:       "kubeconfig with context",
			authMethod: AuthKubeconfig,
			kubeconfig: "/path/to/kubeconfig",
			context:    "my-context",
			args:       []string{"get", "pods"},
			wantLen:    6, // --kubeconfig, path, --context, context, get, pods
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.url, "test", PlatformKubernetes)

			opts := AuthOptions{
				Method:     tt.authMethod,
				Token:      tt.token,
				Kubeconfig: tt.kubeconfig,
				Context:    tt.context,
			}

			// Need to set auth options even if file doesn't exist for this test
			c.auth = &opts

			result := c.buildKubectlArgs(tt.args...)
			if len(result) != tt.wantLen {
				t.Errorf("buildKubectlArgs() len = %v, want %v, got args: %v", len(result), tt.wantLen, result)
			}
		})
	}
}

func TestTestConnection_NotAuthenticated(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.TestConnection(context.Background())
	if err == nil {
		t.Error("TestConnection() should fail when not authenticated")
	}
}

func TestGetServerVersion_NotAuthenticated(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	_, err := c.GetServerVersion(context.Background())
	if err == nil {
		t.Error("GetServerVersion() should fail when not authenticated")
	}
}

func TestCreateNamespace_NotAuthenticated(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.CreateNamespace(context.Background(), "test-ns")
	if err == nil {
		t.Error("CreateNamespace() should fail when not authenticated")
	}
}

func TestApply_NotAuthenticated(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Apply(context.Background(), "apiVersion: v1\nkind: ConfigMap")
	if err == nil {
		t.Error("Apply() should fail when not authenticated")
	}
}

func TestApplyFile_NotAuthenticated(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.ApplyFile(context.Background(), "/path/to/manifest.yaml")
	if err == nil {
		t.Error("ApplyFile() should fail when not authenticated")
	}
}

func TestWaitForDeployment_NotAuthenticated(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.WaitForDeployment(context.Background(), "default", "nginx", 30)
	if err == nil {
		t.Error("WaitForDeployment() should fail when not authenticated")
	}
}

func TestGetPodLogs_NotAuthenticated(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	_, err := c.GetPodLogs(context.Background(), "default", "nginx-pod", 100)
	if err == nil {
		t.Error("GetPodLogs() should fail when not authenticated")
	}
}

func TestRunCommand_NotAuthenticated(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	_, err := c.RunCommand(context.Background(), "get", "pods")
	if err == nil {
		t.Error("RunCommand() should fail when not authenticated")
	}
}

func TestGetKubeEnv(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Kubeconfig: "/custom/kubeconfig",
	}

	env := c.getKubeEnv()

	found := false
	for _, e := range env {
		if e == "KUBECONFIG=/custom/kubeconfig" {
			found = true
			break
		}
	}

	if !found {
		t.Error("getKubeEnv() should include KUBECONFIG environment variable")
	}
}
