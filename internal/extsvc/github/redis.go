package github

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, b []byte)
}

type NewCacheFunc func(key string, ttl int) Cache
