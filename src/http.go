package main

import (
	"fmt"
	"net/http"
)

func HttpHandler(writer http.ResponseWriter, request *http.Request) {
	host := request.Host
	tunnel, err := GetTunnelByHost(host)

	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte(err.Error()))
	}

	fmt.Println(tunnel)
	// write response
}
