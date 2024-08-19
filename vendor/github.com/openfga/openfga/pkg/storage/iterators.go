package storage

import (
	"context"
	"errors"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
)

// ErrIteratorDone is returned when the iterator has finished iterating through all the items.
var ErrIteratorDone = errors.New("iterator done")

// Iterator is a generic interface defining methods for
// iterating over a collection of items of type T.
type Iterator[T any] interface {
	// Next will return the next available
	// item or ErrIteratorDone if no more
	// items are available.
	Next(ctx context.Context) (T, error)

	// Stop terminates iteration over
	// the underlying iterator.
	Stop()
}

// TupleIterator is an iterator for [*openfgav1.Tuple](s).
// It is closed by explicitly calling [Iterator.Stop] or by calling
// [Iterator.Next] until it returns an [ErrIteratorDone] error.
type TupleIterator = Iterator[*openfgav1.Tuple]

// TupleKeyIterator is an iterator for [*openfgav1.TupleKey](s). It is closed by
// explicitly calling [Iterator.Stop] or by calling [Iterator.Next] until it
// returns an [ErrIteratorDone] error.
type TupleKeyIterator = Iterator[*openfgav1.TupleKey]

type emptyTupleIterator struct{}

var _ TupleIterator = (*emptyTupleIterator)(nil)

// Next see [Iterator.Next].
func (e *emptyTupleIterator) Next(ctx context.Context) (*openfgav1.Tuple, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return nil, ErrIteratorDone
}

// Stop see [Iterator.Stop].
func (e *emptyTupleIterator) Stop() {}

type combinedIterator[T any] struct {
	iters []Iterator[T]
}

// Next see [Iterator.Next].
func (c *combinedIterator[T]) Next(ctx context.Context) (T, error) {
	for i, iter := range c.iters {
		if iter == nil {
			continue
		}
		val, err := iter.Next(ctx)
		if err != nil {
			if !errors.Is(err, ErrIteratorDone) {
				return val, err
			}
			c.iters[i] = nil // End of this iterator.
			continue
		}

		return val, nil
	}

	// All iterators ended.
	var val T
	return val, ErrIteratorDone
}

// Stop see [Iterator.Stop].
func (c *combinedIterator[T]) Stop() {
	for _, iter := range c.iters {
		if iter != nil {
			iter.Stop()
		}
	}
}

// NewCombinedIterator takes generic iterators of a given type T
// and combines them into a single iterator that yields all the
// values from all iterators. Duplicates can be returned.
func NewCombinedIterator[T any](iters ...Iterator[T]) Iterator[T] {
	return &combinedIterator[T]{iters}
}

// NewStaticTupleIterator returns a [TupleIterator] that iterates over the provided slice.
func NewStaticTupleIterator(tuples []*openfgav1.Tuple) TupleIterator {
	iter := &staticIterator[*openfgav1.Tuple]{
		items: tuples,
	}

	return iter
}

// NewStaticTupleKeyIterator returns a [TupleKeyIterator] that iterates over the provided slice.
func NewStaticTupleKeyIterator(tupleKeys []*openfgav1.TupleKey) TupleKeyIterator {
	iter := &staticIterator[*openfgav1.TupleKey]{
		items: tupleKeys,
	}

	return iter
}

type tupleKeyIterator struct {
	iter TupleIterator
}

var _ TupleKeyIterator = (*tupleKeyIterator)(nil)

// Next see [Iterator.Next].
func (t *tupleKeyIterator) Next(ctx context.Context) (*openfgav1.TupleKey, error) {
	tuple, err := t.iter.Next(ctx)
	return tuple.GetKey(), err
}

// Stop see [Iterator.Stop].
func (t *tupleKeyIterator) Stop() {
	t.iter.Stop()
}

// NewTupleKeyIteratorFromTupleIterator takes a [TupleIterator] and yields
// all the [*openfgav1.TupleKey](s) from it as a [TupleKeyIterator].
func NewTupleKeyIteratorFromTupleIterator(iter TupleIterator) TupleKeyIterator {
	return &tupleKeyIterator{iter}
}

type staticIterator[T any] struct {
	items []T
}

// Next see [Iterator.Next].
func (s *staticIterator[T]) Next(ctx context.Context) (T, error) {
	var val T

	if ctx.Err() != nil {
		return val, ctx.Err()
	}

	if len(s.items) == 0 {
		return val, ErrIteratorDone
	}

	next, rest := s.items[0], s.items[1:]
	s.items = rest

	return next, nil
}

// Stop see [Iterator.Stop].
func (s *staticIterator[T]) Stop() {}

// TupleKeyFilterFunc is a filter function that is used to filter out
// tuples from a [TupleKeyIterator] that don't meet certain criteria.
// Implementations should return true if the tuple should be returned
// and false if it should be filtered out.
type TupleKeyFilterFunc func(tupleKey *openfgav1.TupleKey) bool

type filteredTupleKeyIterator struct {
	iter   TupleKeyIterator
	filter TupleKeyFilterFunc
}

var _ TupleKeyIterator = &filteredTupleKeyIterator{}

// Next returns the next most tuple in the underlying iterator that meets
// the filter function this iterator was constructed with.
func (f *filteredTupleKeyIterator) Next(ctx context.Context) (*openfgav1.TupleKey, error) {
	for {
		tuple, err := f.iter.Next(ctx)
		if err != nil {
			return nil, err
		}

		if f.filter(tuple) {
			return tuple, nil
		}
	}
}

// Stop see [Iterator.Stop].
func (f *filteredTupleKeyIterator) Stop() {
	f.iter.Stop()
}

// NewFilteredTupleKeyIterator returns a [TupleKeyIterator] that filters out all
// [*openfgav1.Tuple](s) that don't meet the conditions of the provided [TupleKeyFilterFunc].
func NewFilteredTupleKeyIterator(iter TupleKeyIterator, filter TupleKeyFilterFunc) TupleKeyIterator {
	return &filteredTupleKeyIterator{
		iter,
		filter,
	}
}
