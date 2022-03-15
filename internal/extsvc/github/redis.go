package github

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, b []byte)
}

type NewCacheFactory func(key string, ttl int) Cache
