package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultEndpoint = "http://127.0.0.1:8787"
	DefaultTimeout  = 60
)

type Values struct {
	OutputFormat       string `json:"outputFormat"`
	Endpoint           string `json:"endpoint"`
	TimeoutSeconds     int    `json:"timeoutSeconds"`
	SystemAliasEnabled bool   `json:"systemAliasEnabled"`
	SystemAliasName    string `json:"systemAliasName"`
	CollisionPolicy    string `json:"collisionPolicy"`
}

type Overrides struct {
	OutputFormat   string
	Endpoint       string
	TimeoutSeconds int
	HasTimeout     bool
}

type Effective struct {
	Values       Values `json:"values"`
	Profile      string `json:"profile"`
	ConfigDir    string `json:"configDir"`
	DefaultPath  string `json:"defaultPath"`
	ProfilePath  string `json:"profilePath,omitempty"`
	DefaultFound bool   `json:"defaultFound"`
	ProfileFound bool   `json:"profileFound"`
}

func BuiltIn() Values {
	return Values{
		OutputFormat:       "human",
		Endpoint:           DefaultEndpoint,
		TimeoutSeconds:     DefaultTimeout,
		SystemAliasEnabled: false,
		SystemAliasName:    "prt",
		CollisionPolicy:    "refuse",
	}
}

func ConfigDir() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("PRAETORIUM_CONFIG_DIR")); configured != "" {
		return filepath.Clean(configured), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(home, ".praetorium", "config"), nil
}

func Paths(profile string) (dir, defaultPath, profilePath string, err error) {
	if err := ValidateProfile(profile); err != nil {
		return "", "", "", err
	}
	dir, err = ConfigDir()
	if err != nil {
		return "", "", "", err
	}
	defaultPath = filepath.Join(dir, "praetorium-config-default.ini")
	if profile != "default" {
		profilePath = filepath.Join(dir, "praetorium-config-"+profile+".ini")
	}
	return dir, defaultPath, profilePath, nil
}

func ValidateProfile(profile string) error {
	if profile == "" {
		return errors.New("profile name must not be empty")
	}
	for i, r := range profile {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9' && i > 0) || r == '_' || r == '-' || r == '.' {
			continue
		}
		return fmt.Errorf("invalid profile name %q: use letters, digits, '.', '_' or '-' and start with a letter", profile)
	}
	first := profile[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return fmt.Errorf("invalid profile name %q: it must start with a letter", profile)
	}
	return nil
}

func Load(profile string, overrides Overrides) (Effective, error) {
	if profile == "" {
		profile = "default"
	}
	dir, defaultPath, profilePath, err := Paths(profile)
	if err != nil {
		return Effective{}, err
	}
	values := BuiltIn()
	effective := Effective{Values: values, Profile: profile, ConfigDir: dir, DefaultPath: defaultPath, ProfilePath: profilePath}
	if found, err := mergeFile(&values, defaultPath); err != nil {
		return Effective{}, err
	} else {
		effective.DefaultFound = found
	}
	if profile != "default" {
		found, err := mergeFile(&values, profilePath)
		if err != nil {
			return Effective{}, err
		}
		effective.ProfileFound = found
		if !found {
			return Effective{}, fmt.Errorf("selected profile %q does not exist at %s", profile, profilePath)
		}
	}
	if overrides.OutputFormat != "" {
		values.OutputFormat = overrides.OutputFormat
	}
	if overrides.Endpoint != "" {
		values.Endpoint = overrides.Endpoint
	}
	if overrides.HasTimeout {
		values.TimeoutSeconds = overrides.TimeoutSeconds
	}
	if err := Validate(values); err != nil {
		return Effective{}, err
	}
	effective.Values = values
	return effective, nil
}

func Validate(values Values) error {
	if values.OutputFormat != "human" && values.OutputFormat != "json" {
		return fmt.Errorf("output format must be human or json, got %q", values.OutputFormat)
	}
	if err := validateEndpoint(values.Endpoint); err != nil {
		return err
	}
	if values.TimeoutSeconds < 1 || values.TimeoutSeconds > 3600 {
		return fmt.Errorf("timeout_seconds must be between 1 and 3600, got %d", values.TimeoutSeconds)
	}
	if values.SystemAliasEnabled && values.SystemAliasName == "" {
		return errors.New("system alias name is required when the system alias is enabled")
	}
	if values.SystemAliasName != "" {
		if err := validateAliasName(values.SystemAliasName); err != nil {
			return err
		}
	}
	if values.CollisionPolicy != "refuse" {
		return fmt.Errorf("collision_policy must be refuse, got %q", values.CollisionPolicy)
	}
	return nil
}

func mergeFile(values *Values, path string) (bool, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("open config %s: %w", path, err)
	}
	defer file.Close()
	section := ""
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			if section != "cli" && section != "coordinator" && section != "system_alias" {
				return false, fmt.Errorf("config %s line %d: unsupported section %q", path, lineNumber, section)
			}
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 || section == "" {
			return false, fmt.Errorf("config %s line %d: expected section and key = value", path, lineNumber)
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		if err := assign(values, section, key, value); err != nil {
			return false, fmt.Errorf("config %s line %d: %w", path, lineNumber, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("read config %s: %w", path, err)
	}
	return true, nil
}

func assign(values *Values, section, key, value string) error {
	switch section + "." + key {
	case "cli.output_format":
		values.OutputFormat = strings.ToLower(value)
	case "coordinator.endpoint":
		values.Endpoint = value
	case "coordinator.timeout_seconds":
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("timeout_seconds must be an integer")
		}
		values.TimeoutSeconds = parsed
	case "system_alias.enabled":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("enabled must be true or false")
		}
		values.SystemAliasEnabled = parsed
	case "system_alias.name":
		values.SystemAliasName = value
	case "system_alias.collision_policy":
		values.CollisionPolicy = strings.ToLower(value)
	default:
		return fmt.Errorf("unsupported key %s.%s", section, key)
	}
	return nil
}

func validateEndpoint(endpoint string) error {
	if endpoint == "" {
		return errors.New("coordinator endpoint must not be empty")
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return fmt.Errorf("coordinator endpoint must use http:// or https://, got %q", endpoint)
	}
	return nil
}

func validateAliasName(name string) error {
	for i, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9' && i > 0) || r == '_' {
			continue
		}
		return fmt.Errorf("system alias name %q must be a shell identifier", name)
	}
	if name == "" {
		return errors.New("system alias name must not be empty")
	}
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return fmt.Errorf("system alias name %q must start with a letter or underscore", name)
	}
	return nil
}
