package githubutil

import "github.com/sourcegraph/httpcache"

// namespacedCache is a Cache wrapper that prepends namespace + ":" to
// all keys before invoking the corresponding underlying Cache's
// method.
//
// It is used to, for example, store cached items for multiple users
// separately to avoid leaking private information (the user's OAuth2
// token is the namespace).
type namespacedCache struct {
	namespace string
	httpcache.Cache
}

func (c namespacedCache) Get(key string) (responseBytes []byte, ok bool) {
	return c.Cache.Get(c.namespace + ":" + key)
}

func (c namespacedCache) Set(key string, responseBytes []byte) {
	c.Cache.Set(c.namespace+":"+key, responseBytes)
}

func (c namespacedCache) Delete(key string) {
	c.Cache.Delete(c.namespace + ":" + key)
}
