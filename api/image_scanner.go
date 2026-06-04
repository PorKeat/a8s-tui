package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ImageScannerClient struct {
	baseURL string
	client  *http.Client
}

func NewImageScannerClient(baseURL string) ImageScannerClient {
	return ImageScannerClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

type ImageScannerImage struct {
	ID                 string
	Name               string
	Tag                string
	Repository         string
	Architecture       string
	Distro             string
	SizeLabel          string
	LayerCount         int
	VulnerabilityCount int
	Environment        string
	UpdatedLabel       string
	Port               int
	Digest             string
	LastScannedAtISO   string
	ProjectID          string
	ReleaseID          string
	ProjectName        string
	DeployURL          string
}

type ImageScanFinding struct {
	ID             string
	CVEID          string
	PackageName    string
	PackageVersion string
	Severity       string
	FixedIn        string
	CVSSScore      float64
	Fixable        bool
}

type ImageScanJob struct {
	ID               string
	SourceKind       string
	ImageName        string
	ImageTag         string
	FullReference    string
	RegistryLabel    string
	Digest           string
	Architecture     string
	Distro           string
	SizeLabel        string
	LayerCount       int
	EnvironmentLabel string
	ScannerName      string
	ScannerVersion   string
	DurationSeconds  float64
	ScannedAtISO     string
	Vulnerabilities  []ImageScanFinding
	ContainerPort    int
	Status           string
	Progress         int
	StatusMessage    string
	ReusedResult     bool
	CreatedAt        string
	CompletedAt      string
}

type CreateImageScanInput struct {
	SourceKind  string
	ImageID     string
	ImageRef    string
	RegistryURL string
	ImageName   string
	ImageTag    string
	ForceRescan bool
}

func (c ImageScannerClient) ListImages(ctx context.Context, token string) ([]ImageScannerImage, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/image-scanner/images")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
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

	var payload []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode image scanner images response: %w", err)
	}
	images := make([]ImageScannerImage, 0, len(payload))
	for _, item := range payload {
		images = append(images, normalizeImageScannerImage(item))
	}
	return images, nil
}

func (c ImageScannerClient) ListScans(ctx context.Context, token string) ([]ImageScanJob, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/image-scanner/scans")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
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

	var payload []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode image scanner scan history response: %w", err)
	}
	scans := make([]ImageScanJob, 0, len(payload))
	for _, item := range payload {
		scans = append(scans, normalizeImageScanJob(item))
	}
	return scans, nil
}

func (c ImageScannerClient) CreateScan(ctx context.Context, token string, input CreateImageScanInput) (ImageScanJob, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/image-scanner/scans")
	if err != nil {
		return ImageScanJob{}, err
	}
	body := map[string]any{
		"sourceKind":  firstNonEmpty(strings.TrimSpace(input.SourceKind), "harbor"),
		"forceRescan": input.ForceRescan,
	}
	switch strings.ToLower(strings.TrimSpace(input.SourceKind)) {
	case "external":
		body["imageRef"] = strings.TrimSpace(input.ImageRef)
		body["registryUrl"] = strings.TrimSpace(input.RegistryURL)
		body["imageName"] = strings.TrimSpace(input.ImageName)
		body["imageTag"] = strings.TrimSpace(input.ImageTag)
		body["privateRegistry"] = false
	default:
		body["imageId"] = strings.TrimSpace(input.ImageID)
	}
	payloadBytes, err := json.Marshal(body)
	if err != nil {
		return ImageScanJob{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return ImageScanJob{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return ImageScanJob{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return ImageScanJob{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return ImageScanJob{}, fmt.Errorf("decode image scanner scan response: %w", err)
	}
	return normalizeImageScanJob(payload), nil
}

func (c ImageScannerClient) GetScan(ctx context.Context, token, scanID string) (ImageScanJob, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/image-scanner/scans", strings.TrimSpace(scanID))
	if err != nil {
		return ImageScanJob{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ImageScanJob{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return ImageScanJob{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return ImageScanJob{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return ImageScanJob{}, fmt.Errorf("decode image scanner scan detail response: %w", err)
	}
	return normalizeImageScanJob(payload), nil
}

func (c ImageScannerClient) GetScanReport(ctx context.Context, token, scanID string) (string, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/image-scanner/scans", strings.TrimSpace(scanID), "report")
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
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read image scanner report response: %w", err)
	}
	return strings.TrimSpace(string(body)), nil
}

func normalizeImageScannerImage(payload map[string]any) ImageScannerImage {
	return ImageScannerImage{
		ID:                 readString(payload["id"]),
		Name:               readString(payload["name"]),
		Tag:                readString(payload["tag"]),
		Repository:         readString(payload["repository"]),
		Architecture:       firstNonEmpty(readString(payload["architecture"]), "amd64"),
		Distro:             firstNonEmpty(readString(payload["distro"]), "linux"),
		SizeLabel:          firstNonEmpty(readString(payload["sizeLabel"]), "0 MB"),
		LayerCount:         readInt(payload["layerCount"]),
		VulnerabilityCount: readInt(payload["vulnerabilityCount"]),
		Environment:        firstNonEmpty(readString(payload["environment"]), "production"),
		UpdatedLabel:       firstNonEmpty(readString(payload["updatedLabel"]), "Updated recently"),
		Port:               readInt(payload["port"]),
		Digest:             readString(payload["digest"]),
		LastScannedAtISO:   readString(payload["lastScannedAtIso"]),
		ProjectID:          readString(payload["projectId"]),
		ReleaseID:          readString(payload["releaseId"]),
		ProjectName:        readString(payload["projectName"]),
		DeployURL:          readString(payload["deployUrl"]),
	}
}

func normalizeImageScanJob(payload map[string]any) ImageScanJob {
	vulnerabilities := []ImageScanFinding{}
	if items, ok := payload["vulnerabilities"].([]any); ok {
		vulnerabilities = make([]ImageScanFinding, 0, len(items))
		for _, item := range items {
			record, ok := item.(map[string]any)
			if !ok {
				continue
			}
			vulnerabilities = append(vulnerabilities, normalizeImageScanFinding(record))
		}
	}
	return ImageScanJob{
		ID:               readString(payload["id"]),
		SourceKind:       firstNonEmpty(readString(payload["sourceKind"]), "harbor"),
		ImageName:        readString(payload["imageName"]),
		ImageTag:         readString(payload["imageTag"]),
		FullReference:    readString(payload["fullReference"]),
		RegistryLabel:    firstNonEmpty(readString(payload["registryLabel"]), "registry"),
		Digest:           readString(payload["digest"]),
		Architecture:     firstNonEmpty(readString(payload["architecture"]), "amd64"),
		Distro:           firstNonEmpty(readString(payload["distro"]), "linux"),
		SizeLabel:        firstNonEmpty(readString(payload["sizeLabel"]), "0 MB"),
		LayerCount:       readInt(payload["layerCount"]),
		EnvironmentLabel: firstNonEmpty(readString(payload["environmentLabel"]), "image"),
		ScannerName:      firstNonEmpty(readString(payload["scannerName"]), "Trivy"),
		ScannerVersion:   firstNonEmpty(readString(payload["scannerVersion"]), "unknown"),
		DurationSeconds:  readFloat(payload["durationSeconds"]),
		ScannedAtISO:     readString(payload["scannedAtIso"]),
		Vulnerabilities:  vulnerabilities,
		ContainerPort:    readInt(payload["containerPort"]),
		Status:           firstNonEmpty(readString(payload["status"]), "PENDING"),
		Progress:         readInt(payload["progress"]),
		StatusMessage:    readString(payload["statusMessage"]),
		ReusedResult:     readBool(payload["reusedResult"]),
		CreatedAt:        readTimestamp(payload["createdAt"]),
		CompletedAt:      readTimestamp(payload["completedAt"]),
	}
}

func normalizeImageScanFinding(payload map[string]any) ImageScanFinding {
	return ImageScanFinding{
		ID:             readString(payload["id"]),
		CVEID:          readString(payload["cveId"]),
		PackageName:    readString(payload["packageName"]),
		PackageVersion: firstNonEmpty(readString(payload["packageVersion"]), "unknown"),
		Severity:       strings.ToUpper(firstNonEmpty(readString(payload["severity"]), "LOW")),
		FixedIn:        firstNonEmpty(readString(payload["fixedIn"]), "No fix"),
		CVSSScore:      readFloat(payload["cvssScore"]),
		Fixable:        readBool(payload["fixable"]),
	}
}

func ImageScanTerminal(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "COMPLETED", "SUCCESS", "SUCCEEDED", "FAILED", "ERROR", "CANCELLED", "CANCELED":
		return true
	default:
		return false
	}
}

func ImageScanFailed(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "FAILED", "ERROR", "CANCELLED", "CANCELED":
		return true
	default:
		return false
	}
}

func ImageScanSeverityCounts(findings []ImageScanFinding) map[string]int {
	counts := map[string]int{"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0}
	for _, finding := range findings {
		key := strings.ToUpper(strings.TrimSpace(finding.Severity))
		if _, ok := counts[key]; ok {
			counts[key]++
		}
	}
	return counts
}

func readFloat(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		parsed, _ := typed.Float64()
		return parsed
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return parsed
	default:
		return 0
	}
}
