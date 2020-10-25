package jprq

import (
	"net/http"
)

func (j Jprq) HttpHandler(writer http.ResponseWriter, request *http.Request) {
	host := request.Host
	tunnel, err := j.GetTunnelByHost(host)

	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte(err.Error()))
		return
	}

	requestMessage := FromHttpRequest(request)
	tunnel.requestChan <- requestMessage

	responseMessage := <-requestMessage.ResponseChan
	responseMessage.WriteToHttpResponse(writer)
}
