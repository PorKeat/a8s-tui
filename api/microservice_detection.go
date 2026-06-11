package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type MicroserviceDetectionResult struct {
	Repository DetectedMicroserviceRepository `json:"repository"`
	Services   []DetectedMicroserviceService  `json:"services"`
}

type DetectedMicroserviceRepository struct {
	Provider      string `json:"provider"`
	FullName      string `json:"fullName"`
	HTMLURL       string `json:"htmlUrl"`
	DefaultBranch string `json:"defaultBranch"`
	Visibility    string `json:"visibility"`
}

type DetectedMicroserviceService struct {
	Name              string                      `json:"name"`
	ServiceModule     string                      `json:"serviceModule"`
	Path              string                      `json:"path"`
	Manifest          string                      `json:"manifest"`
	RepoProvider      string                      `json:"repoProvider"`
	RepoURL           string                      `json:"repoUrl"`
	RepoFullName      string                      `json:"repoFullName"`
	Branch            string                      `json:"branch"`
	AppPort           int                         `json:"appPort"`
	ServiceType       string                      `json:"serviceType"`
	Framework         string                      `json:"framework"`
	ExposePublic      bool                        `json:"exposePublic"`
	PrimaryPublic     bool                        `json:"primaryPublic"`
	RepoCredentialsID string                      `json:"repoCredentialsId"`
	Env               []MicroserviceEnvInput      `json:"env"`
	Relationships     []MicroserviceRelationInput `json:"relationships"`
	DependsOn         []string                    `json:"dependsOn"`
}

type MicroserviceEnvInput struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Secret bool   `json:"secret"`
}

type MicroserviceRelationInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c ProjectClient) DetectMicroserviceRepository(
	ctx context.Context,
	token, repoURL, branch string,
) (MicroserviceDetectionResult, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/projects/microservices/detect")
	if err != nil {
		return MicroserviceDetectionResult{}, err
	}
	body := map[string]any{
		"repoUrl": strings.TrimSpace(repoURL),
		"branch":  strings.TrimSpace(branch),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return MicroserviceDetectionResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return MicroserviceDetectionResult{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return MicroserviceDetectionResult{}, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		message := decodeErrorMessage(res)
		if res.StatusCode == http.StatusForbidden {
			return MicroserviceDetectionResult{}, fmt.Errorf(
				"GitHub rejected the public repository scan: %s; unauthenticated public scans may be rate-limited",
				firstNonEmpty(message, "forbidden"),
			)
		}
		if res.StatusCode == http.StatusNotFound {
			return MicroserviceDetectionResult{}, fmt.Errorf(
				"public GitHub repository or branch was not found: %s",
				firstNonEmpty(message, "verify the URL and branch"),
			)
		}
		return MicroserviceDetectionResult{}, httpStatusError{status: res.StatusCode, message: message}
	}

	var result MicroserviceDetectionResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return MicroserviceDetectionResult{}, fmt.Errorf("decode microservice repository scan: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(result.Repository.Visibility), "public") {
		return MicroserviceDetectionResult{}, fmt.Errorf("TUI supports public GitHub repositories only")
	}
	return result, nil
}

func (c ProjectClient) ValidateMicroserviceRepositories(
	ctx context.Context,
	token string,
	services []CreateMicroserviceServiceInput,
) error {
	type repository struct {
		url    string
		branch string
	}
	repositories := make([]repository, 0)
	seen := map[string]bool{}
	for _, service := range services {
		repoURL := strings.TrimSpace(service.RepoURL)
		if repoURL == "" {
			return fmt.Errorf("service %s has no Git remote URL", firstNonEmpty(service.Name, "unknown"))
		}
		key := strings.ToLower(repoURL + "|" + strings.TrimSpace(service.Branch))
		if seen[key] {
			continue
		}
		seen[key] = true
		repositories = append(repositories, repository{url: repoURL, branch: service.Branch})
	}
	for _, repository := range repositories {
		if _, err := c.DetectMicroserviceRepository(ctx, token, repository.url, repository.branch); err != nil {
			return fmt.Errorf("Git remote is not accessible: %s: %w", repository.url, err)
		}
	}
	return nil
}

func (service DetectedMicroserviceService) DeploymentInput() CreateMicroserviceServiceInput {
	return CreateMicroserviceServiceInput{
		Name:              service.Name,
		RepoURL:           service.RepoURL,
		RepoFullName:      service.RepoFullName,
		RepoProvider:      service.RepoProvider,
		Path:              service.Path,
		Manifest:          service.Manifest,
		ServiceModule:     service.ServiceModule,
		Branch:            service.Branch,
		AppPort:           service.AppPort,
		ServiceType:       service.ServiceType,
		Framework:         service.Framework,
		ExposePublic:      service.ExposePublic,
		PrimaryPublic:     service.PrimaryPublic,
		RepoCredentialsID: service.RepoCredentialsID,
		Env:               service.Env,
		Relationships:     service.Relationships,
		DependsOn:         service.DependsOn,
	}
}
