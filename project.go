package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type liveProject struct {
	Kind                  string
	ID                    string
	Name                  string
	Status                string
	CreatedAt             string
	UpdatedAt             string
	Engine                string
	DeploymentMode        string
	Version               string
	ProjectName           string
	DatabaseName          string
	Namespace             string
	DeploymentCount       int
	DatabaseDeploymentIDs []string
	ArchitectureType      string
	RepoURL               string
	RepoFullName          string
	RepoProvider          string
	Branch                string
	Framework             string
	AppPort               int
	DeployURL             string
	CurrentReleaseID      string
	AutoDeployEnabled     bool
	AutoDeployTrigger     string
	ReleaseTagPattern     string
	Deletable             bool
	ClusterReleaseName    string
	TargetClusterName     string
	HealthStatus          string
}

type createDatabaseDeploymentInput struct {
	ProjectName    string `json:"projectName"`
	Engine         string `json:"engine"`
	DeploymentMode string `json:"deploymentMode"`
	DatabaseName   string `json:"databaseName"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Version        string `json:"version"`
	SizeProfile    string `json:"sizeProfile"`
}

type databaseDeploymentRecord struct {
	ID            string
	ReleaseName   string
	Namespace     string
	Engine        string
	ProjectName   string
	DatabaseName  string
	Username      string
	Version       string
	StorageSize   string
	Status        string
	StatusMessage string
	StatusLog     string
}

type projectClient struct {
	baseURL string
	client  *http.Client
}

func newProjectClient(baseURL string) projectClient {
	return projectClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (c projectClient) fetchLiveProjects(ctx context.Context, token string) ([]liveProject, error) {
	backendUserID, err := c.fetchBackendUserID(ctx, token)
	if err != nil {
		return nil, err
	}
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/projects/live")
	if err != nil {
		return nil, err
	}
	projectsURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	query := projectsURL.Query()
	query.Set("userId", backendUserID)
	projectsURL.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.URL = projectsURL
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var raw []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode live projects response: %w", err)
	}

	projects := make([]liveProject, 0, len(raw))
	for _, item := range raw {
		project, ok := normalizeLiveProject(item)
		if ok {
			projects = append(projects, project)
		}
	}
	return projects, nil
}

func (c projectClient) createDatabaseDeployment(
	ctx context.Context,
	token string,
	input createDatabaseDeploymentInput,
) (databaseDeploymentRecord, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/database-deployments")
	if err != nil {
		return databaseDeploymentRecord{}, err
	}
	body, err := json.Marshal(input)
	if err != nil {
		return databaseDeploymentRecord{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return databaseDeploymentRecord{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return databaseDeploymentRecord{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return databaseDeploymentRecord{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return databaseDeploymentRecord{}, fmt.Errorf("decode database deployment response: %w", err)
	}
	return normalizeDatabaseDeployment(payload), nil
}

func (c projectClient) fetchDatabaseDeployment(ctx context.Context, token, deploymentID string) (databaseDeploymentRecord, error) {
	if strings.TrimSpace(deploymentID) == "" {
		return databaseDeploymentRecord{}, errors.New("database deployment id is required")
	}
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/database-deployments", deploymentID)
	if err != nil {
		return databaseDeploymentRecord{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return databaseDeploymentRecord{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return databaseDeploymentRecord{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return databaseDeploymentRecord{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return databaseDeploymentRecord{}, fmt.Errorf("decode database deployment detail response: %w", err)
	}
	return normalizeDatabaseDeployment(payload), nil
}

func (c projectClient) fetchBackendUserID(ctx context.Context, token string) (string, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/profile/me")
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode profile response: %w", err)
	}
	userID := readString(payload["userId"])
	if userID == "" {
		return "", errors.New("backend profile response did not include userId")
	}
	return userID, nil
}

type httpStatusError struct {
	status  int
	message string
}

func (e httpStatusError) Error() string {
	if e.message != "" {
		return e.message
	}
	return fmt.Sprintf("backend request failed with HTTP %d", e.status)
}

func isUnauthorized(err error) bool {
	var statusErr httpStatusError
	return errors.As(err, &statusErr) && statusErr.status == http.StatusUnauthorized
}

func decodeErrorMessage(res *http.Response) string {
	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return ""
	}
	for _, key := range []string{"details", "message", "error", "detail"} {
		if value := readString(payload[key]); value != "" {
			return value
		}
	}
	return ""
}

func normalizeLiveProject(payload map[string]any) (liveProject, bool) {
	kind := strings.ToLower(readString(payload["kind"]))
	id := readString(payload["id"])
	name := readString(payload["name"])
	status := readString(payload["status"])
	createdAt := readTimestamp(payload["createdAt"])
	updatedAt := readTimestamp(payload["updatedAt"])
	if !validKind(kind) || id == "" || name == "" || status == "" {
		return liveProject{}, false
	}
	if createdAt == "" {
		createdAt = updatedAt
	}
	if updatedAt == "" {
		updatedAt = createdAt
	}

	return liveProject{
		Kind:                  kind,
		ID:                    id,
		Name:                  name,
		Status:                status,
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
		Engine:                readString(payload["engine"]),
		DeploymentMode:        readString(payload["deploymentMode"]),
		Version:               readString(payload["version"]),
		ProjectName:           readString(payload["projectName"]),
		DatabaseName:          readString(payload["databaseName"]),
		Namespace:             readString(payload["namespace"]),
		DeploymentCount:       readInt(payload["deploymentCount"]),
		DatabaseDeploymentIDs: readStringSlice(payload["databaseDeploymentIds"]),
		ArchitectureType:      readString(payload["architectureType"]),
		RepoURL:               readString(payload["repoUrl"]),
		RepoFullName:          readString(payload["repoFullName"]),
		RepoProvider:          readString(payload["repoProvider"]),
		Branch:                readString(payload["branch"]),
		Framework:             readString(payload["framework"]),
		AppPort:               readInt(payload["appPort"]),
		DeployURL:             readString(payload["deployUrl"]),
		CurrentReleaseID:      readString(payload["currentReleaseId"]),
		AutoDeployEnabled:     readBool(payload["autoDeployEnabled"]),
		AutoDeployTrigger:     readString(payload["autoDeployTrigger"]),
		ReleaseTagPattern:     readString(payload["releaseTagPattern"]),
		Deletable:             readBool(payload["deletable"]),
		ClusterReleaseName:    readString(payload["clusterReleaseName"]),
		TargetClusterName:     readString(payload["targetClusterName"]),
		HealthStatus:          readString(payload["healthStatus"]),
	}, true
}

func normalizeDatabaseDeployment(payload map[string]any) databaseDeploymentRecord {
	return databaseDeploymentRecord{
		ID:            readString(payload["id"]),
		ReleaseName:   readString(payload["releaseName"]),
		Namespace:     readString(payload["namespace"]),
		Engine:        readString(payload["engine"]),
		ProjectName:   readString(payload["projectName"]),
		DatabaseName:  readString(payload["databaseName"]),
		Username:      readString(payload["username"]),
		Version:       readString(payload["version"]),
		StorageSize:   readString(payload["storageSize"]),
		Status:        readString(payload["status"]),
		StatusMessage: readString(payload["statusMessage"]),
		StatusLog:     readString(payload["statusLog"]),
	}
}

func parseDeploymentLogLines(statusLog string) []string {
	parts := strings.Split(strings.ReplaceAll(statusLog, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		line := strings.TrimSpace(part)
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true
		lines = append(lines, line)
	}
	return lines
}

func databaseDeploymentTerminal(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "DEPLOYED", "READY", "RUNNING", "SUCCESS", "SUCCEEDED", "FAILED", "ERROR", "UNHEALTHY", "CANCELLED", "CANCELED":
		return true
	default:
		return false
	}
}

func databaseDeploymentFailed(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "FAILED", "ERROR", "UNHEALTHY", "CANCELLED", "CANCELED":
		return true
	default:
		return false
	}
}

func validKind(kind string) bool {
	switch kind {
	case "database", "monolith", "microservices", "dbcluster":
		return true
	default:
		return false
	}
}

func readString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return ""
	}
}

func readTimestamp(value any) string {
	if value == nil {
		return ""
	}
	if text := readString(value); text != "" {
		return text
	}
	if parts, ok := value.([]any); ok && len(parts) >= 3 {
		numbers := make([]int, 7)
		for i := range numbers {
			if i < len(parts) {
				numbers[i] = readInt(parts[i])
			}
		}
		if numbers[0] > 0 && numbers[1] > 0 && numbers[2] > 0 {
			return time.Date(
				numbers[0],
				time.Month(numbers[1]),
				numbers[2],
				numbers[3],
				numbers[4],
				numbers[5],
				numbers[6],
				time.UTC,
			).Format(time.RFC3339)
		}
	}
	return ""
}

func readInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, _ := typed.Int64()
		return int(parsed)
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(typed))
		return parsed
	default:
		return 0
	}
}

func readBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "1", "true", "yes", "on":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func readStringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if text := readString(item); text != "" {
			out = append(out, text)
		}
	}
	return out
}

func projectMatchesFilter(project liveProject, filter string) bool {
	filter = strings.ToLower(strings.TrimSpace(filter))
	if filter == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		project.Name,
		project.Kind,
		project.Status,
		project.Namespace,
		project.Engine,
		project.RepoFullName,
		project.RepoURL,
		project.Branch,
		project.Framework,
		project.TargetClusterName,
		project.HealthStatus,
	}, " "))
	return strings.Contains(haystack, filter)
}

func filteredProjects(projects []liveProject, filter string) []liveProject {
	out := make([]liveProject, 0, len(projects))
	for _, project := range projects {
		if projectMatchesFilter(project, filter) {
			out = append(out, project)
		}
	}
	return out
}
