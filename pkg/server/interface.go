package server

import (
	"context"
)

type EntriesStore interface {
	GetAddress(
		ctx context.Context,
		domainName string,
	) (address string)

	AddEntry(
		ctx context.Context,
		domainName, address string,
	) error

	ListEntries(ctx context.Context) map[string]string
}
