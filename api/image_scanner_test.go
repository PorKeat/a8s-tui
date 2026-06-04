package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestImageScannerListImages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/image-scanner/images" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id":                 "image-1",
			"name":               "api",
			"tag":                "v1",
			"repository":         "harbor.local/team/api",
			"architecture":       "amd64",
			"distro":             "alpine",
			"sizeLabel":          "42 MB",
			"layerCount":         8,
			"vulnerabilityCount": 3,
			"environment":        "production",
			"updatedLabel":       "Updated 1m ago",
		}})
	}))
	defer server.Close()

	client := NewImageScannerClient(server.URL)
	client.client = server.Client()
	images, err := client.ListImages(context.Background(), "access")
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 || images[0].ID != "image-1" || images[0].VulnerabilityCount != 3 {
		t.Fatalf("images = %#v", images)
	}
}

func TestImageScannerCreateAndGetScan(t *testing.T) {
	var sawCreate bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/api/v1/image-scanner/scans":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %q", r.Method)
			}
			sawCreate = true
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload["sourceKind"] != "harbor" || payload["imageId"] != "image-1" || payload["forceRescan"] != true {
				t.Fatalf("payload = %#v", payload)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(scanPayload("scan-1", "RUNNING"))
		case "/api/v1/image-scanner/scans/scan-1":
			_ = json.NewEncoder(w).Encode(scanPayload("scan-1", "COMPLETED"))
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewImageScannerClient(server.URL)
	client.client = server.Client()
	scan, err := client.CreateScan(context.Background(), "access", CreateImageScanInput{
		SourceKind:  "harbor",
		ImageID:     "image-1",
		ForceRescan: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !sawCreate || scan.ID != "scan-1" || scan.Status != "RUNNING" {
		t.Fatalf("scan = %#v", scan)
	}

	scan, err = client.GetScan(context.Background(), "access", "scan-1")
	if err != nil {
		t.Fatal(err)
	}
	if scan.Status != "COMPLETED" || len(scan.Vulnerabilities) != 1 || scan.Vulnerabilities[0].Severity != "HIGH" {
		t.Fatalf("scan = %#v", scan)
	}
}

func TestImageScannerGetScanReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/image-scanner/scans/scan-1/report" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" || r.Header.Get("Authorization") != "Bearer access" {
			t.Fatalf("headers = %#v", r.Header)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"SchemaVersion":2,"ArtifactName":"api:v1","Results":[{"Target":"api","Vulnerabilities":[{"VulnerabilityID":"CVE-2026-0001"}]}]}`))
	}))
	defer server.Close()

	client := NewImageScannerClient(server.URL)
	client.client = server.Client()
	report, err := client.GetScanReport(context.Background(), "access", "scan-1")
	if err != nil {
		t.Fatal(err)
	}
	if report == "" || !strings.Contains(report, "CVE-2026-0001") {
		t.Fatalf("report = %q", report)
	}
}

func scanPayload(id, status string) map[string]any {
	return map[string]any{
		"id":              id,
		"sourceKind":      "harbor",
		"imageName":       "api",
		"imageTag":        "v1",
		"fullReference":   "harbor.local/team/api:v1",
		"status":          status,
		"progress":        100,
		"scannerName":     "Trivy",
		"scannerVersion":  "0.50.0",
		"durationSeconds": 4.2,
		"vulnerabilities": []map[string]any{{
			"id":             "finding-1",
			"cveId":          "CVE-2026-0001",
			"packageName":    "openssl",
			"packageVersion": "3.0.0",
			"severity":       "high",
			"fixedIn":        "3.0.1",
			"cvssScore":      8.1,
			"fixable":        true,
		}},
	}
}
