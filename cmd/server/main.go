package main

import (
	"github.com/charleszheng44/tolinks/pkg/server"
)

func main() {
	s := server.NewServer("80", "8090", "/Users/charlesz/.tolinks/db")
	s.Serve()
}
