package systemalias

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type State struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Applied    bool     `json:"applied"`
	Collisions []string `json:"collisions"`
}

var shellName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func Inspect(configDir, name string) (State, error) {
	if !shellName.MatchString(name) {
		return State{}, fmt.Errorf("system alias name %q must be a shell identifier", name)
	}
	path := filepath.Join(filepath.Dir(configDir), "aliases", name+".sh")
	state := State{Name: name, Path: path}
	if executable, err := exec.LookPath(name); err == nil {
		state.Collisions = append(state.Collisions, "executable on PATH: "+executable)
	}
	for _, rcPath := range shellRCPaths() {
		if content, err := os.ReadFile(rcPath); err == nil && containsShellDefinition(string(content), name) {
			state.Collisions = append(state.Collisions, "shell definition in "+rcPath)
		}
	}
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			state.Collisions = append(state.Collisions, "managed alias path is a symlink: "+path)
		} else if content, readErr := os.ReadFile(path); readErr != nil {
			state.Collisions = append(state.Collisions, "managed alias path is unreadable: "+path)
		} else if string(content) == expected(name) {
			state.Applied = true
		} else {
			state.Collisions = append(state.Collisions, "managed alias file contains different content: "+path)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return State{}, fmt.Errorf("inspect alias path: %w", err)
	}
	return state, nil
}

func Apply(configDir, name string) (State, error) {
	state, err := Inspect(configDir, name)
	if err != nil {
		return State{}, err
	}
	if len(state.Collisions) > 0 {
		return state, fmt.Errorf("refusing system alias %q because a collision was detected: %s", name, strings.Join(state.Collisions, "; "))
	}
	if state.Applied {
		return state, nil
	}
	aliasDir := filepath.Dir(state.Path)
	if err := os.MkdirAll(aliasDir, 0o700); err != nil {
		return state, fmt.Errorf("create alias directory: %w", err)
	}
	file, err := os.OpenFile(state.Path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return state, fmt.Errorf("create alias file: %w", err)
	}
	if _, err := file.WriteString(expected(name)); err != nil {
		file.Close()
		return state, fmt.Errorf("write alias file: %w", err)
	}
	if err := file.Close(); err != nil {
		return state, fmt.Errorf("close alias file: %w", err)
	}
	state.Applied = true
	return state, nil
}

func expected(name string) string {
	return "# Managed by praetorium config apply. Source this file to enable the alias.\n" +
		"alias " + name + "='praetorium'\n"
}

func shellRCPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".profile"),
	}
}

func containsShellDefinition(content, name string) bool {
	pattern := regexp.MustCompile(`(?m)^\s*(alias\s+` + regexp.QuoteMeta(name) + `\s*=|function\s+` + regexp.QuoteMeta(name) + `\b|` + regexp.QuoteMeta(name) + `\s*\(\s*\)|export\s+` + regexp.QuoteMeta(name) + `\s*=)`)
	return pattern.MatchString(content)
}
