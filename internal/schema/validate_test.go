package schema

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidateMissingScript(t *testing.T) {
	var out, errBuf bytes.Buffer
	err := Validate(t.TempDir(), &out, &errBuf)
	if err == nil || !strings.Contains(err.Error(), "build_catalog.py not found") {
		t.Fatalf("Validate(empty dir) = %v, want script-not-found error", err)
	}
}
