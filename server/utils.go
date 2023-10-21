package main

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var regex = regexp.MustCompile(`^[a-z\d](?:[a-z\d]|-[a-z\d]){0,38}$`)
var blockList = map[string]bool{"www": true, "jprq": true}

func validate(subdomain string) error {
	if len(subdomain) > 38 || len(subdomain) < 3 {
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

func parseHost(r io.Reader) (string, []byte, error) {
	buffer := make([]byte, 2048)
	size, err := r.Read(buffer)
	buffer = buffer[:size]
	if err != nil {
		return "", buffer, err
	}
	text := string(buffer)
	left := strings.Index(text, "Host: ")
	if left < 0 {
		left = strings.Index(text, "host: ")
	}
	if left < 0 {
		return "", buffer, fmt.Errorf("no host detected")
	}
	text = text[left+6:] // drops chars "Host: "
	right := strings.Index(text, "\n")
	if right < 0 {
		return "", buffer, fmt.Errorf("no host detected")
	}
	return strings.TrimSpace(text[:right]), buffer, nil
}

func writeResponse(conn io.WriteCloser, statusCode int, status string, message string) {
	response := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\nContent-Length: %d\r\n\r\n%s", statusCode, status, len(message), message)
	conn.Write([]byte(response))
	conn.Close()
}
