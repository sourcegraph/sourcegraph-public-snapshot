package limiter

func NewPrefixRedisStore(prefix string, store RedisStore) RedisStore {
	return &prefixRedisStore{
		prefix: prefix,
		store:  store,
	}
}

type prefixRedisStore struct {
	prefix string
	store  RedisStore
}

func (s *prefixRedisStore) Incrby(key string, val int) (int, error) {
	return s.store.Incrby(s.prefix+key, val)
}

func (s *prefixRedisStore) GetInt(key string) (int, error) {
	return s.store.GetInt(s.prefix + key)
}

func (s *prefixRedisStore) TTL(key string) (int, error) {
	return s.store.TTL(s.prefix + key)
}

func (s *prefixRedisStore) Expire(key string, ttlSeconds int) error {
	return s.store.Expire(s.prefix+key, ttlSeconds)
}
