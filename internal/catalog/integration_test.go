//go:build integration

package catalog

import (
	"os"
	"testing"
)

// TestRemoteCatalogIntegration exercises the real remote-fetch path against a
// live catalog endpoint (a WireMock service container in CI). It is excluded
// from normal runs; CI runs it with -tags=integration and TANUKI_CATALOG_URL set.
func TestRemoteCatalogIntegration(t *testing.T) {
	url := os.Getenv("TANUKI_CATALOG_URL")
	if url == "" {
		t.Skip("TANUKI_CATALOG_URL not set; skipping integration test")
	}
	svcs, err := loadFromURL(url)
	if err != nil {
		t.Fatalf("loadFromURL(%s): %v", url, err)
	}
	if len(svcs) == 0 {
		t.Fatal("remote catalog returned zero services")
	}
	seen := make(map[string]bool, len(svcs))
	for _, s := range svcs {
		if s.Name == "" || s.Team == "" {
			t.Errorf("service with missing fields: %+v", s)
		}
		if seen[s.Name] {
			t.Errorf("duplicate service name in remote catalog: %s", s.Name)
		}
		seen[s.Name] = true
	}
}
