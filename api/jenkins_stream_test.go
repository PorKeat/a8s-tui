package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchJenkinsLogStreamChunk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/jenkins/logs/stream" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("job") != "deploy-monolith" || r.URL.Query().Get("queueItem") != "42" {
			t.Fatalf("query = %q", r.URL.RawQuery)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "event: queued\ndata: {\"queueItemId\":42,\"message\":\"Waiting for executor\"}\n\n")
		_, _ = fmt.Fprint(w, "event: open\ndata: {\"job\":\"deploy-monolith\",\"build\":9}\n\n")
		_, _ = fmt.Fprint(w, "event: log\ndata: {\"chunk\":\"Building image\\nPushing image\",\"build\":9}\n\n")
		_, _ = fmt.Fprint(w, "event: done\ndata: {\"build\":9,\"result\":\"SUCCESS\"}\n\n")
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	chunk, err := client.FetchJenkinsLogStreamChunk(context.Background(), "access", "deploy-monolith", 42)
	if err != nil {
		t.Fatal(err)
	}
	if !chunk.Completed || chunk.Status != "DEPLOYED" || chunk.BuildNumber != 9 {
		t.Fatalf("chunk = %#v", chunk)
	}
	if len(chunk.Lines) != 5 || chunk.Lines[2] != "Building image" || chunk.Lines[4] != "Jenkins build completed: SUCCESS" {
		t.Fatalf("lines = %#v", chunk.Lines)
	}
}

func TestParseJenkinsStreamError(t *testing.T) {
	var chunk JenkinsLogStreamChunk
	parseJenkinsStreamEvent(&chunk, "error", `{"message":"Build failed","detail":"Docker push denied"}`)
	if !chunk.Completed || chunk.Status != "FAILED" || chunk.Message != "Docker push denied" {
		t.Fatalf("chunk = %#v", chunk)
	}
}
