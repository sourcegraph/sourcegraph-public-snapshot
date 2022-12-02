package rcache

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RecentList holds the most recently inserted items, discarding older ones if the total item count goes over the configured size.
type RecentList struct {
	key  string
	size int
}

// NewRecentList returns a RecentCache, storing only a fixed amount of elements, discarding old ones if needed.
func NewRecentList(key string, size int) *RecentList {
	return &RecentList{
		key:  key,
		size: size,
	}
}

// Insert b in the cache and drops the last recently inserted item if the size exceeds the configured limit.
func (l *RecentList) Insert(b []byte) error {
	c := pool.Get()
	defer c.Close()

	if !utf8.Valid(b) {
		errors.Newf("rcache: keys must be valid utf8", "key", b)
	}
	key := l.globalPrefixKey()

	// O(1) because we're just adding a single element.
	_, err := c.Do("LPUSH", key, b)
	if err != nil {
		return errors.Wrap(err, "failed to execute redis command LPUSH")
	}

	// O(1) because the average case if just about dropping the last element.
	_, err = c.Do("LTRIM", key, 0, l.size-1)
	if err != nil {
		return errors.Wrap(err, "failed to execute redis command LTRIM")
	}
	return nil
}

func (l *RecentList) Size() (int, error) {
	c := pool.Get()
	defer c.Close()

	key := l.globalPrefixKey()
	n, err := redis.Int(c.Do("LLEN", key))
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute redis command LLEN")
	}
	return n, nil
}

// All return all items stored in the RecentCache.
//
// This a O(n) operation, where n is the list size.
func (l *RecentList) All(ctx context.Context) ([][]byte, error) {
	return l.Slice(ctx, 0, -1)
}

// Slice return all items stored in the RecentCache between indexes from and to.
//
// This a O(n) operation, where n is the list size.
func (l *RecentList) Slice(ctx context.Context, from, to int) ([][]byte, error) {
	c, err := pool.GetContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get redis conn")
	}
	defer c.Close()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	key := l.globalPrefixKey()
	res, err := redis.Values(c.Do("LRANGE", key, from, to))
	if err != nil {
		return nil, err
	}
	bs, err := redis.ByteSlices(res, nil)
	if err != nil {
		return nil, err
	}
	if len(bs) > l.size {
		bs = bs[:l.size]
	}
	return bs, nil
}

func (l *RecentList) globalPrefixKey() string {
	return fmt.Sprintf("%s:%s", globalPrefix, l.key)
}
