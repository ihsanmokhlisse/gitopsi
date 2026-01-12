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

func TestGetKubeEnv_NoKubeconfig(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method: AuthToken,
		Token:  "test-token",
	}

	env := c.getKubeEnv()

	// Should still return env but without custom KUBECONFIG
	if env == nil {
		t.Error("getKubeEnv() should return environment variables")
	}
}

func TestAuthenticate_OIDC_NoKubeconfig(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Authenticate(&AuthOptions{
		Method: AuthOIDC,
		// No kubeconfig specified
	})

	if err == nil {
		t.Error("Authenticate() with OIDC should fail without kubeconfig")
	}
}

func TestAuthenticate_OIDC_WithKubeconfig(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "oidc-config")
	err := os.WriteFile(kubeconfigPath, []byte("apiVersion: v1\nkind: Config"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err = c.Authenticate(&AuthOptions{
		Method:     AuthOIDC,
		Kubeconfig: kubeconfigPath,
	})

	if err != nil {
		t.Errorf("Authenticate() with OIDC should succeed with kubeconfig: %v", err)
	}
	if c.GetAuthMethod() != AuthOIDC {
		t.Errorf("GetAuthMethod() = %v, want %v", c.GetAuthMethod(), AuthOIDC)
	}
}

func TestAuthenticate_ServiceAccount(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Authenticate(&AuthOptions{
		Method: AuthServiceAccount,
	})

	// Should fail because we're not running in a cluster
	if err == nil {
		t.Error("Authenticate() with ServiceAccount should fail outside of cluster")
	}
}

func TestGetAuthMethod_NotAuthenticated(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	method := c.GetAuthMethod()
	if method != "" {
		t.Errorf("GetAuthMethod() = %v, want empty string", method)
	}
}

func TestBuildKubectlArgs_TokenWithCACert(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method: AuthToken,
		Token:  "test-token",
		CACert: "/path/to/ca.crt",
	}

	args := c.buildKubectlArgs("get", "pods")

	// Should include --certificate-authority
	found := false
	for i, arg := range args {
		if arg == "--certificate-authority" && i+1 < len(args) && args[i+1] == "/path/to/ca.crt" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("buildKubectlArgs() should include --certificate-authority, got: %v", args)
	}
}

func TestBuildKubectlArgs_TokenWithSkipTLS(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method:  AuthToken,
		Token:   "test-token",
		SkipTLS: true,
	}

	args := c.buildKubectlArgs("get", "pods")

	// Should include --insecure-skip-tls-verify
	found := false
	for _, arg := range args {
		if arg == "--insecure-skip-tls-verify" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("buildKubectlArgs() should include --insecure-skip-tls-verify, got: %v", args)
	}
}

func TestBuildKubectlArgs_NoAuth(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	// No auth configured

	args := c.buildKubectlArgs("get", "pods")

	// Should just have the command args
	if len(args) != 2 || args[0] != "get" || args[1] != "pods" {
		t.Errorf("buildKubectlArgs() without auth should only have command args, got: %v", args)
	}
}

func TestBuildKubectlArgs_ServerURLWithToken(t *testing.T) {
	c := New("https://my-cluster.example.com:6443", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method: AuthToken,
		Token:  "test-token",
	}

	args := c.buildKubectlArgs("get", "pods")

	// Should include --server
	foundServer := false
	for i, arg := range args {
		if arg == "--server" && i+1 < len(args) && args[i+1] == "https://my-cluster.example.com:6443" {
			foundServer = true
			break
		}
	}

	if !foundServer {
		t.Errorf("buildKubectlArgs() with token should include --server, got: %v", args)
	}
}

func TestClusterInfo(t *testing.T) {
	info := ClusterInfo{
		URL:      "https://kubernetes.default.svc",
		Name:     "prod-cluster",
		Context:  "prod-context",
		Platform: PlatformKubernetes,
	}

	if info.URL != "https://kubernetes.default.svc" {
		t.Errorf("URL = %s, want 'https://kubernetes.default.svc'", info.URL)
	}
	if info.Name != "prod-cluster" {
		t.Errorf("Name = %s, want 'prod-cluster'", info.Name)
	}
	if info.Context != "prod-context" {
		t.Errorf("Context = %s, want 'prod-context'", info.Context)
	}
	if info.Platform != PlatformKubernetes {
		t.Errorf("Platform = %v, want %v", info.Platform, PlatformKubernetes)
	}
}

func TestAuthOptions(t *testing.T) {
	opts := AuthOptions{
		Method:     AuthToken,
		Token:      "my-token",
		TokenEnv:   "TOKEN_ENV",
		Kubeconfig: "/path/to/config",
		Context:    "my-context",
		CACert:     "/path/to/ca.crt",
		SkipTLS:    true,
	}

	if opts.Method != AuthToken {
		t.Errorf("Method = %v, want %v", opts.Method, AuthToken)
	}
	if opts.Token != "my-token" {
		t.Errorf("Token = %s, want 'my-token'", opts.Token)
	}
	if opts.TokenEnv != "TOKEN_ENV" {
		t.Errorf("TokenEnv = %s, want 'TOKEN_ENV'", opts.TokenEnv)
	}
	if opts.Kubeconfig != "/path/to/config" {
		t.Errorf("Kubeconfig = %s, want '/path/to/config'", opts.Kubeconfig)
	}
	if opts.Context != "my-context" {
		t.Errorf("Context = %s, want 'my-context'", opts.Context)
	}
	if opts.CACert != "/path/to/ca.crt" {
		t.Errorf("CACert = %s, want '/path/to/ca.crt'", opts.CACert)
	}
	if !opts.SkipTLS {
		t.Error("SkipTLS should be true")
	}
}

func TestAuthenticate_Token_EnvMissing(t *testing.T) {
	// Ensure the env var doesn't exist
	os.Unsetenv("MISSING_TOKEN_ENV")

	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Authenticate(&AuthOptions{
		Method:   AuthToken,
		TokenEnv: "MISSING_TOKEN_ENV",
	})

	if err == nil {
		t.Error("Authenticate() should fail when token env var is missing")
	}
}

func TestAuthenticate_Kubeconfig_Default(t *testing.T) {
	// This test checks that default kubeconfig path is used
	// It may fail if there's no home directory, but that's expected
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Authenticate(&AuthOptions{
		Method: AuthKubeconfig,
		// No kubeconfig specified - should use default
	})

	// Will likely fail because default kubeconfig may not exist
	// But the error should be about file not found, not about missing path
	if err != nil && !filepath.IsAbs(c.auth.Kubeconfig) {
		t.Error("Default kubeconfig path should be absolute")
	}
}

func TestGetKubeEnv_EmptyKubeconfig(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method:     AuthKubeconfig,
		Kubeconfig: "",
	}

	env := c.getKubeEnv()

	// Should not include KUBECONFIG= with empty value
	for _, e := range env {
		if e == "KUBECONFIG=" {
			t.Error("getKubeEnv() should not include empty KUBECONFIG")
		}
	}
}

func TestBuildKubectlArgs_EmptyKubeconfig(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method:     AuthKubeconfig,
		Kubeconfig: "", // Empty - should not add --kubeconfig
	}

	args := c.buildKubectlArgs("get", "pods")

	// Should not include --kubeconfig with empty value
	for _, arg := range args {
		if arg == "--kubeconfig" {
			t.Errorf("buildKubectlArgs() should not include --kubeconfig with empty value, got: %v", args)
		}
	}
}

func TestBuildKubectlArgs_EmptyContext(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method:     AuthKubeconfig,
		Kubeconfig: "/path/to/kubeconfig",
		Context:    "", // Empty - should not add --context
	}

	args := c.buildKubectlArgs("get", "pods")

	// Should not include --context with empty value
	for _, arg := range args {
		if arg == "--context" {
			t.Errorf("buildKubectlArgs() should not include --context with empty value, got: %v", args)
		}
	}
}

func TestNew_EmptyValues(t *testing.T) {
	c := New("", "", "")

	if c.GetURL() != "" {
		t.Errorf("GetURL() = %s, want empty", c.GetURL())
	}
	if c.GetName() != "" {
		t.Errorf("GetName() = %s, want empty", c.GetName())
	}
	if c.GetPlatform() != "" {
		t.Errorf("GetPlatform() = %v, want empty", c.GetPlatform())
	}
}

func TestNewCluster_AllPlatforms(t *testing.T) {
	platforms := []Platform{
		PlatformKubernetes,
		PlatformOpenShift,
		PlatformAKS,
		PlatformEKS,
		PlatformGKE,
	}

	for _, p := range platforms {
		t.Run(string(p), func(t *testing.T) {
			c := New("https://api.example.com", "test-cluster", p)
			if c.GetPlatform() != p {
				t.Errorf("GetPlatform() = %v, want %v", c.GetPlatform(), p)
			}
		})
	}
}

func TestAuthenticate_TokenWithEnvAndDirect(t *testing.T) {
	t.Setenv("TEST_BOTH_TOKEN", "env-token")

	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	// Direct token should take precedence
	err := c.Authenticate(&AuthOptions{
		Method:   AuthToken,
		Token:    "direct-token",
		TokenEnv: "TEST_BOTH_TOKEN",
	})

	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
	}
	if c.auth.Token != "direct-token" {
		t.Errorf("Token = %s, want direct-token (direct should take precedence)", c.auth.Token)
	}
}

func TestBuildKubectlArgs_AllTokenOptions(t *testing.T) {
	c := New("https://api.example.com:6443", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method:  AuthToken,
		Token:   "test-token",
		CACert:  "/path/to/ca.crt",
		SkipTLS: true,
	}

	args := c.buildKubectlArgs("get", "pods")

	// Should have --server, url, --token, token, --certificate-authority, path, --insecure-skip-tls-verify, get, pods
	if len(args) != 9 {
		t.Errorf("Expected 9 args, got %d: %v", len(args), args)
	}

	foundCACert := false
	foundSkipTLS := false
	for _, arg := range args {
		if arg == "--certificate-authority" {
			foundCACert = true
		}
		if arg == "--insecure-skip-tls-verify" {
			foundSkipTLS = true
		}
	}

	if !foundCACert {
		t.Error("Should include --certificate-authority")
	}
	if !foundSkipTLS {
		t.Error("Should include --insecure-skip-tls-verify")
	}
}

func TestAuthenticate_Kubeconfig_WithContext(t *testing.T) {
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
		Context:    "my-context",
	})

	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
	}
	if c.auth.Context != "my-context" {
		t.Errorf("Context = %s, want my-context", c.auth.Context)
	}
}

func TestGetKubeEnv_NilAuth(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	// No auth configured

	env := c.getKubeEnv()

	// Should return current environment
	if env == nil {
		t.Error("getKubeEnv() should return environment even with nil auth")
	}
}

func TestBuildKubectlArgs_ServiceAccountAuth(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method: AuthServiceAccount,
	}

	args := c.buildKubectlArgs("get", "pods")

	// ServiceAccount uses in-cluster config, no special args needed
	if len(args) != 2 || args[0] != "get" || args[1] != "pods" {
		t.Errorf("ServiceAccount auth should only have command args, got: %v", args)
	}
}

func TestBuildKubectlArgs_OIDCAuth(t *testing.T) {
	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method:     AuthOIDC,
		Kubeconfig: "/path/to/oidc-config",
		Context:    "oidc-context",
	}

	args := c.buildKubectlArgs("get", "pods")

	// OIDC uses kubeconfig like regular kubeconfig auth
	foundKubeconfig := false
	foundContext := false
	for i, arg := range args {
		if arg == "--kubeconfig" && i+1 < len(args) {
			foundKubeconfig = true
		}
		if arg == "--context" && i+1 < len(args) {
			foundContext = true
		}
	}

	if !foundKubeconfig {
		t.Error("OIDC auth should include --kubeconfig")
	}
	if !foundContext {
		t.Error("OIDC auth should include --context")
	}
}

func TestClusterInfo_AllFields(t *testing.T) {
	info := ClusterInfo{
		URL:      "https://api.ocp.example.com:6443",
		Name:     "production",
		Context:  "admin@production",
		Platform: PlatformOpenShift,
	}

	if info.URL != "https://api.ocp.example.com:6443" {
		t.Errorf("URL = %s, want https://api.ocp.example.com:6443", info.URL)
	}
	if info.Name != "production" {
		t.Errorf("Name = %s, want production", info.Name)
	}
	if info.Context != "admin@production" {
		t.Errorf("Context = %s, want admin@production", info.Context)
	}
	if info.Platform != PlatformOpenShift {
		t.Errorf("Platform = %v, want %v", info.Platform, PlatformOpenShift)
	}
}

func TestClusterInfo_EmptyFields(t *testing.T) {
	info := ClusterInfo{}

	if info.URL != "" {
		t.Errorf("URL = %s, want empty", info.URL)
	}
	if info.Name != "" {
		t.Errorf("Name = %s, want empty", info.Name)
	}
	if info.Context != "" {
		t.Errorf("Context = %s, want empty", info.Context)
	}
	if info.Platform != "" {
		t.Errorf("Platform = %v, want empty", info.Platform)
	}
}

func TestAuthOptions_AllMethods(t *testing.T) {
	methods := []AuthMethod{
		AuthKubeconfig,
		AuthToken,
		AuthOIDC,
		AuthServiceAccount,
	}

	for _, m := range methods {
		t.Run(string(m), func(t *testing.T) {
			opts := AuthOptions{Method: m}
			if opts.Method != m {
				t.Errorf("Method = %v, want %v", opts.Method, m)
			}
		})
	}
}

func TestBuildKubectlArgs_EmptyServerURL(t *testing.T) {
	c := New("", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method: AuthToken,
		Token:  "test-token",
	}

	args := c.buildKubectlArgs("get", "pods")

	// Should not include --server for empty URL
	for _, arg := range args {
		if arg == "--server" {
			t.Error("Should not include --server for empty URL")
		}
	}
}

func TestBuildKubectlArgs_KubeconfigWithEmptyContext(t *testing.T) {
	c := New("", "test", PlatformKubernetes)
	c.auth = &AuthOptions{
		Method:     AuthKubeconfig,
		Kubeconfig: "/path/to/config",
		Context:    "", // Empty context
	}

	args := c.buildKubectlArgs("get", "pods")

	// Should not include --context for empty context
	for _, arg := range args {
		if arg == "--context" {
			t.Error("Should not include --context for empty context")
		}
	}
}

func TestCluster_MethodsChain(t *testing.T) {
	c := New("https://api.example.com", "test", PlatformEKS)

	// Test getter chain
	if c.GetURL() != "https://api.example.com" {
		t.Errorf("GetURL() = %s", c.GetURL())
	}
	if c.GetName() != "test" {
		t.Errorf("GetName() = %s", c.GetName())
	}
	if c.GetPlatform() != PlatformEKS {
		t.Errorf("GetPlatform() = %v", c.GetPlatform())
	}
	if c.IsAuthenticated() {
		t.Error("Should not be authenticated initially")
	}
	if c.GetAuthMethod() != "" {
		t.Errorf("GetAuthMethod() = %v, want empty", c.GetAuthMethod())
	}
}

func TestAuthenticate_Token_EmptyEnvVar(t *testing.T) {
	// Set empty value for env var
	t.Setenv("EMPTY_TOKEN", "")

	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	err := c.Authenticate(&AuthOptions{
		Method:   AuthToken,
		TokenEnv: "EMPTY_TOKEN",
	})

	if err == nil {
		t.Error("Authenticate() should fail when token env var is empty")
	}
}

func TestAuthenticate_MultipleCallsOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")
	err := os.WriteFile(kubeconfigPath, []byte("apiVersion: v1\nkind: Config"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	c := New("https://kubernetes.default.svc", "test", PlatformKubernetes)

	// First auth with token
	err = c.Authenticate(&AuthOptions{
		Method: AuthToken,
		Token:  "first-token",
	})
	if err != nil {
		t.Fatalf("First Authenticate() error = %v", err)
	}
	if c.GetAuthMethod() != AuthToken {
		t.Errorf("GetAuthMethod() = %v, want AuthToken", c.GetAuthMethod())
	}

	// Second auth with kubeconfig should overwrite
	err = c.Authenticate(&AuthOptions{
		Method:     AuthKubeconfig,
		Kubeconfig: kubeconfigPath,
	})
	if err != nil {
		t.Fatalf("Second Authenticate() error = %v", err)
	}
	if c.GetAuthMethod() != AuthKubeconfig {
		t.Errorf("GetAuthMethod() = %v, want AuthKubeconfig", c.GetAuthMethod())
	}
}

func TestPlatform_StringConversions(t *testing.T) {
	tests := []struct {
		platform Platform
		want     string
	}{
		{PlatformKubernetes, "kubernetes"},
		{PlatformOpenShift, "openshift"},
		{PlatformAKS, "aks"},
		{PlatformEKS, "eks"},
		{PlatformGKE, "gke"},
		{Platform("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.platform) != tt.want {
				t.Errorf("string(%v) = %s, want %s", tt.platform, string(tt.platform), tt.want)
			}
		})
	}
}

func TestAuthMethod_StringConversions(t *testing.T) {
	tests := []struct {
		method AuthMethod
		want   string
	}{
		{AuthKubeconfig, "kubeconfig"},
		{AuthToken, "token"},
		{AuthOIDC, "oidc"},
		{AuthServiceAccount, "service-account"},
		{AuthMethod("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.method) != tt.want {
				t.Errorf("string(%v) = %s, want %s", tt.method, string(tt.method), tt.want)
			}
		})
	}
}
