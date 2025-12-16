package output

import (
	"fmt"
	"os"
	"path/filepath"
)

type Writer struct {
	BaseDir string
	DryRun  bool
	Verbose bool
}

func New(baseDir string, dryRun, verbose bool) *Writer {
	return &Writer{
		BaseDir: baseDir,
		DryRun:  dryRun,
		Verbose: verbose,
	}
}

func (w *Writer) WriteFile(relativePath string, content []byte) error {
	fullPath := filepath.Join(w.BaseDir, relativePath)

	if w.Verbose || w.DryRun {
		fmt.Printf("  → %s\n", relativePath)
	}

	if w.DryRun {
		return nil
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	return nil
}

func (w *Writer) CreateDir(relativePath string) error {
	fullPath := filepath.Join(w.BaseDir, relativePath)

	if w.Verbose || w.DryRun {
		fmt.Printf("  📁 %s/\n", relativePath)
	}

	if w.DryRun {
		return nil
	}

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
	}

	return nil
}

func (w *Writer) Exists(relativePath string) bool {
	fullPath := filepath.Join(w.BaseDir, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

