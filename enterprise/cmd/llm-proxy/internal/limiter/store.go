package limiter

// RedisStore is the backend for tracking limiter state.
type RedisStore interface {
	// Incr incremeents a key's value
	Incr(key string) (int, error)
	// TTL provides seconds TTL
	TTL(key string) (int, error)
	// Expire configures a key's TTL
	Expire(key string, ttlSeconds int) error
}
