package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDetectMicroserviceRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/microservices/detect" || r.Method != http.MethodPost {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["repoUrl"] != "https://github.com/team/platform.git" || payload["branch"] != "develop" {
			t.Fatalf("payload = %#v", payload)
		}
		if _, ok := payload["githubToken"]; ok {
			t.Fatalf("public-only scan must not send githubToken: %#v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"repository": map[string]any{
				"provider":      "github",
				"fullName":      "team/platform",
				"htmlUrl":       "https://github.com/team/platform",
				"defaultBranch": "develop",
				"visibility":    "public",
			},
			"services": []map[string]any{{
				"name":          "api",
				"path":          "services/api",
				"repoProvider":  "github",
				"repoUrl":       "https://github.com/team/platform",
				"repoFullName":  "team/platform",
				"branch":        "develop",
				"serviceType":   "backend",
				"framework":     "spring-boot",
				"appPort":       8080,
				"exposePublic":  false,
				"primaryPublic": false,
			}},
		})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	result, err := client.DetectMicroserviceRepository(
		context.Background(),
		"access",
		"https://github.com/team/platform.git",
		"develop",
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Repository.FullName != "team/platform" || len(result.Services) != 1 {
		t.Fatalf("result = %#v", result)
	}
	input := result.Services[0].DeploymentInput()
	if input.Name != "api" || input.Path != "services/api" || input.AppPort != 8080 {
		t.Fatalf("deployment input = %#v", input)
	}
}

func TestValidateMicroserviceRepositoriesStopsOnMissingRemote(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if strings.Contains(r.URL.Path, "/detect") {
			http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
			return
		}
		t.Fatalf("path = %q", r.URL.Path)
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	err := client.ValidateMicroserviceRepositories(context.Background(), "access", []CreateMicroserviceServiceInput{
		{Name: "api", RepoURL: "https://github.com/team/missing", Branch: "main"},
		{Name: "worker", RepoURL: "https://github.com/team/missing", Branch: "main"},
	})
	if err == nil || !strings.Contains(err.Error(), "Git remote is not accessible") {
		t.Fatalf("err = %v", err)
	}
	if requests != 1 {
		t.Fatalf("expected duplicate repository to be checked once, requests=%d", requests)
	}
}

func TestDetectMicroserviceRepositoryRejectsPrivateRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"repository": map[string]any{
				"provider":   "github",
				"fullName":   "team/private",
				"visibility": "private",
			},
			"services": []map[string]any{},
		})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	_, err := client.DetectMicroserviceRepository(
		context.Background(),
		"access",
		"https://github.com/team/private",
		"main",
	)
	if err == nil || !strings.Contains(err.Error(), "public GitHub repositories only") {
		t.Fatalf("err = %v", err)
	}
}

func TestDetectMicroserviceRepositoryExplainsGitHubRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	_, err := client.DetectMicroserviceRepository(
		context.Background(),
		"access",
		"https://github.com/team/platform",
		"main",
	)
	if err == nil || !strings.Contains(err.Error(), "rate limit exceeded") ||
		!strings.Contains(err.Error(), "unauthenticated public scans may be rate-limited") {
		t.Fatalf("err = %v", err)
	}
}

func TestDetectMicroserviceRepositoryExplainsMissingBranch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	_, err := client.DetectMicroserviceRepository(
		context.Background(),
		"access",
		"https://github.com/team/platform",
		"missing",
	)
	if err == nil || !strings.Contains(err.Error(), "repository or branch was not found") {
		t.Fatalf("err = %v", err)
	}
}
