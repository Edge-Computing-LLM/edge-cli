package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingUsesDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GPU.Vendor != "nvidia" {
		t.Fatalf("vendor = %q, want nvidia", cfg.GPU.Vendor)
	}
	if cfg.Cluster.DefaultNamespace != "llm-observability" {
		t.Fatalf("namespace = %q", cfg.Cluster.DefaultNamespace)
	}
}

func TestLoadRejectsNonNVIDIA(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("gpu:\n  vendor: amd\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load unexpectedly accepted unsupported vendor")
	}
}
