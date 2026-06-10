package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type ObservabilityClient struct {
	baseURL string
	client  *http.Client
}

func NewObservabilityClient(baseURL string) ObservabilityClient {
	return ObservabilityClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

type MonitoringOverview struct {
	Namespace        string
	GeneratedAt      string
	Clusters         []MonitoringClusterMetric
	NamespaceMetrics MonitoringNamespaceMetrics
	NamespaceSeries  []MonitoringMetricSeries
	Projects         []MonitoringProjectMetrics
}

type MonitoringClusterMetric struct {
	Cluster   string
	UpTargets float64
	Nodes     float64
}

type MonitoringNamespaceMetrics struct {
	TotalPods                     float64
	RunningPods                   float64
	PendingPods                   float64
	FailedPods                    float64
	CPUCores                      float64
	MemoryBytes                   float64
	RestartsLastHour              float64
	NetworkReceiveBytesPerSecond  float64
	NetworkTransmitBytesPerSecond float64
	CPURequestsUsed               float64
	CPURequestsLimit              float64
	MemoryRequestsUsed            float64
	MemoryRequestsLimit           float64
	StorageRequestsUsed           float64
	StorageRequestsLimit          float64
	CPULimitsUsed                 float64
	CPULimitsLimit                float64
	MemoryLimitsUsed              float64
	MemoryLimitsLimit             float64
}

type MonitoringMetricSeries struct {
	Key    string
	Label  string
	Unit   string
	Points []MonitoringMetricPoint
}

type MonitoringMetricPoint struct {
	Timestamp string
	Value     float64
}

type MonitoringProjectMetrics struct {
	ID               string
	Name             string
	Kind             string
	Status           string
	Namespace        string
	TotalPods        float64
	RunningPods      float64
	PendingPods      float64
	FailedPods       float64
	CPUCores         float64
	MemoryBytes      float64
	RestartsLastHour float64
}

type PodSummaryResponse struct {
	Namespace string
	Total     int
	Pods      []PodSummary
}

type PodSummary struct {
	Name            string
	Phase           string
	PodIP           string
	NodeName        string
	ReadyContainers int
	TotalContainers int
	RestartCount    int
	AgeSeconds      int
	Labels          map[string]string
}

type LogLine struct {
	Pod     string
	Level   string
	Message string
}

func (c ObservabilityClient) MonitoringOverview(ctx context.Context, token string) (MonitoringOverview, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/monitoring/overview")
	if err != nil {
		return MonitoringOverview{}, err
	}
	overviewURL, err := url.Parse(endpoint)
	if err != nil {
		return MonitoringOverview{}, err
	}
	query := overviewURL.Query()
	query.Set("includeSeries", "true")
	query.Set("includeProjects", "true")
	query.Set("includeNamespaceDetails", "true")
	overviewURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, overviewURL.String(), nil)
	if err != nil {
		return MonitoringOverview{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return MonitoringOverview{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return MonitoringOverview{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return MonitoringOverview{}, fmt.Errorf("decode monitoring overview response: %w", err)
	}
	return normalizeMonitoringOverview(payload), nil
}

func (c ObservabilityClient) ListPods(ctx context.Context, token, namespace, targetClusterName string) (PodSummaryResponse, error) {
	if strings.TrimSpace(namespace) == "" {
		return PodSummaryResponse{}, fmt.Errorf("namespace is required")
	}
	endpoint, err := url.JoinPath(c.baseURL, "/api/kubernetes/namespaces", strings.TrimSpace(namespace), "pods")
	if err != nil {
		return PodSummaryResponse{}, err
	}
	podsURL, err := url.Parse(endpoint)
	if err != nil {
		return PodSummaryResponse{}, err
	}
	if strings.TrimSpace(targetClusterName) != "" {
		query := podsURL.Query()
		query.Set("targetClusterName", strings.TrimSpace(targetClusterName))
		podsURL.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, podsURL.String(), nil)
	if err != nil {
		return PodSummaryResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return PodSummaryResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return PodSummaryResponse{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return PodSummaryResponse{}, fmt.Errorf("decode Kubernetes pod summary response: %w", err)
	}
	return normalizePodSummaryResponse(payload), nil
}

func (c ObservabilityClient) FetchPodLogs(ctx context.Context, token, namespace, podName, targetClusterName string, tailLines int) ([]LogLine, error) {
	if strings.TrimSpace(namespace) == "" || strings.TrimSpace(podName) == "" {
		return nil, fmt.Errorf("namespace and pod name are required")
	}
	endpoint, err := url.JoinPath(c.baseURL, "/api/kubernetes/namespaces", strings.TrimSpace(namespace), "pods", strings.TrimSpace(podName), "logs", "stream")
	if err != nil {
		return nil, err
	}
	logsURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	query := logsURL.Query()
	if tailLines > 0 {
		query.Set("tailLines", fmt.Sprintf("%d", tailLines))
	}
	if strings.TrimSpace(targetClusterName) != "" {
		query.Set("targetClusterName", strings.TrimSpace(targetClusterName))
	}
	logsURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, logsURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Cache-Control", "no-cache")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	return readSSELogLines(ctx, res, strings.TrimSpace(podName)), nil
}

func normalizeMonitoringOverview(payload map[string]any) MonitoringOverview {
	overview := MonitoringOverview{
		Namespace:   readString(payload["namespace"]),
		GeneratedAt: readTimestamp(payload["generatedAt"]),
	}
	if metrics, ok := payload["namespaceMetrics"].(map[string]any); ok {
		overview.NamespaceMetrics = normalizeNamespaceMetrics(metrics)
	}
	if items, ok := payload["clusters"].([]any); ok {
		for _, item := range items {
			if raw, ok := item.(map[string]any); ok {
				overview.Clusters = append(overview.Clusters, MonitoringClusterMetric{
					Cluster:   readString(raw["cluster"]),
					UpTargets: readFloat(raw["upTargets"]),
					Nodes:     readFloat(raw["nodes"]),
				})
			}
		}
	}
	if items, ok := payload["namespaceSeries"].([]any); ok {
		for _, item := range items {
			if raw, ok := item.(map[string]any); ok {
				overview.NamespaceSeries = append(overview.NamespaceSeries, normalizeMetricSeries(raw))
			}
		}
	}
	if items, ok := payload["projects"].([]any); ok {
		for _, item := range items {
			if raw, ok := item.(map[string]any); ok {
				overview.Projects = append(overview.Projects, normalizeMonitoringProjectMetrics(raw))
			}
		}
	}
	return overview
}

func normalizeNamespaceMetrics(raw map[string]any) MonitoringNamespaceMetrics {
	return MonitoringNamespaceMetrics{
		TotalPods:                     readFloat(raw["totalPods"]),
		RunningPods:                   readFloat(raw["runningPods"]),
		PendingPods:                   readFloat(raw["pendingPods"]),
		FailedPods:                    readFloat(raw["failedPods"]),
		CPUCores:                      readFloat(raw["cpuCores"]),
		MemoryBytes:                   readFloat(raw["memoryBytes"]),
		RestartsLastHour:              readFloat(raw["restartsLastHour"]),
		NetworkReceiveBytesPerSecond:  readFloat(raw["networkReceiveBytesPerSecond"]),
		NetworkTransmitBytesPerSecond: readFloat(raw["networkTransmitBytesPerSecond"]),
		CPURequestsUsed:               readFloat(raw["cpuRequestsUsed"]),
		CPURequestsLimit:              readFloat(raw["cpuRequestsLimit"]),
		MemoryRequestsUsed:            readFloat(raw["memoryRequestsUsed"]),
		MemoryRequestsLimit:           readFloat(raw["memoryRequestsLimit"]),
		StorageRequestsUsed:           readFloat(raw["storageRequestsUsed"]),
		StorageRequestsLimit:          readFloat(raw["storageRequestsLimit"]),
		CPULimitsUsed:                 readFloat(raw["cpuLimitsUsed"]),
		CPULimitsLimit:                readFloat(raw["cpuLimitsLimit"]),
		MemoryLimitsUsed:              readFloat(raw["memoryLimitsUsed"]),
		MemoryLimitsLimit:             readFloat(raw["memoryLimitsLimit"]),
	}
}

func normalizeMetricSeries(raw map[string]any) MonitoringMetricSeries {
	series := MonitoringMetricSeries{
		Key:   readString(raw["key"]),
		Label: readString(raw["label"]),
		Unit:  readString(raw["unit"]),
	}
	if items, ok := raw["points"].([]any); ok {
		for _, item := range items {
			if point, ok := item.(map[string]any); ok {
				series.Points = append(series.Points, MonitoringMetricPoint{
					Timestamp: readTimestamp(point["timestamp"]),
					Value:     readFloat(point["value"]),
				})
			}
		}
	}
	return series
}

func normalizeMonitoringProjectMetrics(raw map[string]any) MonitoringProjectMetrics {
	return MonitoringProjectMetrics{
		ID:               readString(raw["id"]),
		Name:             readString(raw["name"]),
		Kind:             readString(raw["kind"]),
		Status:           readString(raw["status"]),
		Namespace:        readString(raw["namespace"]),
		TotalPods:        readFloat(raw["totalPods"]),
		RunningPods:      readFloat(raw["runningPods"]),
		PendingPods:      readFloat(raw["pendingPods"]),
		FailedPods:       readFloat(raw["failedPods"]),
		CPUCores:         readFloat(raw["cpuCores"]),
		MemoryBytes:      readFloat(raw["memoryBytes"]),
		RestartsLastHour: readFloat(raw["restartsLastHour"]),
	}
}

func normalizePodSummaryResponse(payload map[string]any) PodSummaryResponse {
	out := PodSummaryResponse{
		Namespace: readString(payload["namespace"]),
		Total:     readInt(payload["total"]),
	}
	if items, ok := payload["pods"].([]any); ok {
		for _, item := range items {
			if raw, ok := item.(map[string]any); ok {
				out.Pods = append(out.Pods, normalizePodSummary(raw))
			}
		}
	}
	return out
}

func normalizePodSummary(raw map[string]any) PodSummary {
	labels := map[string]string{}
	if rawLabels, ok := raw["labels"].(map[string]any); ok {
		for key, value := range rawLabels {
			if label := readString(value); label != "" {
				labels[key] = label
			}
		}
	}
	return PodSummary{
		Name:            readString(raw["name"]),
		Phase:           readString(raw["phase"]),
		PodIP:           readString(raw["podIP"]),
		NodeName:        readString(raw["nodeName"]),
		ReadyContainers: readInt(raw["readyContainers"]),
		TotalContainers: readInt(raw["totalContainers"]),
		RestartCount:    readInt(raw["restartCount"]),
		AgeSeconds:      readInt(raw["ageSeconds"]),
		Labels:          labels,
	}
}

func readSSELogLines(ctx context.Context, res *http.Response, podName string) []LogLine {
	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	var lines []LogLine
	var eventName string
	var dataParts []string
	flush := func() {
		if len(dataParts) == 0 {
			eventName = ""
			return
		}
		message := sanitizeLogLine(strings.Join(dataParts, "\n"))
		if message != "" && (eventName == "" || eventName == "log") {
			lines = append(lines, LogLine{
				Pod:     podName,
				Level:   InferLogLevel(message),
				Message: message,
			})
		}
		eventName = ""
		dataParts = nil
	}
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return lines
		default:
		}
		line := scanner.Text()
		if line == "" {
			flush()
			continue
		}
		switch {
		case strings.HasPrefix(line, "event:"):
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			dataParts = append(dataParts, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		case !strings.HasPrefix(line, ":"):
			dataParts = append(dataParts, line)
		}
		if len(lines) >= 200 {
			return lines
		}
	}
	flush()
	return lines
}

var ansiEscapePattern = regexp.MustCompile(`\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)

func sanitizeLogLine(value string) string {
	cleaned := ansiEscapePattern.ReplaceAllString(value, "")
	cleaned = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r >= ' ' {
			return r
		}
		return -1
	}, cleaned)
	return strings.TrimSpace(cleaned)
}

func InferLogLevel(line string) string {
	normalized := strings.ToLower(line)
	switch {
	case strings.Contains(normalized, "error"),
		strings.Contains(normalized, "exception"),
		strings.Contains(normalized, "failed"),
		strings.Contains(normalized, "failure"),
		strings.Contains(normalized, "fatal"),
		strings.Contains(normalized, "unhealthy"),
		strings.Contains(normalized, "denied"):
		return "error"
	case strings.Contains(normalized, "warn"),
		strings.Contains(normalized, "retry"),
		strings.Contains(normalized, "waiting"),
		strings.Contains(normalized, "pending"),
		strings.Contains(normalized, "back-off"):
		return "warn"
	case strings.Contains(normalized, "ready"),
		strings.Contains(normalized, "started"),
		strings.Contains(normalized, "listening"),
		strings.Contains(normalized, "success"),
		strings.Contains(normalized, "succeeded"),
		strings.Contains(normalized, "completed"),
		strings.Contains(normalized, "deployed"),
		strings.Contains(normalized, "healthy"):
		return "success"
	default:
		return "info"
	}
}
