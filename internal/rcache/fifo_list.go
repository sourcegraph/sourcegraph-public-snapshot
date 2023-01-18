package rcache

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"go.uber.org/atomic"
)

// FIFOList holds the most recently inserted items, discarding older ones if the total item count goes over the configured size.
type FIFOList struct {
	key     string
	maxSize *atomic.Int64
}

// NewFIFOList returns a FIFOList, storing only a fixed amount of elements, discarding old ones if needed.
func NewFIFOList(key string, size int) *FIFOList {
	return &FIFOList{
		key:     key,
		maxSize: atomic.NewInt64(int64(size)),
	}
}

// Insert b in the cache and drops the oldest inserted item if the size exceeds the configured limit.
func (l *FIFOList) Insert(b []byte) error {
	c := poolGet()
	defer c.Close()

	if !utf8.Valid(b) {
		errors.Newf("rcache: keys must be valid utf8", "key", b)
	}
	key := l.globalPrefixKey()

	// Special case maxSize 0 to mean keep the list empty. Used to handle
	// disabling.
	if l.maxSize.Load() == 0 {
		_, err := c.Do("LTRIM", key, 0, 0)
		if err != nil {
			return errors.Wrap(err, "failed to execute redis command LTRIM")
		}
		return nil
	}

	// O(1) because we're just adding a single element.
	_, err := c.Do("LPUSH", key, b)
	if err != nil {
		return errors.Wrap(err, "failed to execute redis command LPUSH")
	}

	// O(1) because the average case if just about dropping the last element.
	_, err = c.Do("LTRIM", key, 0, l.maxSize.Load()-1)
	if err != nil {
		return errors.Wrap(err, "failed to execute redis command LTRIM")
	}
	return nil
}

func (l *FIFOList) Size() (int, error) {
	c := poolGet()
	defer c.Close()

	key := l.globalPrefixKey()
	n, err := redis.Int(c.Do("LLEN", key))
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute redis command LLEN")
	}
	return n, nil
}

func (l *FIFOList) MaxSize() int {
	return int(l.maxSize.Load())
}

// SetMaxSize will change the size we truncate at.
//
// Note: this won't cause truncation to happen, instead truncation is done on
// the next insert.
func (l *FIFOList) SetMaxSize(maxSize int) {
	l.maxSize.Store(int64(maxSize))
}

// All return all items stored in the FIFOList.
//
// This a O(n) operation, where n is the list size.
func (l *FIFOList) All(ctx context.Context) ([][]byte, error) {
	return l.Slice(ctx, 0, -1)
}

// Slice return all items stored in the FIFOlist between indexes from and to.
//
// This a O(n) operation, where n is the list size.
func (l *FIFOList) Slice(ctx context.Context, from, to int) ([][]byte, error) {
	c, err := poolGetContext(ctx)
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
	if maxSize := int(l.maxSize.Load()); len(bs) > maxSize {
		bs = bs[:maxSize]
	}
	return bs, nil
}

func (l *FIFOList) globalPrefixKey() string {
	return fmt.Sprintf("%s:%s", globalPrefix, l.key)
}
