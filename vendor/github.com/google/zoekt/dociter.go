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
	"log"
	"sort"

	"github.com/google/zoekt/query"
)

var _ = log.Println

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

func (m *candidateMatch) String() string {
	return fmt.Sprintf("%d:%d", m.file, m.runeOffset)
}

func (m *candidateMatch) matchContent(content []byte) bool {
	if m.caseSensitive {
		comp := bytes.Compare(m.substrBytes, content[m.byteOffset:m.byteOffset+uint32(len(m.substrBytes))]) == 0
		return comp
	} else {
		// It is tempting to try a simple ASCII based
		// comparison if possible, but we need more
		// information. Simple ASCII chars have unicode upper
		// case variants (the ASCII 'k' has the Kelvin symbol
		// as upper case variant). We can only degrade to
		// ASCII if we are sure that both the corpus and the
		// query is ASCII only
		return caseFoldingEqualsRunes(m.substrLowered, content[m.byteOffset:])
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

type ngramDocIterator struct {
	query *query.Substring

	leftPad  uint32
	rightPad uint32
	distance uint32
	first    []uint32
	last     []uint32
	ends     []uint32
}

func (s *ngramDocIterator) candidates() []*candidateMatch {
	patBytes := []byte(s.query.Pattern)
	lowerPatBytes := toLower(patBytes)

	fileIdx := 0
	var candidates []*candidateMatch
	for {
		if len(s.first) == 0 || len(s.last) == 0 {
			break
		}
		p1 := s.first[0]
		p2 := s.last[0]

		for fileIdx < len(s.ends) && s.ends[fileIdx] <= p1 {
			fileIdx++
		}

		if p1+s.distance < p2 {
			s.first = s.first[1:]
		} else if p1+s.distance > p2 {
			s.last = s.last[1:]
		} else {
			s.first = s.first[1:]
			s.last = s.last[1:]

			var fileStart uint32
			if fileIdx > 0 {
				fileStart = s.ends[fileIdx-1]
			}
			if p1 < s.leftPad+fileStart || p1+s.distance+ngramSize+s.rightPad > s.ends[fileIdx] {
				continue
			}

			cand := &candidateMatch{
				caseSensitive: s.query.CaseSensitive,
				fileName:      s.query.FileName,
				substrBytes:   patBytes,
				substrLowered: lowerPatBytes,
				// TODO - this is wrong for casefolding searches.
				byteMatchSz: uint32(len(lowerPatBytes)),
				file:        uint32(fileIdx),
				runeOffset:  p1 - fileStart - s.leftPad,
			}
			candidates = append(candidates, cand)
		}
	}
	return candidates
}
