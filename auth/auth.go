package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/PorKeat/a8s-tui/config"
)

const tokenRefreshSkew = 30 * time.Second

//go:embed a8s-logo.avif
var callbackLogoAVIF []byte

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
	values.Set("prompt", "login select_account")
	values.Set("max_age", "0")
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
	if strings.TrimSpace(current.IDToken) == "" && strings.TrimSpace(current.AccessToken) == "" && strings.TrimSpace(current.RefreshToken) == "" {
		return nil
	}
	logoutURL, err := a.buildLogoutURL(current)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, logoutURL, nil)
	if err != nil {
		return err
	}
	res, err := a.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("call Keycloak logout: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("Keycloak logout failed with HTTP %d", res.StatusCode)
	}
	return nil
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

func (a AuthClient) httpClient() *http.Client {
	if a.client != nil {
		return a.client
	}
	return &http.Client{Timeout: 20 * time.Second}
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

	res, err := a.httpClient().Do(req)
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
			writeLoginCallbackHTML(w, http.StatusBadRequest, loginCallbackPage{
				Status:      "Login blocked",
				Title:       "Invalid login state",
				Description: "This browser callback did not match the login session started by your terminal.",
				Detail:      "Close this tab and start login again from A8S TUI.",
				Tone:        "error",
			})
			result <- callbackResult{err: errors.New("Keycloak callback state did not match")}
			return
		}
		if authErr := query.Get("error"); authErr != "" {
			message := query.Get("error_description")
			if message == "" {
				message = authErr
			}
			writeLoginCallbackHTML(w, http.StatusBadRequest, loginCallbackPage{
				Status:      "Login failed",
				Title:       "Keycloak login failed",
				Description: "Keycloak returned an error before A8S TUI could finish authentication.",
				Detail:      message,
				Tone:        "error",
			})
			result <- callbackResult{err: errors.New(message)}
			return
		}
		code := strings.TrimSpace(query.Get("code"))
		if code == "" {
			writeLoginCallbackHTML(w, http.StatusBadRequest, loginCallbackPage{
				Status:      "Login incomplete",
				Title:       "Missing authorization code",
				Description: "The callback reached A8S TUI, but Keycloak did not include an authorization code.",
				Detail:      "Close this tab and retry login from your terminal.",
				Tone:        "error",
			})
			result <- callbackResult{err: errors.New("Keycloak callback did not include an authorization code")}
			return
		}
		writeLoginCallbackHTML(w, http.StatusOK, loginCallbackPage{
			Status:      "Authenticated",
			Title:       "Login successful",
			Description: "You are already logged in to A8S TUI.",
			Detail:      "You can close this tab and return to your terminal.",
			Tone:        "success",
		})
		result <- callbackResult{code: code}
	})
}

type loginCallbackPage struct {
	Status      string
	Title       string
	Description string
	Detail      string
	Tone        string
}

func writeLoginCallbackHTML(w http.ResponseWriter, statusCode int, page loginCallbackPage) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	fmt.Fprint(w, loginCallbackHTML(page))
}

func loginCallbackHTML(page loginCallbackPage) string {
	if page.Tone == "" {
		page.Tone = "success"
	}
	status := html.EscapeString(page.Status)
	title := html.EscapeString(page.Title)
	description := html.EscapeString(page.Description)
	detail := html.EscapeString(page.Detail)
	tone := html.EscapeString(page.Tone)
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>A8S TUI - %s</title>
  <style>
    :root {
      color-scheme: dark;
      --page: #16110d;
      --panel: #211812;
      --panel-soft: #2b2018;
      --text: #fff8f0;
      --muted: #c4aa98;
      --subtle: #8c7465;
      --accent: #ff6b00;
      --accent-soft: rgba(255, 107, 0, .14);
      --accent-border: rgba(255, 107, 0, .42);
      --border: rgba(255, 248, 240, .12);
      --ok: #42e57b;
      --error: #ff6f83;
    }
    * { box-sizing: border-box; }
    html { min-height: 100vh; background: var(--page); }
    body {
      margin: 0;
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 24px;
      background:
        linear-gradient(135deg, rgba(255, 107, 0, .16), transparent 340px),
        linear-gradient(180deg, #1d1510 0, var(--page) 420px);
      color: var(--text);
      font: 16px/1.5 Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    body::before {
      content: "";
      position: fixed;
      inset: 0;
      pointer-events: none;
      background-image:
        linear-gradient(rgba(255,255,255,.025) 1px, transparent 1px),
        linear-gradient(90deg, rgba(255,255,255,.025) 1px, transparent 1px);
      background-size: 44px 44px;
      mask-image: linear-gradient(to bottom, #000, transparent);
    }
    .shell {
      position: relative;
      width: min(920px, calc(100vw - 48px));
    }
    .brandbar {
      display: flex;
      align-items: center;
      gap: 14px;
      margin-bottom: 18px;
      color: var(--muted);
      font-weight: 800;
    }
    .brandbar img {
      width: 46px;
      height: 46px;
      border-radius: 14px;
      object-fit: contain;
      background: #fff8f0;
      padding: 7px;
    }
    .brandbar span {
      color: var(--accent);
      letter-spacing: .18em;
      text-transform: uppercase;
    }
    main {
      overflow: hidden;
      border: 1px solid var(--accent-border);
      background: var(--panel);
      border-radius: 26px;
      box-shadow: 0 28px 90px rgba(0, 0, 0, .38);
    }
    .header {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 26px 30px;
      background:
        linear-gradient(90deg, rgba(255, 107, 0, .22), transparent),
        var(--panel-soft);
      border-bottom: 1px solid var(--border);
    }
    .mark {
      display: grid;
      place-items: center;
      width: 68px;
      height: 68px;
      border-radius: 20px;
      background: #fff8f0;
      border: 1px solid rgba(255, 107, 0, .34);
      box-shadow: inset 0 0 0 1px rgba(255,255,255,.36);
    }
    .mark img {
      width: 46px;
      height: 46px;
      object-fit: contain;
    }
    .eyebrow {
      margin: 0 0 4px;
      color: var(--muted);
      font-size: 13px;
      font-weight: 800;
      letter-spacing: .14em;
      text-transform: uppercase;
    }
    .product {
      margin: 0;
      color: var(--text);
      font-size: 25px;
      font-weight: 800;
    }
    .status {
      margin-left: auto;
      display: inline-flex;
      align-items: center;
      gap: 8px;
      white-space: nowrap;
      padding: 9px 14px;
      border-radius: 999px;
      background: var(--accent-soft);
      color: var(--accent);
      font-size: 13px;
      font-weight: 800;
    }
    .status::before {
      content: "";
      width: 8px;
      height: 8px;
      border-radius: 999px;
      background: var(--ok);
      box-shadow: 0 0 18px var(--ok);
    }
    main[data-tone="error"] .status {
      background: rgba(255, 111, 131, .14);
      color: var(--error);
    }
    main[data-tone="error"] .status::before {
      background: var(--error);
      box-shadow: 0 0 18px var(--error);
    }
    .content {
      display: grid;
      grid-template-columns: 1.25fr .75fr;
      gap: 24px;
      padding: 34px 30px 30px;
    }
    h1 {
      margin: 0 0 16px;
      color: var(--text);
      font-size: 48px;
      line-height: 1.02;
    }
    .lead {
      max-width: 600px;
      margin: 0;
      color: var(--muted);
      font-size: 18px;
    }
    .detail {
      margin-top: 26px;
      padding: 18px 20px;
      border: 1px solid var(--border);
      border-radius: 18px;
      background: rgba(255,255,255,.04);
      color: var(--text);
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
      font-size: 15px;
    }
    .next {
      border: 1px solid var(--border);
      border-radius: 20px;
      background: rgba(255,255,255,.035);
      padding: 22px;
      align-self: stretch;
    }
    .next h2 {
      margin: 0 0 14px;
      color: var(--text);
      font-size: 16px;
    }
    .next p {
      margin: 0 0 18px;
      color: var(--muted);
    }
    .terminal {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 14px 16px;
      border-radius: 14px;
      background: var(--accent-soft);
      color: var(--text);
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
      font-weight: 800;
    }
    .terminal span {
      color: var(--accent);
    }
    .footer {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      padding: 16px 30px;
      border-top: 1px solid var(--border);
      color: var(--subtle);
      font-size: 13px;
    }
    .footer strong { color: var(--accent); }
    @media (max-width: 760px) {
      body { padding: 16px; }
      .shell { width: calc(100vw - 32px); }
      .header { align-items: flex-start; flex-wrap: wrap; padding: 22px; }
      .status { margin-left: 0; }
      .content { grid-template-columns: 1fr; padding: 28px 22px 22px; }
      h1 { font-size: 36px; }
      .footer { flex-direction: column; padding: 16px 22px; }
    }
  </style>
</head>
<body>
  <div class="shell">
    <div class="brandbar">
      <img src="%s" alt="A8S logo">
      <div><span>A8S</span> Autonomous</div>
    </div>
    <main data-tone="%s">
      <section class="header">
        <div class="mark"><img src="%s" alt="A8S logo"></div>
        <div>
          <p class="eyebrow">Keycloak callback</p>
          <p class="product">A8S TUI</p>
        </div>
        <div class="status">%s</div>
      </section>
      <section class="content">
        <div>
          <h1>%s</h1>
          <p class="lead">%s</p>
          <div class="detail">%s</div>
        </div>
        <aside class="next">
          <h2>Return to terminal</h2>
          <p>Your browser login is complete. A8S TUI will continue from the terminal session that opened this page.</p>
          <div class="terminal"><span>$</span> a8s-cli</div>
        </aside>
      </section>
      <section class="footer">
        <span>localhost callback complete</span>
        <span><strong>Next:</strong> return to your terminal</span>
      </section>
    </main>
  </div>
</body>
</html>`, title, callbackLogoDataURI(), tone, callbackLogoDataURI(), status, title, description, detail)
}

func callbackLogoDataURI() string {
	return "data:image/avif;base64," + base64.StdEncoding.EncodeToString(callbackLogoAVIF)
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
