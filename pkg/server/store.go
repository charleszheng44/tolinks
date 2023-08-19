package server

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
)

var (
	ErrDomainNameNotFound = errors.New("Domain Name not found")
	ErrDomainNameExist    = errors.New("Domain Name exist")
)

type route struct {
	addr string
	hdlr http.Handler
}

type routes struct {
	sync.RWMutex
	m map[string]route
}

type store struct {
	file   *os.File
	routes *routes
}

func newStore(fileName string) *store {
	initServerErr := func(err error) {
		log.Fatalf("fail to initialize the server: %v", err)
	}
	// deserialize the data
	fs, openErr := os.OpenFile(
		fileName,
		os.O_APPEND|
			os.O_CREATE|
			os.O_RDWR,
		0644,
	)
	if openErr != nil {
		initServerErr(openErr)
	}
	routes, err := readDnsEntries(fs)
	if err != nil {
		initServerErr(openErr)
	}
	// construct the cache
	return &store{
		file:   fs,
		routes: routes,
	}
}

func readDnsEntries(f *os.File) (*routes, error) {
	m := make(map[string]route)
	fScnr := bufio.NewScanner(f)
	fScnr.Split(bufio.ScanLines)
	for fScnr.Scan() {
		tks := strings.Split(fScnr.Text(), ":")
		if len(tks) != 2 {
			return nil, errors.New("invalid input format")
		}
		url, err := url.Parse(tks[1])
		if err != nil {
			return nil, err
		}
		route := route{
			addr: tks[1],
			hdlr: httputil.NewSingleHostReverseProxy(url),
		}
		m[tks[0]] = route
	}
	return &routes{
		m: m,
	}, nil
}

func (ds *store) getHandler(
	ctx context.Context,
	domainName string,
) (http.Handler, error) {
	ds.routes.RLock()
	defer ds.routes.RUnlock()
	r, exist := ds.routes.m[domainName]
	if !exist {
		return nil, ErrDomainNameNotFound
	}
	return r.hdlr, nil
}

func (ds *store) getAddress(
	ctx context.Context,
	domainName string,
) string {
	// read the cache only
	ds.routes.RLock()
	defer ds.routes.Unlock()
	r, exist := ds.routes.m[domainName]
	if !exist {
		return ""
	}
	return r.addr
}

func (ds *store) addEntry(
	ctx context.Context,
	domainName, address string,
) error {
	// write through
	ds.routes.Lock()
	defer ds.routes.Unlock()
	if _, exist := ds.routes.m[domainName]; exist {
		return ErrDomainNameExist
	}

	u, err := url.Parse(address)
	if err != nil {
		return err
	}

	if _, err := ds.file.WriteString(domainName + ":" + address); err != nil {
		return err
	}

	// update the cache
	ds.routes.m[domainName] = route{
		addr: address,
		hdlr: httputil.NewSingleHostReverseProxy(u),
	}
	return nil
}

func (ds *store) listAddrs() map[string]string {
	ds.routes.RLock()
	defer ds.routes.RUnlock()
	ret := make(map[string]string)
	for n, r := range ds.routes.m {
		ret[n] = r.addr
	}
	return ret
}
