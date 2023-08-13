package serve

import (
	"log"
	"strconv"
)

func Handle(serve string) (string, string, string) {
	switch serve {
	case "ssh":
		port := handleSSH()
		return "ssh", "tcp", strconv.Itoa(port)
	default:
		log.Fatalf("invalid serve type, available: {ssh}")
	}
	return "http", "http", "8000"
}
