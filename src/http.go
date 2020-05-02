package main

import (
	"net/http"
)

func HttpHandler(writer http.ResponseWriter, request *http.Request) {
	host := request.Host
	tunnel, error := GetTunnelByHost(host)

	if error != nil {
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte(error.Error()))
	}

	// todo: receive response from websocket and write response
}
