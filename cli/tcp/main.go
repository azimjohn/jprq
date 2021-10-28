package main

import (
	"flag"
	"fmt"
	"github.com/azimjohn/jprq.io/jprq_tcp"
	"github.com/gorilla/mux"
	"github.com/labstack/gommon/log"
	"net/http"
)

func main() {
	var baseHost string
	flag.StringVar(&baseHost, "host", "tcp.jprq.io", "Base Host")
	flag.Parse()

	j := jprq_tcp.New(baseHost)
	r := mux.NewRouter()
	r.HandleFunc("/_ws/", j.WebsocketHandler)

	fmt.Println("Server is running on Port 4500")
	log.Fatal(http.ListenAndServe(":4500", r))
}
