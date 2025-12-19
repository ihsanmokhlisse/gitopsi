package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// FileStore implements Store using file-based storage.
type FileStore struct {
	path        string
	mu          sync.RWMutex
	credentials map[string]*Credential
}

// NewFileStore creates a new file-based credential store.
func NewFileStore(path string) (*FileStore, error) {
	store := &FileStore{
		path:        path,
		credentials: make(map[string]*Credential),
	}

	// Load existing credentials if file exists
	if _, err := os.Stat(path); err == nil {
		if err := store.load(); err != nil {
			return nil, fmt.Errorf("failed to load credentials: %w", err)
		}
	}

	return store, nil
}

// Save stores a credential.
func (s *FileStore) Save(ctx context.Context, cred *Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.credentials[cred.Name] = cred
	return s.persist()
}

// Get retrieves a credential by name.
func (s *FileStore) Get(ctx context.Context, name string) (*Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cred, ok := s.credentials[name]
	if !ok {
		return nil, fmt.Errorf("credential %q not found", name)
	}
	return cred, nil
}

// List returns all credentials of a given type.
func (s *FileStore) List(ctx context.Context, credType CredentialType) ([]*Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Credential
	for _, cred := range s.credentials {
		if credType == "" || cred.Type == credType {
			result = append(result, cred)
		}
	}
	return result, nil
}

// Delete removes a credential.
func (s *FileStore) Delete(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.credentials[name]; !ok {
		return fmt.Errorf("credential %q not found", name)
	}

	delete(s.credentials, name)
	return s.persist()
}

// Exists checks if a credential exists.
func (s *FileStore) Exists(ctx context.Context, name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.credentials[name]
	return ok, nil
}

func (s *FileStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	var fileData struct {
		Credentials []*Credential `yaml:"credentials"`
	}

	if err := yaml.Unmarshal(data, &fileData); err != nil {
		return err
	}

	for _, cred := range fileData.Credentials {
		s.credentials[cred.Name] = cred
	}

	return nil
}

func (s *FileStore) persist() error {
	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	creds := make([]*Credential, 0, len(s.credentials))
	for _, cred := range s.credentials {
		creds = append(creds, cred)
	}

	fileData := struct {
		Credentials []*Credential `yaml:"credentials"`
	}{
		Credentials: creds,
	}

	data, err := yaml.Marshal(fileData)
	if err != nil {
		return err
	}

	// Write with restricted permissions (owner only)
	return os.WriteFile(s.path, data, 0600)
}

// MemoryStore implements Store using in-memory storage.
type MemoryStore struct {
	mu          sync.RWMutex
	credentials map[string]*Credential
}

// NewMemoryStore creates a new in-memory credential store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		credentials: make(map[string]*Credential),
	}
}

// Save stores a credential.
func (s *MemoryStore) Save(ctx context.Context, cred *Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.credentials[cred.Name] = cred
	return nil
}

// Get retrieves a credential by name.
func (s *MemoryStore) Get(ctx context.Context, name string) (*Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cred, ok := s.credentials[name]
	if !ok {
		return nil, fmt.Errorf("credential %q not found", name)
	}
	return cred, nil
}

// List returns all credentials of a given type.
func (s *MemoryStore) List(ctx context.Context, credType CredentialType) ([]*Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Credential
	for _, cred := range s.credentials {
		if credType == "" || cred.Type == credType {
			result = append(result, cred)
		}
	}
	return result, nil
}

// Delete removes a credential.
func (s *MemoryStore) Delete(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.credentials[name]; !ok {
		return fmt.Errorf("credential %q not found", name)
	}

	delete(s.credentials, name)
	return nil
}

// Exists checks if a credential exists.
func (s *MemoryStore) Exists(ctx context.Context, name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.credentials[name]
	return ok, nil
}

// GetDefaultStorePath returns the default path for storing credentials.
func GetDefaultStorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gitopsi/credentials.yaml"
	}
	return filepath.Join(home, ".gitopsi", "credentials.yaml")
}
