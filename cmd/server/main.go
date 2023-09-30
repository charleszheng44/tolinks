package main

import (
	"log"

	"github.com/charleszheng44/tolinks/pkg/server"
)

const (
	proxyPort = "8091"
	adminPort = "8090"
)

func main() {
	s := server.NewServer(proxyPort, adminPort, "/Users/charlesz/.tolinks/db")
	log.Printf("tolink server listening on :%s ...", proxyPort)
	s.Serve()
}
