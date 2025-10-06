package schema

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Validate runs the Python build_catalog script in validate-only mode.
// repoRoot is the path to the repo root (where go.mod and scripts/ live).
// Prefers .venv/bin/python3 if present (local dev); CI should have deps in path.
func Validate(repoRoot string, stdout, stderr io.Writer) error {
	script := filepath.Join(repoRoot, "scripts", "build_catalog.py")
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("build_catalog.py not found at %s: %w", script, err)
	}
	python := "python3"
	if venv := filepath.Join(repoRoot, ".venv", "bin", "python3"); pathExists(venv) {
		python = venv
	}
	cmd := exec.Command(python, script, "--validate")
	cmd.Dir = repoRoot
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
