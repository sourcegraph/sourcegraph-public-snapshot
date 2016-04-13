// Copyright 2013 Sean Treadway, SoundCloud Ltd. All rights reserved.  Use of
// this source code is governed by a BSD-style license that can be found in the
// LICENSE file.

/*
Package quantile implements a streaming quantile estimator. The implementation
is based on "Effective Computation of Biased Quantiles over Data Streams"
(Cormode, Korn, Muthukrishnan, Srivastava) to provide a space and time
efficient estimator for online quantile estimation.

For the normal distribution of 10^9 elements, a tolerance for 0.99th percentile
at 0.001 uses under 1000 bins at 32 bytes per bin.
*/
package quantile

import (
	"math"
	"sort"
)

type Estimate interface {
	// Delta calculates the acceptable difference in ranks between two values.
	// It is used to remove redundant values during compression.
	Delta(rank, observations float64) float64
}

type bias struct {
	tolerance float64
}

func (b bias) Delta(rank, observations float64) float64 {
	return 2 * b.tolerance * rank
}

// Unknown produces estimations for all possible quantiles at this error tolerance.
// It uses significantly more space and time than when you know the quantiles
// you wish to estimate.
//
// The Known estimation should be used when you know which quantiles you will be
// querying.
func Unknown(tolerance float64) Estimate {
	return bias{tolerance: tolerance}
}

type target struct {
	q  float64 // targeted quantile
	f1 float64 // cached coefficient for fi  q*n <= rank <= n
	f2 float64 // cached coefficient for fii 0 <= rank <= q*n
}

func (t target) Delta(rank, observations float64) float64 {
	if rank <= math.Floor(t.q*observations) {
		return t.f2 * (observations - rank)
	}
	return t.f1 * rank
}

// Known produces a optimal space usage for estimations at the given quantile and error tolerance.
//
// Quantiles not known ahead of time can also be queried, but at a lower accuracy.
func Known(quantile, tolerance float64) Estimate {
	return target{
		q:  quantile,
		f1: 2 * tolerance / quantile,
		f2: 2 * tolerance / (1 - quantile),
	}
}

// the tuple and list element
type item struct {
	v     float64
	rank  float64
	delta float64
	next  *item
}

type Estimator struct {
	// linked list data structure "S", bookeeping in observe/recycle
	head  *item
	items int

	// float64 avoids conversion during invariant checks
	observations float64

	// used to calculate ƒ(r,n)
	invariants []Estimate

	// batching of updates
	buffer []float64

	// free list
	pool chan *item
}

var defaultInvariants = []Estimate{Unknown(0.1)}

// New allocates a new estimator tolerating the minimum of the invariants provided.
//
// When you know how much error you can tolerate in the quantiles you will
// query, use a Known estimation for each quantile you will query.  For
// example:
//
//    quantile.New(quantile.Known(0.50, 0.01), quantile.Known(0.95, 0.001), quantile.Known(0.99, 0.0005))
//
// When you will query for multiple different quantiles, and know the error
// tolerance, use the Bias invariant.  For example:
//
//    quantile.New(quantile.Unknown(0.01))
//
// Targeted estimators consume significantly less resources than Biased estimators.
//
// Passing no parameters will create an estimator that has a tolerance of 0.1, equivalent to:
//
//   quantile.New(quantile.Unknown(0.1))
//
// Estimators are not safe to use from multiple goroutines.
func New(invariants ...Estimate) *Estimator {
	if len(invariants) == 0 {
		invariants = defaultInvariants
	}

	return &Estimator{
		invariants: invariants,
		buffer:     make([]float64, 0, 512),
		pool:       make(chan *item, 1024),
	}
}

// Add buffers a new sample, committing and compressing the data structure
// when the buffer is full.
func (est *Estimator) Add(value float64) {
	est.buffer = append(est.buffer, value)
	if len(est.buffer) == cap(est.buffer) {
		est.flush()
	}
}

// Get finds a value within (quantile - tolerance) * n <= value <= (quantile + tolerance) * n
// or 0 if no values have been observed.
func (est *Estimator) Get(quantile float64) float64 {
	if est.observations == 0 && len(est.buffer) == 0 {
		return 0
	}

	est.flush()

	cur := est.head
	if cur == nil {
		return 0
	}

	midrank := math.Floor(quantile * est.observations)
	maxrank := midrank + math.Floor(est.invariant(midrank, est.observations)/2)

	rank := 0.0
	for cur.next != nil {
		rank += cur.rank
		if rank+cur.next.rank+cur.next.delta > maxrank {
			return cur.v
		}
		cur = cur.next
	}
	return cur.v
}

// Samples returns the number of values this estimator has sampled.
func (est *Estimator) Samples() int {
	return int(est.observations) + len(est.buffer)
}

// ƒ(r,n) = minⁱ(ƒⁱ(r,n))
func (est *Estimator) invariant(rank float64, n float64) float64 {
	min := (n + 1)
	for _, f := range est.invariants {
		if delta := f.Delta(rank, n); delta < min {
			min = delta
		}
	}
	return math.Floor(min)
}

func (est *Estimator) observe(v float64, rank, delta float64, next *item) *item {
	est.observations++
	est.items++

	// reuse or allocate
	select {
	case old := <-est.pool:
		old.v = v
		old.rank = rank
		old.delta = delta
		old.next = next
		return old
	default:
		return &item{
			v:     v,
			rank:  rank,
			delta: delta,
			next:  next,
		}
	}

	panic("unreachable")
}

func (est *Estimator) recycle(old *item) {
	est.items--
	select {
	case est.pool <- old:
	default:
	}
}

// merges the batch
func (est *Estimator) update(batch []float64) {
	// initial data
	if est.head == nil {
		est.head = est.observe(batch[0], 1, 0, nil)
		batch = batch[1:]
	}

	rank := 0.0
	cur := est.head
	for _, v := range batch {
		// min
		if v < est.head.v {
			est.head = est.observe(v, 1, 0, est.head)
			cur = est.head
			continue
		}

		// cursor
		for cur.next != nil && cur.next.v < v {
			rank += cur.rank
			cur = cur.next
		}

		// max
		if cur.next == nil {
			cur.next = est.observe(v, 1, 0, nil)
			continue
		}

		cur.next = est.observe(v, 1, est.invariant(rank, est.observations)-1, cur.next)
	}
}

func (est *Estimator) compress() {
	rank := 0.0
	cur := est.head
	for cur != nil && cur.next != nil {
		if cur.rank+cur.next.rank+cur.next.delta <= est.invariant(rank, est.observations) {
			// merge with previous/head
			removed := cur.next

			cur.v = removed.v
			cur.rank += removed.rank
			cur.delta = removed.delta
			cur.next = removed.next

			est.recycle(removed)
		}
		rank += cur.rank
		cur = cur.next
	}
}

func (est *Estimator) flush() {
	sort.Float64Slice(est.buffer).Sort()
	est.update(est.buffer)
	est.buffer = est.buffer[0:0]
	est.compress()
}
