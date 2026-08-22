package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Oliveszn/Schema-Watch/internal/schema"
)

func TestLoad_ParsesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
target: "http://localhost:8080"
port: "9090"
ignore_fields:
  - updated_at
  - created_at
ignore_endpoints:
  - "GET /health"
  - "GET /metrics*"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Target != "http://localhost:8080" || cfg.Port != "9090" {
		t.Fatalf("unexpected target/port: %+v", cfg)
	}
	if len(cfg.IgnoreFields) != 2 || len(cfg.IgnoreEndpoints) != 2 {
		t.Fatalf("unexpected list lengths: %+v", cfg)
	}
}

func TestLoad_MissingFileReturnsNotExistError(t *testing.T) {
	_, err := Load("/definitely/does/not/exist.yaml")
	if !os.IsNotExist(err) {
		t.Fatalf("expected an IsNotExist error, got %v", err)
	}
}

func TestIsEndpointIgnored_ExactMatch(t *testing.T) {
	cfg := &Config{IgnoreEndpoints: []string{"GET /health"}}

	if !cfg.IsEndpointIgnored("GET /health") {
		t.Error("expected exact match to be ignored")
	}
	if cfg.IsEndpointIgnored("GET /healthcheck") {
		t.Error("expected non-exact match to NOT be ignored")
	}
}

func TestIsEndpointIgnored_WildcardPrefix(t *testing.T) {
	cfg := &Config{IgnoreEndpoints: []string{"GET /metrics*"}}

	if !cfg.IsEndpointIgnored("GET /metrics/cpu") {
		t.Error("expected prefix match to be ignored")
	}
	if cfg.IsEndpointIgnored("GET /users/1") {
		t.Error("expected unrelated endpoint to NOT be ignored")
	}
}

func TestIsEndpointIgnored_NilConfigNeverIgnores(t *testing.T) {
	var cfg *Config
	if cfg.IsEndpointIgnored("GET /anything") {
		t.Error("expected nil config to never ignore an endpoint")
	}
}

func TestFilterSchema_RemovesMatchingLeafNames(t *testing.T) {
	cfg := &Config{IgnoreFields: []string{"updated_at"}}
	in := schema.Schema{
		"id":                schema.TypeNumber,
		"updated_at":        schema.TypeString,
		"user.updated_at":   schema.TypeString,
		"user.address.city": schema.TypeString,
	}

	out := cfg.FilterSchema(in)

	if _, ok := out["updated_at"]; ok {
		t.Error("expected top-level updated_at to be filtered out")
	}
	if _, ok := out["user.updated_at"]; ok {
		t.Error("expected nested user.updated_at to be filtered out (leaf match)")
	}
	if _, ok := out["id"]; !ok {
		t.Error("expected id to survive filtering")
	}
	if _, ok := out["user.address.city"]; !ok {
		t.Error("expected unrelated nested field to survive filtering")
	}
}

func TestFilterSchema_NilConfigReturnsUnchanged(t *testing.T) {
	var cfg *Config
	in := schema.Schema{"id": schema.TypeNumber}

	out := cfg.FilterSchema(in)
	if len(out) != 1 || out["id"] != schema.TypeNumber {
		t.Fatalf("expected schema unchanged for nil config, got %+v", out)
	}
}

func TestFilterSchema_EmptyIgnoreListReturnsUnchanged(t *testing.T) {
	cfg := &Config{}
	in := schema.Schema{"id": schema.TypeNumber}

	out := cfg.FilterSchema(in)
	if len(out) != 1 {
		t.Fatalf("expected schema unchanged for empty ignore list, got %+v", out)
	}
}
