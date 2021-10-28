package main

import (
	"flag"
	"fmt"
	"github.com/azimjohn/jprq.io/jprq_http"
	"github.com/gorilla/mux"
	"github.com/labstack/gommon/log"
	"net/http"
)

var baseHost string

func main() {
	flag.StringVar(&baseHost, "host", "jprq.io", "Base Host")
	flag.Parse()

	j := jprq_http.New(baseHost)
	r := mux.NewRouter()
	r.HandleFunc("/_ws/", j.WebsocketHandler)
	r.PathPrefix("/").HandlerFunc(j.HttpHandler)

	fmt.Println("Server is running on Port 4200")
	log.Fatal(http.ListenAndServe(":4200", r))
}
