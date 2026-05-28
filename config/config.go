package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var backendURLKeys = []string{
	"BACKEND_API_BASE_URL",
	"BACKEND_API_URL",
	"API_URL",
	"NEXT_PUBLIC_BACKEND_API_BASE_URL",
	"NEXT_PUBLIC_BACKEND_API_URL",
	"NEXT_PUBLIC_API_URL",
}

type AppConfig struct {
	BackendBaseURL       string
	KeycloakURL          string
	KeycloakRealm        string
	KeycloakClientID     string
	KeycloakClientSecret string
	KeycloakRedirectURL  string
}

func LoadConfig() (AppConfig, error) {
	values := map[string]string{}

	for _, path := range envCandidates("tui/.env") {
		mergeEnvFile(values, path)
	}

	for _, pair := range os.Environ() {
		key, value, ok := strings.Cut(pair, "=")
		if ok && strings.TrimSpace(key) != "" {
			values[key] = value
		}
	}

	cfg := AppConfig{
		BackendBaseURL:       resolveBackendBaseURL(values),
		KeycloakURL:          fallback(trimTrailingSlash(values["KEYCLOAK_URL"]), DefaultKeycloakURL),
		KeycloakRealm:        fallback(strings.TrimSpace(values["KEYCLOAK_REALM"]), DefaultKeycloakRealm),
		KeycloakClientID:     fallback(strings.TrimSpace(values["KEYCLOAK_CLIENT_ID"]), DefaultKeycloakClientID),
		KeycloakClientSecret: fallback(strings.TrimSpace(values["KEYCLOAK_CLIENT_SECRET"]), DefaultKeycloakClientSecret),
		KeycloakRedirectURL:  fallback(strings.TrimSpace(values["KEYCLOAK_REDIRECT_URL"]), DefaultKeycloakRedirectURL),
	}

	var missing []string
	if cfg.KeycloakURL == "" { missing = append(missing, "KEYCLOAK_URL") }
	if cfg.KeycloakRealm == "" { missing = append(missing, "KEYCLOAK_REALM") }
	if cfg.KeycloakClientID == "" { missing = append(missing, "KEYCLOAK_CLIENT_ID") }
	if cfg.KeycloakRedirectURL == "" { missing = append(missing, "KEYCLOAK_REDIRECT_URL") }
	if len(missing) > 0 {
		return cfg, fmt.Errorf("missing required TUI env: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func envCandidates(path string) []string {
	cwd, err := os.Getwd()
	if err != nil {
		return []string{path}
	}

	base := filepath.Base(cwd)
	candidates := []string{
		filepath.Join(cwd, path),
		filepath.Join(cwd, filepath.Base(path)),
		filepath.Join(cwd, "..", path),
	}
	if base == "tui" {
		candidates = append(candidates, filepath.Join(cwd, "..", path))
	}
	
	// Add global home directory config
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".a8s-cli.env"))
	}
	
	return uniquePaths(candidates)
}

func uniquePaths(paths []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, path := range paths {
		cleaned := filepath.Clean(path)
		if !seen[cleaned] {
			seen[cleaned] = true
			out = append(out, cleaned)
		}
	}
	return out
}

func mergeEnvFile(values map[string]string, path string) {
	fileValues, err := parseEnvFile(path)
	if err != nil {
		return
	}
	for key, value := range fileValues {
		values[key] = value
	}
}

func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		return nil, fmt.Errorf("read env file %s: %w", path, err)
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		values[key] = unquoteEnvValue(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan env file %s: %w", path, err)
	}
	return values, nil
}

func unquoteEnvValue(value string) string {
	if len(value) >= 2 {
		first := value[0]
		last := value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func resolveBackendBaseURL(values map[string]string) string {
	for _, key := range backendURLKeys {
		value := trimTrailingSlash(values[key])
		if value != "" && !strings.Contains(value, "REPLACE_WITH_PROD_BACKEND_ORIGIN") {
			return value
		}
	}
	return "http://localhost:8080"
}

func trimTrailingSlash(value string) string {
	value = strings.TrimSpace(value)
	for strings.HasSuffix(value, "/") {
		value = strings.TrimSuffix(value, "/")
	}
	return value
}

func (c AppConfig) KeycloakIssuer() string {
	return c.KeycloakURL + "/realms/" + strings.Trim(c.KeycloakRealm, "/")
}

func (c AppConfig) AuthURL() string {
	return c.KeycloakIssuer() + "/protocol/openid-connect/auth"
}

func (c AppConfig) TokenURL() string {
	return c.KeycloakIssuer() + "/protocol/openid-connect/token"
}

func fallback(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
