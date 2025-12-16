package output

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriter_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	writer := New(tmpDir, false, false)

	content := []byte("test content")
	relativePath := "subdir/test.txt"

	if err := writer.WriteFile(relativePath, content); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	fullPath := filepath.Join(tmpDir, relativePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("content mismatch: expected %s, got %s", content, data)
	}
}

func TestWriter_WriteFile_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	writer := New(tmpDir, true, false)

	content := []byte("test content")
	relativePath := "subdir/test.txt"

	if err := writer.WriteFile(relativePath, content); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	fullPath := filepath.Join(tmpDir, relativePath)
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Error("file should not exist in dry-run mode")
	}
}

func TestWriter_CreateDir(t *testing.T) {
	tmpDir := t.TempDir()
	writer := New(tmpDir, false, false)

	relativePath := "nested/directory/path"

	if err := writer.CreateDir(relativePath); err != nil {
		t.Fatalf("CreateDir() error = %v", err)
	}

	fullPath := filepath.Join(tmpDir, relativePath)
	info, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected directory, got file")
	}
}

func TestWriter_CreateDir_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	writer := New(tmpDir, true, false)

	relativePath := "nested/directory/path"

	if err := writer.CreateDir(relativePath); err != nil {
		t.Fatalf("CreateDir() error = %v", err)
	}

	fullPath := filepath.Join(tmpDir, relativePath)
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Error("directory should not exist in dry-run mode")
	}
}

func TestWriter_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	writer := New(tmpDir, false, false)

	testFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if !writer.Exists("exists.txt") {
		t.Error("Exists() should return true for existing file")
	}

	if writer.Exists("nonexistent.txt") {
		t.Error("Exists() should return false for non-existent file")
	}
}

