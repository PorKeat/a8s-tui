package api

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

type LiveProject struct {
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
	DatabaseUsername      string
	Namespace             string
	DeploymentCount       int
	DatabaseDeploymentIDs []string
	ServiceName           string
	ConnectionServiceName string
	ServiceHost           string
	ServicePort           int
	RequireSSL            bool
	ConnectionTLSEnabled  bool
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

type CreateDatabaseDeploymentInput struct {
	ProjectName    string `json:"projectName"`
	Engine         string `json:"engine"`
	DeploymentMode string `json:"deploymentMode"`
	DatabaseName   string `json:"databaseName"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Version        string `json:"version"`
	SizeProfile    string `json:"sizeProfile"`
}

type CreateClusterDeploymentInput struct {
	Namespace       string
	ProjectName     string
	Engine          string
	DatabaseName    string
	Username        string
	Password        string
	Version         string
	SizeProfile     string
	TargetCluster   string
	StorageSize     string
	PublicHostnames []string
}

type CreateMonolithicDeploymentInput struct {
	ProjectName       string `json:"projectName"`
	RepoURL           string `json:"repoUrl"`
	RepoFullName      string `json:"repoFullName,omitempty"`
	Branch            string `json:"branch,omitempty"`
	AppPort           int    `json:"appPort,omitempty"`
	ArchitectureType  string `json:"architectureType"`
	AutoDeployEnabled bool   `json:"autoDeployEnabled,omitempty"`
}

type MonolithicDeploymentRecord struct {
	ProjectID         string
	Name              string
	Status            string
	RepoProvider      string
	DeployURL         string
	QueueURL          string
	QueueItemID       int
	JenkinsJobName    string
	AutoDeployEnabled bool
}

type DatabaseDeploymentRecord struct {
	ID                    string
	ReleaseName           string
	Namespace             string
	Engine                string
	DeploymentMode        string
	ProjectName           string
	DatabaseName          string
	Username              string
	Version               string
	StorageSize           string
	ServiceName           string
	ConnectionServiceName string
	ServiceHost           string
	ServicePort           int
	RequireSSL            bool
	ConnectionTLSEnabled  bool
	Status                string
	StatusMessage         string
	StatusLog             string
}

type ClusterDeploymentRecord struct {
	ClusterID         string
	ReleaseName       string
	Name              string
	Namespace         string
	TargetClusterName string
	Engine            string
	Status            string
	StatusMessage     string
	ServiceHost       string
	ServicePort       int
	TLSEnabled        bool
	Command           []string
	Stdout            string
	Stderr            string
	Successful        bool
	ExitCode          int
	StartedAt         string
	FinishedAt        string
}

type ProjectClient struct {
	baseURL string
	client  *http.Client
}

func NewProjectClient(baseURL string) ProjectClient {
	return ProjectClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (c ProjectClient) FetchLiveProjects(ctx context.Context, token string) ([]LiveProject, string, error) {
	backendUserID, userName, err := c.fetchBackendUserID(ctx, token)
	if err != nil {
		return nil, "", err
	}
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/projects/live")
	if err != nil {
		return nil, "", err
	}
	projectsURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, "", err
	}
	query := projectsURL.Query()
	query.Set("userId", backendUserID)
	projectsURL.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}
	req.URL = projectsURL
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, "", httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var raw []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, "", fmt.Errorf("decode live projects response: %w", err)
	}

	projects := make([]LiveProject, 0, len(raw))
	for _, item := range raw {
		project, ok := normalizeLiveProject(item)
		if ok {
			projects = append(projects, project)
		}
	}
	projects = c.hydrateLiveProjectConnectionDetails(ctx, token, projects)
	return projects, userName, nil
}

func (c ProjectClient) hydrateLiveProjectConnectionDetails(ctx context.Context, token string, projects []LiveProject) []LiveProject {
	for index := range projects {
		project := projects[index]
		if project.Kind != "database" || len(project.DatabaseDeploymentIDs) == 0 {
			continue
		}
		deployment, err := c.FetchDatabaseDeployment(ctx, token, project.DatabaseDeploymentIDs[0])
		if err != nil {
			continue
		}
		projects[index] = mergeDatabaseDeploymentDetail(project, deployment)
	}
	return projects
}

func mergeDatabaseDeploymentDetail(project LiveProject, deployment DatabaseDeploymentRecord) LiveProject {
	project.Engine = firstNonEmpty(deployment.Engine, project.Engine)
	project.DeploymentMode = firstNonEmpty(deployment.DeploymentMode, project.DeploymentMode)
	project.Version = firstNonEmpty(deployment.Version, project.Version)
	project.ProjectName = firstNonEmpty(deployment.ProjectName, project.ProjectName)
	project.DatabaseName = firstNonEmpty(deployment.DatabaseName, project.DatabaseName)
	project.DatabaseUsername = firstNonEmpty(deployment.Username, project.DatabaseUsername)
	project.Namespace = firstNonEmpty(deployment.Namespace, project.Namespace)
	project.ServiceName = firstNonEmpty(deployment.ServiceName, project.ServiceName)
	project.ConnectionServiceName = firstNonEmpty(deployment.ConnectionServiceName, project.ConnectionServiceName)
	project.ServiceHost = firstNonEmpty(deployment.ServiceHost, project.ServiceHost)
	if deployment.ServicePort > 0 {
		project.ServicePort = deployment.ServicePort
	}
	project.RequireSSL = deployment.RequireSSL
	project.ConnectionTLSEnabled = deployment.ConnectionTLSEnabled
	return project
}

func (c ProjectClient) CreateDatabaseDeployment(
	ctx context.Context,
	token string,
	input CreateDatabaseDeploymentInput,
) (DatabaseDeploymentRecord, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/database-deployments")
	if err != nil {
		return DatabaseDeploymentRecord{}, err
	}
	body, err := json.Marshal(input)
	if err != nil {
		return DatabaseDeploymentRecord{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return DatabaseDeploymentRecord{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return DatabaseDeploymentRecord{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return DatabaseDeploymentRecord{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return DatabaseDeploymentRecord{}, fmt.Errorf("decode database deployment response: %w", err)
	}
	return normalizeDatabaseDeployment(payload), nil
}

func (c ProjectClient) CreateClusterDeployment(
	ctx context.Context,
	token string,
	input CreateClusterDeploymentInput,
) (ClusterDeploymentRecord, error) {
	namespace := firstNonEmpty(input.Namespace, "default")
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/cluster/namespaces", namespace, "cluster-deployments")
	if err != nil {
		return ClusterDeploymentRecord{}, err
	}
	body := buildClusterDeploymentPayload(input, namespace)
	payloadBytes, err := json.Marshal(body)
	if err != nil {
		return ClusterDeploymentRecord{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return ClusterDeploymentRecord{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return ClusterDeploymentRecord{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return ClusterDeploymentRecord{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return ClusterDeploymentRecord{}, fmt.Errorf("decode cluster deployment response: %w", err)
	}
	return normalizeClusterDeployment(payload), nil
}

func (c ProjectClient) CreateMonolithicDeployment(
	ctx context.Context,
	token string,
	input CreateMonolithicDeploymentInput,
) (MonolithicDeploymentRecord, error) {
	backendUserID, _, err := c.fetchBackendUserID(ctx, token)
	if err != nil {
		return MonolithicDeploymentRecord{}, err
	}
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/projects")
	if err != nil {
		return MonolithicDeploymentRecord{}, err
	}
	body := map[string]any{
		"userId":           backendUserID,
		"projectName":      strings.TrimSpace(input.ProjectName),
		"repoUrl":          strings.TrimSpace(input.RepoURL),
		"repoFullName":     strings.TrimSpace(input.RepoFullName),
		"branch":           strings.TrimSpace(input.Branch),
		"architectureType": firstNonEmpty(strings.TrimSpace(input.ArchitectureType), "monolithic"),
	}
	if input.AppPort > 0 {
		body["appPort"] = input.AppPort
	}
	if input.AutoDeployEnabled {
		body["autoDeployEnabled"] = true
	}
	payloadBytes, err := json.Marshal(body)
	if err != nil {
		return MonolithicDeploymentRecord{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return MonolithicDeploymentRecord{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return MonolithicDeploymentRecord{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return MonolithicDeploymentRecord{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return MonolithicDeploymentRecord{}, fmt.Errorf("decode monolithic deployment response: %w", err)
	}
	return normalizeMonolithicDeployment(payload), nil
}

func (c ProjectClient) DeleteLiveProject(ctx context.Context, token string, project LiveProject) error {
	switch strings.ToLower(strings.TrimSpace(project.Kind)) {
	case "database":
		deploymentIDs := project.DatabaseDeploymentIDs
		if len(deploymentIDs) == 0 && strings.TrimSpace(project.ID) != "" {
			deploymentIDs = []string{project.ID}
		}
		if len(deploymentIDs) == 0 {
			return errors.New("database project does not include a deployment id")
		}
		for _, deploymentID := range deploymentIDs {
			endpoint, err := url.JoinPath(c.baseURL, "/api/v1/database-deployments", deploymentID)
			if err != nil {
				return err
			}
			if err := c.delete(ctx, token, endpoint); err != nil {
				return err
			}
		}
		return nil
	case "monolith":
		endpoint, err := url.JoinPath(c.baseURL, "/api/v1/projects", project.ID)
		if err != nil {
			return err
		}
		return c.delete(ctx, token, endpoint)
	case "microservices":
		endpoint, err := url.JoinPath(c.baseURL, "/api/v1/projects/microservices", project.ID)
		if err != nil {
			return err
		}
		return c.delete(ctx, token, endpoint)
	case "dbcluster":
		namespace := firstNonEmpty(project.Namespace, "default")
		endpoint, err := url.JoinPath(c.baseURL, "/api/v1/cluster/namespaces", namespace, "clusters", project.ID)
		if err != nil {
			return err
		}
		deleteURL, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		query := deleteURL.Query()
		query.Set("deleteData", "true")
		deleteURL.RawQuery = query.Encode()
		return c.delete(ctx, token, deleteURL.String())
	default:
		return fmt.Errorf("delete is not available for %s projects", firstNonEmpty(project.Kind, "unknown"))
	}
}

func ProjectKindSupportsDelete(kind string) bool {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "database", "monolith", "microservices", "dbcluster":
		return true
	default:
		return false
	}
}

func (c ProjectClient) delete(ctx context.Context, token, endpoint string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}
	return nil
}

func (c ProjectClient) FetchDatabaseDeployment(ctx context.Context, token, deploymentID string) (DatabaseDeploymentRecord, error) {
	if strings.TrimSpace(deploymentID) == "" {
		return DatabaseDeploymentRecord{}, errors.New("database deployment id is required")
	}
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/database-deployments", deploymentID)
	if err != nil {
		return DatabaseDeploymentRecord{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return DatabaseDeploymentRecord{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return DatabaseDeploymentRecord{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return DatabaseDeploymentRecord{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return DatabaseDeploymentRecord{}, fmt.Errorf("decode database deployment detail response: %w", err)
	}
	return normalizeDatabaseDeployment(payload), nil
}

func (c ProjectClient) fetchBackendUserID(ctx context.Context, token string) (string, string, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/profile/me")
	if err != nil {
		return "", "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", "", httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", "", fmt.Errorf("decode profile response: %w", err)
	}
	userID := readString(payload["userId"])
	if userID == "" {
		return "", "", errors.New("backend profile response did not include userId")
	}

	var userName string
	if personal, ok := payload["personal"].(map[string]any); ok {
		userName = readString(personal["displayName"])
		if userName == "" {
			firstName := readString(personal["firstName"])
			lastName := readString(personal["lastName"])
			userName = strings.TrimSpace(firstName + " " + lastName)
		}
	}

	return userID, userName, nil
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

func IsUnauthorized(err error) bool {
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

func normalizeLiveProject(payload map[string]any) (LiveProject, bool) {
	kind := strings.ToLower(readString(payload["kind"]))
	id := readString(payload["id"])
	name := readString(payload["name"])
	status := readString(payload["status"])
	createdAt := readTimestamp(payload["createdAt"])
	updatedAt := readTimestamp(payload["updatedAt"])
	if !validKind(kind) || id == "" || name == "" || status == "" {
		return LiveProject{}, false
	}
	if createdAt == "" {
		createdAt = updatedAt
	}
	if updatedAt == "" {
		updatedAt = createdAt
	}

	return LiveProject{
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
		DatabaseUsername:      readString(payload["username"]),
		Namespace:             readString(payload["namespace"]),
		DeploymentCount:       readInt(payload["deploymentCount"]),
		DatabaseDeploymentIDs: readStringSlice(payload["databaseDeploymentIds"]),
		ServiceName:           readString(payload["serviceName"]),
		ConnectionServiceName: readString(payload["connectionServiceName"]),
		ServiceHost:           readString(payload["serviceHost"]),
		ServicePort:           readInt(payload["servicePort"]),
		RequireSSL:            readBool(payload["requireSsl"]),
		ConnectionTLSEnabled:  readBool(payload["connectionTlsEnabled"]),
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

func normalizeDatabaseDeployment(payload map[string]any) DatabaseDeploymentRecord {
	return DatabaseDeploymentRecord{
		ID:                    readString(payload["id"]),
		ReleaseName:           readString(payload["releaseName"]),
		Namespace:             readString(payload["namespace"]),
		Engine:                readString(payload["engine"]),
		DeploymentMode:        readString(payload["deploymentMode"]),
		ProjectName:           readString(payload["projectName"]),
		DatabaseName:          readString(payload["databaseName"]),
		Username:              readString(payload["username"]),
		Version:               readString(payload["version"]),
		StorageSize:           readString(payload["storageSize"]),
		ServiceName:           readString(payload["serviceName"]),
		ConnectionServiceName: readString(payload["connectionServiceName"]),
		ServiceHost:           readString(payload["serviceHost"]),
		ServicePort:           readInt(payload["servicePort"]),
		RequireSSL:            readBool(payload["requireSsl"]),
		ConnectionTLSEnabled:  readBool(payload["connectionTlsEnabled"]),
		Status:                readString(payload["status"]),
		StatusMessage:         readString(payload["statusMessage"]),
		StatusLog:             readString(payload["statusLog"]),
	}
}

func normalizeClusterDeployment(payload map[string]any) ClusterDeploymentRecord {
	return ClusterDeploymentRecord{
		ClusterID:         firstNonEmpty(readString(payload["clusterId"]), readString(payload["id"])),
		ReleaseName:       readString(payload["releaseName"]),
		Name:              readString(payload["name"]),
		Namespace:         readString(payload["namespace"]),
		TargetClusterName: readString(payload["targetClusterName"]),
		Engine:            readString(payload["engine"]),
		Status:            readString(payload["status"]),
		StatusMessage:     readString(payload["statusMessage"]),
		ServiceHost:       readString(payload["serviceHost"]),
		ServicePort:       readInt(payload["servicePort"]),
		TLSEnabled:        readBool(payload["tlsEnabled"]),
		Command:           readStringSlice(payload["command"]),
		Stdout:            readString(payload["stdout"]),
		Stderr:            readString(payload["stderr"]),
		Successful:        readBool(payload["successful"]),
		ExitCode:          readInt(payload["exitCode"]),
		StartedAt:         readTimestamp(payload["startedAt"]),
		FinishedAt:        readTimestamp(payload["finishedAt"]),
	}
}

type clusterResourceProfile struct {
	storageSize string
	cpuRequest  string
	cpuLimit    string
	memRequest  string
	memLimit    string
}

func buildClusterDeploymentPayload(input CreateClusterDeploymentInput, namespace string) map[string]any {
	engine := normalizeClusterEngine(input.Engine)
	sizeProfile := normalizeClusterSizeProfile(input.SizeProfile)
	resource := clusterResource(engine, sizeProfile)
	storageSize := firstNonEmpty(input.StorageSize, resource.storageSize)
	projectName := strings.TrimSpace(input.ProjectName)
	databaseName := firstNonEmpty(input.DatabaseName, defaultClusterDatabase(engine))
	username := firstNonEmpty(input.Username, defaultClusterUsername(engine))
	publicHostnames := normalizeClusterHostnames(input.PublicHostnames)

	database := map[string]any{
		"engine":            strings.ToUpper(engine),
		"enabled":           true,
		"instances":         3,
		"storageSize":       storageSize,
		"publicHostnames":   publicHostnames,
		"monitoringEnabled": true,
		"notes":             "",
		"resource": map[string]any{
			"cpuRequest":      resource.cpuRequest,
			"cpuLimit":        resource.cpuLimit,
			"memRequest":      resource.memRequest,
			"memLimit":        resource.memLimit,
			"resourceProfile": strings.ToUpper(sizeProfile),
		},
		"version":      strings.TrimSpace(input.Version),
		"databaseName": databaseName,
		"username":     username,
	}
	switch engine {
	case "postgresql":
		database["postgresql"] = map[string]any{
			"walEnabled":        false,
			"bootstrapDatabase": databaseName,
			"bootstrapOwner":    username,
		}
	case "mongodb":
		database["mongo"] = map[string]any{
			"replicaSetHorizonsEnabled": len(publicHostnames) >= 3,
		}
	case "mysql":
		database["mysql"] = map[string]any{
			"haproxySize": 2,
		}
	case "redis":
		database["redis"] = map[string]any{
			"exporterEnabled":  false,
			"aclPermissions":   []string{"+@read", "+@write", "+@keyspace", "+@connection"},
			"followersEnabled": true,
		}
	case "cassandra":
		database["cassandra"] = map[string]any{
			"clusterName":       projectName,
			"datacenter":        "dc1",
			"requireClientAuth": false,
		}
	}

	return map[string]any{
		"releaseName": firstNonEmpty(projectName, "database-cluster"),
		"namespace":   namespace,
		"projectName": projectName,
		"cluster": map[string]any{
			"name":        projectName,
			"environment": "DEVELOPMENT",
			"notes":       "",
			"platformConfig": map[string]any{
				"targetClusterName": firstNonEmpty(input.TargetCluster, "k8s-cluster2"),
			},
		},
		"database": database,
		"secrets": map[string]any{
			"pgPassword":        secretForEngine(engine, "postgresql", input.Password),
			"mongoPassword":     secretForEngine(engine, "mongodb", input.Password),
			"mysqlPassword":     secretForEngine(engine, "mysql", input.Password),
			"redisPassword":     secretForEngine(engine, "redis", input.Password),
			"cassandraPassword": secretForEngine(engine, "cassandra", input.Password),
		},
	}
}

func secretForEngine(engine, target, password string) any {
	if engine != target {
		return nil
	}
	return password
}

func normalizeClusterEngine(engine string) string {
	switch strings.ToLower(strings.TrimSpace(engine)) {
	case "mongodb", "mongo":
		return "mongodb"
	case "mysql":
		return "mysql"
	case "redis":
		return "redis"
	case "cassandra":
		return "cassandra"
	default:
		return "postgresql"
	}
}

func normalizeClusterSizeProfile(size string) string {
	switch strings.ToLower(strings.TrimSpace(size)) {
	case "medium":
		return "medium"
	case "large":
		return "large"
	default:
		return "small"
	}
}

func defaultClusterDatabase(engine string) string {
	switch engine {
	case "redis":
		return "0"
	case "cassandra":
		return "app_keyspace"
	default:
		return "appdb"
	}
}

func defaultClusterUsername(engine string) string {
	switch engine {
	case "postgresql":
		return "app_postgres"
	case "mongodb":
		return "databaseAdmin"
	case "mysql":
		return "appuser"
	case "redis":
		return "default"
	case "cassandra":
		return "cassandra"
	default:
		return "app_user"
	}
}

func normalizeClusterHostnames(values []string) []string {
	const wildcardZone = "cluster-db.autonomous-istad.com"
	hostnames := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		hostname := strings.ToLower(strings.TrimSpace(value))
		if hostname == "" {
			continue
		}
		if !strings.Contains(hostname, ".") {
			hostname += "." + wildcardZone
		}
		if seen[hostname] {
			continue
		}
		seen[hostname] = true
		hostnames = append(hostnames, hostname)
	}
	return hostnames
}

func clusterResource(engine, size string) clusterResourceProfile {
	profiles := map[string]map[string]clusterResourceProfile{
		"postgresql": {
			"small":  {"4Gi", "200m", "500m", "512Mi", "1Gi"},
			"medium": {"8Gi", "350m", "1000m", "1Gi", "2Gi"},
			"large":  {"16Gi", "500m", "1500m", "2Gi", "4Gi"},
		},
		"mongodb": {
			"small":  {"4Gi", "200m", "600m", "768Mi", "1536Mi"},
			"medium": {"8Gi", "350m", "1000m", "1Gi", "2Gi"},
			"large":  {"16Gi", "500m", "1500m", "2Gi", "4Gi"},
		},
		"mysql": {
			"small":  {"4Gi", "250m", "600m", "768Mi", "1536Mi"},
			"medium": {"8Gi", "400m", "1200m", "1Gi", "2Gi"},
			"large":  {"16Gi", "600m", "1500m", "2Gi", "4Gi"},
		},
		"redis": {
			"small":  {"3Gi", "75m", "200m", "128Mi", "512Mi"},
			"medium": {"5Gi", "150m", "350m", "256Mi", "768Mi"},
			"large":  {"10Gi", "250m", "500m", "512Mi", "1Gi"},
		},
		"cassandra": {
			"small":  {"4Gi", "300m", "1000m", "768Mi", "2Gi"},
			"medium": {"8Gi", "500m", "1500m", "2Gi", "4Gi"},
			"large":  {"16Gi", "750m", "2000m", "3Gi", "6Gi"},
		},
	}
	if bySize, ok := profiles[engine]; ok {
		if profile, ok := bySize[size]; ok {
			return profile
		}
		return bySize["small"]
	}
	return profiles["postgresql"]["small"]
}

func normalizeMonolithicDeployment(payload map[string]any) MonolithicDeploymentRecord {
	return MonolithicDeploymentRecord{
		ProjectID:         readString(payload["projectId"]),
		Name:              readString(payload["name"]),
		Status:            readString(payload["status"]),
		RepoProvider:      readString(payload["repoProvider"]),
		DeployURL:         readString(payload["deployUrl"]),
		QueueURL:          readString(payload["queueUrl"]),
		QueueItemID:       readInt(payload["queueItemId"]),
		JenkinsJobName:    readString(payload["jenkinsJobName"]),
		AutoDeployEnabled: readBool(payload["autoDeployEnabled"]),
	}
}

func ParseDeploymentLogLines(statusLog string) []string {
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

func DatabaseDeploymentTerminal(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "DEPLOYED", "READY", "RUNNING", "SUCCESS", "SUCCEEDED", "FAILED", "ERROR", "UNHEALTHY", "CANCELLED", "CANCELED":
		return true
	default:
		return false
	}
}

func DatabaseDeploymentFailed(status string) bool {
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

func ProjectMatchesFilter(project LiveProject, filter string) bool {
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

func FilteredProjects(projects []LiveProject, filter string) []LiveProject {
	out := make([]LiveProject, 0, len(projects))
	for _, project := range projects {
		if ProjectMatchesFilter(project, filter) {
			out = append(out, project)
		}
	}
	return out
}

func firstNonEmpty(strings ...string) string {
	for _, s := range strings {
		if s != "" {
			return s
		}
	}
	return ""
}
