package sqlite

import (
	"context"
	"errors"
	"os"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/cache"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/util"
)

// ErrUnknownDatabase occurs when a request for an unknown database is made.
var ErrUnknownDatabase = errors.New("unknown database")

// NewReaderCache creates a new reader cache. All readers share the same data cache with the
// given maximum capacity.
func NewReaderCache(dataCacheSize int) (cache.ReaderCache, error) {
	readerDataCache, err := cache.NewDataCache(dataCacheSize)
	if err != nil {
		return nil, err
	}

	cache := cache.NewReaderCache(func(filename string) (persistence.Reader, error) {
		// Ensure database exists prior to opening
		if exists, err := util.PathExists(filename); err != nil {
			return nil, err
		} else if !exists {
			return nil, ErrUnknownDatabase
		}

		reader, err := NewReader(context.Background(), filename, readerDataCache)
		if err != nil {
			return nil, err
		}

		// Check to see if the database exists after opening it. If it doesn't, then
		// the database file was deleted between the exists check and opening the
		// database and SQLite has created a new, empty database that is not yet been
		// written to disk.
		if exists, err := util.PathExists(filename); err != nil {
			return nil, err
		} else if !exists {
			reader.Close(nil)
			os.Remove(filename) // Possibly created on close
			return nil, ErrUnknownDatabase
		}

		return reader, nil
	})

	return cache, nil
}
