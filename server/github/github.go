package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const tokenPrefix = "gho_"

type User struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Login      string `json:"login"`
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
	redirectUri  string
}

func New(clientId, clientSecret string) Authenticator {
	return github{
		clientId:     clientId,
		clientSecret: clientSecret,
		defaultScope: "user:email",
		userEndpoint: "https://api.github.com/user",
		redirectUri:  "https://jprq.io/oauth-callback",
	}
}

func (g github) OAuthUrl() string {
	return fmt.Sprintf("https://github.com/login/oauth/authorize?"+
		"client_id=%s&redirect_uri=%s&scope=%s", g.clientId, url.QueryEscape(g.redirectUri), g.defaultScope)
}

func (g github) ObtainToken(code string) (string, error) {
	client := &http.Client{}

	payload := url.Values{}
	payload.Add("code", code)
	payload.Add("client_id", g.clientId)
	payload.Add("client_secret", g.clientSecret)

	req, err := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		strings.NewReader(payload.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
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
	return strings.TrimLeft(response.AccessToken, tokenPrefix), nil
}

func (g github) Authenticate(token string) (User, error) {
	user := User{}
	client := &http.Client{}

	req, _ := http.NewRequest("GET", g.userEndpoint, nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s%s", tokenPrefix, token))
	resp, err := client.Do(req)

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
