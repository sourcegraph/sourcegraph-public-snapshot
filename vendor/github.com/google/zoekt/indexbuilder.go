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
	"encoding/binary"
	"fmt"
	"hash/crc64"
	"html/template"
	"log"
	"path/filepath"
	"sort"
	"unicode/utf8"
)

var _ = log.Println

const ngramSize = 3

type searchableString struct {
	data []byte
}

// Filled by the linker (see build-deploy.sh)
var Version string

// Store character (unicode codepoint) offset (in bytes) this often.
const runeOffsetFrequency = 100

type postingsBuilder struct {
	postings    map[ngram][]byte
	lastOffsets map[ngram]uint32

	// To support UTF-8 searching, we must map back runes to byte
	// offsets. As a first attempt, we sample regularly. The
	// precise offset can be found by walking from the recorded
	// offset to the desired rune.
	runeOffsets []uint32
	runeCount   uint32

	isPlainASCII bool

	endRunes []uint32
	endByte  uint32
}

func newPostingsBuilder() *postingsBuilder {
	return &postingsBuilder{
		postings:     map[ngram][]byte{},
		lastOffsets:  map[ngram]uint32{},
		isPlainASCII: true,
	}
}

// Store trigram offsets for the given UTF-8 data. The
// DocumentSections must correspond to rune boundaries in the UTF-8
// data.
func (s *postingsBuilder) newSearchableString(data []byte, byteSections []DocumentSection) (*searchableString, []DocumentSection, error) {
	dest := searchableString{
		data: data,
	}
	var buf [8]byte
	var runeGram [3]rune

	var runeIndex uint32
	byteCount := 0
	dataSz := uint32(len(data))

	byteSectionBoundaries := make([]uint32, 0, 2*len(byteSections))
	for _, s := range byteSections {
		byteSectionBoundaries = append(byteSectionBoundaries, s.Start, s.End)
	}
	var runeSectionBoundaries []uint32

	endRune := s.runeCount
	for ; len(data) > 0; runeIndex++ {
		c, sz := utf8.DecodeRune(data)
		if sz > 1 {
			s.isPlainASCII = false
		}
		data = data[sz:]

		runeGram[0], runeGram[1], runeGram[2] = runeGram[1], runeGram[2], c

		if idx := s.runeCount + runeIndex; idx%runeOffsetFrequency == 0 {
			s.runeOffsets = append(s.runeOffsets, s.endByte+uint32(byteCount))
		}
		for len(byteSectionBoundaries) > 0 && byteSectionBoundaries[0] == uint32(byteCount) {
			runeSectionBoundaries = append(runeSectionBoundaries,
				endRune+uint32(runeIndex))
			byteSectionBoundaries = byteSectionBoundaries[1:]
		}

		byteCount += sz

		if runeIndex < 2 {
			continue
		}

		ng := runesToNGram(runeGram)
		lastOff := s.lastOffsets[ng]
		newOff := endRune + uint32(runeIndex) - 2

		m := binary.PutUvarint(buf[:], uint64(newOff-lastOff))
		s.postings[ng] = append(s.postings[ng], buf[:m]...)
		s.lastOffsets[ng] = newOff
	}
	s.runeCount += runeIndex

	for len(byteSectionBoundaries) > 0 && byteSectionBoundaries[0] < uint32(byteCount) {
		return nil, nil, fmt.Errorf("no rune for section boundary at byte %d", byteSectionBoundaries[0])
	}

	// Handle symbol definition that ends at file end. This can
	// happen for labels at the end of .bat files.

	for len(byteSectionBoundaries) > 0 && byteSectionBoundaries[0] == uint32(byteCount) {
		runeSectionBoundaries = append(runeSectionBoundaries,
			endRune+runeIndex)
		byteSectionBoundaries = byteSectionBoundaries[1:]
	}
	runeSecs := make([]DocumentSection, 0, len(byteSections))
	for i := 0; i < len(runeSectionBoundaries); i += 2 {
		runeSecs = append(runeSecs, DocumentSection{
			Start: runeSectionBoundaries[i],
			End:   runeSectionBoundaries[i+1],
		})
	}

	s.endRunes = append(s.endRunes, s.runeCount)
	s.endByte += dataSz
	return &dest, runeSecs, nil
}

// IndexBuilder builds a single index shard.
type IndexBuilder struct {
	contentStrings  []*searchableString
	nameStrings     []*searchableString
	docSections     [][]DocumentSection
	runeDocSections []DocumentSection

	checksums []byte

	branchMasks []uint64
	subRepos    []uint32

	contentPostings *postingsBuilder
	namePostings    *postingsBuilder

	// root repository
	repo Repository

	// name to index.
	subRepoIndices map[string]uint32

	// language => language code
	languageMap map[string]byte

	// languages codes
	languages []byte
}

func (d *Repository) verify() error {
	for _, t := range []string{d.FileURLTemplate, d.LineFragmentTemplate, d.CommitURLTemplate} {
		if _, err := template.New("").Parse(t); err != nil {
			return err
		}
	}
	return nil
}

// ContentSize returns the number of content bytes so far ingested.
func (b *IndexBuilder) ContentSize() uint32 {
	// Add the name too so we don't skip building index if we have
	// lots of empty files.
	return b.contentPostings.endByte + b.namePostings.endByte
}

// NewIndexBuilder creates a fresh IndexBuilder. The passed in
// Repository contains repo metadata, and may be set to nil.
func NewIndexBuilder(r *Repository) (*IndexBuilder, error) {
	b := &IndexBuilder{
		contentPostings: newPostingsBuilder(),
		namePostings:    newPostingsBuilder(),
		languageMap:     map[string]byte{},
	}

	if r == nil {
		r = &Repository{}
	}
	if err := b.setRepository(r); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *IndexBuilder) setRepository(desc *Repository) error {
	if len(b.contentStrings) > 0 {
		return fmt.Errorf("AddSubRepository called after adding files.")
	}
	if err := desc.verify(); err != nil {
		return err
	}

	for _, subrepo := range desc.SubRepoMap {
		branchEqual := len(subrepo.Branches) == len(desc.Branches)
		if branchEqual {
			for i, b := range subrepo.Branches {
				branchEqual = branchEqual && (b.Name == desc.Branches[i].Name)
			}
		}
	}

	if len(desc.Branches) > 64 {
		return fmt.Errorf("too many branches.")
	}

	b.repo = *desc
	repoCopy := *desc
	repoCopy.SubRepoMap = nil

	if b.repo.SubRepoMap == nil {
		b.repo.SubRepoMap = map[string]*Repository{}
	}
	b.repo.SubRepoMap[""] = &repoCopy

	b.populateSubRepoIndices()
	return nil
}

type DocumentSection struct {
	Start, End uint32
}

// Document holds a document (file) to index.
type Document struct {
	Name              string
	Content           []byte
	Branches          []string
	SubRepositoryPath string
	Language          string

	// If set, something is wrong with the file contents, and this
	// is the reason it wasn't indexed.
	SkipReason string

	// Document sections for symbols. Offsets should use bytes.
	Symbols []DocumentSection
}

type docSectionSlice []DocumentSection

func (m docSectionSlice) Len() int           { return len(m) }
func (m docSectionSlice) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m docSectionSlice) Less(i, j int) bool { return m[i].Start < m[j].Start }

// AddFile is a convenience wrapper for Add
func (b *IndexBuilder) AddFile(name string, content []byte) error {
	return b.Add(Document{Name: name, Content: content})
}

const maxTrigramCount = 20000

// CheckText returns a reason why the given contents are probably not source texts.
func CheckText(content []byte) error {
	if len(content) == 0 {
		return nil
	}

	if len(content) < ngramSize {
		return fmt.Errorf("file size smaller than %d", ngramSize)
	}

	trigrams := map[ngram]struct{}{}

	var cur [3]rune
	byteCount := 0
	for len(content) > 0 {
		if content[0] == 0 {
			return fmt.Errorf("binary data at byte offset %d", byteCount)
		}

		r, sz := utf8.DecodeRune(content)
		content = content[sz:]
		byteCount += sz

		cur[0], cur[1], cur[2] = cur[1], cur[2], r
		if cur[0] == 0 {
			// start of file.
			continue
		}

		trigrams[runesToNGram(cur)] = struct{}{}
		if len(trigrams) > maxTrigramCount {
			// probably not text.
			return fmt.Errorf("number of trigrams exceeds %d", maxTrigramCount)
		}
	}
	return nil
}

func (b *IndexBuilder) populateSubRepoIndices() {
	if b.subRepoIndices != nil {
		return
	}
	var paths []string
	for k := range b.repo.SubRepoMap {
		paths = append(paths, k)
	}
	sort.Strings(paths)
	b.subRepoIndices = make(map[string]uint32, len(paths))
	for i, p := range paths {
		b.subRepoIndices[p] = uint32(i)
	}
}

const notIndexedMarker = "NOT-INDEXED: "

// Add a file which only occurs in certain branches.
func (b *IndexBuilder) Add(doc Document) error {
	hasher := crc64.New(crc64.MakeTable(crc64.ISO))

	if idx := bytes.IndexByte(doc.Content, 0); idx >= 0 {
		doc.SkipReason = fmt.Sprintf("binary content at byte offset %d", idx)
		doc.Language = "binary"
	}

	if doc.SkipReason != "" {
		doc.Content = []byte(notIndexedMarker + doc.SkipReason)
		doc.Symbols = nil
	}

	sort.Sort(docSectionSlice(doc.Symbols))
	var last DocumentSection
	for i, s := range doc.Symbols {
		if i > 0 {
			if last.End > s.Start {
				return fmt.Errorf("sections overlap")
			}
		}
		last = s
	}
	if last.End > uint32(len(doc.Content)) {
		return fmt.Errorf("section goes past end of content")
	}

	if doc.SubRepositoryPath != "" {
		rel, err := filepath.Rel(doc.SubRepositoryPath, doc.Name)
		if err != nil || rel == doc.Name {
			return fmt.Errorf("path %q must start subrepo path %q", doc.Name, doc.SubRepositoryPath)
		}
	}
	docStr, runeSecs, err := b.contentPostings.newSearchableString(doc.Content, doc.Symbols)
	if err != nil {
		return err
	}
	nameStr, _, err := b.namePostings.newSearchableString([]byte(doc.Name), nil)
	if err != nil {
		return err
	}

	subRepoIdx, ok := b.subRepoIndices[doc.SubRepositoryPath]
	if !ok {
		return fmt.Errorf("unknown subrepo path %q", doc.SubRepositoryPath)
	}

	var mask uint64
	for _, br := range doc.Branches {
		m := b.branchMask(br)
		if m == 0 {
			return fmt.Errorf("no branch found for %s", br)
		}
		mask |= m
	}

	b.subRepos = append(b.subRepos, subRepoIdx)

	hasher.Write(doc.Content)

	b.contentStrings = append(b.contentStrings, docStr)
	b.runeDocSections = append(b.runeDocSections, runeSecs...)

	b.nameStrings = append(b.nameStrings, nameStr)
	b.docSections = append(b.docSections, doc.Symbols)
	b.branchMasks = append(b.branchMasks, mask)
	b.checksums = append(b.checksums, hasher.Sum(nil)...)

	langCode, ok := b.languageMap[doc.Language]
	if !ok {
		if len(b.languageMap) >= 255 {
			return fmt.Errorf("too many languages")
		}
		langCode = byte(len(b.languageMap))
		b.languageMap[doc.Language] = langCode
	}
	b.languages = append(b.languages, langCode)

	return nil
}

func (b *IndexBuilder) branchMask(br string) uint64 {
	for i, b := range b.repo.Branches {
		if b.Name == br {
			return uint64(1) << uint(i)
		}
	}
	return 0
}
