package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResolveEffectiveClusterNamespaceUsesWorkspaceBootstrap(t *testing.T) {
	var postCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/bootstrap" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Method == http.MethodPost {
			postCalls++
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"onboarded": true, "namespace": "workspace-cheng"})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	namespace, err := client.ResolveEffectiveClusterNamespace(context.Background(), "access", "default")
	if err != nil {
		t.Fatal(err)
	}
	if namespace != "workspace-cheng" || postCalls != 0 {
		t.Fatalf("namespace = %q, postCalls = %d", namespace, postCalls)
	}
}

func TestResolveEffectiveClusterNamespaceBootstrapsWhenMissing(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "workspace missing"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"onboarded": true, "namespace": "workspace-new"})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	namespace, err := client.ResolveEffectiveClusterNamespace(context.Background(), "access", "")
	if err != nil {
		t.Fatal(err)
	}
	if namespace != "workspace-new" || calls != 2 {
		t.Fatalf("namespace = %q, calls = %d", namespace, calls)
	}
}

func TestResolveEffectiveClusterNamespaceReportsOnboarding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{"onboarded": false, "message": "Provisioning your workspace"})
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	_, err := client.ResolveEffectiveClusterNamespace(context.Background(), "access", "default")
	if err == nil || !strings.Contains(err.Error(), "Provisioning your workspace") {
		t.Fatalf("err = %v", err)
	}
}

func TestResolveEffectiveClusterNamespaceKeepsExplicitNamespace(t *testing.T) {
	client := NewProjectClient("http://unused")
	namespace, err := client.ResolveEffectiveClusterNamespace(context.Background(), "access", "workspace-existing")
	if err != nil || namespace != "workspace-existing" {
		t.Fatalf("namespace = %q, err = %v", namespace, err)
	}
}
