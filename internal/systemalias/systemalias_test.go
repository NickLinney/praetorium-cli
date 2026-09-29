package systemalias

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyCreatesCollisionSafeManagedAlias(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "config")
	state, err := Apply(configDir, "prt_test")
	if err != nil {
		t.Fatal(err)
	}
	if !state.Applied {
		t.Fatal("expected alias to be applied")
	}
	if _, err := os.Stat(state.Path); err != nil {
		t.Fatal(err)
	}
	second, err := Inspect(configDir, "prt_test")
	if err != nil {
		t.Fatal(err)
	}
	if !second.Applied || len(second.Collisions) != 0 {
		t.Fatalf("unexpected repeat state: %+v", second)
	}
}

func TestApplyRefusesDifferentExistingManagedFile(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "config")
	path := filepath.Join(filepath.Dir(configDir), "aliases", "prt_test.sh")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not managed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(configDir, "prt_test"); err == nil {
		t.Fatal("expected existing different alias file to be refused")
	}
}
