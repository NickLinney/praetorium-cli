package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesProfileOverDefaultAndBuiltIn(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PRAETORIUM_CONFIG_DIR", dir)
	if err := os.WriteFile(filepath.Join(dir, "praetorium-config-default.ini"), []byte("[coordinator]\nendpoint = http://default:8787\ntimeout_seconds = 30\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "praetorium-config-local.ini"), []byte("[coordinator]\nendpoint = http://profile:8787\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load("local", Overrides{Endpoint: "http://override:8787"})
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Values.Endpoint != "http://override:8787" || loaded.Values.TimeoutSeconds != 30 {
		t.Fatalf("unexpected effective configuration: %+v", loaded.Values)
	}
}

func TestLoadRejectsPathTraversalProfile(t *testing.T) {
	if _, err := Load("../outside", Overrides{}); err == nil {
		t.Fatal("expected path traversal profile to fail")
	}
}
