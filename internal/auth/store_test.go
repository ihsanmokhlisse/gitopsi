package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMemoryStore_Save(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	cred := &Credential{
		Name:      "test-cred",
		Type:      CredentialTypeGit,
		Provider:  "github",
		Method:    MethodToken,
		CreatedAt: time.Now(),
	}

	err := store.Save(ctx, cred)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify saved
	exists, err := store.Exists(ctx, "test-cred")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Credential should exist after save")
	}
}

func TestMemoryStore_Get(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Save credential
	cred := &Credential{
		Name:     "test-cred",
		Type:     CredentialTypeGit,
		Provider: "github",
		Method:   MethodToken,
		Data: CredentialData{
			Token: "test-token",
		},
	}
	_ = store.Save(ctx, cred)

	// Get existing
	retrieved, err := store.Get(ctx, "test-cred")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != "test-cred" {
		t.Errorf("Name: got %s, want test-cred", retrieved.Name)
	}
	if retrieved.Data.Token != "test-token" {
		t.Errorf("Token: got %s, want test-token", retrieved.Data.Token)
	}

	// Get non-existing
	_, err = store.Get(ctx, "non-existing")
	if err == nil {
		t.Error("Get should fail for non-existing credential")
	}
}

func TestMemoryStore_List(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Add multiple credentials
	_ = store.Save(ctx, &Credential{
		Name:     "git-1",
		Type:     CredentialTypeGit,
		Provider: "github",
	})
	_ = store.Save(ctx, &Credential{
		Name:     "git-2",
		Type:     CredentialTypeGit,
		Provider: "gitlab",
	})
	_ = store.Save(ctx, &Credential{
		Name:     "platform-1",
		Type:     CredentialTypePlatform,
		Provider: "openshift",
	})

	// List all
	all, err := store.List(ctx, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("List all: got %d, want 3", len(all))
	}

	// List by type
	gitCreds, err := store.List(ctx, CredentialTypeGit)
	if err != nil {
		t.Fatalf("List git failed: %v", err)
	}
	if len(gitCreds) != 2 {
		t.Errorf("List git: got %d, want 2", len(gitCreds))
	}

	platformCreds, err := store.List(ctx, CredentialTypePlatform)
	if err != nil {
		t.Fatalf("List platform failed: %v", err)
	}
	if len(platformCreds) != 1 {
		t.Errorf("List platform: got %d, want 1", len(platformCreds))
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Save and delete
	_ = store.Save(ctx, &Credential{Name: "to-delete"})

	err := store.Delete(ctx, "to-delete")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	exists, _ := store.Exists(ctx, "to-delete")
	if exists {
		t.Error("Credential should not exist after delete")
	}

	// Delete non-existing
	err = store.Delete(ctx, "non-existing")
	if err == nil {
		t.Error("Delete should fail for non-existing credential")
	}
}

func TestMemoryStore_Exists(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Check non-existing
	exists, err := store.Exists(ctx, "test")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Should not exist initially")
	}

	// Save and check
	_ = store.Save(ctx, &Credential{Name: "test"})
	exists, err = store.Exists(ctx, "test")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Should exist after save")
	}
}

func TestFileStore_NewFileStore(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "credentials.yaml")

	store, err := NewFileStore(storePath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	if store == nil {
		t.Error("Store should not be nil")
	}
}

func TestFileStore_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "credentials.yaml")
	ctx := context.Background()

	// Create store and save credentials
	store1, err := NewFileStore(storePath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}

	cred := &Credential{
		Name:     "test-cred",
		Type:     CredentialTypeGit,
		Provider: "github",
		Method:   MethodToken,
		Data: CredentialData{
			Token:    "secret-token",
			Username: "testuser",
		},
		Metadata: CredentialMetadata{
			URL:       "https://github.com/test/repo",
			Namespace: "default",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = store1.Save(ctx, cred)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new store instance to test loading
	store2, err := NewFileStore(storePath)
	if err != nil {
		t.Fatalf("NewFileStore (reload) failed: %v", err)
	}

	loaded, err := store2.Get(ctx, "test-cred")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if loaded.Name != cred.Name {
		t.Errorf("Name: got %s, want %s", loaded.Name, cred.Name)
	}
	if loaded.Data.Token != cred.Data.Token {
		t.Errorf("Token: got %s, want %s", loaded.Data.Token, cred.Data.Token)
	}
	if loaded.Metadata.URL != cred.Metadata.URL {
		t.Errorf("URL: got %s, want %s", loaded.Metadata.URL, cred.Metadata.URL)
	}
}

func TestFileStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "credentials.yaml")
	ctx := context.Background()

	store, _ := NewFileStore(storePath)

	// Save
	_ = store.Save(ctx, &Credential{Name: "to-delete"})

	// Delete
	err := store.Delete(ctx, "to-delete")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify persistence
	store2, _ := NewFileStore(storePath)
	exists, _ := store2.Exists(ctx, "to-delete")
	if exists {
		t.Error("Deleted credential should not persist")
	}
}

func TestFileStore_Permissions(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "credentials.yaml")
	ctx := context.Background()

	store, _ := NewFileStore(storePath)
	_ = store.Save(ctx, &Credential{Name: "test"})

	// Check file permissions (should be 0600)
	info, err := os.Stat(storePath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("File permissions: got %o, want 0600", perm)
	}
}

func TestFileStore_List(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "credentials.yaml")
	ctx := context.Background()

	store, _ := NewFileStore(storePath)

	// Add credentials
	_ = store.Save(ctx, &Credential{Name: "git-1", Type: CredentialTypeGit})
	_ = store.Save(ctx, &Credential{Name: "git-2", Type: CredentialTypeGit})
	_ = store.Save(ctx, &Credential{Name: "platform-1", Type: CredentialTypePlatform})

	// List all
	all, _ := store.List(ctx, "")
	if len(all) != 3 {
		t.Errorf("List all: got %d, want 3", len(all))
	}

	// List by type
	gitCreds, _ := store.List(ctx, CredentialTypeGit)
	if len(gitCreds) != 2 {
		t.Errorf("List git: got %d, want 2", len(gitCreds))
	}
}

func TestGetDefaultStorePath(t *testing.T) {
	path := GetDefaultStorePath()
	if path == "" {
		t.Error("GetDefaultStorePath should not return empty string")
	}
	if !filepath.IsAbs(path) && path != ".gitopsi/credentials.yaml" {
		t.Errorf("GetDefaultStorePath should return absolute path or fallback, got: %s", path)
	}
}

func TestFileStore_InvalidPath(t *testing.T) {
	// Test with path in non-writable location (Unix only)
	if os.Getuid() == 0 {
		t.Skip("Skipping as root user")
	}

	ctx := context.Background()
	store, err := NewFileStore("/root/test/credentials.yaml")
	if err != nil {
		// This is expected if the path is not accessible
		return
	}

	// If we got here, try to save - should fail
	err = store.Save(ctx, &Credential{Name: "test"})
	if err == nil {
		t.Error("Save to non-writable path should fail")
	}
}

func TestFileStore_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "credentials.yaml")
	ctx := context.Background()

	store, _ := NewFileStore(storePath)

	// Concurrent saves
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			cred := &Credential{
				Name:     "cred-" + string(rune('a'+idx)),
				Type:     CredentialTypeGit,
				Provider: "github",
			}
			_ = store.Save(ctx, cred)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all saved
	all, _ := store.List(ctx, "")
	if len(all) != 10 {
		t.Errorf("Concurrent save: got %d credentials, want 10", len(all))
	}
}
