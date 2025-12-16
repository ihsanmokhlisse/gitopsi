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

func TestWriter_WriteFile_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	writer := New(tmpDir, false, true)

	content := []byte("verbose test content")
	relativePath := "verbose/test.txt"

	if err := writer.WriteFile(relativePath, content); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	fullPath := filepath.Join(tmpDir, relativePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Error("File should be created in verbose mode")
	}
}

func TestWriter_WriteFile_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	writer := New(tmpDir, false, false)

	content := []byte("deeply nested content")
	relativePath := "a/b/c/d/e/test.txt"

	if err := writer.WriteFile(relativePath, content); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	fullPath := filepath.Join(tmpDir, relativePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != string(content) {
		t.Error("Content mismatch in nested file")
	}
}

func TestWriter_CreateDir_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	writer := New(tmpDir, false, true)

	relativePath := "verbose/dir/path"

	if err := writer.CreateDir(relativePath); err != nil {
		t.Fatalf("CreateDir() error = %v", err)
	}

	fullPath := filepath.Join(tmpDir, relativePath)
	info, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected directory")
	}
}

func TestWriter_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	writer := New(tmpDir, false, false)

	files := map[string]string{
		"file1.txt":     "content1",
		"dir/file2.txt": "content2",
		"dir/file3.txt": "content3",
	}

	for path, content := range files {
		if err := writer.WriteFile(path, []byte(content)); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", path, err)
		}
	}

	for path, expectedContent := range files {
		fullPath := filepath.Join(tmpDir, path)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", path, err)
		}
		if string(data) != expectedContent {
			t.Errorf("Content mismatch for %s", path)
		}
	}
}

func TestWriter_ExistsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	writer := New(tmpDir, false, false)

	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	if !writer.Exists("testdir") {
		t.Error("Exists() should return true for existing directory")
	}
}

func TestNewWriter(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		dryRun  bool
		verbose bool
	}{
		{"basic", "/tmp", false, false},
		{"dry-run", "/tmp", true, false},
		{"verbose", "/tmp", false, true},
		{"both", "/tmp", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := New(tt.baseDir, tt.dryRun, tt.verbose)
			if w.BaseDir != tt.baseDir {
				t.Errorf("BaseDir = %s, want %s", w.BaseDir, tt.baseDir)
			}
			if w.DryRun != tt.dryRun {
				t.Errorf("DryRun = %v, want %v", w.DryRun, tt.dryRun)
			}
			if w.Verbose != tt.verbose {
				t.Errorf("Verbose = %v, want %v", w.Verbose, tt.verbose)
			}
		})
	}
}
