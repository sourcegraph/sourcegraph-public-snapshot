package cache

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
)

// TODO(efritz) - document
type ReaderCache interface {
	// TODO(efritz) - document
	WithReader(ctx context.Context, key string, f HandlerFunc) error
}

// TODO(efritz) - document
type HandlerFunc func(reader persistence.Reader) error

// TODO(efritz) - document
type OpenReaderFunc func(key string) (persistence.Reader, error)

// TODO(efritz) - document
func NewReaderCache(factory OpenReaderFunc) ReaderCache {
	// TODO(efritz) - implement
	return nil
}
