package database

import (
	"sync"

	"github.com/dgraph-io/ristretto"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// DatabaseCache is an in-memory LRU cache of Database instances.
type DatabaseCache struct {
	cache *ristretto.Cache
}

type databaseCacheEntry struct {
	filename string
	db       *Database
	wg       sync.WaitGroup // user ref count
	once     sync.Once      // guards db.Close()
}

// close closes the cached database value after its refcount has dropped to zero.
// The underlying database's close method is guaranteed to be invoked only once.
func (entry *databaseCacheEntry) close() {
	entry.once.Do(func() {
		go func() {
			entry.wg.Wait()

			if err := entry.db.Close(); err != nil {
				log15.Error("Failed to close database", "filename", entry.filename, "err", err)
			}
		}()
	})
}

// NewDatabaseCache creates a Database instance cache with the given maximum size.
func NewDatabaseCache(size int64) (*DatabaseCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: size * 10,
		MaxCost:     size,
		BufferItems: 64,
		OnEvict: func(_, _ uint64, value interface{}, _ int64) {
			value.(*databaseCacheEntry).close()
		},
	})
	if err != nil {
		return nil, err
	}

	return &DatabaseCache{cache: cache}, nil
}

// WithDatabase invokes the given handler function with a Database instance either
// cached with the given filename, or created with the given openDatabase function.
// This method is goroutine-safe and the database instance is guaranteed to remain
// open until the handler has returned, regardless of the cache entry's eviction
// status.
func (c *DatabaseCache) WithDatabase(filename string, openDatabase func() (*Database, error), handler func(db *Database) error) error {
	if value, ok := c.cache.Get(filename); ok {
		entry := value.(*databaseCacheEntry)
		entry.wg.Add(1)
		defer entry.wg.Done()
		return handler(entry.db)
	}

	db, err := openDatabase()
	if err != nil {
		return err
	}

	entry := &databaseCacheEntry{filename: filename, db: db}
	entry.wg.Add(1)
	defer entry.wg.Done()

	if !c.cache.Set(filename, entry, 1) {
		defer entry.close()
	}

	return handler(entry.db)
}

// DocumentDataCache is an in-memory LRU cache of unmarshalled DocumentData instances.
type DocumentDataCache struct {
	cache *ristretto.Cache
}

// NewDocumentDataCache creates a DocumentData instance cache with the given maximum size.
// The size of the cache is determined by the number of field in each DocumentData value.
func NewDocumentDataCache(size int64) (*DocumentDataCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: size * 10,
		MaxCost:     size,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	return &DocumentDataCache{cache: cache}, nil
}

// GetOrCreate returns the document data cached at the given key or calls the given factory
// to create it. This method is goroutine-safe.
func (c *DocumentDataCache) GetOrCreate(key string, factory func() (types.DocumentData, error)) (types.DocumentData, error) {
	if value, ok := c.cache.Get(key); ok {
		return value.(types.DocumentData), nil
	}

	data, err := factory()
	if err != nil {
		return types.DocumentData{}, err
	}

	c.cache.Set(key, data, int64(1+len(data.HoverResults)+len(data.Monikers)+len(data.PackageInformation)+len(data.Ranges)))
	return data, nil
}

// ResultChunkDataCache is an in-memory LRU cache of unmarshalled ResultChunkData instances.
type ResultChunkDataCache struct {
	cache *ristretto.Cache
}

// ResultChunkDataCache creates a ResultChunkData instance cache with the given maximum size.
// The size of the cache is determined by the number of field in each ResultChunkData value.
func NewResultChunkDataCache(size int64) (*ResultChunkDataCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: size * 10,
		MaxCost:     size,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	return &ResultChunkDataCache{cache: cache}, nil
}

// GetOrCreate returns the result chunk data cached at the given key or calls the given factory
// to create it. This method is goroutine-safe.
func (c *ResultChunkDataCache) GetOrCreate(key string, factory func() (types.ResultChunkData, error)) (types.ResultChunkData, error) {
	if value, ok := c.cache.Get(key); ok {
		return value.(types.ResultChunkData), nil
	}

	data, err := factory()
	if err != nil {
		return types.ResultChunkData{}, err
	}

	c.cache.Set(key, data, int64(1+len(data.DocumentPaths)+len(data.DocumentIDRangeIDs)))
	return data, nil
}
