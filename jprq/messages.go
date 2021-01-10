package jprq

import (
	"bytes"
	"encoding/base64"
	"github.com/gofrs/uuid"
	"github.com/labstack/gommon/log"
	"io"
	"io/ioutil"
	"net/http"
)

type TunnelMessage struct {
	Host  string `json:"host"`
	Token string `json:"token"`
}

type RequestMessage struct {
	ID           uuid.UUID            `json:"id"`
	Method       string               `json:"method"`
	URL          string               `json:"url"`
	Body         string               `json:"body"`
	Header       map[string]string    `json:"header"`
	ResponseChan chan ResponseMessage `json:"-"`
}

type ResponseMessage struct {
	RequestId uuid.UUID         `json:"request_id"`
	Token     string            `json:"token"`
	Body      string            `json:"body"`
	Status    int               `json:"status"`
	Header    map[string]string `json:"header"`
}

func FromHttpRequest(httpRequest *http.Request) RequestMessage {
	requestMessage := RequestMessage{}
	requestMessage.ID, _ = uuid.NewV4()
	requestMessage.Method = httpRequest.Method
	requestMessage.URL = httpRequest.URL.RequestURI()

	if httpRequest.Body != nil {
		body, _ := ioutil.ReadAll(httpRequest.Body)
		requestMessage.Body = base64.StdEncoding.EncodeToString(body)
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
	writer.Header().Set("Access-Control-Max-Age", "*")

	for name, value := range responseMessage.Header {
		writer.Header().Set(name, value)
	}
	if 100 > responseMessage.Status || responseMessage.Status > 600 {
		responseMessage.Status = http.StatusInternalServerError
	}
	writer.WriteHeader(responseMessage.Status)

	decoded, err := base64.StdEncoding.DecodeString(responseMessage.Body)
	if err != nil {
		log.Error("Error Decoding Response Body: ", err)
	}
	io.Copy(writer, bytes.NewBuffer(decoded))
}
