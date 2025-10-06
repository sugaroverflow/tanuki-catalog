package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testCatalog = `[
  {"name": "payments-api", "version": "2.3.0", "owner": "john@tanuki-services.example", "team": "payments", "health_url": "h", "repo_url": "r", "on_call": "#payments-oncall"},
  {"name": "auth-service", "version": "1.6.1", "owner": "cesar@tanuki-services.example", "team": "platform", "health_url": "h", "repo_url": "r"}
]`

func chdirWithCatalog(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "catalog.json"), []byte(testCatalog), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
}

func TestRunNoArgs(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run(nil, &out, &errBuf); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
	if !strings.Contains(errBuf.String(), "Usage:") {
		t.Errorf("stderr missing usage:\n%s", errBuf.String())
	}
}

func TestRunHelp(t *testing.T) {
	for _, arg := range []string{"help", "--help", "-h"} {
		var out, errBuf bytes.Buffer
		if code := run([]string{arg}, &out, &errBuf); code != 0 {
			t.Errorf("run(%s) = %d, want 0", arg, code)
		}
		if !strings.Contains(out.String(), "Usage:") {
			t.Errorf("run(%s) stdout missing usage:\n%s", arg, out.String())
		}
	}
}

func TestRunVersion(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"version"}, &out, &errBuf); code != 0 {
		t.Fatalf("run(version) = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "tanuki dev") {
		t.Errorf("run(version) = %q, want tanuki dev", out.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"frobnicate"}, &out, &errBuf); code != 2 {
		t.Fatalf("run(frobnicate) = %d, want 2", code)
	}
	if !strings.Contains(errBuf.String(), "Unknown command: frobnicate") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRunList(t *testing.T) {
	chdirWithCatalog(t)
	var out, errBuf bytes.Buffer
	if code := run([]string{"list"}, &out, &errBuf); code != 0 {
		t.Fatalf("run(list) = %d, stderr: %s", code, errBuf.String())
	}
	for _, want := range []string{"payments-api", "auth-service"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("list output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunStatusNotFound(t *testing.T) {
	chdirWithCatalog(t)
	var out, errBuf bytes.Buffer
	if code := run([]string{"status", "missing-service"}, &out, &errBuf); code != 1 {
		t.Fatalf("run(status missing) = %d, want 1", code)
	}
	if out.Len() != 0 {
		t.Errorf("not-found message leaked to stdout: %q", out.String())
	}
	if !strings.Contains(errBuf.String(), "not found") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRunSearch(t *testing.T) {
	chdirWithCatalog(t)
	var out, errBuf bytes.Buffer
	if code := run([]string{"search", "--team", "payments"}, &out, &errBuf); code != 0 {
		t.Fatalf("run(search --team payments) = %d, stderr: %s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "payments-api") || strings.Contains(out.String(), "auth-service") {
		t.Errorf("search output wrong:\n%s", out.String())
	}

	// --team=payments form
	out.Reset()
	errBuf.Reset()
	if code := run([]string{"search", "--team=payments"}, &out, &errBuf); code != 0 {
		t.Fatalf("run(search --team=payments) = %d", code)
	}
	if !strings.Contains(out.String(), "payments-api") {
		t.Errorf("search --team= output wrong:\n%s", out.String())
	}
}

func TestRunSearchNoMatchExitsNonzero(t *testing.T) {
	chdirWithCatalog(t)
	var out, errBuf bytes.Buffer
	if code := run([]string{"search", "--team", "nonexistent"}, &out, &errBuf); code != 1 {
		t.Fatalf("run(search --team nonexistent) = %d, want 1", code)
	}
}

func TestRunSearchFlagValueMissing(t *testing.T) {
	chdirWithCatalog(t)
	// A flag-looking token must not be swallowed as the team value.
	var out, errBuf bytes.Buffer
	if code := run([]string{"search", "--team", "--json"}, &out, &errBuf); code != 2 {
		t.Fatalf("run(search --team --json) = %d, want 2", code)
	}
	if !strings.Contains(errBuf.String(), "--team requires a value") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRunSearchUnknownFlag(t *testing.T) {
	chdirWithCatalog(t)
	var out, errBuf bytes.Buffer
	if code := run([]string{"search", "--owner", "x"}, &out, &errBuf); code != 2 {
		t.Fatalf("run(search --owner) = %d, want 2", code)
	}
}

func TestRunListRejectsExtraArgs(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"list", "--json"}, &out, &errBuf); code != 2 {
		t.Fatalf("run(list --json) = %d, want 2", code)
	}
}
