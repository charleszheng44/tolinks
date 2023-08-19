package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
		hlr, err := store.getHandler(context.TODO(), r.URL.Path)
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
		if r.Method == http.MethodGet && len(vals) == 0 {
			// 1. list all entries
			m := store.listAddrs()
			for dn, adr := range m {
				fmt.Fprintf(w, "%s:%s\n", dn, adr)
			}
			return
		}

		if r.Method == http.MethodGet && len(vals) != 0 {
			// 2. return desired address
			dn := vals.Get("domainName")
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

		// 3. TODO(charleszheng44): add new entry
		if r.Method == http.MethodPost {
			return
		}

		// 4. TODO (charleszheng44): update an existing entry
		if r.Method == http.MethodPut {
			return
		}

		// 5. TODO(charleszheng44): delete an entry
		if r.Method == http.MethodDelete {
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
			log.Println("tolinks proxy server listening on 127.0.0.1:80")
			return s.proxyServer.ListenAndServe()
		},
	)
	g.Go(
		func() error {
			log.Println("tolinks admin server listening on 127.0.0.1:8090")
			return s.adminServer.ListenAndServe()
		},
	)
	if err := g.Wait(); err != nil {
		log.Fatalf("server exit unexpectly: %v", err)
	}
}
