package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		case "/api/v1/database-deployments/db-1/credentials":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"releaseName":          "orders",
				"namespace":            "ns-orders",
				"engine":               "postgresql",
				"databaseName":         "ordersdb",
				"username":             "orders_user",
				"password":             "runtime-secret",
				"serviceHost":          "db-orders.db.autonomous-istad.com",
				"servicePort":          float64(5432),
				"connectionTlsEnabled": true,
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
	if project.DatabaseUsername != "orders_user" || project.DatabasePassword != "runtime-secret" || !project.RequireSSL || !project.ConnectionTLSEnabled {
		t.Fatalf("database details = %#v", project)
	}
}

func TestFetchLiveProjectsHydratesClusterConnectionDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/api/v1/profile/me":
			_ = json.NewEncoder(w).Encode(map[string]any{"userId": "user-1"})
		case "/api/v1/projects/live":
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"kind":      "dbcluster",
				"id":        "cluster-1",
				"name":      "orders-ha",
				"status":    "DEPLOYED",
				"namespace": "ns-orders",
				"createdAt": "2026-05-25T01:00:00Z",
				"updatedAt": "2026-05-25T01:30:00Z",
			}})
		case "/api/v1/cluster/namespaces/ns-orders/clusters/cluster-1/console/credentials":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"namespace":            "ns-orders",
				"engine":               "postgresql",
				"databaseName":         "ordersdb",
				"username":             "orders_user",
				"password":             "cluster-secret",
				"serviceHost":          "orders-ha.autonomous-istad.com",
				"servicePort":          float64(15432),
				"connectionTlsEnabled": true,
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
	if project.ServiceHost != "orders-ha.autonomous-istad.com" || project.ServicePort != 15432 {
		t.Fatalf("connection details = %#v", project)
	}
	if project.DatabaseUsername != "orders_user" || project.DatabasePassword != "cluster-secret" || !project.ConnectionTLSEnabled {
		t.Fatalf("cluster credentials = %#v", project)
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

func TestCreateClusterDeployment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/namespaces/default/cluster-deployments" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		database, ok := payload["database"].(map[string]any)
		if !ok {
			t.Fatalf("database payload = %#v", payload["database"])
		}
		cluster, ok := payload["cluster"].(map[string]any)
		if !ok {
			t.Fatalf("cluster payload = %#v", payload["cluster"])
		}
		secrets, ok := payload["secrets"].(map[string]any)
		if !ok {
			t.Fatalf("secrets payload = %#v", payload["secrets"])
		}
		if payload["projectName"] != "orders-ha" || database["engine"] != "POSTGRESQL" || database["databaseName"] != "ordersdb" || database["username"] != "orders_user" {
			t.Fatalf("payload = %#v", payload)
		}
		if cluster["name"] != "orders-ha" || secrets["pgPassword"] != "secret" {
			t.Fatalf("payload = %#v", payload)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"clusterId":   "cluster-1",
			"releaseName": "orders-ha",
			"namespace":   "default",
			"engine":      "POSTGRESQL",
			"status":      "DEPLOYED",
			"stdout":      "helm upgrade ok",
			"successful":  true,
		})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	deployment, err := client.CreateClusterDeployment(context.Background(), "access", CreateClusterDeploymentInput{
		Namespace:    "default",
		ProjectName:  "orders-ha",
		Engine:       "postgresql",
		DatabaseName: "ordersdb",
		Username:     "orders_user",
		Password:     "secret",
		Version:      "18",
		SizeProfile:  "small",
	})
	if err != nil {
		t.Fatal(err)
	}
	if deployment.ClusterID != "cluster-1" || deployment.Status != "DEPLOYED" || deployment.Stdout != "helm upgrade ok" {
		t.Fatalf("deployment = %#v", deployment)
	}
}

func TestFetchClusterDeployment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/namespaces/workspace-a/clusters/cluster-1" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"clusterId":     "cluster-1",
			"releaseName":   "orders-db",
			"namespace":     "workspace-a",
			"status":        "DEPLOYED",
			"statusMessage": "All pods ready",
		})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	deployment, err := client.FetchClusterDeployment(context.Background(), "access", "workspace-a", "cluster-1")
	if err != nil {
		t.Fatal(err)
	}
	if deployment.ClusterID != "cluster-1" || deployment.Status != "DEPLOYED" {
		t.Fatalf("deployment = %#v", deployment)
	}
}

func TestResolveClusterDeploymentFallsBackToNamespaceList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cluster/namespaces/workspace-a/clusters/orders-db":
			http.Error(w, "invalid UUID", http.StatusBadRequest)
		case "/api/v1/cluster/namespaces/workspace-a/clusters":
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"clusterId":   "cluster-uuid",
				"name":        "orders",
				"releaseName": "orders-db",
				"namespace":   "workspace-a",
				"status":      "DEPLOYED",
			}})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	deployment, err := client.ResolveClusterDeployment(context.Background(), "access", "workspace-a", "orders-db", "orders-db", "orders")
	if err != nil {
		t.Fatal(err)
	}
	if deployment.ClusterID != "cluster-uuid" || deployment.Status != "DEPLOYED" {
		t.Fatalf("deployment = %#v", deployment)
	}
}

func TestDownloadClusterCertificate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/namespaces/workspace-a/clusters/cluster-1/certificate" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Disposition", `attachment; filename="postgresql-ca.crt"`)
		_, _ = w.Write([]byte("-----BEGIN CERTIFICATE-----\ncertificate\n-----END CERTIFICATE-----\n"))
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	certificate, err := client.DownloadClusterCertificate(context.Background(), "access", "workspace-a", "cluster-1")
	if err != nil {
		t.Fatal(err)
	}
	if certificate.Filename != "postgresql-ca.crt" || !strings.Contains(string(certificate.Content), "BEGIN CERTIFICATE") {
		t.Fatalf("certificate = %#v", certificate)
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

func TestCreateMicroserviceDeployment(t *testing.T) {
	var sawWorkspace bool
	var sawProfile bool
	var sawCreate bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/api/v1/workspaces/bootstrap":
			sawWorkspace = true
			_ = json.NewEncoder(w).Encode(map[string]any{"onboarded": true, "namespace": "workspace-team"})
		case "/api/v1/profile/me":
			sawProfile = true
			_ = json.NewEncoder(w).Encode(map[string]any{"userId": "user-1"})
		case "/api/v1/projects/microservices":
			sawCreate = true
			if r.Method != http.MethodPost {
				t.Fatalf("method = %q", r.Method)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			services, ok := payload["services"].([]any)
			if !ok || len(services) != 1 {
				t.Fatalf("services = %#v", payload["services"])
			}
			service, ok := services[0].(map[string]any)
			if !ok {
				t.Fatalf("service = %#v", services[0])
			}
			if payload["userId"] != "user-1" || payload["projectName"] != "commerce" || service["name"] != "api" || service["repoUrl"] != "https://github.com/team/api.git" {
				t.Fatalf("payload = %#v", payload)
			}
			if service["serviceType"] != "backend" || service["repoProvider"] != "github" || service["primaryPublic"] != true {
				t.Fatalf("service payload = %#v", service)
			}
			dependsOn, ok := service["dependsOn"].([]any)
			if !ok || len(dependsOn) != 1 || dependsOn[0] != "eureka-server" {
				t.Fatalf("dependsOn = %#v", service["dependsOn"])
			}
			relationships, ok := service["relationships"].([]any)
			if !ok || len(relationships) != 1 {
				t.Fatalf("relationships = %#v", service["relationships"])
			}
			relationship, ok := relationships[0].(map[string]any)
			if !ok || relationship["name"] != "EUREKA_URL" || relationship["value"] != "eureka-server" {
				t.Fatalf("relationship = %#v", relationships[0])
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"projectId":      "project-1",
				"name":           "commerce",
				"status":         "CREATED",
				"queueItemId":    "43",
				"jenkinsJobName": "deploy-microservices",
			})
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	deployment, err := client.CreateMicroserviceDeployment(context.Background(), "access", CreateMicroserviceDeploymentInput{
		ProjectName: "commerce",
		Branch:      "main",
		Services: []CreateMicroserviceServiceInput{{
			Name:          "api",
			RepoURL:       "https://github.com/team/api.git",
			RepoFullName:  "team/api",
			Branch:        "main",
			AppPort:       8080,
			ServiceType:   "backend",
			Framework:     "Go",
			ExposePublic:  true,
			PrimaryPublic: true,
			Relationships: []MicroserviceRelationInput{{Name: "EUREKA_URL", Value: "eureka-server"}},
			DependsOn:     []string{"eureka-server"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !sawWorkspace || !sawProfile || !sawCreate {
		t.Fatalf("expected workspace, profile, and create calls")
	}
	if deployment.ProjectID != "project-1" || deployment.QueueItemID != 43 || deployment.JenkinsJobName != "deploy-microservices" {
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

func TestFetchDatabaseDeploymentMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/database-deployments/db-1/metrics" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"deploymentId":     "db-1",
			"podName":          "orders-db-0",
			"podPhase":         "Running",
			"readyReplicas":    1,
			"replicas":         1,
			"cpuRequest":       "250m",
			"memoryRequest":    "512Mi",
			"storageRequested": "10Gi",
		})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	metrics, err := client.FetchDatabaseDeploymentMetrics(context.Background(), "access", "db-1")
	if err != nil {
		t.Fatal(err)
	}
	if metrics.DeploymentID != "db-1" || metrics.PodPhase != "Running" ||
		metrics.ReadyReplicas != 1 || metrics.CPURequest != "250m" {
		t.Fatalf("metrics = %#v", metrics)
	}
}

func TestDeleteLiveProjectRoutesByKind(t *testing.T) {
	tests := []struct {
		name        string
		project     LiveProject
		wantPaths   []string
		wantQueries []string
	}{
		{
			name:      "database",
			project:   LiveProject{Kind: "database", ID: "fallback-db", DatabaseDeploymentIDs: []string{"db-1", "db-2"}},
			wantPaths: []string{"/api/v1/database-deployments/db-1", "/api/v1/database-deployments/db-2"},
		},
		{
			name:      "monolith",
			project:   LiveProject{Kind: "monolith", ID: "project-1"},
			wantPaths: []string{"/api/v1/projects/project-1"},
		},
		{
			name:      "microservices",
			project:   LiveProject{Kind: "microservices", ID: "project-2"},
			wantPaths: []string{"/api/v1/projects/microservices/project-2"},
		},
		{
			name:        "dbcluster",
			project:     LiveProject{Kind: "dbcluster", ID: "cluster-1", Namespace: "ns-a"},
			wantPaths:   []string{"/api/v1/cluster/namespaces/ns-a/clusters/cluster-1"},
			wantQueries: []string{"deleteData=true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPaths []string
			var gotQueries []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Fatalf("method = %q", r.Method)
				}
				if r.Header.Get("Authorization") != "Bearer access" {
					t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
				}
				gotPaths = append(gotPaths, r.URL.Path)
				gotQueries = append(gotQueries, r.URL.RawQuery)
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := NewProjectClient(server.URL)
			client.client = server.Client()
			if err := client.DeleteLiveProject(context.Background(), "access", tt.project); err != nil {
				t.Fatal(err)
			}
			if len(gotPaths) != len(tt.wantPaths) {
				t.Fatalf("paths = %#v", gotPaths)
			}
			for i := range tt.wantPaths {
				if gotPaths[i] != tt.wantPaths[i] {
					t.Fatalf("paths = %#v", gotPaths)
				}
			}
			for i := range tt.wantQueries {
				if gotQueries[i] != tt.wantQueries[i] {
					t.Fatalf("queries = %#v", gotQueries)
				}
			}
		})
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

func TestInferDeploymentLogLevel(t *testing.T) {
	tests := map[string]string{
		"Deployment completed successfully.": "success",
		"Release is pending readiness.":      "warn",
		"Access denied by cluster policy.":   "error",
		"Applying GitOps values.":            "info",
	}
	for line, want := range tests {
		if got := InferLogLevel(line); got != want {
			t.Fatalf("InferLogLevel(%q) = %q, want %q", line, got, want)
		}
	}
}

func TestDatabaseDeploymentTerminalSuccessAliases(t *testing.T) {
	for _, status := range []string{"SUCCESSFUL", "COMPLETED", "DEPLOYED"} {
		if !DatabaseDeploymentTerminal(status) || DatabaseDeploymentFailed(status) {
			t.Fatalf("expected successful terminal status for %q", status)
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
