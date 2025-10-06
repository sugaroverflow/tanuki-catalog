package catalog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// Service represents a single entry in the catalog.
type Service struct {
	Name        string   `json:"name" yaml:"name"`
	Version     string   `json:"version" yaml:"version"`
	Owner       string   `json:"owner" yaml:"owner"`
	Owners      []string `json:"owners,omitempty" yaml:"owners,omitempty"`
	Team        string   `json:"team" yaml:"team"`
	HealthURL   string   `json:"health_url" yaml:"health_url"`
	RepoURL     string   `json:"repo_url" yaml:"repo_url"`
	OnCall      string   `json:"on_call,omitempty" yaml:"on_call,omitempty"`
	LastDeploy  string   `json:"last_deploy,omitempty" yaml:"last_deploy,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// httpClient is used for remote catalog fetches. The timeout guards against a
// black-holed TANUKI_CATALOG_URL hanging the CLI (and CI smoke tests) forever.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// Load reads the catalog from TANUKI_CATALOG_URL (if set), from a local
// catalog.json or dist/catalog.json, or — so a fresh clone works without the
// Python toolchain — builds it directly from the registry manifests.
func Load() ([]Service, error) {
	if source := os.Getenv("TANUKI_CATALOG_URL"); source != "" {
		return loadFromURL(source)
	}
	for _, path := range []string{"catalog.json", "dist/catalog.json"} {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		return parseCatalog(data, path)
	}
	if root, err := RepoRoot(); err == nil {
		if svcs, err := LoadFromRegistry(filepath.Join(root, "registry")); err == nil {
			return svcs, nil
		}
	}
	return nil, fmt.Errorf("no catalog found: set TANUKI_CATALOG_URL, or run from a directory containing catalog.json or registry/")
}

func loadFromURL(url string) ([]Service, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if key := os.Getenv("TANUKI_CATALOG_KEY"); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch catalog: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("catalog URL returned %d: %s", resp.StatusCode, string(body))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read catalog: %w", err)
	}
	return parseCatalog(data, url)
}

func parseCatalog(data []byte, source string) ([]Service, error) {
	var catalog []Service
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("parse %s: %w", source, err)
	}
	if catalog == nil {
		return nil, fmt.Errorf("%s contains no services", source)
	}
	return catalog, nil
}

// LoadFromRegistry builds the catalog directly from the YAML manifests in dir.
func LoadFromRegistry(dir string) ([]Service, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, err
	}
	more, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	files = append(files, more...)
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no manifests found in %s", dir)
	}
	var catalog []Service
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var s Service
		if err := yaml.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		catalog = append(catalog, s)
	}
	return catalog, nil
}

// FindByName returns the service with the given name, or nil.
func FindByName(catalog []Service, name string) *Service {
	for i := range catalog {
		if catalog[i].Name == name {
			return &catalog[i]
		}
	}
	return nil
}

// FilterByTeam returns services whose team matches (case-sensitive).
func FilterByTeam(catalog []Service, team string) []Service {
	var out []Service
	for _, s := range catalog {
		if s.Team == team {
			out = append(out, s)
		}
	}
	return out
}

// RepoRoot returns the repo root by walking up for go.mod or registry.
func RepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "registry")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root not found")
		}
		dir = parent
	}
}
