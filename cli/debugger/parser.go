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
		req, err := http.ReadRequest(bufio.NewReader(r))
		if err != nil {
			fmt.Println("[debugger] error parsing http request", err)
			break
		}
		r := request{
			Id:      conId + "000" + strconv.Itoa(i),
			Method:  req.Method,
			URL:     req.URL.String(),
			Headers: req.Header,
		}

		if length := parseContentLength(r.Headers); length > 0 && length < 65536 {
			body, _ := io.ReadAll(req.Body)
			r.Body = string(body)
		} else {
			io.Copy(io.Discard, req.Body)
		}
		process(r)
	}
}

func parseResponses(r io.Reader, conId string, process func(interface{})) {
	for i := 0; ; i++ {
		resp, err := http.ReadResponse(bufio.NewReader(r), nil)
		if err != nil {
			fmt.Println("[debugger] error parsing http response", err)
			break
		}
		r := response{
			RequestId: conId + "000" + strconv.Itoa(i),
			Status:    resp.StatusCode,
			Headers:   resp.Header,
		}

		if length := parseContentLength(r.Headers); length > 0 && length < 65536 {
			body, _ := io.ReadAll(resp.Body)
			r.Body = string(body)
		} else {
			io.Copy(io.Discard, resp.Body)
		}
		process(r)
	}
}

func parseContentLength(headers http.Header) int {
	if header := headers["Content-Length"]; len(header) > 0 {
		if length, err := strconv.Atoi(header[0]); err == nil {
			return length
		}
	}
	return 0
}
