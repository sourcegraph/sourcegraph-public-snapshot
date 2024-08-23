// Copyright 2016 Google Inc. All rights reserved.
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
	"bytes"
	"fmt"
)

// candidateMatch is a candidate match for a substring.
//
// Note: a lot of these can be in memory, so think about fieldalignment when
// modify the fields of this structure.
type candidateMatch struct {
	substrBytes   []byte
	substrLowered []byte

	scoreWeight float64

	file      uint32
	symbolIdx uint32

	// Offsets are relative to the start of the filename or file contents.
	runeOffset  uint32
	byteOffset  uint32
	byteMatchSz uint32

	// bools at end for struct field alignment
	caseSensitive bool
	fileName      bool
	symbol        bool
}

// Matches content against the substring, and populates byteMatchSz on success
func (m *candidateMatch) matchContent(content []byte) bool {
	if m.caseSensitive {
		comp := bytes.Equal(m.substrBytes, content[m.byteOffset:m.byteOffset+uint32(len(m.substrBytes))])

		m.byteMatchSz = uint32(len(m.substrBytes))
		return comp
	} else {
		// It is tempting to try a simple ASCII based
		// comparison if possible, but we need more
		// information. Simple ASCII chars have unicode upper
		// case variants (the ASCII 'k' has the Kelvin symbol
		// as upper case variant). We can only degrade to
		// ASCII if we are sure that both the corpus and the
		// query is ASCII only
		sz, ok := caseFoldingEqualsRunes(m.substrLowered, content[m.byteOffset:])
		m.byteMatchSz = uint32(sz)
		return ok
	}
}

// matchIterator is a docIterator that produces candidateMatches for a given document
type matchIterator interface {
	docIterator

	candidates() []*candidateMatch

	// updateStats is called twice. After matchtree construction and after
	// searching is done. Implementations must take care to not report
	// statistics twice.
	updateStats(*Stats)
}

// noMatchTree is both matchIterator and matchTree that matches nothing.
type noMatchTree struct {
	Why string

	// Stats captures the work done to create the noMatchTree.
	Stats Stats
}

func (t *noMatchTree) String() string {
	return fmt.Sprintf("not(%q)", t.Why)
}

func (t *noMatchTree) candidates() []*candidateMatch {
	return nil
}

func (t *noMatchTree) nextDoc() uint32 {
	return maxUInt32
}

func (t *noMatchTree) prepare(uint32) {}

func (t *noMatchTree) matches(cp *contentProvider, cost int, known map[matchTree]bool) matchesState {
	return matchesNone
}

func (t *noMatchTree) updateStats(s *Stats) {
	s.Add(t.Stats)
	t.Stats = Stats{}
}

func (m *candidateMatch) String() string {
	return fmt.Sprintf("%d:%d", m.file, m.runeOffset)
}

type ngramDocIterator struct {
	leftPad  uint32
	rightPad uint32

	iter hitIterator
	ends []uint32

	// ngramLookups is how many lookups we did to create this iterator.
	ngramLookups int

	// mutable
	fileIdx    uint32
	matchCount int
}

// nextFileIndex returns the smallest index j of ends such that
// ends[j] > offset, assuming ends[f] <= offset.
func nextFileIndex(offset, f uint32, ends []uint32) uint32 {
	d := uint32(1)
	for f < uint32(len(ends)) && ends[f] <= offset {
		if f+d < uint32(len(ends)) && ends[f+d] <= offset {
			f += d
			d *= 2
		} else if d > 1 {
			d = d/4 + 1
		} else {
			f++
		}
	}
	return f
}

func (i *ngramDocIterator) nextDoc() uint32 {
	i.fileIdx = nextFileIndex(i.iter.first(), i.fileIdx, i.ends)
	if i.fileIdx >= uint32(len(i.ends)) {
		return maxUInt32
	}
	return i.fileIdx
}

func (i *ngramDocIterator) String() string {
	return fmt.Sprintf("ngram(L=%d,R=%d,%v)", i.leftPad, i.rightPad, i.iter)
}

func (i *ngramDocIterator) prepare(nextDoc uint32) {
	var start uint32
	if nextDoc > 0 {
		start = i.ends[nextDoc-1]
	}
	if start > 0 {
		i.iter.next(start + i.leftPad - 1)
	}
	i.fileIdx = nextDoc
}

func (i *ngramDocIterator) updateStats(s *Stats) {
	i.iter.updateStats(s)
	s.NgramMatches += i.matchCount
	s.NgramLookups += i.ngramLookups
	i.matchCount = 0
	i.ngramLookups = 0
}

func (i *ngramDocIterator) candidates() []*candidateMatch {
	if i.fileIdx >= uint32(len(i.ends)) {
		return nil
	}

	var fileStart uint32
	if i.fileIdx > 0 {
		fileStart = i.ends[i.fileIdx-1]
	}
	fileEnd := i.ends[i.fileIdx]

	var candidates []*candidateMatch
	for {
		p1 := i.iter.first()
		if p1 == maxUInt32 || p1 >= i.ends[i.fileIdx] {
			break
		}
		i.iter.next(p1)

		if p1 < i.leftPad+fileStart || p1+i.rightPad > fileEnd {
			continue
		}

		candidates = append(candidates, &candidateMatch{
			file:       uint32(i.fileIdx),
			runeOffset: p1 - fileStart - i.leftPad,
		})
	}
	i.matchCount += len(candidates)
	return candidates
}
