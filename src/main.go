package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/labstack/gommon/log"
	"net/http"
)

var baseHost string

func main() {
	flag.StringVar(&baseHost, "host", "jprq.live", "Base Host")
	flag.Parse()

	tunnels = make(map[string]Tunnel)

	fmt.Println(baseHost)
	r := mux.NewRouter()
	r.HandleFunc("/_ws/", WebsocketHandler)
	r.PathPrefix("/").HandlerFunc(HttpHandler)

	fmt.Println("Server is running on Port 4200")
	log.Fatal(http.ListenAndServe(":4200", r))
}
