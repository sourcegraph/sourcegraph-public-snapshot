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
	"log"
	"sort"
	"unicode/utf8"
)

var _ = log.Println

// contentProvider is an abstraction to treat matches for names and
// content with the same code.
type contentProvider struct {
	id    *indexData
	stats *Stats

	// mutable
	err          error
	idx          uint32
	_data        []byte
	_nl          []uint32
	_nlBuf       []uint32
	_runeOffsets []uint32
	_sects       []DocumentSection
	_sectBuf     []DocumentSection
	fileSize     uint32
	bytesRead    uint32
}

// setDocument skips to the given document.
func (p *contentProvider) setDocument(docID uint32) {
	fileStart := p.id.boundaries[docID]

	p.idx = docID
	p.fileSize = p.id.boundaries[docID+1] - fileStart

	p._nl = nil
	p._sects = nil
	p._data = nil
}

func (p *contentProvider) docSections() []DocumentSection {
	if p._sects == nil {
		var sz uint32
		p._sects, sz, p.err = p.id.readDocSections(p.idx, p._sectBuf)
		p.stats.ContentBytesLoaded += int64(sz)
		p._sectBuf = p._sects
	}
	return p._sects
}

func (p *contentProvider) newlines() []uint32 {
	if p._nl == nil {
		var sz uint32
		p._nl, sz, p.err = p.id.readNewlines(p.idx, p._nlBuf)
		p._nlBuf = p._nl
		p.stats.ContentBytesLoaded += int64(sz)
	}
	return p._nl
}

func (p *contentProvider) data(fileName bool) []byte {
	if fileName {
		return p.id.fileNameContent[p.id.fileNameIndex[p.idx]:p.id.fileNameIndex[p.idx+1]]
	}

	if p._data == nil {
		p._data, p.err = p.id.readContents(p.idx)
		p.stats.FilesLoaded++
		p.stats.ContentBytesLoaded += int64(len(p._data))
	}
	return p._data
}

// Find offset in bytes (relative to corpus start) for an offset in
// runes (relative to document start). If filename is set, the corpus
// is the set of filenames, with the document being the name itself.
func (p *contentProvider) findOffset(filename bool, r uint32) uint32 {
	if p.id.metaData.PlainASCII {
		return r
	}

	sample := p.id.runeOffsets
	runeEnds := p.id.fileEndRunes
	fileStartByte := p.id.boundaries[p.idx]
	if filename {
		sample = p.id.fileNameRuneOffsets
		runeEnds = p.id.fileNameEndRunes
		fileStartByte = p.id.fileNameIndex[p.idx]
	}

	absR := r
	if p.idx > 0 {
		absR += runeEnds[p.idx-1]
	}

	byteOff := sample[absR/runeOffsetFrequency]
	left := absR % runeOffsetFrequency

	var data []byte

	if filename {
		data = p.id.fileNameContent[byteOff:]
	} else {
		data, p.err = p.id.readContentSlice(byteOff, 3*runeOffsetFrequency)
		if p.err != nil {
			return 0
		}
	}
	for left > 0 {
		_, sz := utf8.DecodeRune(data)
		byteOff += uint32(sz)
		data = data[sz:]
		left--
	}

	byteOff -= fileStartByte
	return byteOff
}

func (p *contentProvider) fillMatches(ms []*candidateMatch) []LineMatch {
	var result []LineMatch
	if ms[0].fileName {
		// There is only "line" in a filename.
		res := LineMatch{
			Line:     p.id.fileName(p.idx),
			FileName: true,
		}

		for _, m := range ms {
			res.LineFragments = append(res.LineFragments, LineFragmentMatch{
				LineOffset:  int(m.byteOffset),
				MatchLength: int(m.byteMatchSz),
				Offset:      m.byteOffset,
			})

			result = []LineMatch{res}
		}
	} else {
		result = p.fillContentMatches(ms)
	}

	sects := p.docSections()
	for i, m := range result {
		result[i].Score = matchScore(sects, &m)
	}

	return result
}

func (p *contentProvider) fillContentMatches(ms []*candidateMatch) []LineMatch {
	var result []LineMatch
	for len(ms) > 0 {
		m := ms[0]
		num, lineStart, lineEnd := m.line(p.newlines(), p.fileSize)

		var lineCands []*candidateMatch

		endMatch := m.byteOffset + m.byteMatchSz

		for len(ms) > 0 {
			m := ms[0]
			if int(m.byteOffset) <= lineEnd {
				endMatch = m.byteOffset + m.byteMatchSz
				lineCands = append(lineCands, m)
				ms = ms[1:]
			} else {
				break
			}
		}

		if len(lineCands) == 0 {
			log.Panicf(
				"%s %v infinite loop: num %d start,end %d,%d, offset %d",
				p.id.fileName(p.idx), p.id.metaData,
				num, lineStart, lineEnd,
				m.byteOffset)
		}

		data := p.data(false)

		// Due to merging matches, we may have a match that
		// crosses a line boundary. Prevent confusion by
		// taking lines until we pass the last match
		for lineEnd < len(data) && endMatch > uint32(lineEnd) {
			next := bytes.IndexByte(data[lineEnd+1:], '\n')
			if next == -1 {
				lineEnd = len(data)
			} else {
				// TODO(hanwen): test that checks "+1" part here.
				lineEnd += next + 1
			}
		}

		finalMatch := LineMatch{
			LineStart:  lineStart,
			LineEnd:    lineEnd,
			LineNumber: num,
		}
		finalMatch.Line = p.data(false)[lineStart:lineEnd]

		for _, m := range lineCands {
			fragment := LineFragmentMatch{
				Offset:      m.byteOffset,
				LineOffset:  int(m.byteOffset) - lineStart,
				MatchLength: int(m.byteMatchSz),
			}
			finalMatch.LineFragments = append(finalMatch.LineFragments, fragment)
		}
		result = append(result, finalMatch)
	}
	return result
}

const (
	// TODO - how to scale this relative to rank?
	scorePartialWordMatch   = 50.0
	scoreWordMatch          = 500.0
	scoreImportantThreshold = 2000.0
	scorePartialSymbol      = 4000.0
	scoreSymbol             = 7000.0
	scoreFactorAtomMatch    = 400.0
	scoreShardRankFactor    = 20.0
	scoreFileOrderFactor    = 10.0
	scoreLineOrderFactor    = 1.0
)

func findSection(secs []DocumentSection, off, sz uint32) *DocumentSection {
	j := sort.Search(len(secs), func(i int) bool {
		return secs[i].End >= off+sz
	})

	if j == len(secs) {
		return nil
	}

	if secs[j].Start <= off && off+sz <= secs[j].End {
		return &secs[j]
	}
	return nil
}

func matchScore(secs []DocumentSection, m *LineMatch) float64 {
	var maxScore float64
	for _, f := range m.LineFragments {
		startBoundary := f.LineOffset < len(m.Line) && (f.LineOffset == 0 || byteClass(m.Line[f.LineOffset-1]) != byteClass(m.Line[f.LineOffset]))

		end := int(f.LineOffset) + f.MatchLength
		endBoundary := end > 0 && (end == len(m.Line) || byteClass(m.Line[end-1]) != byteClass(m.Line[end]))

		score := 0.0
		if startBoundary && endBoundary {
			score = scoreWordMatch
		} else if startBoundary || endBoundary {
			score = scorePartialWordMatch
		}

		sec := findSection(secs, f.Offset, uint32(f.MatchLength))
		if sec != nil {
			startMatch := sec.Start == f.Offset
			endMatch := sec.End == f.Offset+uint32(f.MatchLength)
			if startMatch && endMatch {
				score += scoreSymbol
			} else if startMatch || endMatch {
				score += (scoreSymbol + scorePartialSymbol) / 2
			} else {
				score += scorePartialSymbol
			}
		}
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}

type matchScoreSlice []LineMatch

func (m matchScoreSlice) Len() int           { return len(m) }
func (m matchScoreSlice) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m matchScoreSlice) Less(i, j int) bool { return m[i].Score > m[j].Score }

type fileMatchSlice []FileMatch

func (m fileMatchSlice) Len() int           { return len(m) }
func (m fileMatchSlice) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m fileMatchSlice) Less(i, j int) bool { return m[i].Score > m[j].Score }

func sortMatchesByScore(ms []LineMatch) {
	sort.Sort(matchScoreSlice(ms))
}

// Sort a slice of results.
func SortFilesByScore(ms []FileMatch) {
	sort.Sort(fileMatchSlice(ms))
}
