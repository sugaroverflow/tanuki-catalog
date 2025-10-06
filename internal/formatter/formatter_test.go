package formatter

import (
	"bytes"
	"strings"
	"testing"

	"gitlab.com/gitlab-da/projects/tanuki-services/tanuki-catalog/internal/catalog"
)

func TestListEmpty(t *testing.T) {
	var buf bytes.Buffer
	List(&buf, nil)
	if got := buf.String(); got != "No services found.\n" {
		t.Fatalf("List(nil) = %q, want %q", got, "No services found.\n")
	}
}

func TestListWithServices(t *testing.T) {
	svcs := []catalog.Service{
		{Name: "payments-api", Version: "2.3.0", Team: "payments", Owner: "john@tanuki-services.example"},
		{Name: "webhook-dispatcher", Version: "0.5.1", Team: "payments", Owner: "mei@tanuki-services.example"},
	}
	var buf bytes.Buffer
	List(&buf, svcs)
	out := buf.String()

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("List printed %d lines, want 3 (header + 2 rows):\n%s", len(lines), out)
	}
	if !strings.HasPrefix(lines[0], "NAME") || !strings.Contains(lines[0], "VERSION") {
		t.Errorf("header = %q, want NAME/VERSION columns", lines[0])
	}
	for _, want := range []string{"payments-api", "2.3.0", "webhook-dispatcher", "0.5.1"} {
		if !strings.Contains(out, want) {
			t.Errorf("List output missing %q:\n%s", want, out)
		}
	}
	// tabwriter must align columns even when names differ in length.
	versionCol := strings.Index(lines[1], "2.3.0")
	if versionCol == -1 || !strings.HasPrefix(lines[2][versionCol:], "0.5.1") {
		t.Errorf("columns not aligned:\n%s", out)
	}
}

func TestStatus(t *testing.T) {
	s := catalog.Service{
		Name:        "auth-service",
		Version:     "1.6.1",
		Owner:       "cesar@tanuki-services.example",
		Team:        "platform",
		HealthURL:   "https://auth-service.tanuki.internal/health",
		RepoURL:     "https://github.com/sugaroverflow/auth-service",
		LastDeploy:  "2026-07-21T09:14:00Z",
		Description: "Token issuing and session validation",
	}
	var buf bytes.Buffer
	Status(&buf, s)
	out := buf.String()
	for _, want := range []string{
		"Name:        auth-service",
		"Version:     1.6.1",
		"Team:        platform",
		"Last deploy: 2026-07-21T09:14:00Z",
		"Description: Token issuing and session validation",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Status output missing %q:\n%s", want, out)
		}
	}
}

func TestStatusOmitsEmptyDescription(t *testing.T) {
	var buf bytes.Buffer
	Status(&buf, catalog.Service{Name: "x-service"})
	if strings.Contains(buf.String(), "Description:") {
		t.Errorf("Status printed Description for empty field:\n%s", buf.String())
	}
}

func TestOwners(t *testing.T) {
	s := catalog.Service{
		Name:   "payments-api",
		Owner:  "john@tanuki-services.example",
		Owners: []string{"mei@tanuki-services.example"},
		OnCall: "#payments-oncall",
	}
	var buf bytes.Buffer
	Owners(&buf, s)
	out := buf.String()
	for _, want := range []string{"payments-api", "john@tanuki-services.example", "Others:   mei@tanuki-services.example", "On-call:  #payments-oncall"} {
		if !strings.Contains(out, want) {
			t.Errorf("Owners output missing %q:\n%s", want, out)
		}
	}
}

func TestOwnersNoOnCall(t *testing.T) {
	var buf bytes.Buffer
	Owners(&buf, catalog.Service{Name: "x-service", Owner: "a@tanuki-services.example"})
	out := buf.String()
	if !strings.Contains(out, "On-call:  (none listed)") {
		t.Errorf("Owners output missing on-call placeholder:\n%s", out)
	}
	if strings.Count(out, "On-call") != 1 {
		t.Errorf("Owners printed on-call more than once:\n%s", out)
	}
}
