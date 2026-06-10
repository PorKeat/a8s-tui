package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type ClusterDeploymentStreamChunk struct {
	Lines     []string
	Completed bool
}

func (c ProjectClient) FetchClusterDeploymentStreamChunk(
	ctx context.Context,
	token string,
	namespace string,
	releaseName string,
	targetClusterName string,
) (ClusterDeploymentStreamChunk, error) {
	endpoint, err := url.JoinPath(
		c.baseURL,
		"/api/kubernetes/namespaces",
		strings.TrimSpace(namespace),
		"releases",
		strings.TrimSpace(releaseName),
		"deployment-stream",
	)
	if err != nil {
		return ClusterDeploymentStreamChunk{}, err
	}
	streamURL, err := url.Parse(endpoint)
	if err != nil {
		return ClusterDeploymentStreamChunk{}, err
	}
	if target := strings.TrimSpace(targetClusterName); target != "" {
		query := streamURL.Query()
		query.Set("targetClusterName", target)
		streamURL.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL.String(), nil)
	if err != nil {
		return ClusterDeploymentStreamChunk{}, err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Cache-Control", "no-cache")

	client := *c.client
	client.Timeout = 0
	res, err := client.Do(req)
	if err != nil {
		return ClusterDeploymentStreamChunk{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return ClusterDeploymentStreamChunk{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}
	return readClusterDeploymentStreamChunk(ctx, res), nil
}

func readClusterDeploymentStreamChunk(ctx context.Context, res *http.Response) ClusterDeploymentStreamChunk {
	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	chunk := ClusterDeploymentStreamChunk{}
	var eventName string
	var dataParts []string

	flush := func() {
		if len(dataParts) == 0 {
			eventName = ""
			return
		}
		line := clusterDeploymentStreamLine(eventName, strings.Join(dataParts, "\n"))
		if line != "" && !containsString(chunk.Lines, line) {
			chunk.Lines = append(chunk.Lines, line)
			if clusterDeploymentStreamCompleted(line) {
				chunk.Completed = true
			}
		}
		eventName = ""
		dataParts = nil
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			flush()
			return chunk
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
	}
	flush()
	return chunk
}

func clusterDeploymentStreamLine(eventName string, eventData string) string {
	data := strings.TrimSpace(eventData)
	if data == "" {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return data
	}

	switch eventName {
	case "deployment-event":
		reason := firstNonEmpty(readString(payload["reason"]), "Event")
		message := readString(payload["message"])
		if message == "" {
			return ""
		}
		return fmt.Sprintf("[%s] %s", reason, message)
	case "pod-log":
		pod := firstNonEmpty(readString(payload["pod"]), "pod")
		content := readString(payload["content"])
		if content == "" {
			return ""
		}
		return pod + ": " + content
	case "console-log", "":
		return readString(payload["content"])
	default:
		return readString(payload["content"])
	}
}

func clusterDeploymentStreamCompleted(line string) bool {
	normalized := strings.ToLower(strings.TrimSpace(line))
	if strings.Contains(normalized, "deployment stream completed.") {
		return true
	}
	matches := releaseReadinessPattern.FindStringSubmatch(normalized)
	if len(matches) != 5 {
		return false
	}
	ready, readyTotalErr := strconv.Atoi(matches[1])
	total, totalErr := strconv.Atoi(matches[2])
	running, runningErr := strconv.Atoi(matches[3])
	runningTotal, runningTotalErr := strconv.Atoi(matches[4])
	return readyTotalErr == nil && totalErr == nil && runningErr == nil && runningTotalErr == nil &&
		total > 0 && ready == total && running == runningTotal && total == runningTotal
}

var releaseReadinessPattern = regexp.MustCompile(`release pod readiness:\s*(\d+)/(\d+)\s+ready,\s*(\d+)/(\d+)\s+running`)

func containsString(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
