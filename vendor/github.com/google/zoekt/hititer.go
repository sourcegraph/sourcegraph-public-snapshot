// Copyright 2018 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zoekt

import (
	"encoding/binary"
	"fmt"
)

// hitIterator finds potential search matches, measured in offsets of
// the concatenation of all documents.
type hitIterator interface {
	// Return the first hit, or maxUInt32 if none.
	first() uint32

	// Skip until past limit. The argument maxUInt32 should be
	// treated specially.
	next(limit uint32)

	// Return how many bytes were read.
	updateStats(s *Stats)
}

// distanceHitIterator looks for hits at a fixed distance apart.
type distanceHitIterator struct {
	started  bool
	distance uint32
	i1       hitIterator
	i2       hitIterator
}

func (i *distanceHitIterator) String() string {
	return fmt.Sprintf("dist(%d, %v, %v)", i.distance, i.i1, i.i2)
}

func (i *distanceHitIterator) findNext() {
	for {
		var p1, p2 uint32
		p1 = i.i1.first()
		p2 = i.i2.first()
		if p1 == maxUInt32 || p2 == maxUInt32 {
			i.i1.next(maxUInt32)
			break
		}

		if p1+i.distance < p2 {
			i.i1.next(p2 - i.distance - 1)
		} else if p1+i.distance > p2 {
			i.i2.next(p1 + i.distance - 1)
		} else {
			break
		}
	}
}

func (i *distanceHitIterator) first() uint32 {
	if !i.started {
		i.findNext()
		i.started = true
	}
	return i.i1.first()
}

func (i *distanceHitIterator) updateStats(s *Stats) {
	i.i1.updateStats(s)
	i.i2.updateStats(s)
}

func (i *distanceHitIterator) next(limit uint32) {
	i.i1.next(limit)
	l2 := limit + i.distance

	if l2 < limit { // overflow.
		l2 = maxUInt32
	}
	i.i2.next(l2)
	i.findNext()
}

func (d *indexData) newDistanceTrigramIter(ng1, ng2 ngram, dist uint32, caseSensitive, fileName bool) (hitIterator, error) {
	if dist == 0 {
		return nil, fmt.Errorf("d == 0")
	}

	i1, err := d.trigramHitIterator(ng1, caseSensitive, fileName)
	if err != nil {
		return nil, err
	}
	i2, err := d.trigramHitIterator(ng2, caseSensitive, fileName)
	if err != nil {
		return nil, err
	}
	return &distanceHitIterator{
		i1:       i1,
		i2:       i2,
		distance: dist,
	}, nil
}

func (d *indexData) trigramHitIterator(ng ngram, caseSensitive, fileName bool) (hitIterator, error) {
	variants := []ngram{ng}
	if !caseSensitive {
		variants = generateCaseNgrams(ng)
	}

	iters := make([]hitIterator, 0, len(variants))
	for _, v := range variants {
		if fileName {
			blob := d.fileNameNgrams[v]
			if len(blob) > 0 {
				iters = append(iters, &inMemoryIterator{
					d.fileNameNgrams[v],
					v,
				})
			}
			continue
		}

		sec := d.ngrams[v]
		blob, err := d.readSectionBlob(sec)
		if err != nil {
			return nil, err
		}
		if len(blob) > 0 {
			iters = append(iters, newCompressedPostingIterator(blob, v))
		}
	}

	if len(iters) == 1 {
		return iters[0], nil
	}
	return &mergingIterator{
		iters: iters,
	}, nil
}

// inMemoryIterator is hitIterator that goes over an in-memory uint32 posting list.
type inMemoryIterator struct {
	postings []uint32
	what     ngram
}

func (i *inMemoryIterator) String() string {
	return fmt.Sprintf("mem(%s):%v", i.what, i.postings)
}

func (i *inMemoryIterator) first() uint32 {
	if len(i.postings) > 0 {
		return i.postings[0]
	}
	return maxUInt32
}

func (i *inMemoryIterator) updateStats(s *Stats) {
}

func (i *inMemoryIterator) next(limit uint32) {
	if limit == maxUInt32 {
		i.postings = nil
	}

	for len(i.postings) > 0 && i.postings[0] <= limit {
		i.postings = i.postings[1:]
	}
}

// compressedPostingIterator goes over a delta varint encoded posting
// list.
type compressedPostingIterator struct {
	blob, orig []byte
	_first     uint32
	what       ngram
}

func newCompressedPostingIterator(b []byte, w ngram) *compressedPostingIterator {
	d, sz := binary.Uvarint(b)
	return &compressedPostingIterator{
		_first: uint32(d),
		blob:   b[sz:],
		orig:   b,
		what:   w,
	}
}

func (i *compressedPostingIterator) String() string {
	return fmt.Sprintf("compressed(%s, %d, [%d bytes])", i.what, i._first, len(i.blob))
}

func (i *compressedPostingIterator) first() uint32 {
	return i._first
}

func (i *compressedPostingIterator) next(limit uint32) {
	if limit == maxUInt32 {
		i.blob = nil
		i._first = maxUInt32
		return
	}

	if i._first <= limit && len(i.blob) == 0 {
		i._first = maxUInt32
		return
	}

	for i._first <= limit && len(i.blob) > 0 {
		delta, sz := binary.Uvarint(i.blob)
		i._first += uint32(delta)
		i.blob = i.blob[sz:]
	}
}

func (i *compressedPostingIterator) updateStats(s *Stats) {
	s.IndexBytesLoaded += int64(len(i.orig) - len(i.blob))
}

// mergingIterator forms the merge of a set of hitIterators, to
// implement an OR operation at the hit level.
type mergingIterator struct {
	iters []hitIterator
}

func (i *mergingIterator) String() string {
	return fmt.Sprintf("merge:%v", i.iters)
}

func (i *mergingIterator) updateStats(s *Stats) {
	for _, j := range i.iters {
		j.updateStats(s)
	}
}

func (i *mergingIterator) first() uint32 {
	r := uint32(maxUInt32)
	for _, j := range i.iters {
		f := j.first()
		if f < r {
			r = f
		}
	}

	return r
}

func (i *mergingIterator) next(limit uint32) {
	for _, j := range i.iters {
		j.next(limit)
	}
}
