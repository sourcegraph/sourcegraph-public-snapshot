package cache

import "github.com/golang/groupcache/lru"

type Cache interface {
	Get(key lru.Key) (value interface{}, ok bool)
	Add(key lru.Key, value interface{})
}
