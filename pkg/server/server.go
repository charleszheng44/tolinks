package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	// proxyServer forwards requests to correct destinations.
	proxyServer *http.Server
	// adminServer handles admin requests sent from client,
	// like add/list/delete route entries.
	adminServer *http.Server
}

func newProxyServer(port string, store *store) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dn := strings.TrimPrefix(r.URL.Path, "/")
		log.Printf("handling proxy request for domain %s", dn)
		hlr, err := store.getHandler(context.TODO(), dn)
		// domain name not found
		if err != nil && err == ErrDomainNameNotFound {
			http.NotFound(w, r)
			return
		}

		// all other errors
		if err != nil {
			http.Error(
				w, fmt.Sprintf("%v", err),
				http.StatusInternalServerError,
			)
			return
		}

		// forward the request
		hlr.ServeHTTP(w, r)
	})
	return &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: mux,
	}
}

func newAdminServer(port string, store *store) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		vals := r.URL.Query()
		// 1. list all entries
		if r.Method == http.MethodGet && len(vals) == 0 {
			log.Printf("handling request of listing address from %s", r.RemoteAddr)
			m := store.listAddrs()
			for dn, adr := range m {
				fmt.Fprintf(w, "%s:%s\n", dn, adr)
			}
			return
		}

		// 2. return desired address
		if r.Method == http.MethodGet && len(vals) != 0 {
			dn := vals.Get("domainName")
			log.Printf(
				"handling request of getting address for %s from %s",
				dn, r.RemoteAddr,
			)
			adr := store.getAddress(context.TODO(), dn)
			if adr == "" {
				http.Error(
					w, ErrDomainNameNotFound.Error(),
					http.StatusInternalServerError,
				)
				return
			}
			fmt.Fprintf(w, "%s:%s", dn, adr)
			return
		}

		// 3. add new entry
		if r.Method == http.MethodPost {
			dn := vals.Get("domainName")
			adr := vals.Get("address")
			log.Printf(
				"handling request of add/update entry %s:%s from %s",
				dn, adr, r.RemoteAddr,
			)
			if err := store.addOrUpdateEntry(
				context.TODO(),
				dn,
				adr,
			); err != nil {
				http.Error(
					w,
					fmt.Sprintf("failed to add new entry: %v", err),
					http.StatusInternalServerError,
				)
			}
			return
		}

		// 4. delete an entry
		if r.Method == http.MethodDelete {
			if err := store.deleteEntry(
				context.TODO(),
				vals.Get("domainName"),
			); err != nil {
				http.Error(
					w,
					fmt.Sprintf("failed to delete the entry: %v", err),
					http.StatusInternalServerError,
				)
			}
			return
		}
	})
	return &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: mux,
	}
}

func NewServer(proxyPort, adminPort, fileName string) *Server {
	store := newStore(fileName)
	return &Server{
		proxyServer: newProxyServer(proxyPort, store),
		adminServer: newAdminServer(adminPort, store),
	}
}

func (s *Server) Serve() {
	var g errgroup.Group
	g.Go(
		func() error {
			log.Printf(
				"tolinks proxy server listening on %v\n",
				s.proxyServer.Addr,
			)
			return s.proxyServer.ListenAndServe()
		},
	)
	g.Go(
		func() error {
			log.Printf(
				"tolinks admin server listening on %v\n",
				s.adminServer.Addr,
			)
			return s.adminServer.ListenAndServe()
		},
	)
	if err := g.Wait(); err != nil {
		log.Fatalf("server exit unexpectly: %v", err)
	}
}
