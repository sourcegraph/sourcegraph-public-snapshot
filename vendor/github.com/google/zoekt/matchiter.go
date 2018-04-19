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
	"sort"
	"unicode/utf8"

	"github.com/google/zoekt/query"
)

// candidateMatch is a candidate match for a substring.
type candidateMatch struct {
	caseSensitive bool
	fileName      bool

	substrBytes   []byte
	substrLowered []byte

	file uint32

	// Offsets are relative to the start of the filename or file contents.
	runeOffset  uint32
	byteOffset  uint32
	byteMatchSz uint32
}

// Matches content against the substring, and populates byteMatchSz on success
func (m *candidateMatch) matchContent(content []byte) bool {
	if m.caseSensitive {
		comp := bytes.Compare(m.substrBytes, content[m.byteOffset:m.byteOffset+uint32(len(m.substrBytes))]) == 0

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

// line returns the line holding the match. If the match starts with
// the newline ending line M, we return M.  The line is characterized
// by its linenumber (base-1, byte index of line start, byte index of
// line end).  The line end is the index of a newline, or the filesize
// (if matching the last line of the file.)
func (m *candidateMatch) line(newlines []uint32, fileSize uint32) (lineNum, lineStart, lineEnd int) {
	idx := sort.Search(len(newlines), func(n int) bool {
		return newlines[n] >= m.byteOffset
	})

	end := int(fileSize)
	if idx < len(newlines) {
		end = int(newlines[idx])
	}

	start := 0
	if idx > 0 {
		start = int(newlines[idx-1] + 1)
	}

	return idx + 1, start, end
}

// matchIterator is a docIterator that produces candidateMatches for a given document
type matchIterator interface {
	docIterator

	candidates() []*candidateMatch
}

// noMatchTree is both matchIterator and matchTree that matches nothing.
type noMatchTree struct {
	Why string
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

func (t *noMatchTree) matches(cp *contentProvider, cost int, known map[matchTree]bool) (bool, bool) {
	return false, true
}

func (m *candidateMatch) String() string {
	return fmt.Sprintf("%d:%d", m.file, m.runeOffset)
}

type ngramDocIterator struct {
	leftPad  uint32
	rightPad uint32

	iter hitIterator
	ends []uint32

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

type trimBySectionMatchIter struct {
	matchIterator

	patternSize  uint32
	fileEndRunes []uint32

	// mutable
	doc      uint32
	sections []DocumentSection
}

func (i *trimBySectionMatchIter) String() string {
	return fmt.Sprintf("trimSection(sz=%d, %v)", i.patternSize, i.matchIterator)
}

func (d *indexData) newTrimByDocSectionIter(q *query.Substring, iter matchIterator) *trimBySectionMatchIter {
	return &trimBySectionMatchIter{
		matchIterator: iter,
		patternSize:   uint32(utf8.RuneCountInString(q.Pattern)),
		fileEndRunes:  d.fileEndRunes,
		sections:      d.runeDocSections,
	}
}

func (i *trimBySectionMatchIter) prepare(doc uint32) {
	i.matchIterator.prepare(doc)
	i.doc = doc

	var fileStart uint32
	if doc > 0 {
		fileStart = i.fileEndRunes[doc-1]
	}

	for len(i.sections) > 0 && i.sections[0].Start < fileStart {
		i.sections = i.sections[1:]
	}
}

func (i *trimBySectionMatchIter) candidates() []*candidateMatch {
	var fileStart uint32
	if i.doc > 0 {
		fileStart = i.fileEndRunes[i.doc-1]
	}

	ms := i.matchIterator.candidates()
	trimmed := ms[:0]
	for len(i.sections) > 0 && len(ms) > 0 {
		start := fileStart + ms[0].runeOffset
		end := start + i.patternSize
		if start >= i.sections[0].End {
			i.sections = i.sections[1:]
			continue
		}

		if start < i.sections[0].Start {
			ms = ms[1:]
			continue
		}

		// here we have: sec.Start <= start < sec.End
		if end <= i.sections[0].End {
			// complete match falls inside section.
			trimmed = append(trimmed, ms[0])
		}

		ms = ms[1:]
	}
	return trimmed
}
