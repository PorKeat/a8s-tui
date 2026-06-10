package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RouteCheckInput struct {
	MaxRoutes int `json:"maxRoutes,omitempty"`
	MaxDepth  int `json:"maxDepth,omitempty"`
	TimeoutMS int `json:"timeoutMs,omitempty"`
}

type RouteCheckJob struct {
	JobID        string
	ProjectID    string
	Status       string
	BaseURL      string
	Summary      RouteCheckSummary
	Routes       []RouteCheckResult
	ErrorMessage string
	CreatedAt    string
	UpdatedAt    string
}

type RouteCheckSummary struct {
	Discovered int
	Passed     int
	Failed     int
	Warnings   int
}

type RouteCheckResult struct {
	Path          string
	URL           string
	FinalURL      string
	HTTPStatus    int
	BrowserOK     bool
	DurationMS    int
	ConsoleErrors []string
	Error         string
	Warning       bool
}

func (c ProjectClient) StartRouteCheck(ctx context.Context, token, projectID string, input RouteCheckInput) (RouteCheckJob, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/projects/live", strings.TrimSpace(projectID), "route-check")
	if err != nil {
		return RouteCheckJob{}, err
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return RouteCheckJob{}, err
	}
	return c.requestRouteCheck(ctx, token, http.MethodPost, endpoint, strings.NewReader(string(payload)))
}

func (c ProjectClient) GetRouteCheck(ctx context.Context, token, projectID, jobID string) (RouteCheckJob, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/projects/live", strings.TrimSpace(projectID), "route-check", strings.TrimSpace(jobID))
	if err != nil {
		return RouteCheckJob{}, err
	}
	return c.requestRouteCheck(ctx, token, http.MethodGet, endpoint, nil)
}

func (c ProjectClient) requestRouteCheck(ctx context.Context, token, method, endpoint string, body io.Reader) (RouteCheckJob, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return RouteCheckJob{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.client.Do(req)
	if err != nil {
		return RouteCheckJob{}, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return RouteCheckJob{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return RouteCheckJob{}, fmt.Errorf("decode route check response: %w", err)
	}
	return normalizeRouteCheckJob(payload), nil
}

func normalizeRouteCheckJob(payload map[string]any) RouteCheckJob {
	job := RouteCheckJob{
		JobID:        readString(payload["jobId"]),
		ProjectID:    readString(payload["projectId"]),
		Status:       readString(payload["status"]),
		BaseURL:      readString(payload["baseUrl"]),
		ErrorMessage: readString(payload["errorMessage"]),
		CreatedAt:    readTimestamp(payload["createdAt"]),
		UpdatedAt:    readTimestamp(payload["updatedAt"]),
	}
	if summary, ok := payload["summary"].(map[string]any); ok {
		job.Summary = RouteCheckSummary{
			Discovered: readInt(summary["discovered"]),
			Passed:     readInt(summary["passed"]),
			Failed:     readInt(summary["failed"]),
			Warnings:   readInt(summary["warnings"]),
		}
	}
	if routes, ok := payload["routes"].([]any); ok {
		for _, item := range routes {
			route, ok := item.(map[string]any)
			if !ok {
				continue
			}
			result := RouteCheckResult{
				Path:       readString(route["path"]),
				URL:        readString(route["url"]),
				FinalURL:   readString(route["finalUrl"]),
				HTTPStatus: readInt(route["httpStatus"]),
				BrowserOK:  readBool(route["browserOk"]),
				DurationMS: readInt(route["durationMs"]),
				Error:      readString(route["error"]),
				Warning:    readBool(route["warning"]),
			}
			if errors, ok := route["consoleErrors"].([]any); ok {
				for _, item := range errors {
					if message := readString(item); message != "" {
						result.ConsoleErrors = append(result.ConsoleErrors, message)
					}
				}
			}
			job.Routes = append(job.Routes, result)
		}
	}
	return job
}

func RouteCheckTerminal(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "COMPLETED", "FAILED":
		return true
	default:
		return false
	}
}
