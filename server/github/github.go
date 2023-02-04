package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	Name  string `json:"name"`
	Login string `json:"login"`
}

type Authenticator interface {
	Authenticate(token string) (User, error)
}

type github struct {
	clientId     string
	clientSecret string
	authEndpoint string
}

func New(clientId, clientSecret string) Authenticator {
	return github{
		clientId:     clientId,
		clientSecret: clientSecret,
		authEndpoint: "https://api.github.com/user",
	}
}

func (g github) Authenticate(token string) (User, error) {
	user := User{}
	client := &http.Client{}

	req, _ := http.NewRequest("GET", g.authEndpoint, nil)
	req.Header.Set("Authorization", "token "+token)
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
		return user, fmt.Errorf("decoding user data failed: %v", err)
	}

	return user, nil
}
