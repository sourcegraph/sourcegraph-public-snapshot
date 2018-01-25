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
	"fmt"
	"hash/crc64"
	"unicode/utf8"

	"github.com/google/zoekt/query"
)

// indexData holds the pattern-independent data that we have to have
// in memory to search. Most of the memory is taken up by the ngram =>
// offset index.
type indexData struct {
	file IndexFile

	ngrams map[ngram]simpleSection

	newlinesStart uint32
	newlinesIndex []uint32

	docSectionsStart uint32
	docSectionsIndex []uint32

	runeDocSections []DocumentSection

	// rune offset=>byte offset mapping, relative to the start of the content corpus
	runeOffsets []uint32

	// offsets of file contents; includes end of last file
	boundariesStart uint32
	boundaries      []uint32

	// rune offsets for the file content boundaries
	fileEndRunes []uint32

	fileNameContent []byte
	fileNameIndex   []uint32
	fileNameNgrams  map[ngram][]uint32

	// rune offset=>byte offset mapping, relative to the start of the filename corpus
	fileNameRuneOffsets []uint32

	// rune offsets for the file name boundaries
	fileNameEndRunes []uint32

	fileBranchMasks []uint64

	// mask (power of 2) => name
	branchNames map[uint]string

	// name => mask (power of 2)
	branchIDs map[string]uint

	metaData     IndexMetadata
	repoMetaData Repository

	subRepos     []uint32
	subRepoPaths []string

	// Checksums for all the files, at 8-byte intervals
	checksums []byte

	// languages for all the files.
	languages []byte

	// inverse of LanguageMap in metaData
	languageMap map[byte]string

	repoListEntry RepoListEntry
}

func (d *indexData) getChecksum(idx uint32) []byte {
	start := crc64.Size * idx
	return d.checksums[start : start+crc64.Size]
}

func (d *indexData) calculateStats() {
	var last uint32
	if len(d.boundaries) > 0 {
		last += d.boundaries[len(d.boundaries)-1]
	}

	lastFN := last
	if len(d.fileNameIndex) > 0 {
		lastFN = d.fileNameIndex[len(d.fileNameIndex)-1]
	}

	stats := RepoStats{
		IndexBytes:   int64(d.memoryUse()),
		ContentBytes: int64(int(last) + int(lastFN)),
		Documents:    len(d.newlinesIndex) - 1,
	}
	d.repoListEntry = RepoListEntry{
		Repository:    d.repoMetaData,
		IndexMetadata: d.metaData,
		Stats:         stats,
	}
}

func (d *indexData) String() string {
	return fmt.Sprintf("shard(%s)", d.file.Name())
}

func (d *indexData) memoryUse() int {
	sz := 0
	for _, a := range [][]uint32{
		d.newlinesIndex, d.docSectionsIndex,
		d.boundaries, d.fileNameIndex,
		d.runeOffsets, d.fileNameRuneOffsets,
		d.fileEndRunes, d.fileNameEndRunes,
	} {
		sz += 4 * len(a)
	}
	sz += 8 * len(d.runeDocSections)
	sz += 8 * len(d.fileBranchMasks)
	sz += 12 * len(d.ngrams)
	for _, v := range d.fileNameNgrams {
		sz += 4*len(v) + 4
	}
	return sz
}

const maxUInt32 = 0xffffffff

func minarg(xs []uint32) uint32 {
	m := uint32(maxUInt32)
	j := len(xs)
	for i, x := range xs {
		if x <= m {
			m = x
			j = i
		}
	}
	return uint32(j)
}

func (data *indexData) ngramFrequency(ng ngram, filename bool) uint32 {
	if filename {
		return uint32(len(data.fileNameNgrams[ng]))
	}

	return data.ngrams[ng].sz
}

type ngramIterationResults struct {
	cands []*candidateMatch

	// The ngram matches cover the pattern, so no need to check
	// contents.
	coversContent bool

	// The number of posting bytes
	bytesRead uint32
}

func (data *indexData) iterateNgrams(query *query.Substring) (*ngramIterationResults, error) {
	iter := &ngramDocIterator{
		query: query,
	}

	if query.FileName {
		iter.ends = data.fileNameEndRunes
	} else {
		iter.ends = data.fileEndRunes
	}

	str := query.Pattern

	// Find the 2 least common ngrams from the string.
	ngramOffs := splitNGrams([]byte(query.Pattern))
	frequencies := make([]uint32, 0, len(ngramOffs))
	for _, o := range ngramOffs {
		var freq uint32
		if query.CaseSensitive {
			freq = data.ngramFrequency(o.ngram, query.FileName)
		} else {
			for _, v := range generateCaseNgrams(o.ngram) {
				freq += data.ngramFrequency(v, query.FileName)
			}
		}

		if freq == 0 {
			return &ngramIterationResults{}, nil
		}

		frequencies = append(frequencies, freq)
	}

	firstI := minarg(frequencies)
	frequencies[firstI] = maxUInt32
	lastI := minarg(frequencies)
	if firstI > lastI {
		lastI, firstI = firstI, lastI
	}

	firstNG := ngramOffs[firstI].ngram
	lastNG := ngramOffs[lastI].ngram
	iter.distance = lastI - firstI
	iter.leftPad = firstI
	iter.rightPad = uint32(utf8.RuneCountInString(str)-ngramSize) - lastI

	postings, bytesRead, err := data.readPostings(firstNG, query.CaseSensitive, query.FileName)
	if err != nil {
		return nil, err
	}

	iter.first = postings

	if firstI != lastI {
		postings, sz, err := data.readPostings(lastNG, query.CaseSensitive, query.FileName)
		bytesRead += sz
		if err != nil {
			return nil, err
		}
		iter.last = postings
	} else {
		// TODO - we could be a little faster and skip the
		// list intersection
		iter.last = iter.first
	}

	result := ngramIterationResults{
		bytesRead:     bytesRead,
		cands:         iter.candidates(),
		coversContent: lastI-firstI <= ngramSize && iter.leftPad == 0 && iter.rightPad == 0,
	}
	return &result, nil
}

func (d *indexData) readPostings(ng ngram, caseSensitive, fileName bool) ([]uint32, uint32, error) {
	variants := []ngram{ng}
	if !caseSensitive {
		variants = generateCaseNgrams(ng)
	}

	// TODO - generate less garbage.
	var sz uint32
	postings := make([][]uint32, 0, len(variants))
	for _, v := range variants {
		if fileName {
			postings = append(postings, d.fileNameNgrams[v])
			continue
		}

		sec := d.ngrams[v]
		blob, err := d.readSectionBlob(sec)
		if err != nil {
			return nil, 0, err
		}
		sz += sec.sz
		ps := fromDeltas(blob, nil)
		if len(ps) > 0 {
			postings = append(postings, ps)
		}
	}

	result := mergeUint32(postings)
	return result, sz, nil
}

func mergeUint32(in [][]uint32) []uint32 {
	sz := 0
	for _, i := range in {
		sz += len(i)
	}
	out := make([]uint32, 0, sz)
	for len(in) > 0 {
		minVal := uint32(maxUInt32)
		for _, n := range in {
			if len(n) > 0 && n[0] < minVal {
				minVal = n[0]
			}
		}

		next := in[:0]
		for _, n := range in {
			if len(n) == 0 {
				continue
			}
			if n[0] == minVal {
				out = append(out, minVal)
				n = n[1:]
			}
			if len(n) > 0 {
				next = append(next, n)
			}
		}

		in = next
	}

	return out
}

func (d *indexData) fileName(i uint32) []byte {
	return d.fileNameContent[d.fileNameIndex[i]:d.fileNameIndex[i+1]]
}

func (s *indexData) Close() {
	s.file.Close()
}
