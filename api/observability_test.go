package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMonitoringOverviewFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/monitoring/overview" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("includeSeries") != "true" || r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("query/header mismatch: %s %s", r.URL.RawQuery, r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"namespace":"workspace-dev",
			"generatedAt":"2026-06-04T02:00:00Z",
			"clusters":[{"cluster":"primary","upTargets":8,"nodes":3}],
			"namespaceMetrics":{"totalPods":4,"runningPods":3,"failedPods":1,"cpuCores":1.5,"memoryBytes":1048576},
			"namespaceSeries":[{"key":"cpu","label":"CPU","unit":"cores","points":[{"timestamp":"2026-06-04T02:00:00Z","value":1.5}]}],
			"projects":[{"id":"p1","name":"api","kind":"monolith","status":"DEPLOYED","namespace":"workspace-dev","totalPods":2,"runningPods":2}]
		}`)
	}))
	defer server.Close()

	overview, err := NewObservabilityClient(server.URL).MonitoringOverview(context.Background(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if overview.Namespace != "workspace-dev" || len(overview.Clusters) != 1 || len(overview.Projects) != 1 {
		t.Fatalf("overview = %#v", overview)
	}
	if overview.NamespaceMetrics.RunningPods != 3 || overview.NamespaceSeries[0].Points[0].Value != 1.5 {
		t.Fatalf("metrics = %#v", overview.NamespaceMetrics)
	}
}

func TestListPodsAndFetchPodLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/kubernetes/namespaces/workspace-dev/pods":
			if r.URL.Query().Get("targetClusterName") != "primary" {
				t.Fatalf("targetClusterName = %q", r.URL.Query().Get("targetClusterName"))
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"namespace":"workspace-dev","total":1,"pods":[{"name":"api-0","phase":"Running","readyContainers":1,"totalContainers":1,"restartCount":0,"ageSeconds":90}]}`)
		case "/api/kubernetes/namespaces/workspace-dev/pods/api-0/logs/stream":
			if r.URL.Query().Get("tailLines") != "40" {
				t.Fatalf("tailLines = %q", r.URL.Query().Get("tailLines"))
			}
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: log\ndata: server started\n\n")
			fmt.Fprint(w, "event: log\ndata: error opening file\n\n")
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewObservabilityClient(server.URL)
	pods, err := client.ListPods(context.Background(), "token", "workspace-dev", "primary")
	if err != nil {
		t.Fatal(err)
	}
	if len(pods.Pods) != 1 || pods.Pods[0].Name != "api-0" {
		t.Fatalf("pods = %#v", pods)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lines, err := client.FetchPodLogs(ctx, "token", "workspace-dev", "api-0", "primary", 40)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 || lines[0].Level != "success" || lines[1].Level != "error" {
		t.Fatalf("lines = %#v", lines)
	}
	if strings.Contains(lines[0].Message, "\x1b") {
		t.Fatalf("line was not sanitized: %q", lines[0].Message)
	}
}
