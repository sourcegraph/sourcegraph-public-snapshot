package updater

import "src.sourcegraph.com/sourcegraph/platform/storage"

const (
	cacheBucket = "httpCache"
)

// kvCache implements httpcache.Cache using storage.System.
type kvCache struct {
	kv storage.System
}

func (c kvCache) Get(key string) (responseBytes []byte, ok bool) {
	b, err := c.kv.Get(cacheBucket, key)
	return b, err == nil
}

func (c kvCache) Set(key string, responseBytes []byte) {
	c.kv.Put(cacheBucket, key, responseBytes)
}

func (c kvCache) Delete(key string) {
	c.kv.Delete(cacheBucket, key)
}
