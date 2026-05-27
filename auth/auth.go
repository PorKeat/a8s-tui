package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"a8s-tui/config"
)

const tokenRefreshSkew = 30 * time.Second

type TokenSet struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	TokenType    string
	ExpiresAt    time.Time
}

func (t TokenSet) CanRefresh() bool {
	return strings.TrimSpace(t.RefreshToken) != ""
}

func (t TokenSet) ExpiresSoon(now time.Time) bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	return now.Add(tokenRefreshSkew).After(t.ExpiresAt)
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

type AuthClient struct {
	config  config.AppConfig
	client  *http.Client
	OpenURL func(string) error
}

func NewAuthClient(config config.AppConfig) AuthClient {
	return AuthClient{
		config: config,
		client: &http.Client{Timeout: 20 * time.Second},
		OpenURL: func(target string) error {
			return openBrowser(target)
		},
	}
}

func (a AuthClient) Login(ctx context.Context) (TokenSet, error) {
	redirectURL, err := url.Parse(a.config.KeycloakRedirectURL)
	if err != nil {
		return TokenSet{}, fmt.Errorf("invalid KEYCLOAK_REDIRECT_URL: %w", err)
	}
	if redirectURL.Scheme != "http" || redirectURL.Host == "" || redirectURL.Path == "" {
		return TokenSet{}, errors.New("KEYCLOAK_REDIRECT_URL must be an http localhost callback URL")
	}

	state, err := randomURLToken(24)
	if err != nil {
		return TokenSet{}, err
	}
	verifier, err := randomURLToken(48)
	if err != nil {
		return TokenSet{}, err
	}
	challenge := codeChallenge(verifier)

	listener, err := net.Listen("tcp", redirectURL.Host)
	if err != nil {
		return TokenSet{}, fmt.Errorf("start callback listener on %s: %w", redirectURL.Host, err)
	}
	defer listener.Close()

	result := make(chan callbackResult, 1)
	server := &http.Server{
		Handler: callbackHandler(redirectURL.Path, state, result),
	}
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			result <- callbackResult{err: serveErr}
		}
	}()
	defer server.Shutdown(context.Background())

	loginURL, err := a.buildLoginURL(state, challenge)
	if err != nil {
		return TokenSet{}, err
	}
	if err := a.OpenURL(loginURL); err != nil {
		return TokenSet{}, fmt.Errorf("open browser for Keycloak login: %w", err)
	}

	select {
	case <-ctx.Done():
		return TokenSet{}, ctx.Err()
	case received := <-result:
		if received.err != nil {
			return TokenSet{}, received.err
		}
		return a.exchangeCode(ctx, received.code, verifier)
	case <-time.After(5 * time.Minute):
		return TokenSet{}, errors.New("login timed out waiting for Keycloak callback")
	}
}

func (a AuthClient) buildLoginURL(state, challenge string) (string, error) {
	authURL, err := url.Parse(a.config.AuthURL())
	if err != nil {
		return "", err
	}
	values := authURL.Query()
	values.Set("response_type", "code")
	values.Set("client_id", a.config.KeycloakClientID)
	values.Set("redirect_uri", a.config.KeycloakRedirectURL)
	values.Set("scope", "openid profile email")
	values.Set("state", state)
	values.Set("code_challenge", challenge)
	values.Set("code_challenge_method", "S256")
	authURL.RawQuery = values.Encode()
	return authURL.String(), nil
}

func (a AuthClient) exchangeCode(ctx context.Context, code, verifier string) (TokenSet, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("client_id", a.config.KeycloakClientID)
	values.Set("client_secret", a.config.KeycloakClientSecret)
	values.Set("redirect_uri", a.config.KeycloakRedirectURL)
	values.Set("code", code)
	values.Set("code_verifier", verifier)
	return a.tokenRequest(ctx, values)
}

func (a AuthClient) Refresh(ctx context.Context, current TokenSet) (TokenSet, error) {
	if !current.CanRefresh() {
		return TokenSet{}, errors.New("no refresh token is available")
	}
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("client_id", a.config.KeycloakClientID)
	values.Set("client_secret", a.config.KeycloakClientSecret)
	values.Set("refresh_token", current.RefreshToken)
	return a.tokenRequest(ctx, values)
}

func (a AuthClient) Logout(current TokenSet) error {
	logoutURL, err := a.buildLogoutURL(current)
	if err != nil {
		return err
	}
	return a.OpenURL(logoutURL)
}

func (a AuthClient) buildLogoutURL(current TokenSet) (string, error) {
	logoutURL, err := url.Parse(a.config.KeycloakIssuer() + "/protocol/openid-connect/logout")
	if err != nil {
		return "", err
	}
	values := logoutURL.Query()
	values.Set("client_id", a.config.KeycloakClientID)
	if strings.TrimSpace(current.IDToken) != "" {
		values.Set("id_token_hint", current.IDToken)
	}
	logoutURL.RawQuery = values.Encode()
	return logoutURL.String(), nil
}

func (a AuthClient) tokenRequest(ctx context.Context, values url.Values) (TokenSet, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.config.TokenURL(),
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return TokenSet{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := a.client.Do(req)
	if err != nil {
		return TokenSet{}, err
	}
	defer res.Body.Close()

	var payload tokenResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return TokenSet{}, fmt.Errorf("decode Keycloak token response: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		message := strings.TrimSpace(payload.ErrorDesc)
		if message == "" {
			message = strings.TrimSpace(payload.Error)
		}
		if message == "" {
			message = fmt.Sprintf("Keycloak token request failed with HTTP %d", res.StatusCode)
		}
		return TokenSet{}, errors.New(message)
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return TokenSet{}, errors.New("Keycloak token response did not include an access token")
	}

	tokenType := strings.TrimSpace(payload.TokenType)
	if tokenType == "" {
		tokenType = "Bearer"
	}
	expiresAt := time.Time{}
	if payload.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	}

	return TokenSet{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		IDToken:      payload.IDToken,
		TokenType:    tokenType,
		ExpiresAt:    expiresAt,
	}, nil
}

type callbackResult struct {
	code string
	err  error
}

func callbackHandler(path, expectedState string, result chan<- callbackResult) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		query := r.URL.Query()
		if query.Get("state") != expectedState {
			http.Error(w, "Invalid login state. You can close this tab and retry in the TUI.", http.StatusBadRequest)
			result <- callbackResult{err: errors.New("Keycloak callback state did not match")}
			return
		}
		if authErr := query.Get("error"); authErr != "" {
			message := query.Get("error_description")
			if message == "" {
				message = authErr
			}
			http.Error(w, "Keycloak login failed. You can close this tab and retry in the TUI.", http.StatusBadRequest)
			result <- callbackResult{err: errors.New(message)}
			return
		}
		code := strings.TrimSpace(query.Get("code"))
		if code == "" {
			http.Error(w, "Missing authorization code. You can close this tab and retry in the TUI.", http.StatusBadRequest)
			result <- callbackResult{err: errors.New("Keycloak callback did not include an authorization code")}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, loginSuccessHTML())
		result <- callbackResult{code: code}
	})
}

func loginSuccessHTML() string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>A8S TUI Login Successful</title>
  <style>
    :root {
      color-scheme: dark;
      --bg: #1f2130;
      --panel: #25283a;
      --text: #c8f7ff;
      --muted: #8f9ab8;
      --accent: #f56618;
      --ok: #a6da95;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      display: grid;
      place-items: center;
      background: var(--bg);
      color: var(--text);
      font: 16px/1.5 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
    }
    main {
      width: min(92vw, 680px);
      padding: 48px;
      border: 1px solid rgba(245, 102, 24, .22);
      background: var(--panel);
      text-align: center;
    }
    .mark {
      width: 56px;
      height: 56px;
      margin: 0 auto 24px;
      border: 1px solid rgba(166, 218, 149, .45);
      display: grid;
      place-items: center;
      color: var(--ok);
      font-size: 30px;
    }
    h1 {
      margin: 0 0 12px;
      color: var(--accent);
      font-size: 28px;
      letter-spacing: 0;
    }
    p {
      margin: 0;
      color: var(--muted);
    }
    strong {
      color: var(--text);
      font-weight: 700;
    }
  </style>
</head>
<body>
  <main>
    <div class="mark">OK</div>
    <h1>Login successful</h1>
    <p>You are already logged in to <strong>A8S TUI</strong>.</p>
    <p>You can close this tab and return to your terminal.</p>
  </main>
</body>
</html>`
}

func randomURLToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func codeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func openBrowser(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	return cmd.Start()
}
