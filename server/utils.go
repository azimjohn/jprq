package main

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var regex = regexp.MustCompile(`^[a-z0-9]+[a-z0-9\-]+[a-z0-9]$`)
var blockList = map[string]bool{"www": true, "jprq": true}

func validate(hostname string) error {
	domains := strings.Split(hostname, ".")
	if len(domains) != 3 {
		return errors.New("3.level.domain is expected")
	}
	subdomain := domains[0]
	if len(subdomain) > 42 || len(subdomain) < 3 {
		return errors.New("subdomain length must be between 3 and 42")
	}
	if blockList[subdomain] {
		return errors.New("subdomain is in deny list")
	}
	if !regex.MatchString(subdomain) {
		return errors.New("subdomain must be lowercase & alphanumeric")
	}
	return nil
}

func readLine(r io.Reader) (string, error) {
	var line []byte
	buffer := make([]byte, 1)
	for {
		if _, err := r.Read(buffer); err != nil {
			return "", err
		}
		line = append(line, buffer[0])
		if buffer[0] == '\n' {
			break
		}
		if len(buffer) > 4096 {
			return "", errors.New("host search limit reached")
		}
	}
	return string(line), nil
}

func writeResponse(conn io.WriteCloser, statusCode int, status string, message string) {
	response := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\nContent-Length: %d\r\n\r\n%s", statusCode, status, len(message), message)
	conn.Write([]byte(response))
	conn.Close()
}
