package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeLiveProject(t *testing.T) {
	project, ok := normalizeLiveProject(map[string]any{
		"kind":              "monolith",
		"id":                "project-1",
		"name":              "API",
		"status":            "DEPLOYED",
		"repoFullName":      "team/api",
		"deploymentCount":   float64(2),
		"autoDeployEnabled": true,
	})
	if !ok {
		t.Fatal("expected project to normalize")
	}
	if project.Kind != "monolith" || project.Name != "API" || project.DeploymentCount != 2 || !project.AutoDeployEnabled {
		t.Fatalf("unexpected project: %#v", project)
	}
}

func TestFilteredProjects(t *testing.T) {
	projects := []LiveProject{
		{Name: "Frontend", Kind: "monolith", Status: "DEPLOYED"},
		{Name: "Data", Kind: "database", Engine: "postgres"},
	}
	got := FilteredProjects(projects, "post")
	if len(got) != 1 || got[0].Name != "Data" {
		t.Fatalf("filtered projects = %#v", got)
	}
}

func TestFetchLiveProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/api/v1/profile/me":
			_ = json.NewEncoder(w).Encode(map[string]any{"userId": "user-1"})
			return
		case "/api/v1/projects/live":
			if r.URL.Query().Get("userId") != "user-1" {
				t.Fatalf("userId = %q", r.URL.Query().Get("userId"))
			}
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"kind":      "database",
			"id":        "db-1",
			"name":      "Orders",
			"status":    "RUNNING",
			"createdAt": "2026-05-25T01:00:00Z",
			"updatedAt": "2026-05-25T01:30:00Z",
		}})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	projects, _, err := client.FetchLiveProjects(context.Background(), "access")
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Name != "Orders" {
		t.Fatalf("projects = %#v", projects)
	}
}

func TestFetchLiveProjectsHydratesDatabaseConnectionDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/api/v1/profile/me":
			_ = json.NewEncoder(w).Encode(map[string]any{"userId": "user-1"})
		case "/api/v1/projects/live":
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"kind":                  "database",
				"id":                    "orders",
				"name":                  "Orders",
				"status":                "DEPLOYED",
				"createdAt":             "2026-05-25T01:00:00Z",
				"updatedAt":             "2026-05-25T01:30:00Z",
				"databaseDeploymentIds": []string{"db-1"},
			}})
		case "/api/v1/database-deployments/db-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                   "db-1",
				"projectName":          "orders",
				"engine":               "postgresql",
				"deploymentMode":       "single-instance",
				"databaseName":         "ordersdb",
				"username":             "orders_user",
				"version":              "18",
				"namespace":            "ns-orders",
				"serviceHost":          "db-orders.db.autonomous-istad.com",
				"servicePort":          float64(5432),
				"requireSsl":           true,
				"connectionTlsEnabled": true,
				"status":               "DEPLOYED",
			})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	projects, _, err := client.FetchLiveProjects(context.Background(), "access")
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("projects = %#v", projects)
	}
	project := projects[0]
	if project.ServiceHost != "db-orders.db.autonomous-istad.com" || project.ServicePort != 5432 {
		t.Fatalf("connection details = %#v", project)
	}
	if project.DatabaseUsername != "orders_user" || !project.RequireSSL || !project.ConnectionTLSEnabled {
		t.Fatalf("database details = %#v", project)
	}
}

func TestFetchLiveProjectsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "nope"})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	_, _, err := client.FetchLiveProjects(context.Background(), "bad")
	if !IsUnauthorized(err) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestCreateDatabaseDeployment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/database-deployments" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		var payload CreateDatabaseDeploymentInput
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.ProjectName != "orders" || payload.Engine != "postgresql" || payload.DeploymentMode != "single-instance" {
			t.Fatalf("payload = %#v", payload)
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":           "db-1",
			"projectName":  "orders",
			"engine":       "postgresql",
			"databaseName": "ordersdb",
			"status":       "DEPLOYING",
		})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	deployment, err := client.CreateDatabaseDeployment(context.Background(), "access", CreateDatabaseDeploymentInput{
		ProjectName:    "orders",
		Engine:         "postgresql",
		DeploymentMode: "single-instance",
		DatabaseName:   "ordersdb",
		Username:       "orders_user",
		Password:       "secret",
		Version:        "18",
		SizeProfile:    "small",
	})
	if err != nil {
		t.Fatal(err)
	}
	if deployment.ID != "db-1" || deployment.Status != "DEPLOYING" {
		t.Fatalf("deployment = %#v", deployment)
	}
}

func TestCreateMonolithicDeployment(t *testing.T) {
	var sawProfile bool
	var sawCreate bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/api/v1/profile/me":
			sawProfile = true
			_ = json.NewEncoder(w).Encode(map[string]any{"userId": "user-1"})
		case "/api/v1/projects":
			sawCreate = true
			if r.Method != http.MethodPost {
				t.Fatalf("method = %q", r.Method)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload["userId"] != "user-1" || payload["projectName"] != "web" || payload["repoUrl"] != "https://github.com/team/web.git" {
				t.Fatalf("payload = %#v", payload)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"projectId":      "project-1",
				"name":           "web",
				"status":         "CREATED",
				"deployUrl":      "https://web.example.com",
				"queueItemId":    float64(42),
				"jenkinsJobName": "deploy-monolith",
			})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	deployment, err := client.CreateMonolithicDeployment(context.Background(), "access", CreateMonolithicDeploymentInput{
		ProjectName:      "web",
		RepoURL:          "https://github.com/team/web.git",
		RepoFullName:     "team/web",
		Branch:           "main",
		AppPort:          3000,
		ArchitectureType: "monolithic",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !sawProfile || !sawCreate {
		t.Fatalf("expected profile and create calls")
	}
	if deployment.ProjectID != "project-1" || deployment.QueueItemID != 42 {
		t.Fatalf("deployment = %#v", deployment)
	}
}

func TestFetchDatabaseDeployment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/database-deployments/db-1" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "db-1",
			"projectName":   "orders",
			"status":        "DEPLOYING",
			"statusMessage": "Creating database resources",
			"statusLog":     "queued\ncreating namespace\ncreating namespace\nready",
		})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	deployment, err := client.FetchDatabaseDeployment(context.Background(), "access", "db-1")
	if err != nil {
		t.Fatal(err)
	}
	if deployment.ID != "db-1" || deployment.StatusLog == "" || deployment.StatusMessage == "" {
		t.Fatalf("deployment = %#v", deployment)
	}
}

func TestParseDeploymentLogLines(t *testing.T) {
	lines := ParseDeploymentLogLines("queued\r\ncreating namespace\n\ncreating namespace\nready")
	want := []string{"queued", "creating namespace", "ready"}
	if len(lines) != len(want) {
		t.Fatalf("lines = %#v", lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("lines = %#v", lines)
		}
	}
}

func TestFetchBackendUserID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/profile/me" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"userId": "user-123"})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	userID, _, err := client.fetchBackendUserID(context.Background(), "access")
	if err != nil {
		t.Fatal(err)
	}
	if userID != "user-123" {
		t.Fatalf("userID = %q", userID)
	}
}
