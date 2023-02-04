package main

import (
	_ "embed"
	"fmt"
	"github.com/azimjohn/jprq/server/github"
	"log"
	"net/http"
	"os"
)

var oauth github.Authenticator

//go:embed static/index.html
var html string

//go:embed static/config.json
var config string

//go:embed static/install.sh
var installer string

func main() {
	clientId := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if clientId == "" || clientSecret == "" {
		log.Fatalf("missing github client id/secret")
	}
	oauth = github.New(clientId, clientSecret)

	http.HandleFunc("/", contentHandler([]byte(html), "text/html"))
	http.HandleFunc("/config.json", contentHandler([]byte(config), "application/json"))
	http.HandleFunc("/install.sh", contentHandler([]byte(installer), "text/x-shellscript"))
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/oauth-callback", oauthCallback)

	log.Print("Listening on 127.0.0.1:3300")
	log.Fatal(http.ListenAndServe(":3300", nil))
}

func contentHandler(content []byte, contentType string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write(content)
	}
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, oauth.OAuthUrl(), http.StatusFound)
}

func oauthCallback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil || r.FormValue("code") == "" {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	token, err := oauth.ObtainToken(r.FormValue("code"))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	redirect := fmt.Sprintf("http://127.0.0.1:4444/store-token?token=%s", token)
	http.Redirect(w, r, redirect, http.StatusFound)
}
