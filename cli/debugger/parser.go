package debugger

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type request struct {
	Id      string              `json:"id"`
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Body    string              `json:"body"`
	Headers map[string][]string `json:"headers"`
}

type response struct {
	RequestId string              `json:"request_id"`
	Status    int                 `json:"status"`
	Headers   map[string][]string `json:"headers"`
	Body      string              `json:"body"`
}

func parseRequests(r io.Reader, conId string, process func(interface{})) {
	for i := 0; ; i++ {
		fmt.Println("parsing requests")
		req, err := http.ReadRequest(bufio.NewReader(r))
		fmt.Println("parsing requests finished")
		if err != nil {
			break
		}
		r := request{
			Id:      conId + strconv.Itoa(i),
			Method:  req.Method,
			URL:     req.URL.String(),
			Body:    "<todo>",
			Headers: req.Header,
		}
		process(r)
	}
}

func parseResponses(r io.Reader, conId string, process func(interface{})) {
	for i := 0; ; i++ {
		resp, err := http.ReadResponse(bufio.NewReader(r), nil)
		if err != nil {
			break
		}
		r := response{
			RequestId: conId + strconv.Itoa(i),
			Status:    resp.StatusCode,
			Body:      "<todo>",
			Headers:   resp.Header,
		}
		process(r)
	}
}
