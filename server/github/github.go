package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const tokenPrefix = "gho_"

type User struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Login      string `json:"login"`
	Allowed    bool   `json:"allowed"`
	Tier       string `json:"tier,omitempty"`
	JoinedDate string `json:"created_at"`
}

type Authenticator interface {
	OAuthUrl() string
	ObtainToken(code string) (string, error)
	Authenticate(token string) (User, error)
}

type github struct {
	clientId     string
	clientSecret string
	defaultScope string
	userEndpoint string
	qir2Endpoint string
	authURL      string
	redirectUri  string
	httpClient   *http.Client
}

func New(clientId, clientSecret string) Authenticator {
	return github{
		clientId:     clientId,
		clientSecret: clientSecret,
		defaultScope: "user:email",
		userEndpoint: "https://api.github.com/user",
		redirectUri:  "https://jprq.io/oauth-callback",
		qir2Endpoint: "https://api.42.uz/api/profile/jprq/",
		authURL:      "https://web.jprq.io/api/auth/validate",
		httpClient:   &http.Client{Timeout: 2 * time.Second},
	}
}

// client returns the configured HTTP client, or http.DefaultClient if the
// struct was instantiated directly (e.g. in tests) without going through New.
func (g github) client() *http.Client {
	if g.httpClient != nil {
		return g.httpClient
	}
	return http.DefaultClient
}

func (g github) OAuthUrl() string {
	return fmt.Sprintf("https://github.com/login/oauth/authorize?"+
		"client_id=%s&redirect_uri=%s&scope=%s", g.clientId, url.QueryEscape(g.redirectUri), g.defaultScope)
}

func (g github) ObtainToken(code string) (string, error) {
	payload := url.Values{}
	payload.Add("code", code)
	payload.Add("client_id", g.clientId)
	payload.Add("client_secret", g.clientSecret)

	req, err := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		strings.NewReader(payload.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to perform http request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	resp, err := g.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform obtain token request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to obtain access token: http %d", resp.StatusCode)
	}

	var response struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode github response: %v", err)
	}
	return strings.TrimPrefix(response.AccessToken, tokenPrefix), nil
}

func (g github) Authenticate(token string) (User, error) {
	user, err := g.authenticate(g.userEndpoint, token)
	if err != nil {
		user, err = g.authenticate(g.qir2Endpoint, token)
	}

	if user.Allowed {
		return user, nil
	}

	result, err := g.validateWithAuth(user.Login)
	if err != nil {
		log.Printf("auth validate failed for %s: %v", user.Login, err)
		return user, nil
	}
	if result.Allowed {
		user.Allowed = true
		user.Tier = result.Tier
	}
	return user, nil
}

func (g github) authenticate(endpoint, token string) (User, error) {
	user := User{}

	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s%s", tokenPrefix, token))
	resp, err := g.client().Do(req)

	if err != nil {
		return user, fmt.Errorf("authentication request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return user, fmt.Errorf("invalid token %v", token)
	}
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return user, fmt.Errorf("failed to decode user data: %v", err)
	}
	user.Login = strings.ToLower(user.Login)
	return user, nil
}

type authValidateResult struct {
	Allowed bool   `json:"allowed"`
	Tier    string `json:"tier"`
}

// validateWithAuth asks the auth service whether the given GitHub login has
// an active jprq subscription.
func (g github) validateWithAuth(login string) (authValidateResult, error) {
	var result authValidateResult

	body, _ := json.Marshal(map[string]string{"github_login": login})
	req, _ := http.NewRequest(http.MethodPost, g.authURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := g.client().Do(req)
	if err != nil {
		return result, fmt.Errorf("call auth service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("auth service returned http %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}
