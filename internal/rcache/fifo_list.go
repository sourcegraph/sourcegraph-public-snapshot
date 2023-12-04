package rcache

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// FIFOList holds the most recently inserted items, discarding older ones if the total item count goes over the configured size.
type FIFOList struct {
	key     string
	maxSize func() int
}

// NewFIFOList returns a FIFOList, storing only a fixed amount of elements, discarding old ones if needed.
func NewFIFOList(key string, size int) *FIFOList {
	return &FIFOList{
		key:     key,
		maxSize: func() int { return size },
	}
}

// NewFIFOListDynamic is like NewFIFOList except size will be called each time
// we enforce list size invariants.
func NewFIFOListDynamic(key string, size func() int) *FIFOList {
	l := &FIFOList{
		key:     key,
		maxSize: size,
	}
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
		if err := kv().LTrim(key, 0, 0); err != nil {
			return errors.Wrap(err, "failed to execute redis command LTRIM")
		}
		return nil
	}

	// O(1) because we're just adding a single element.
	if err := kv().LPush(key, b); err != nil {
		return errors.Wrap(err, "failed to execute redis command LPUSH")
	}

	// O(1) because the average case if just about dropping the last element.
	if err := kv().LTrim(key, 0, maxSize-1); err != nil {
		return errors.Wrap(err, "failed to execute redis command LTRIM")
	}
	return nil
}

// Size returns the number of elements in the list.
func (l *FIFOList) Size() (int, error) {
	key := l.globalPrefixKey()
	n, err := kv().LLen(key)
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute redis command LLEN")
	}
	return n, nil
}

// IsEmpty returns true if the number of elements in the list is 0.
func (l *FIFOList) IsEmpty() (bool, error) {
	size, err := l.Size()
	if err != nil {
		return false, err
	}
	return size == 0, nil
}

// MaxSize returns the capacity of the list.
func (l *FIFOList) MaxSize() int {
	maxSize := l.maxSize()
	if maxSize < 0 {
		return 0
	}
	return maxSize
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
	bs, err := kv().WithContext(ctx).LRange(key, from, to).ByteSlices()
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
