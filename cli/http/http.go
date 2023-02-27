package http

import (
	"bufio"
	"io"
	"net/http"
	"strconv"
)

type Request struct {
	Id      string            `json:"id"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

type Response struct {
	RequestId string            `json:"request_id"`
	Status    int               `json:"status"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
}

func ParseRequests(conn io.Reader, connID string) <-chan Request {
	i := 0
	ch := make(chan Request)
	reader := bufio.NewReader(conn)
	go func() {
		for {
			req, err := http.ReadRequest(reader)
			if err != nil {
				close(ch)
				break
			}
			r := Request{
				Id:     connID + strconv.Itoa(i),
				Method: req.Method,
				URL:    req.URL.String(),
				Body:   "<todo>",
			}
			r.Headers = make(map[string]string)
			for key, value := range req.Header {
				r.Headers[key] = value[0]
			}
			ch <- r
			i++
		}
	}()
	return ch
}

func ParseResponses(conn io.Reader, connID string) <-chan Response {
	i := 0
	ch := make(chan Response)
	reader := bufio.NewReader(conn)
	go func() {
		for {
			resp, err := http.ReadResponse(reader, nil)
			if err != nil {
				close(ch)
				break
			}
			r := Response{
				RequestId: connID + strconv.Itoa(i),
				Status:    resp.StatusCode,
				Body:      "<todo>",
			}
			r.Headers = make(map[string]string)
			for key, value := range resp.Header {
				r.Headers[key] = value[0]
			}
			ch <- r
			i++
		}
	}()
	return ch
}
