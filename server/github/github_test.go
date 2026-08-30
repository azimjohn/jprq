package github

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGithub_Authenticate(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
		wantUser   User
	}{
		{
			name:       "valid token",
			wantErr:    false,
			statusCode: http.StatusOK,
			response:   `{"login": "torvalds", "name": "Linus Torvalds"}`,
			wantUser:   User{Login: "torvalds", Name: "Linus Torvalds"},
		},
		{
			name:       "invalid token",
			wantErr:    true,
			statusCode: http.StatusUnauthorized,
			response:   `{}`,
			wantUser:   User{},
		},
		{
			name:       "request failed",
			wantErr:    true,
			statusCode: 0,
			response:   `{}`,
			wantUser:   User{},
		},
		{
			name:       "decoding response failed",
			wantErr:    true,
			statusCode: http.StatusOK,
			response:   `invalid json`,
			wantUser:   User{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			g := github{userEndpoint: server.URL}
			user, err := g.Authenticate("token")

			if (err != nil) != tt.wantErr {
				t.Logf("Github.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				t.Fail()
			}
			if err == nil && user.Login != tt.wantUser.Login {
				t.Logf("Github.Authenticate() = %v, want %v", user, tt.wantUser)
				t.Fail()
			}
		})
	}
}

func TestGithub_ObtainToken(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantToken  string
		wantErr    bool
	}{
		{
			name:       "token prefix is stripped",
			wantErr:    false,
			statusCode: http.StatusOK,
			response:   `{"access_token": "gho_abc123"}`,
			wantToken:  "abc123",
		},
		{
			name:       "only the prefix is stripped",
			wantErr:    false,
			statusCode: http.StatusOK,
			response:   `{"access_token": "gho_ghost99"}`,
			wantToken:  "ghost99",
		},
		{
			name:       "obtaining token failed",
			wantErr:    true,
			statusCode: http.StatusUnauthorized,
			response:   `{}`,
			wantToken:  "",
		},
		{
			name:       "decoding response failed",
			wantErr:    true,
			statusCode: http.StatusOK,
			response:   `invalid json`,
			wantToken:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			g := github{tokenEndpoint: server.URL}
			token, err := g.ObtainToken("code")

			if (err != nil) != tt.wantErr {
				t.Logf("Github.ObtainToken() error = %v, wantErr %v", err, tt.wantErr)
				t.Fail()
			}
			if token != tt.wantToken {
				t.Logf("Github.ObtainToken() = %q, want %q", token, tt.wantToken)
				t.Fail()
			}
		})
	}
}
