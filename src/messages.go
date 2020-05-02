package main

import (
	"bytes"
	"encoding/json"
	"github.com/gofrs/uuid"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	tunnelCreated = "tunnel"
	request       = "request"
	response      = "response"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (m Message) Bytes() []byte {
	bytes, _ := json.Marshal(m)
	return bytes
}

type TunnelMessage struct {
	Host  string `json:"host"`
	Token string `json:"token"`
}

type RequestMessage struct {
	ID     uuid.UUID           `json:"id"`
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Body   []byte              `json:"body"`
	Header map[string][]string `json:"header"`
}

type ResponseMessage struct {
	RequestId uuid.UUID           `json:"request_id"`
	Body      []byte              `json:"body"`
	Status    int                 `json:"status"`
	Header    map[string][]string `json:"header"`
}

func FromHttpRequest(httpRequest *http.Request) RequestMessage {
	requestMessage := RequestMessage{}
	requestMessage.ID, _ = uuid.NewV4()
	requestMessage.Method = httpRequest.Method
	requestMessage.URL = httpRequest.URL.RequestURI()

	if httpRequest.Body != nil {
		requestMessage.Body, _ = ioutil.ReadAll(httpRequest.Body)
	}

	requestMessage.Header = make(map[string][]string)
	for name, values := range httpRequest.Header {
		requestMessage.Header[name] = values
	}

	return requestMessage
}

func (responseMessage ResponseMessage) WriteToHttpResponse(writer http.ResponseWriter) {
	writer.WriteHeader(responseMessage.Status)
	for name, values := range responseMessage.Header {
		writer.Header()[name] = values
	}

	io.Copy(writer, bytes.NewBuffer(responseMessage.Body))
}
