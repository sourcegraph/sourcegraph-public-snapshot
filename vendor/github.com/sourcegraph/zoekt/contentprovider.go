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
	"path"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sourcegraph/zoekt/ctags"
)

var _ = log.Println

// contentProvider is an abstraction to treat matches for names and
// content with the same code.
type contentProvider struct {
	id    *indexData
	stats *Stats

	// mutable
	err      error
	idx      uint32
	_data    []byte
	_nl      []uint32
	_nlBuf   []uint32
	_sects   []DocumentSection
	_sectBuf []DocumentSection
	fileSize uint32
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

func (p *contentProvider) newlines() newlines {
	if p._nl == nil {
		var sz uint32
		p._nl, sz, p.err = p.id.readNewlines(p.idx, p._nlBuf)
		p._nlBuf = p._nl
		p.stats.ContentBytesLoaded += int64(sz)
	}
	return newlines{locs: p._nl, fileSize: p.fileSize}
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

	byteOff, left := sample.lookup(absR)

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

// fillMatches converts the internal candidateMatch slice into our API's LineMatch.
// It only ever returns content XOR filename matches, not both. If there are any
// content matches, these are always returned, and we omit filename matches.
//
// Performance invariant: ms is sorted and non-overlapping.
//
// Note: the byte slices may be backed by mmapped data, so before being
// returned by the API it needs to be copied.
func (p *contentProvider) fillMatches(ms []*candidateMatch, numContextLines int, language string, debug bool) []LineMatch {
	var filenameMatches []*candidateMatch
	contentMatches := make([]*candidateMatch, 0, len(ms))

	for _, m := range ms {
		if m.fileName {
			filenameMatches = append(filenameMatches, m)
		} else {
			contentMatches = append(contentMatches, m)
		}
	}

	// If there are any content matches, we only return these and skip filename matches.
	if len(contentMatches) > 0 {
		contentMatches = breakMatchesOnNewlines(contentMatches, p.data(false))
		return p.fillContentMatches(contentMatches, numContextLines, language, debug)
	}

	// Otherwise, we return a single line containing the filematch match.
	score, debugScore, _ := p.candidateMatchScore(filenameMatches, language, debug)
	res := LineMatch{
		Line:       p.id.fileName(p.idx),
		FileName:   true,
		Score:      score,
		DebugScore: debugScore,
	}

	for _, m := range ms {
		res.LineFragments = append(res.LineFragments, LineFragmentMatch{
			LineOffset:  int(m.byteOffset),
			MatchLength: int(m.byteMatchSz),
			Offset:      m.byteOffset,
		})
	}

	return []LineMatch{res}

}

// fillChunkMatches converts the internal candidateMatch slice into our API's ChunkMatch.
// It only ever returns content XOR filename matches, not both. If there are any content
// matches, these are always returned, and we omit filename matches.
//
// Performance invariant: ms is sorted and non-overlapping.
//
// Note: the byte slices may be backed by mmapped data, so before being
// returned by the API it needs to be copied.
func (p *contentProvider) fillChunkMatches(ms []*candidateMatch, numContextLines int, language string, debug bool) []ChunkMatch {
	var filenameMatches []*candidateMatch
	contentMatches := make([]*candidateMatch, 0, len(ms))

	for _, m := range ms {
		if m.fileName {
			filenameMatches = append(filenameMatches, m)
		} else {
			contentMatches = append(contentMatches, m)
		}
	}

	// If there are any content matches, we only return these and skip filename matches.
	if len(contentMatches) > 0 {
		return p.fillContentChunkMatches(contentMatches, numContextLines, language, debug)
	}

	// Otherwise, we return a single chunk representing the filename match.
	score, debugScore, _ := p.candidateMatchScore(filenameMatches, language, debug)
	fileName := p.id.fileName(p.idx)
	ranges := make([]Range, 0, len(ms))
	for _, m := range ms {
		ranges = append(ranges, Range{
			Start: Location{
				ByteOffset: m.byteOffset,
				LineNumber: 1,
				Column:     uint32(utf8.RuneCount(fileName[:m.byteOffset]) + 1),
			},
			End: Location{
				ByteOffset: m.byteOffset + m.byteMatchSz,
				LineNumber: 1,
				Column:     uint32(utf8.RuneCount(fileName[:m.byteOffset+m.byteMatchSz]) + 1),
			},
		})
	}

	return []ChunkMatch{{
		Content:      fileName,
		ContentStart: Location{ByteOffset: 0, LineNumber: 1, Column: 1},
		Ranges:       ranges,
		FileName:     true,

		Score:      score,
		DebugScore: debugScore,
	}}
}

func (p *contentProvider) fillContentMatches(ms []*candidateMatch, numContextLines int, language string, debug bool) []LineMatch {
	var result []LineMatch
	for len(ms) > 0 {
		m := ms[0]
		num := p.newlines().atOffset(m.byteOffset)
		lineStart := int(p.newlines().lineStart(num))
		nextLineStart := int(p.newlines().lineStart(num + 1))

		var lineCands []*candidateMatch

		endMatch := m.byteOffset + m.byteMatchSz

		for len(ms) > 0 {
			m := ms[0]
			if int(m.byteOffset) < nextLineStart {
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
				num, lineStart, nextLineStart,
				m.byteOffset)
		}

		data := p.data(false)

		// Due to merging matches, we may have a match that
		// crosses a line boundary. Prevent confusion by
		// taking lines until we pass the last match
		for nextLineStart < len(data) && endMatch > uint32(nextLineStart) {
			next := bytes.IndexByte(data[nextLineStart:], '\n')
			if next == -1 {
				nextLineStart = len(data)
			} else {
				// TODO(hanwen): test that checks "+1" part here.
				nextLineStart += next + 1
			}
		}

		finalMatch := LineMatch{
			LineStart:  lineStart,
			LineEnd:    nextLineStart,
			LineNumber: num,
		}
		finalMatch.Line = data[lineStart:nextLineStart]

		if numContextLines > 0 {
			finalMatch.Before = p.newlines().getLines(data, num-numContextLines, num)
			finalMatch.After = p.newlines().getLines(data, num+1, num+1+numContextLines)
		}

		score, debugScore, symbolInfo := p.candidateMatchScore(lineCands, language, debug)
		finalMatch.Score = score
		finalMatch.DebugScore = debugScore

		for i, m := range lineCands {
			fragment := LineFragmentMatch{
				Offset:      m.byteOffset,
				LineOffset:  int(m.byteOffset) - lineStart,
				MatchLength: int(m.byteMatchSz),
			}
			if i < len(symbolInfo) && symbolInfo[i] != nil {
				fragment.SymbolInfo = symbolInfo[i]
			}

			finalMatch.LineFragments = append(finalMatch.LineFragments, fragment)
		}
		result = append(result, finalMatch)
	}
	return result
}

func (p *contentProvider) fillContentChunkMatches(ms []*candidateMatch, numContextLines int, language string, debug bool) []ChunkMatch {
	newlines := p.newlines()
	data := p.data(false)

	// columnHelper prevents O(len(ms) * len(data)) lookups for all columns.
	// However, it depends on ms being sorted by byteOffset and non-overlapping.
	// This invariant is true at the time of writing, but we conservatively
	// enforce this. Note: chunkCandidates preserves the sorting so safe to
	// transform now.
	columnHelper := columnHelper{data: data}
	if !sort.IsSorted((sortByOffsetSlice)(ms)) {
		log.Printf("WARN: performance invariant violated. candidate matches are not sorted in fillContentChunkMatches. Report to developers.")
		sort.Sort((sortByOffsetSlice)(ms))
	}

	chunks := chunkCandidates(ms, newlines, numContextLines)
	chunkMatches := make([]ChunkMatch, 0, len(chunks))
	for _, chunk := range chunks {
		score, debugScore, symbolInfo := p.candidateMatchScore(chunk.candidates, language, debug)

		ranges := make([]Range, 0, len(chunk.candidates))
		for _, cm := range chunk.candidates {
			startOffset := cm.byteOffset
			endOffset := cm.byteOffset + cm.byteMatchSz
			startLine, endLine := newlines.offsetRangeToLineRange(startOffset, endOffset)

			ranges = append(ranges, Range{
				Start: Location{
					ByteOffset: startOffset,
					LineNumber: uint32(startLine),
					Column:     columnHelper.get(int(newlines.lineStart(startLine)), startOffset),
				},
				End: Location{
					ByteOffset: endOffset,
					LineNumber: uint32(endLine),
					Column:     columnHelper.get(int(newlines.lineStart(endLine)), endOffset),
				},
			})
		}

		firstLineNumber := int(chunk.firstLine) - numContextLines
		if firstLineNumber < 1 {
			firstLineNumber = 1
		}
		firstLineStart := newlines.lineStart(firstLineNumber)

		chunkMatches = append(chunkMatches, ChunkMatch{
			Content: newlines.getLines(data, firstLineNumber, int(chunk.lastLine)+numContextLines+1),
			ContentStart: Location{
				ByteOffset: firstLineStart,
				LineNumber: uint32(firstLineNumber),
				Column:     1,
			},
			FileName:   false,
			Ranges:     ranges,
			SymbolInfo: symbolInfo,
			Score:      score,
			DebugScore: debugScore,
		})
	}
	return chunkMatches
}

type candidateChunk struct {
	candidates []*candidateMatch
	firstLine  uint32 // 1-based, inclusive
	lastLine   uint32 // 1-based, inclusive
	minOffset  uint32 // 0-based, inclusive
	maxOffset  uint32 // 0-based, exclusive
}

// chunkCandidates groups a set of sorted, non-overlapping candidate matches by line number. Adjacent
// chunks will be merged if adding `numContextLines` to the beginning and end of the chunk would cause
// it to overlap with an adjacent chunk.
//
// input invariants: ms is sorted by byteOffset and is non overlapping with respect to endOffset.
// output invariants: if you flatten candidates the input invariant is retained.
func chunkCandidates(ms []*candidateMatch, newlines newlines, numContextLines int) []candidateChunk {
	var chunks []candidateChunk
	for _, m := range ms {
		startOffset := m.byteOffset
		endOffset := m.byteOffset + m.byteMatchSz
		firstLine, lastLine := newlines.offsetRangeToLineRange(startOffset, endOffset)

		if len(chunks) > 0 && int(chunks[len(chunks)-1].lastLine)+numContextLines >= firstLine-numContextLines {
			// If a new chunk created with the current candidateMatch would
			// overlap with the previous chunk, instead add the candidateMatch
			// to the last chunk and extend end of the last chunk.
			last := &chunks[len(chunks)-1]
			last.candidates = append(last.candidates, m)
			if last.maxOffset < endOffset {
				last.lastLine = uint32(lastLine)
				last.maxOffset = uint32(endOffset)
			}
		} else {
			chunks = append(chunks, candidateChunk{
				firstLine:  uint32(firstLine),
				lastLine:   uint32(lastLine),
				minOffset:  startOffset,
				maxOffset:  endOffset,
				candidates: []*candidateMatch{m},
			})
		}
	}
	return chunks
}

// columnHelper is a helper struct which caches the number of runes last
// counted. If we naively use utf8.RuneCount for each match on a line, this
// leads to an O(nm) algorithm where m is the number of matches and n is the
// length of the line. Aassuming we our candidates are increasing in offset
// makes this operation O(n) instead.
type columnHelper struct {
	data []byte

	// 0 values for all these are valid values
	lastLineOffset int
	lastOffset     uint32
	lastRuneCount  uint32
}

// get returns the line column for offset. offset is the byte offset of the
// rune in data. lineOffset is the byte offset inside of data for the line
// containing offset.
func (c *columnHelper) get(lineOffset int, offset uint32) uint32 {
	var runeCount uint32

	if lineOffset == c.lastLineOffset && offset >= c.lastOffset {
		// Can count from last calculation
		runeCount = c.lastRuneCount + uint32(utf8.RuneCount(c.data[c.lastOffset:offset]))
	} else {
		// Need to count from the beginning of line
		runeCount = uint32(utf8.RuneCount(c.data[lineOffset:offset]))
	}

	c.lastLineOffset = lineOffset
	c.lastOffset = offset
	c.lastRuneCount = runeCount

	return runeCount + 1
}

type newlines struct {
	// locs is the sorted set of byte offsets of the newlines in the file
	locs []uint32

	// fileSize is just the number of bytes in the file. It is stored
	// on this struct so we can safely know the length of the last line
	// in the file since not all files end in a newline.
	fileSize uint32
}

// atOffset returns the line containing the offset. If the offset lands on
// the newline ending line M, we return M.
func (nls newlines) atOffset(offset uint32) (lineNumber int) {
	idx := sort.Search(len(nls.locs), func(n int) bool {
		return nls.locs[n] >= offset
	})
	return idx + 1
}

// lineStart returns the byte offset of the beginning of the given line.
// lineNumber is 1-based. If lineNumber is out of range of the lines in the
// file, the return value will be clamped to [0,fileSize].
func (nls newlines) lineStart(lineNumber int) uint32 {
	// nls.locs[0] + 1 is the start of the 2nd line of data.
	startIdx := lineNumber - 2

	if startIdx < 0 {
		return 0
	} else if startIdx >= len(nls.locs) {
		return nls.fileSize
	} else {
		return nls.locs[startIdx] + 1
	}
}

// offsetRangeToLineRange returns range of lines that fully contains the given byte range.
// The inputs are 0-based byte offsets into the file representing the (exclusive) range [startOffset, endOffset).
// The return values are 1-based line numbers representing the (inclusive) range [startLine, endLine].
func (nls newlines) offsetRangeToLineRange(startOffset, endOffset uint32) (startLine, endLine int) {
	startLine = nls.atOffset(startOffset)
	endLine = nls.atOffset(
		max(startOffset, max(endOffset, 1)-1), // clamp endOffset and prevent underflow
	)
	return startLine, endLine
}

// getLines returns a slice of data containing the lines [low, high).
// low is 1-based and inclusive. high is 1-based and exclusive.
func (nls newlines) getLines(data []byte, low, high int) []byte {
	if low >= high {
		return nil
	}

	return data[nls.lineStart(low):nls.lineStart(high)]
}

const (
	// Query-dependent scoring signals. All of these together are bounded at ~9000
	// (scoreWordMatch + scoreSymbol + scoreKindMatch * 10 + scoreFactorAtomMatch).
	scorePartialWordMatch = 50.0
	scoreWordMatch        = 500.0
	scoreBase             = 7000.0
	scorePartialBase      = 4000.0
	scoreSymbol           = 7000.0
	scorePartialSymbol    = 4000.0
	scoreKindMatch        = 100.0
	scoreFactorAtomMatch  = 400.0

	// File-only scoring signals. For now these are also bounded ~9000 to give them
	// equal weight with the query-dependent signals.
	scoreFileRankFactor  = 9000.0
	scoreFileOrderFactor = 10.0
	scoreRepoRankFactor  = 20.0

	// Used for ordering line and chunk matches within a file.
	scoreLineOrderFactor = 1.0
)

// findMaxOverlappingSection returns the index of the section in secs that
// overlaps the most with the area defined by off and sz, relative to the size
// of the section. If no section overlaps, it returns 0, false. If multiple
// sections overlap the same amount, the first one is returned.
//
// The implementation assumes that sections do not overlap and are sorted by
// DocumentSection.Start.
func findMaxOverlappingSection(secs []DocumentSection, off, sz uint32) (uint32, bool) {
	start := off
	end := off + sz

	// Find the first section that might overlap
	j := sort.Search(len(secs), func(i int) bool { return secs[i].End > start })

	if j == len(secs) || secs[j].Start >= end {
		// No overlap.
		return 0, false
	}

	relOverlap := func(j int) float64 {
		secSize := secs[j].End - secs[j].Start
		if secSize == 0 {
			return 0
		}
		// This cannot overflow because we make sure there is overlap before calling relOverlap
		overlap := min(secs[j].End, end) - max(secs[j].Start, start)
		return float64(overlap) / float64(secSize)
	}

	ol1 := relOverlap(j)
	if epsilonEqualsOne(ol1) || j == len(secs)-1 || secs[j+1].Start >= end {
		return uint32(j), ol1 > 0
	}

	// We know that [off,off+sz[ overlaps with at least 2 sections. We only have to check
	// if the second section overlaps more than the first one, because a third
	// section can only overlap if the overlap with the second section is complete.
	ol2 := relOverlap(j + 1)
	if ol2 > ol1 {
		return uint32(j + 1), ol2 > 0
	}

	return uint32(j), ol1 > 0
}

func (p *contentProvider) findSymbol(cm *candidateMatch) (DocumentSection, *Symbol, bool) {
	if cm.fileName {
		return DocumentSection{}, nil, false
	}

	secs := p.docSections()

	secIdx, ok := cm.symbolIdx, cm.symbol
	if !ok {
		// Not from a symbol matchTree. Let's see if it overlaps with a symbol.
		secIdx, ok = findMaxOverlappingSection(secs, cm.byteOffset, cm.byteMatchSz)
	}
	if !ok {
		return DocumentSection{}, nil, false
	}

	sec := secs[secIdx]

	// Now lets hydrate in the SymbolInfo. We do not hydrate in SymbolInfo.Sym
	// since some callsites do not need it stored, and that incurs an extra
	// copy.
	//
	// 2024-01-08 we are refactoring this and the code path indicates this can
	// fail, so callers need to handle nil symbol. However, it would be
	// surprising that we have a matching section but not symbol data.
	start := p.id.fileEndSymbol[p.idx]
	si := p.id.symbols.data(start + secIdx)

	return sec, si, true
}

func (p *contentProvider) candidateMatchScore(ms []*candidateMatch, language string, debug bool) (float64, string, []*Symbol) {
	type debugScore struct {
		what  string
		score float64
	}

	score := &debugScore{}
	maxScore := &debugScore{}

	addScore := func(what string, s float64) {
		if s != 0 && debug {
			score.what += fmt.Sprintf("%s:%.2f, ", what, s)
		}
		score.score += s
	}

	filename := p.data(true)
	var symbolInfo []*Symbol

	for i, m := range ms {
		data := p.data(m.fileName)

		endOffset := m.byteOffset + m.byteMatchSz
		startBoundary := m.byteOffset < uint32(len(data)) && (m.byteOffset == 0 || byteClass(data[m.byteOffset-1]) != byteClass(data[m.byteOffset]))
		endBoundary := endOffset > 0 && (endOffset == uint32(len(data)) || byteClass(data[endOffset-1]) != byteClass(data[endOffset]))

		score.score = 0
		score.what = ""

		if startBoundary && endBoundary {
			addScore("WordMatch", scoreWordMatch)
		} else if startBoundary || endBoundary {
			addScore("PartialWordMatch", scorePartialWordMatch)
		}

		if m.fileName {
			sep := bytes.LastIndexByte(data, '/')
			startMatch := int(m.byteOffset) == sep+1
			endMatch := endOffset == uint32(len(data))
			if startMatch && endMatch {
				addScore("Base", scoreBase)
			} else if startMatch || endMatch {
				addScore("EdgeBase", (scoreBase+scorePartialBase)/2)
			} else if sep < int(m.byteOffset) {
				addScore("InnerBase", scorePartialBase)
			}
		} else if sec, si, ok := p.findSymbol(m); ok {
			startMatch := sec.Start == m.byteOffset
			endMatch := sec.End == endOffset
			if startMatch && endMatch {
				addScore("Symbol", scoreSymbol)
			} else if startMatch || endMatch {
				addScore("EdgeSymbol", (scoreSymbol+scorePartialSymbol)/2)
			} else {
				addScore("OverlapSymbol", scorePartialSymbol)
			}

			// Score based on symbol data
			if si != nil {
				symbolKind := ctags.ParseSymbolKind(si.Kind)
				sym := sectionSlice(data, sec)

				addScore(fmt.Sprintf("kind:%s:%s", language, si.Kind), scoreSymbolKind(language, filename, sym, symbolKind))

				// This is from a symbol tree, so we need to store the symbol
				// information.
				if m.symbol {
					if symbolInfo == nil {
						symbolInfo = make([]*Symbol, len(ms))
					}
					// findSymbols does not hydrate in Sym. So we need to store it.
					si.Sym = string(sym)
					symbolInfo[i] = si
				}
			}
		}

		// scoreWeight != 1 means it affects score
		if !epsilonEqualsOne(m.scoreWeight) {
			score.score = score.score * m.scoreWeight
			if debug {
				score.what += fmt.Sprintf("boost:%.2f, ", m.scoreWeight)
			}
		}

		if score.score > maxScore.score {
			maxScore.score = score.score
			maxScore.what = score.what
		}
	}

	if debug {
		maxScore.what = fmt.Sprintf("score:%.2f <- %s", maxScore.score, strings.TrimSuffix(maxScore.what, ", "))
	}

	return maxScore.score, maxScore.what, symbolInfo
}

// sectionSlice will return data[sec.Start:sec.End] but will clip Start and
// End such that it won't be out of range.
func sectionSlice(data []byte, sec DocumentSection) []byte {
	l := uint32(len(data))
	if sec.Start >= l {
		return nil
	}
	if sec.End > l {
		sec.End = l
	}
	return data[sec.Start:sec.End]
}

// scoreSymbolKind boosts a match based on the combination of language, symbol
// and kind. The language string comes from go-enry, the symbol and kind from
// ctags.
func scoreSymbolKind(language string, filename []byte, sym []byte, kind ctags.SymbolKind) float64 {
	var factor float64

	// Generic ranking which will be overriden by language specific ranking
	switch kind {
	case ctags.Type: // scip-ctags regression workaround https://github.com/sourcegraph/sourcegraph/issues/57659
		factor = 8
	case ctags.Class:
		factor = 10
	case ctags.Struct:
		factor = 9.5
	case ctags.Enum:
		factor = 9
	case ctags.Interface:
		factor = 8
	case ctags.Function, ctags.Method:
		factor = 7
	case ctags.Field:
		factor = 5.5
	case ctags.Constant:
		factor = 5
	case ctags.Variable:
		factor = 4
	default:
		// For all other kinds, assign a low score by default.
		factor = 1
	}

	switch language {
	case "Java", "java":
		switch kind {
		// 2022-03-30: go-ctags contains a regex rule for Java classes that sets "kind"
		// to "classes" instead of "c". We have to cover both cases to support existing
		// indexes.
		case ctags.Class:
			factor = 10
		case ctags.Enum:
			factor = 9
		case ctags.Interface:
			factor = 8
		case ctags.Method:
			factor = 7
		case ctags.Field:
			factor = 6
		case ctags.EnumConstant:
			factor = 5
		}
	case "Kotlin", "kotlin":
		switch kind {
		case ctags.Class:
			factor = 10
		case ctags.Interface:
			factor = 9
		case ctags.Method:
			factor = 8
		case ctags.TypeAlias:
			factor = 7
		case ctags.Constant:
			factor = 6
		case ctags.Variable:
			factor = 5
		}
	case "Go", "go":
		switch kind {
		// scip-ctags regression workaround https://github.com/sourcegraph/sourcegraph/issues/57659
		// for each case a description of the fields in ctags in the comment
		case ctags.Type: // interface struct talias
			factor = 9
		case ctags.Interface: // interfaces
			factor = 10
		case ctags.Struct: // structs
			factor = 9
		case ctags.TypeAlias: // type aliases
			factor = 9
		case ctags.MethodSpec: // interface method specification
			factor = 8.5
		case ctags.Method, ctags.Function: // functions
			factor = 8
		case ctags.Field: // struct fields
			factor = 7
		case ctags.Constant: // constants
			factor = 6
		case ctags.Variable: // variables
			factor = 5
		}

		// Boost exported go symbols. Same implementation as token.IsExported
		if ch, _ := utf8.DecodeRune(sym); unicode.IsUpper(ch) {
			factor += 0.5
		}

		if bytes.HasSuffix(filename, []byte("_test.go")) {
			factor *= 0.8
		}

		// Could also rank on:
		//
		//   - anonMember  struct anonymous members
		//   - packageName name for specifying imported package
		//   - receiver    receivers
		//   - package     packages
		//   - type        types
		//   - unknown     unknown
	case "C++", "c++":
		switch kind {
		case ctags.Class: // classes
			factor = 10
		case ctags.Enum: // enumeration names
			factor = 9
		case ctags.Function: // function definitions
			factor = 8
		case ctags.Struct: // structure names
			factor = 7
		case ctags.Union: // union names
			factor = 6
		case ctags.TypeAlias: // typedefs
			factor = 5
		case ctags.Field: // class, struct, and union members
			factor = 4
		case ctags.Variable: // varialbe definitions
			factor = 3
		}
	// Could also rank on:
	// NAME        DESCRIPTION
	// macro       macro definitions
	// enumerator  enumerators (values inside an enumeration)
	// header      included header files
	// namespace   namespaces
	// variable    variable definitions
	case "Scala", "scala":
		switch kind {
		case ctags.Class:
			factor = 10
		case ctags.Interface:
			factor = 9
		case ctags.Object:
			factor = 8
		case ctags.Function:
			factor = 7
		case ctags.Type:
			factor = 6
		case ctags.Variable:
			factor = 5
		case ctags.Package:
			factor = 4
		}
	case "Python", "python":
		switch kind {
		case ctags.Class: // classes
			factor = 10
		case ctags.Function, ctags.Method: // function definitions
			factor = 8
		case ctags.Field: // class, struct, and union members
			factor = 4
		case ctags.Variable: // variable definitions
			factor = 3
		case ctags.Local: // local variables
			factor = 2
		}
		// Could also rank on:
		//
		//   - namespace name referring a module defined in other file
		//   - module    modules
		//   - unknown   name referring a class/variable/function/module defined in other module
		//   - parameter function parameters
	case "Ruby", "ruby":
		switch kind {
		case ctags.Class:
			factor = 10
		case ctags.Method:
			factor = 9
		case ctags.MethodAlias:
			factor = 8
		case ctags.Module:
			factor = 7
		case ctags.SingletonMethod:
			factor = 6
		case ctags.Constant:
			factor = 5
		case ctags.Accessor:
			factor = 4
		case ctags.Library:
			factor = 3
		}
	case "PHP", "php":
		switch kind {
		case ctags.Class:
			factor = 10
		case ctags.Interface:
			factor = 9
		case ctags.Function:
			factor = 8
		case ctags.Trait:
			factor = 7
		case ctags.Define:
			factor = 6
		case ctags.Namespace:
			factor = 5
		case ctags.MethodAlias:
			factor = 4
		case ctags.Variable:
			factor = 3
		case ctags.Local:
			factor = 3
		}
	case "GraphQL", "graphql":
		switch kind {
		case ctags.Type:
			factor = 10
		}
	case "Markdown", "markdown":
		// Headers are good signal in docs, but do not rank as highly as code.
		switch kind {
		case ctags.Chapter: // #
			factor = 4
		case ctags.Section: // ##
			factor = 3
		case ctags.Subsection: // ###
			factor = 2
		}
	}

	return factor * scoreKindMatch
}

type matchScoreSlice []LineMatch

func (m matchScoreSlice) Len() int           { return len(m) }
func (m matchScoreSlice) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m matchScoreSlice) Less(i, j int) bool { return m[i].Score > m[j].Score }

type chunkMatchScoreSlice []ChunkMatch

func (m chunkMatchScoreSlice) Len() int           { return len(m) }
func (m chunkMatchScoreSlice) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m chunkMatchScoreSlice) Less(i, j int) bool { return m[i].Score > m[j].Score }

type fileMatchesByScore []FileMatch

func (m fileMatchesByScore) Len() int           { return len(m) }
func (m fileMatchesByScore) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m fileMatchesByScore) Less(i, j int) bool { return m[i].Score > m[j].Score }

func sortMatchesByScore(ms []LineMatch) {
	sort.Sort(matchScoreSlice(ms))
}

func sortChunkMatchesByScore(ms []ChunkMatch) {
	sort.Sort(chunkMatchScoreSlice(ms))
}

// SortFiles sorts files matches in the order we want to present results to
// users. The order depends on the match score, which includes both
// query-dependent signals like word overlap, and file-only signals like the
// file ranks (if file ranks are enabled).
//
// We don't only use the scores, we will also boost some results to present
// files with novel extensions.
func SortFiles(ms []FileMatch) {
	sort.Sort(fileMatchesByScore(ms))

	// Boost a file extension not in the top 3 to the third filematch.
	boostNovelExtension(ms, 2, 0.9)
}

func boostNovelExtension(ms []FileMatch, boostOffset int, minScoreRatio float64) {
	if len(ms) <= boostOffset+1 {
		return
	}

	top := ms[:boostOffset]
	candidates := ms[boostOffset:]

	// Don't bother boosting something which is significantly different to the
	// result it replaces.
	minScoreForNovelty := candidates[0].Score * minScoreRatio

	// We want to look for an ext that isn't in the top exts
	exts := make([]string, len(top))
	for i := range top {
		exts[i] = path.Ext(top[i].FileName)
	}

	for i := range candidates {
		// Do not assume sorted due to boostNovelExtension being called on subsets
		if candidates[i].Score < minScoreForNovelty {
			continue
		}

		if slices.Contains(exts, path.Ext(candidates[i].FileName)) {
			continue
		}

		// Found what we are looking for, now boost to front of candidates (which
		// is ms[boostOffset])
		for ; i > 0; i-- {
			candidates[i], candidates[i-1] = candidates[i-1], candidates[i]
		}
		return
	}
}
