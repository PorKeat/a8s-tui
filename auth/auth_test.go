package auth

import (
	"a8s-tui/config"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCodeChallenge(t *testing.T) {
	verifier := "test-verifier"
	challenge := codeChallenge(verifier)
	if challenge == "" || challenge == verifier {
		t.Fatalf("unexpected challenge %q", challenge)
	}
	if strings.Contains(challenge, "=") {
		t.Fatalf("challenge should be raw URL encoding, got %q", challenge)
	}
}

func TestBuildLoginURL(t *testing.T) {
	client := AuthClient{
		config: config.AppConfig{
			KeycloakURL:         "https://keycloak.example.com",
			KeycloakRealm:       "a8s",
			KeycloakClientID:    "a8s-tui",
			KeycloakRedirectURL: "http://localhost:8250/callback",
		},
	}

	rawURL, err := client.buildLoginURL("state-value", "challenge-value")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/realms/a8s/protocol/openid-connect/auth" {
		t.Fatalf("path = %q", parsed.Path)
	}
	query := parsed.Query()
	for key, expected := range map[string]string{
		"response_type":         "code",
		"client_id":             "a8s-tui",
		"redirect_uri":          "http://localhost:8250/callback",
		"scope":                 "openid profile email",
		"prompt":                "login select_account",
		"max_age":               "0",
		"state":                 "state-value",
		"code_challenge":        "challenge-value",
		"code_challenge_method": "S256",
	} {
		if query.Get(key) != expected {
			t.Fatalf("%s = %q", key, query.Get(key))
		}
	}
}

func TestBuildLogoutURL(t *testing.T) {
	client := AuthClient{
		config: config.AppConfig{
			KeycloakURL:      "https://keycloak.example.com",
			KeycloakRealm:    "a8s",
			KeycloakClientID: "a8s-tui",
		},
	}

	rawURL, err := client.buildLogoutURL(TokenSet{IDToken: "id-token"})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/realms/a8s/protocol/openid-connect/logout" {
		t.Fatalf("path = %q", parsed.Path)
	}
	query := parsed.Query()
	if query.Get("client_id") != "a8s-tui" {
		t.Fatalf("client_id = %q", query.Get("client_id"))
	}
	if query.Get("id_token_hint") != "id-token" {
		t.Fatalf("id_token_hint = %q", query.Get("id_token_hint"))
	}
}

func TestCallbackHandlerRejectsInvalidState(t *testing.T) {
	results := make(chan callbackResult, 1)
	handler := callbackHandler("/callback", "expected", results)
	req := httptest.NewRequest(http.MethodGet, "/callback?state=wrong&code=abc", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	got := <-results
	if got.err == nil {
		t.Fatal("expected callback error")
	}
}

func TestCallbackHandlerShowsSuccessHTML(t *testing.T) {
	results := make(chan callbackResult, 1)
	handler := callbackHandler("/callback", "expected", results)
	req := httptest.NewRequest(http.MethodGet, "/callback?state=expected&code=abc", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Login successful") || !strings.Contains(body, "return to your terminal") {
		t.Fatalf("success page did not include expected text: %s", body)
	}
	got := <-results
	if got.err != nil || got.code != "abc" {
		t.Fatalf("callback result = %#v", got)
	}
}

func TestTokenRequestParsesToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/realms/a8s/protocol/openid-connect/token" {
			t.Fatalf("unexpected token path %q", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("client_secret") != "secret" {
			t.Fatal("missing client secret in token request")
		}
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken:  "access",
			RefreshToken: "refresh",
			IDToken:      "id-token",
			TokenType:    "Bearer",
			ExpiresIn:    60,
		})
	}))
	defer server.Close()

	client := AuthClient{
		config: config.AppConfig{
			KeycloakURL:          server.URL,
			KeycloakRealm:        "a8s",
			KeycloakClientID:     "client",
			KeycloakClientSecret: "secret",
			KeycloakRedirectURL:  "http://localhost:8250/callback",
		},
		client: server.Client(),
	}
	values := url.Values{"client_secret": []string{"secret"}}
	tokens, err := client.tokenRequest(context.Background(), values)
	if err != nil {
		t.Fatal(err)
	}
	if tokens.AccessToken != "access" || tokens.RefreshToken != "refresh" || tokens.IDToken != "id-token" || tokens.ExpiresAt.IsZero() {
		t.Fatalf("unexpected tokens: %#v", tokens)
	}
}
