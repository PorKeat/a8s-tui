package api

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type JenkinsLogStreamChunk struct {
	Lines       []string
	BuildNumber int
	Status      string
	Message     string
	Completed   bool
}

func (c ProjectClient) FetchJenkinsLogStreamChunk(ctx context.Context, token, jobName string, queueItemID int) (JenkinsLogStreamChunk, error) {
	if queueItemID <= 0 {
		return JenkinsLogStreamChunk{}, errors.New("Jenkins queue item id is required")
	}
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/jenkins/logs/stream")
	if err != nil {
		return JenkinsLogStreamChunk{}, err
	}
	streamURL, err := url.Parse(endpoint)
	if err != nil {
		return JenkinsLogStreamChunk{}, err
	}
	query := streamURL.Query()
	query.Set("queueItem", strconv.Itoa(queueItemID))
	if strings.TrimSpace(jobName) != "" {
		query.Set("job", strings.TrimSpace(jobName))
	}
	streamURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL.String(), nil)
	if err != nil {
		return JenkinsLogStreamChunk{}, err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Cache-Control", "no-cache")

	res, err := c.client.Do(req)
	if err != nil {
		return JenkinsLogStreamChunk{}, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return JenkinsLogStreamChunk{}, httpStatusError{status: res.StatusCode, message: decodeErrorMessage(res)}
	}

	var chunk JenkinsLogStreamChunk
	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	var eventName string
	var data strings.Builder
	flush := func() {
		if data.Len() > 0 {
			parseJenkinsStreamEvent(&chunk, eventName, data.String())
		}
		eventName = ""
		data.Reset()
	}
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			flush()
			if chunk.Completed {
				return chunk, nil
			}
		case strings.HasPrefix(line, "event:"):
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	flush()
	if err := scanner.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		return chunk, fmt.Errorf("read Jenkins log stream: %w", err)
	}
	return chunk, nil
}

func parseJenkinsStreamEvent(chunk *JenkinsLogStreamChunk, eventName, raw string) {
	var payload map[string]any
	if json.Unmarshal([]byte(raw), &payload) != nil {
		return
	}
	if build := readInt(payload["build"]); build > 0 {
		chunk.BuildNumber = build
	}
	switch strings.ToLower(strings.TrimSpace(eventName)) {
	case "queued":
		chunk.Status = "QUEUED"
		chunk.Message = firstNonEmpty(readString(payload["message"]), "Queued in Jenkins")
		chunk.Lines = append(chunk.Lines, chunk.Message)
	case "open":
		chunk.Status = "DEPLOYING"
		chunk.Message = fmt.Sprintf("Streaming Jenkins build #%d", chunk.BuildNumber)
		chunk.Lines = append(chunk.Lines, chunk.Message)
	case "log":
		chunk.Status = "DEPLOYING"
		chunk.Lines = append(chunk.Lines, ParseDeploymentLogLines(readString(payload["chunk"]))...)
	case "done":
		result := strings.ToUpper(firstNonEmpty(readString(payload["result"]), "SUCCESS"))
		chunk.Completed = true
		chunk.Message = "Jenkins build completed: " + result
		if result == "SUCCESS" || result == "SUCCESSFUL" {
			chunk.Status = "DEPLOYED"
		} else {
			chunk.Status = "FAILED"
		}
		chunk.Lines = append(chunk.Lines, chunk.Message)
	case "error":
		chunk.Completed = true
		chunk.Status = "FAILED"
		chunk.Message = firstNonEmpty(readString(payload["detail"]), readString(payload["message"]), "Jenkins log stream failed")
		chunk.Lines = append(chunk.Lines, chunk.Message)
	}
}
