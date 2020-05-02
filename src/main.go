package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/labstack/gommon/log"
	"net/http"
)

func main() {
	config.Load()
	tunnels = make(map[string]Tunnel)

	r := mux.NewRouter()
	r.HandleFunc("/_ws/", WebsocketHandler)
	r.PathPrefix("/").HandlerFunc(HttpHandler)

	fmt.Println("Server is running on Port 4200")
	log.Fatal(http.ListenAndServe(":4200", r))
}
