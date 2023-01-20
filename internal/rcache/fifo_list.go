package rcache

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"go.uber.org/atomic"
)

// FIFOList holds the most recently inserted items, discarding older ones if the total item count goes over the configured size.
type FIFOList struct {
	key     string
	maxSize atomic.Int64 // invariant: non-negative integer
}

// NewFIFOList returns a FIFOList, storing only a fixed amount of elements, discarding old ones if needed.
func NewFIFOList(key string, size int) *FIFOList {
	l := &FIFOList{
		key: key,
	}
	// SetMaxSize will adjust size to be a non-negative integer.
	l.SetMaxSize(size)
	return l
}

// Insert b in the cache and drops the oldest inserted item if the size exceeds the configured limit.
func (l *FIFOList) Insert(b []byte) error {
	if !utf8.Valid(b) {
		errors.Newf("rcache: keys must be valid utf8", "key", b)
	}
	key := l.globalPrefixKey()

	// Special case maxSize 0 to mean keep the list empty. Used to handle
	// disabling.
	maxSize := l.MaxSize()
	if maxSize == 0 {
		if err := pool.LTrim(key, 0, 0); err != nil {
			return errors.Wrap(err, "failed to execute redis command LTRIM")
		}
		return nil
	}

	// O(1) because we're just adding a single element.
	if err := pool.LPush(key, b); err != nil {
		return errors.Wrap(err, "failed to execute redis command LPUSH")
	}

	// O(1) because the average case if just about dropping the last element.
	if err := pool.LTrim(key, 0, maxSize-1); err != nil {
		return errors.Wrap(err, "failed to execute redis command LTRIM")
	}
	return nil
}

func (l *FIFOList) Size() (int, error) {
	key := l.globalPrefixKey()
	n, err := pool.LLen(key)
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute redis command LLEN")
	}
	return n, nil
}

func (l *FIFOList) MaxSize() int {
	return int(l.maxSize.Load())
}

// SetMaxSize will change the size we truncate at. If maxSize is <= 0 the list
// will remain empty.
//
// Note: this won't cause truncation to happen, instead truncation is done on
// the next insert.
func (l *FIFOList) SetMaxSize(maxSize int) {
	if maxSize < 0 {
		maxSize = 0
	}
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
	// Return early if context is already cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	key := l.globalPrefixKey()
	bs, err := pool.WithContext(ctx).LRange(key, from, to).ByteSlices()
	if err != nil {
		// Return ctx error if it expired
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	if maxSize := l.MaxSize(); len(bs) > maxSize {
		bs = bs[:maxSize]
	}
	return bs, nil
}

func (l *FIFOList) globalPrefixKey() string {
	return fmt.Sprintf("%s:%s", globalPrefix, l.key)
}
