package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c ProjectClient) ResolveEffectiveClusterNamespace(ctx context.Context, token, candidate string) (string, error) {
	candidate = strings.TrimSpace(candidate)
	if candidate != "" && !strings.EqualFold(candidate, "default") {
		return candidate, nil
	}

	namespace, status, message, err := c.workspaceBootstrapRequest(ctx, token, http.MethodGet)
	if err != nil {
		return "", err
	}
	if status >= 200 && status < 300 && namespace != "" {
		return namespace, nil
	}

	namespace, bootstrapStatus, bootstrapMessage, err := c.workspaceBootstrapRequest(ctx, token, http.MethodPost)
	if err != nil {
		return "", err
	}
	if bootstrapStatus >= 200 && bootstrapStatus < 300 && namespace != "" {
		return namespace, nil
	}
	if bootstrapStatus == http.StatusAccepted || status == http.StatusAccepted {
		return "", fmt.Errorf("%s", firstNonEmpty(bootstrapMessage, message, "workspace onboarding is still in progress; wait a moment and try again"))
	}
	if bootstrapStatus < 200 || bootstrapStatus >= 300 {
		return "", httpStatusError{status: bootstrapStatus, message: firstNonEmpty(bootstrapMessage, "could not resolve the workspace namespace for this account")}
	}
	if status < 200 || status >= 300 {
		return "", httpStatusError{status: status, message: firstNonEmpty(message, "could not resolve the workspace namespace for this account")}
	}
	return "", fmt.Errorf("workspace namespace is not ready for this account yet")
}

func (c ProjectClient) workspaceBootstrapRequest(ctx context.Context, token, method string) (string, int, string, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/v1/workspaces/bootstrap")
	if err != nil {
		return "", 0, "", err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return "", 0, "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return "", 0, "", err
	}
	defer res.Body.Close()

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			return "", res.StatusCode, "", fmt.Errorf("decode workspace bootstrap response: %w", err)
		}
		return "", res.StatusCode, "", nil
	}
	message := firstNonEmpty(readString(payload["details"]), readString(payload["message"]), readString(payload["error"]))
	return readString(payload["namespace"]), res.StatusCode, message, nil
}
