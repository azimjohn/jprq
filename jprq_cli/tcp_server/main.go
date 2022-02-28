package main

import (
	"fmt"
	"github.com/azimjohn/jprq/jprq_tcp"
	"github.com/gorilla/mux"
	"github.com/labstack/gommon/log"
	"net/http"
)

func main() {
	j := jprq_tcp.New()
	r := mux.NewRouter()
	r.HandleFunc("/_ws/", j.WebsocketHandler)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	fmt.Println("JPRQ TCP Server is running on Port 4500")
	log.Fatal(http.ListenAndServe(":4500", r))
}
