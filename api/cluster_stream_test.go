package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchClusterDeploymentStreamChunk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/kubernetes/namespaces/workspace-a/releases/orders-db/deployment-stream" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("targetClusterName") != "primary" {
			t.Fatalf("query = %q", r.URL.RawQuery)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("accept = %q", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "event: console-log\ndata: {\"content\":\"Waiting for release pods.\"}\n\n")
		_, _ = fmt.Fprint(w, "event: deployment-event\ndata: {\"reason\":\"Scheduled\",\"message\":\"Assigned pod\"}\n\n")
		_, _ = fmt.Fprint(w, "event: console-log\ndata: {\"content\":\"Deployment stream completed. All release pods are Running and Ready.\"}\n\n")
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	chunk, err := client.FetchClusterDeploymentStreamChunk(ctx, "access", "workspace-a", "orders-db", "primary")
	if err != nil {
		t.Fatal(err)
	}
	if !chunk.Completed {
		t.Fatalf("expected completed chunk: %#v", chunk)
	}
	want := []string{
		"Waiting for release pods.",
		"[Scheduled] Assigned pod",
		"Deployment stream completed. All release pods are Running and Ready.",
	}
	if len(chunk.Lines) != len(want) {
		t.Fatalf("lines = %#v", chunk.Lines)
	}
	for index := range want {
		if chunk.Lines[index] != want[index] {
			t.Fatalf("lines = %#v", chunk.Lines)
		}
	}
}

func TestClusterDeploymentStreamLineFallsBackToPlainText(t *testing.T) {
	if got := clusterDeploymentStreamLine("console-log", "plain progress"); got != "plain progress" {
		t.Fatalf("line = %q", got)
	}
}

func TestClusterDeploymentStreamCompletesFromReadinessSummary(t *testing.T) {
	if !clusterDeploymentStreamCompleted("Release pod readiness: 3/3 ready, 3/3 running. {orders=Running ready}") {
		t.Fatal("expected full readiness to complete the stream")
	}
	if clusterDeploymentStreamCompleted("Release pod readiness: 2/3 ready, 3/3 running.") {
		t.Fatal("partial readiness should not complete the stream")
	}
}
