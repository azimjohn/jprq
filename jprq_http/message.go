package jprq_http

import (
	"bytes"
	"github.com/gofrs/uuid"
	"io"
	"io/ioutil"
	"net/http"
)

type ErrorMessage struct {
	Error string `bson:"error"`
}

type TunnelMessage struct {
	Host  string `bson:"host"`
	Token string `bson:"token"`
}

type RequestMessage struct {
	ID           uuid.UUID            `bson:"id"`
	Method       string               `bson:"method"`
	URL          string               `bson:"url"`
	Body         []byte               `bson:"body"`
	Header       map[string]string    `bson:"header"`
	ResponseChan chan ResponseMessage `bson:"-"`
}

type ResponseMessage struct {
	RequestId uuid.UUID         `bson:"request_id"`
	Token     string            `bson:"token"`
	Body      []byte            `bson:"body"`
	Status    int               `bson:"status"`
	Header    map[string]string `bson:"header"`
}

func FromHttpRequest(httpRequest *http.Request) RequestMessage {
	requestMessage := RequestMessage{}
	requestMessage.ID, _ = uuid.NewV4()
	requestMessage.Method = httpRequest.Method
	requestMessage.URL = httpRequest.URL.RequestURI()

	if httpRequest.Body != nil {
		body, _ := ioutil.ReadAll(httpRequest.Body)
		requestMessage.Body = body
	}

	requestMessage.Header = make(map[string]string)
	for name, values := range httpRequest.Header {
		requestMessage.Header[name] = values[0]
	}

	requestMessage.ResponseChan = make(chan ResponseMessage)

	return requestMessage
}

func (responseMessage ResponseMessage) WriteToHttpResponse(writer http.ResponseWriter) {
	// Set CORS Headers
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Headers", "*")
	writer.Header().Set("Access-Control-Max-Age", "86400")

	for name, value := range responseMessage.Header {
		writer.Header().Set(name, value)
	}
	if 100 > responseMessage.Status || responseMessage.Status > 600 {
		responseMessage.Status = http.StatusInternalServerError
	}
	writer.WriteHeader(responseMessage.Status)

	decoded := responseMessage.Body
	io.Copy(writer, bytes.NewBuffer(decoded))
}
