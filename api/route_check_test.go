package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStartAndGetRouteCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/api/v1/projects/live/project-1/route-check":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %q", r.Method)
			}
			var input RouteCheckInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				t.Fatal(err)
			}
			if input.MaxRoutes != 50 || input.MaxDepth != 2 || input.TimeoutMS != 10000 {
				t.Fatalf("input = %#v", input)
			}
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(routeCheckPayload("RUNNING"))
		case "/api/v1/projects/live/project-1/route-check/job-1":
			_ = json.NewEncoder(w).Encode(routeCheckPayload("COMPLETED"))
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewProjectClient(server.URL)
	client.client = server.Client()
	job, err := client.StartRouteCheck(context.Background(), "access", "project-1", RouteCheckInput{
		MaxRoutes: 50,
		MaxDepth:  2,
		TimeoutMS: 10000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if job.JobID != "job-1" || job.Status != "RUNNING" || RouteCheckTerminal(job.Status) {
		t.Fatalf("job = %#v", job)
	}

	job, err = client.GetRouteCheck(context.Background(), "access", "project-1", "job-1")
	if err != nil {
		t.Fatal(err)
	}
	if !RouteCheckTerminal(job.Status) || job.Summary.Passed != 1 || len(job.Routes) != 1 || !job.Routes[0].BrowserOK {
		t.Fatalf("job = %#v", job)
	}
}

func routeCheckPayload(status string) map[string]any {
	return map[string]any{
		"jobId":     "job-1",
		"projectId": "project-1",
		"status":    status,
		"baseUrl":   "https://example.com",
		"summary": map[string]any{
			"discovered": 1,
			"passed":     1,
			"failed":     0,
			"warnings":   0,
		},
		"routes": []map[string]any{{
			"path":       "/",
			"url":        "https://example.com/",
			"finalUrl":   "https://example.com/",
			"httpStatus": 200,
			"browserOk":  true,
			"durationMs": 25,
			"warning":    false,
		}},
	}
}
