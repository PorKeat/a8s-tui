package deploy

import (
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type LocalProject struct {
	Name         string
	Directory    string
	RepoURL      string
	RepoFullName string
	Branch       string
	Framework    string
	AppPort      int
}

func DetectLocalProject() LocalProject {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	repoURL := gitOutput(cwd, "remote", "get-url", "origin")
	branch := gitOutput(cwd, "branch", "--show-current")
	if branch == "" {
		branch = "main"
	}
	repoFullName := RepoFullNameFromURL(repoURL)
	name := SlugifyProjectName(filepath.Base(cwd))
	if repoFullName != "" {
		parts := strings.Split(repoFullName, "/")
		name = SlugifyProjectName(parts[len(parts)-1])
	}
	framework, port := detectFrameworkAndPort(cwd)
	return LocalProject{
		Name:         name,
		Directory:    cwd,
		RepoURL:      repoURL,
		RepoFullName: repoFullName,
		Branch:       branch,
		Framework:    framework,
		AppPort:      port,
	}
}

func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func RepoFullNameFromURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "git@") {
		_, after, ok := strings.Cut(raw, ":")
		if !ok {
			return ""
		}
		return trimGitSuffix(after)
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Path == "" {
		return ""
	}
	return trimGitSuffix(strings.TrimPrefix(parsed.Path, "/"))
}

func trimGitSuffix(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".git")
	parts := strings.Split(value, "/")
	if len(parts) < 2 {
		return value
	}
	return strings.Join(parts[len(parts)-2:], "/")
}

func detectFrameworkAndPort(dir string) (string, int) {
	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
			Scripts         map[string]string `json:"scripts"`
		}
		_ = json.Unmarshal(data, &pkg)
		hasDep := func(name string) bool {
			_, ok := pkg.Dependencies[name]
			if ok {
				return true
			}
			_, ok = pkg.DevDependencies[name]
			return ok
		}
		scripts := strings.ToLower(strings.Join(mapValues(pkg.Scripts), " "))
		switch {
		case hasDep("next"):
			return "Next.js", 3000
		case hasDep("@vitejs/plugin-react") || hasDep("vite") || strings.Contains(scripts, "vite"):
			return "Vite", 5173
		case hasDep("react"):
			return "React", 3000
		case hasDep("express") || strings.Contains(scripts, "node"):
			return "Node.js", 3000
		default:
			return "Node.js", 3000
		}
	}
	if fileExists(filepath.Join(dir, "pom.xml")) || fileExists(filepath.Join(dir, "build.gradle")) || fileExists(filepath.Join(dir, "build.gradle.kts")) {
		return "Spring Boot", 8080
	}
	if fileExists(filepath.Join(dir, "go.mod")) {
		return "Go", 8080
	}
	if fileExists(filepath.Join(dir, "requirements.txt")) || fileExists(filepath.Join(dir, "pyproject.toml")) {
		return "Python", 8000
	}
	return "Docker", 3000
}

func mapValues(values map[string]string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

var nonSlugCharPattern = regexp.MustCompile(`[^a-z0-9-]+`)

func SlugifyProjectName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimSuffix(value, ".git")
	value = strings.ReplaceAll(value, "_", "-")
	value = nonSlugCharPattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "monolith-app"
	}
	return value
}

func ParsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
