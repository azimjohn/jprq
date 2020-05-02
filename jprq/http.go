package main

import (
	"net/http"
)

func HttpHandler(writer http.ResponseWriter, request *http.Request) {
	host := request.Host
	tunnel, err := GetTunnelByHost(host)

	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte(err.Error()))
	}

	requestMessage := FromHttpRequest(request)
	tunnel.requestChan <- requestMessage

	responseMessage := <-requestMessage.ResponseChan
	responseMessage.WriteToHttpResponse(writer)
}
