package main

import (
	"flag"
	"fmt"
	"github.com/azimjohn/jprq.live/jprq"
	"github.com/gorilla/mux"
	"github.com/labstack/gommon/log"
	"net/http"
)

var baseHost string

func main() {
	flag.StringVar(&baseHost, "host", "jprq.live", "Base Host")
	flag.Parse()

	j := jprq.New(baseHost)
	r := mux.NewRouter()
	r.HandleFunc("/_ws/", j.WebsocketHandler)
	r.PathPrefix("/").HandlerFunc(j.HttpHandler)

	fmt.Println("Server is running on Port 4200")
	log.Fatal(http.ListenAndServe(":4200", r))
}
