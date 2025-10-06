package catalog

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindByName(t *testing.T) {
	svcs := []Service{
		{Name: "payments-api", Team: "payments"},
		{Name: "auth-service", Team: "platform"},
	}
	if s := FindByName(svcs, "auth-service"); s == nil || s.Team != "platform" {
		t.Fatalf("FindByName(auth-service) = %v, want platform", s)
	}
	if s := FindByName(svcs, "missing"); s != nil {
		t.Fatalf("FindByName(missing) = %v, want nil", s)
	}
}

func TestFilterByTeam(t *testing.T) {
	svcs := []Service{
		{Name: "a", Team: "platform"},
		{Name: "b", Team: "payments"},
		{Name: "c", Team: "platform"},
	}
	out := FilterByTeam(svcs, "platform")
	if len(out) != 2 {
		t.Fatalf("FilterByTeam(platform) len = %d, want 2", len(out))
	}
	if out[0].Name != "a" || out[1].Name != "c" {
		t.Fatalf("FilterByTeam(platform) = %v", out)
	}
}

func TestLoadFromRegistry(t *testing.T) {
	dir := t.TempDir()
	manifest := `name: payments-api
version: 2.3.0
owner: john@tanuki-services.example
team: payments
health_url: https://payments-api.tanuki.internal/health
repo_url: https://github.com/sugaroverflow/payments-api
on_call: "#payments-oncall"
last_deploy: "2026-07-18T14:02:00Z"
description: Card processing and payment orchestration
`
	if err := os.WriteFile(filepath.Join(dir, "payments-api.yml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	svcs, err := LoadFromRegistry(dir)
	if err != nil {
		t.Fatalf("LoadFromRegistry: %v", err)
	}
	if len(svcs) != 1 {
		t.Fatalf("LoadFromRegistry len = %d, want 1", len(svcs))
	}
	s := svcs[0]
	if s.Name != "payments-api" || s.Team != "payments" || s.OnCall != "#payments-oncall" {
		t.Errorf("LoadFromRegistry parsed %+v", s)
	}
	if s.HealthURL != "https://payments-api.tanuki.internal/health" {
		t.Errorf("health_url = %q", s.HealthURL)
	}
}

func TestLoadFromRegistryEmptyDir(t *testing.T) {
	if _, err := LoadFromRegistry(t.TempDir()); err == nil {
		t.Fatal("LoadFromRegistry(empty) = nil error, want error")
	}
}

func TestLoadFromURLSendsKeyWhenSet(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Write([]byte(`[{"name":"auth-service","version":"1.6.1","owner":"c@tanuki-services.example","team":"platform","health_url":"h","repo_url":"r"}]`))
	}))
	defer srv.Close()

	t.Setenv("TANUKI_CATALOG_KEY", "test-key-123")
	svcs, err := loadFromURL(srv.URL)
	if err != nil {
		t.Fatalf("loadFromURL: %v", err)
	}
	if gotAuth != "Bearer test-key-123" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key-123")
	}
	if len(svcs) != 1 || svcs[0].Name != "auth-service" {
		t.Errorf("loadFromURL parsed %+v", svcs)
	}
}

func TestLoadFromURLNoKeyNoHeader(t *testing.T) {
	var gotAuth string
	sawHeader := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, sawHeader = r.Header["Authorization"]
		w.Write([]byte(`[{"name":"a","version":"1.0.0","owner":"o","team":"platform","health_url":"h","repo_url":"r"}]`))
	}))
	defer srv.Close()

	t.Setenv("TANUKI_CATALOG_KEY", "")
	if _, err := loadFromURL(srv.URL); err != nil {
		t.Fatalf("loadFromURL: %v", err)
	}
	if sawHeader {
		t.Errorf("Authorization header sent without a key: %q", gotAuth)
	}
}

func TestLoadFromURLErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := loadFromURL(srv.URL)
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("loadFromURL error = %v, want 401 mention", err)
	}
}

func TestLoadRejectsNullCatalog(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "catalog.json"), []byte("null"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if _, err := Load(); err == nil {
		t.Fatal("Load() with null catalog = nil error, want error")
	}
}

func TestLoadReadsLocalCatalog(t *testing.T) {
	dir := t.TempDir()
	data := `[{"name":"search-api","version":"4.1.2","owner":"colleen@tanuki-services.example","team":"experience","health_url":"h","repo_url":"r"}]`
	if err := os.WriteFile(filepath.Join(dir, "catalog.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	svcs, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(svcs) != 1 || svcs[0].Name != "search-api" {
		t.Errorf("Load parsed %+v", svcs)
	}
}

func benchmarkCatalog() []Service {
	svcs := make([]Service, 0, 200)
	teams := []string{"platform", "payments", "experience", "data"}
	for i := 0; i < 200; i++ {
		svcs = append(svcs, Service{
			Name: fmt.Sprintf("service-%03d", i),
			Team: teams[i%len(teams)],
		})
	}
	return svcs
}

func BenchmarkFindByName(b *testing.B) {
	svcs := benchmarkCatalog()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if s := FindByName(svcs, "service-199"); s == nil {
			b.Fatal("service-199 not found")
		}
	}
}

func BenchmarkFilterByTeam(b *testing.B) {
	svcs := benchmarkCatalog()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if out := FilterByTeam(svcs, "payments"); len(out) == 0 {
			b.Fatal("no payments services")
		}
	}
}
