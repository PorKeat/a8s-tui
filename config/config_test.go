package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "KEY=value\nQUOTED=\"hello world\"\n# ignored\nEMPTY=\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	values, err := parseEnvFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if values["KEY"] != "value" {
		t.Fatalf("KEY = %q", values["KEY"])
	}
	if values["QUOTED"] != "hello world" {
		t.Fatalf("QUOTED = %q", values["QUOTED"])
	}
	if _, ok := values["EMPTY"]; !ok {
		t.Fatal("expected EMPTY key to be preserved")
	}
}

func TestResolveBackendBaseURLPrecedence(t *testing.T) {
	values := map[string]string{
		"NEXT_PUBLIC_API_URL": "https://public.example.com/",
		"BACKEND_API_URL":     "https://backend.example.com/",
	}

	got := resolveBackendBaseURL(values)
	if got != "https://backend.example.com" {
		t.Fatalf("backend URL = %q", got)
	}
}

func TestResolveBackendBaseURLFallback(t *testing.T) {
	got := resolveBackendBaseURL(map[string]string{})
	if got != "" {
		t.Fatalf("fallback URL = %q", got)
	}
}

func TestLoadConfigUsesOnlyTUIEnvFile(t *testing.T) {
	root := t.TempDir()
	tuiDir := filepath.Join(root, "tui")
	frontendDir := filepath.Join(root, "a8s-frontend")
	if err := os.MkdirAll(tuiDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(frontendDir, 0o700); err != nil {
		t.Fatal(err)
	}

	for _, key := range backendURLKeys {
		os.Unsetenv(key)
	}
	os.Unsetenv("KEYCLOAK_URL")
	os.Unsetenv("KEYCLOAK_REALM")
	os.Unsetenv("KEYCLOAK_CLIENT_ID")
	os.Unsetenv("KEYCLOAK_CLIENT_SECRET")
	os.Unsetenv("KEYCLOAK_REDIRECT_URL")

	frontendEnv := strings.Join([]string{
		"KEYCLOAK_URL=https://wrong-keycloak.example.com",
		"KEYCLOAK_REALM=wrong",
		"KEYCLOAK_CLIENT_ID=wrong-client",
		"KEYCLOAK_CLIENT_SECRET=wrong-secret",
		"KEYCLOAK_REDIRECT_URL=http://localhost:9999/callback",
		"BACKEND_API_URL=https://wrong-backend.example.com",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(frontendDir, ".env"), []byte(frontendEnv), 0o600); err != nil {
		t.Fatal(err)
	}

	tuiEnv := strings.Join([]string{
		"KEYCLOAK_URL=https://keycloak.example.com/",
		"KEYCLOAK_REALM=a8s",
		"KEYCLOAK_CLIENT_ID=a8s-frontend",
		"KEYCLOAK_CLIENT_SECRET=tui-secret",
		"KEYCLOAK_REDIRECT_URL=http://localhost:8250/callback",
		"BACKEND_API_URL=https://backend.example.com",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(tuiDir, ".env"), []byte(tuiEnv), 0o600); err != nil {
		t.Fatal(err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.Chdir(tuiDir); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BackendBaseURL != "https://backend.example.com" {
		t.Fatalf("backend URL = %q", cfg.BackendBaseURL)
	}
	if cfg.KeycloakURL != "https://keycloak.example.com" {
		t.Fatalf("keycloak URL = %q", cfg.KeycloakURL)
	}
	if cfg.KeycloakClientSecret != "tui-secret" {
		t.Fatal("expected TUI env secret to be used")
	}
}
