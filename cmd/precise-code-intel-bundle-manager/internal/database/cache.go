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
	db       Database
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
func NewDatabaseCache(size int64) (*DatabaseCache, *ristretto.Metrics, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: size * 10,
		MaxCost:     size,
		BufferItems: 64,
		Metrics:     true,
		OnEvict: func(_, _ uint64, value interface{}, _ int64) {
			value.(*databaseCacheEntry).close()
		},
	})
	if err != nil {
		return nil, nil, err
	}

	return &DatabaseCache{cache: cache}, cache.Metrics, nil
}

// WithDatabase invokes the given handler function with a Database instance either
// cached with the given filename, or created with the given openDatabase function.
// This method is goroutine-safe and the database instance is guaranteed to remain
// open until the handler has returned, regardless of the cache entry's eviction
// status.
func (c *DatabaseCache) WithDatabase(filename string, openDatabase func() (Database, error), handler func(db Database) error) error {
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

// DocumentCache is an in-memory LRU cache of unmarshalled DocumentData instances.
type DocumentCache struct {
	cache *ristretto.Cache
}

// NewDocumentCache creates a DocumentData instance cache with the given maximum size.
// The size of the cache is determined by the number of field in each DocumentData value.
func NewDocumentCache(size int64) (*DocumentCache, *ristretto.Metrics, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: size * 10,
		MaxCost:     size,
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		return nil, nil, err
	}

	return &DocumentCache{cache: cache}, cache.Metrics, nil
}

// GetOrCreate returns the document data cached at the given key or calls the given factory
// to create it. This method is goroutine-safe.
func (c *DocumentCache) GetOrCreate(key string, factory func() (types.DocumentData, error)) (types.DocumentData, error) {
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

// ResultChunkCache is an in-memory LRU cache of unmarshalled ResultChunkData instances.
type ResultChunkCache struct {
	cache *ristretto.Cache
}

// ResultChunkCache creates a ResultChunkData instance cache with the given maximum size.
// The size of the cache is determined by the number of field in each ResultChunkData value.
func NewResultChunkCache(size int64) (*ResultChunkCache, *ristretto.Metrics, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: size * 10,
		MaxCost:     size,
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		return nil, nil, err
	}

	return &ResultChunkCache{cache: cache}, cache.Metrics, nil
}

// GetOrCreate returns the result chunk data cached at the given key or calls the given factory
// to create it. This method is goroutine-safe.
func (c *ResultChunkCache) GetOrCreate(key string, factory func() (types.ResultChunkData, error)) (types.ResultChunkData, error) {
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
