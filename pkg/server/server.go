package server

import (
	"log"
	"net/http"

	"github.com/miekg/dns"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	EntriesStore
	dnsServer   *dns.Server
	adminServer *http.Server
}

func newDnsServer(port string) *dns.Server {
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		// TODO:
		// only accept request with the prefix "to/"
	})
	return &dns.Server{
		Addr: ":" + port,
		Net:  "udp",
	}
}

func newAdminServer(port string) *http.Server {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// TODO: GET, POST, DELETE
	})
	return &http.Server{Addr: ":" + port}
}

func NewServer(dnsPort, adminPort, fileName string) *Server {
	// init the dns server
	// init the admin http server
	return &Server{
		EntriesStore: newDefaultStore(fileName),
		dnsServer:    newDnsServer(dnsPort),
		adminServer:  newAdminServer(adminPort),
	}
}

func (s *Server) Serve() {
	var g errgroup.Group
	g.Go(s.dnsServer.ListenAndServe)
	g.Go(s.adminServer.ListenAndServe)
	if err := g.Wait(); err != nil {
		log.Fatalf("server exit unexpectly: %v", err)
	}
}
