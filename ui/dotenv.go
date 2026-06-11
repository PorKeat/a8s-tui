package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PorKeat/a8s-tui/api"
)

const maxDotEnvFileSize = 1024 * 1024

var dotEnvNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func loadDotEnvFile(path string) ([]api.MicroserviceEnvInput, string, error) {
	resolvedPath, err := resolveLocalPath(path)
	if err != nil {
		return nil, "", err
	}
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", filepath.Base(resolvedPath), err)
	}
	if info.IsDir() {
		return nil, "", fmt.Errorf("%s is a directory", resolvedPath)
	}
	if info.Size() > maxDotEnvFileSize {
		return nil, "", fmt.Errorf("%s is larger than 1 MiB", filepath.Base(resolvedPath))
	}
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", filepath.Base(resolvedPath), err)
	}
	envVars, err := parseDotEnv(string(content))
	if err != nil {
		return nil, "", err
	}
	if len(envVars) == 0 {
		return nil, "", fmt.Errorf("%s contains no valid environment variables", filepath.Base(resolvedPath))
	}
	return envVars, resolvedPath, nil
}

func resolveLocalPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("environment file path is required")
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve environment file path: %w", err)
	}
	return filepath.Clean(resolved), nil
}

func parseDotEnv(content string) ([]api.MicroserviceEnvInput, error) {
	values := make([]api.MicroserviceEnvInput, 0)
	indexByName := map[string]int{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 1024), maxDotEnvFileSize)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		separator := strings.IndexByte(line, '=')
		if separator <= 0 {
			continue
		}
		name := strings.TrimSpace(line[:separator])
		if !dotEnvNamePattern.MatchString(name) {
			return nil, fmt.Errorf("invalid environment variable name %q on line %d", name, lineNumber)
		}
		value := strings.TrimSpace(line[separator+1:])
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		entry := api.MicroserviceEnvInput{Name: name, Value: value, Secret: isSecretEnvName(name)}
		if index, ok := indexByName[name]; ok {
			values[index] = entry
			continue
		}
		indexByName[name] = len(values)
		values = append(values, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read environment file: %w", err)
	}
	return values, nil
}

func isSecretEnvName(name string) bool {
	normalized := strings.ToUpper(name)
	return strings.Contains(normalized, "SECRET") ||
		strings.Contains(normalized, "PASSWORD") ||
		strings.Contains(normalized, "TOKEN") ||
		strings.Contains(normalized, "PRIVATE_KEY")
}

func mergeMicroserviceEnv(current, imported []api.MicroserviceEnvInput) []api.MicroserviceEnvInput {
	merged := append([]api.MicroserviceEnvInput(nil), current...)
	indexByName := make(map[string]int, len(merged))
	for index, entry := range merged {
		indexByName[entry.Name] = index
	}
	for _, entry := range imported {
		if index, ok := indexByName[entry.Name]; ok {
			merged[index] = entry
			continue
		}
		indexByName[entry.Name] = len(merged)
		merged = append(merged, entry)
	}
	return merged
}
