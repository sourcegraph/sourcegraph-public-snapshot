package gitlab

import "time"

type cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, b []byte)
	Delete(key string)
}

type cacheVal struct {
	// ProjIDs is the set of project IDs to which a GitLab user has access.
	ProjIDs map[int]struct{} `json:"repos"`

	// TTL is the ttl of the cache entry. This must be checked for equality in case the TTL has
	// changed (and the cache entry should therefore be invalidated).
	TTL time.Duration `json:"ttl"`
}
