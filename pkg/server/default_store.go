package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

var _ EntriesStore = (*defaultStore)(nil)

type defaultStore struct {
	file  *os.File
	cache map[string]string
}

func newDefaultStore(fileName string) EntriesStore {
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
	cache, err := readDnsEntries(fs)
	if err != nil {
		initServerErr(openErr)
	}
	// construct the cache
	return &defaultStore{
		file:  fs,
		cache: cache,
	}
}

func readDnsEntries(f *os.File) (map[string]string, error) {
	ret := make(map[string]string)
	fScnr := bufio.NewScanner(f)
	fScnr.Split(bufio.ScanLines)
	for fScnr.Scan() {
		tks := strings.Split(fScnr.Text(), ",")
		if len(tks) != 2 {
			return nil, errors.New("incorrect line format")
		}
		ret[tks[0]] = tks[1]
	}
	return ret, nil
}

func (ds *defaultStore) GetAddress(
	ctx context.Context,
	domainName string,
) (address string) {
	// read the cache only
	addr, _ := ds.cache[domainName]
	return addr
}

func (ds *defaultStore) AddEntry(
	ctx context.Context,
	domainName, address string,
) error {
	// write through
	if addr, exist := ds.cache[domainName]; exist {
		return fmt.Errorf(
			"entry (%s:%s) already exist",
			domainName,
			addr,
		)
	}
	if _, err := ds.file.WriteString(
		domainName + ":" + address,
	); err != nil {
		return err
	}

	// update the cache
	ds.cache[domainName] = address
	return nil
}

func (ds *defaultStore) ListEntries(
	ctx context.Context,
) map[string]string {
	return ds.cache
}
