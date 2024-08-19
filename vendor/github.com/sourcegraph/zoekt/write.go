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
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"
)

func (w *writer) writeTOC(toc *indexTOC) {
	// Tagged sections are indicated with a 0 section count.
	// Tagged sections allow easier forwards and backwards
	// compatibility when evolving zoekt index files with new
	// sections.
	//
	// A tagged section is:
	// Varint TagLen, Tag String, Varint Kind, Section
	//
	// Section kind is indicated because simpleSections and
	// compoundSections have different lengths.
	w.U32(0)
	secs := toc.sectionsTaggedList()
	for _, s := range secs {
		w.String(s.tag)
		w.Varint(uint32(s.sec.kind()))
		s.sec.write(w)
	}
}

func (s *compoundSection) writeStrings(w *writer, strs []*searchableString) {
	s.start(w)
	for _, f := range strs {
		s.addItem(w, f.data)
	}
	s.end(w)
}

func (s *compoundSection) writeMap(w *writer, m map[string]uint32) {
	keys := make([]*searchableString, 0, len(m))
	for k := range m {
		keys = append(keys, &searchableString{
			data: []byte(k),
		})
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[string(keys[i].data)] < m[string(keys[j].data)]
	})
	s.writeStrings(w, keys)
}

func writePostings(w *writer, s *postingsBuilder, ngramText *simpleSection,
	charOffsets *simpleSection, postings *compoundSection, endRunes *simpleSection,
) {
	keys := make(ngramSlice, 0, len(s.postings))
	for k := range s.postings {
		keys = append(keys, k)
	}
	sort.Sort(keys)

	ngramText.start(w)
	for _, k := range keys {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(k))
		w.Write(buf[:])
	}
	ngramText.end(w)

	postings.start(w)
	for _, k := range keys {
		postings.addItem(w, s.postings[k])
	}
	postings.end(w)

	charOffsets.start(w)
	w.Write(toSizedDeltas(s.runeOffsets))
	charOffsets.end(w)

	endRunes.start(w)
	w.Write(toSizedDeltas(s.endRunes))
	endRunes.end(w)
}

func (b *IndexBuilder) Write(out io.Writer) error {
	next := b.indexFormatVersion == NextIndexFormatVersion

	buffered := bufio.NewWriterSize(out, 1<<20)
	defer buffered.Flush()

	w := &writer{w: buffered}
	toc := indexTOC{}

	toc.fileContents.writeStrings(w, b.contentStrings)
	toc.newlines.start(w)
	for _, f := range b.contentStrings {
		toc.newlines.addItem(w, toSizedDeltas(newLinesIndices(f.data)))
	}
	toc.newlines.end(w)

	toc.fileEndSymbol.start(w)
	for _, m := range b.fileEndSymbol {
		w.U32(m)
	}
	toc.fileEndSymbol.end(w)

	toc.symbolMap.writeMap(w, b.symIndex)
	toc.symbolKindMap.writeMap(w, b.symKindIndex)
	toc.symbolMetaData.start(w)
	for _, m := range b.symMetaData {
		w.U32(m)
	}
	toc.symbolMetaData.end(w)

	toc.branchMasks.start(w)
	for _, m := range b.branchMasks {
		w.U64(m)
	}
	toc.branchMasks.end(w)

	toc.fileSections.start(w)
	for _, s := range b.docSections {
		toc.fileSections.addItem(w, marshalDocSections(s))
	}
	toc.fileSections.end(w)

	writePostings(w, b.contentPostings, &toc.ngramText, &toc.runeOffsets, &toc.postings, &toc.fileEndRunes)

	// names.
	toc.fileNames.writeStrings(w, b.nameStrings)

	writePostings(w, b.namePostings, &toc.nameNgramText, &toc.nameRuneOffsets, &toc.namePostings, &toc.nameEndRunes)

	toc.subRepos.start(w)
	w.Write(toSizedDeltas(b.subRepos))
	toc.subRepos.end(w)

	toc.contentChecksums.start(w)
	w.Write(b.checksums)
	toc.contentChecksums.end(w)

	toc.languages.start(w)
	w.Write(b.languages)
	toc.languages.end(w)

	toc.runeDocSections.start(w)
	w.Write(marshalDocSections(b.runeDocSections))
	toc.runeDocSections.end(w)

	if next {
		toc.repos.start(w)
		w.Write(toSizedDeltas16(b.repos))
		toc.repos.end(w)
	}

	indexTime := b.IndexTime
	if indexTime.IsZero() {
		indexTime = time.Now().UTC()
	}

	if err := b.writeJSON(&IndexMetadata{
		IndexFormatVersion:    b.indexFormatVersion,
		IndexTime:             indexTime,
		IndexFeatureVersion:   b.featureVersion,
		IndexMinReaderVersion: WriteMinFeatureVersion,
		PlainASCII:            b.contentPostings.isPlainASCII && b.namePostings.isPlainASCII,
		LanguageMap:           b.languageMap,
		ZoektVersion:          Version,
		ID:                    b.ID,
	}, &toc.metaData, w); err != nil {
		return err
	}

	if next {
		if err := b.writeJSON(b.repoList, &toc.repoMetaData, w); err != nil {
			return err
		}
	} else {
		if len(b.repoList) != 1 {
			return fmt.Errorf("have %d repos, but only support 1 in index format version %d", len(b.repoList), b.indexFormatVersion)
		}
		if err := b.writeJSON(b.repoList[0], &toc.repoMetaData, w); err != nil {
			return err
		}
	}

	toc.ranks.start(w)
	if err := encodeRanks(w, b.ranks); err != nil {
		return err
	}
	toc.ranks.end(w)

	var tocSection simpleSection

	tocSection.start(w)
	w.writeTOC(&toc)
	tocSection.end(w)
	tocSection.write(w)
	return w.err
}

func (b *IndexBuilder) writeJSON(data interface{}, sec *simpleSection, w *writer) error {
	blob, err := json.Marshal(data)
	if err != nil {
		return err
	}
	sec.start(w)
	w.Write(blob)
	sec.end(w)
	return nil
}

func newLinesIndices(in []byte) []uint32 {
	out := make([]uint32, 0, bytes.Count(in, []byte{'\n'}))
	for i, c := range in {
		if c == '\n' {
			out = append(out, uint32(i))
		}
	}
	return out
}
