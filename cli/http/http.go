package http

import (
	"bufio"
	"net/http"
	"strconv"
)

type Request struct {
	Id      string              `json:"id"`
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Body    string              `json:"body"`
	Headers map[string][]string `json:"headers"`
}

type Response struct {
	RequestId string              `json:"request_id"`
	Status    int                 `json:"status"`
	Headers   map[string][]string `json:"headers"`
	Body      string              `json:"body"`
}

func ParseRequests(r *bufio.Reader, connID string) <-chan Request {
	ch := make(chan Request)
	go func() {
		for i := 0; ; i++ {
			req, err := http.ReadRequest(r)
			if err != nil {
				close(ch)
				break
			}
			ch <- Request{
				Id:      connID + strconv.Itoa(i),
				Method:  req.Method,
				URL:     req.URL.String(),
				Body:    "<todo>",
				Headers: req.Header,
			}
		}
	}()
	return ch
}

func ParseResponses(r *bufio.Reader, connID string) <-chan Response {
	ch := make(chan Response)
	go func() {
		for i := 0; ; i++ {
			resp, err := http.ReadResponse(r, nil)
			if err != nil {
				close(ch)
				break
			}
			ch <- Response{
				RequestId: connID + strconv.Itoa(i),
				Status:    resp.StatusCode,
				Body:      "<todo>",
				Headers:   resp.Header,
			}
		}
	}()
	return ch
}
