package main

import (
	"fmt"
	"github.com/azimjohn/jprq/jprq_http"
	"github.com/gorilla/mux"
	"github.com/labstack/gommon/log"
	"net/http"
)

func main() {
	j, err := jprq_http.New()
	if err != nil {
		log.Fatal(err.Error())
	}

	r := mux.NewRouter()
	r.HandleFunc("/_ws/", j.WebsocketHandler)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	fmt.Println("JPRQ HTTP Server is running on Port 8080 (ws ctl) & 8081 (public)")
	log.Fatal(http.ListenAndServe(":8080", r))
}
