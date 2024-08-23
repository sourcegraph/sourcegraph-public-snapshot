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
	"os"
	"path/filepath"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/sourcegraph/zoekt/internal/languages"
)

var _ = log.Println

const ngramSize = 3

type searchableString struct {
	data []byte
}

// Filled by the linker
var Version string

func HostnameBestEffort() string {
	if h := os.Getenv("NODE_NAME"); h != "" {
		return h
	}
	if h := os.Getenv("HOSTNAME"); h != "" {
		return h
	}
	hostname, _ := os.Hostname()
	return hostname
}

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
	// The version we will write to disk. Sourcegraph Specific. This is to
	// enable feature flagging new format versions.
	indexFormatVersion int
	featureVersion     int

	contentStrings  []*searchableString
	nameStrings     []*searchableString
	docSections     [][]DocumentSection
	runeDocSections []DocumentSection

	symID        uint32
	symIndex     map[string]uint32
	symKindID    uint32
	symKindIndex map[string]uint32
	symMetaData  []uint32

	fileEndSymbol []uint32

	checksums []byte

	branchMasks []uint64
	subRepos    []uint32

	// docID => repoID
	repos []uint16

	// Experimental: docID => rank vec
	ranks [][]float64

	contentPostings *postingsBuilder
	namePostings    *postingsBuilder

	// root repositories
	repoList []Repository

	// name to index.
	subRepoIndices []map[string]uint32

	// language => language code
	languageMap map[string]uint16

	// language codes, uint16 encoded as little-endian
	languages []uint8

	// IndexTime will be used as the time if non-zero. Otherwise
	// time.Now(). This is useful for doing reproducible builds in tests.
	IndexTime time.Time

	// a sortable 20 chars long id.
	ID string
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

// NumFiles returns the number of files added to this builder
func (b *IndexBuilder) NumFiles() int {
	return len(b.contentStrings)
}

// NewIndexBuilder creates a fresh IndexBuilder. The passed in
// Repository contains repo metadata, and may be set to nil.
func NewIndexBuilder(r *Repository) (*IndexBuilder, error) {
	b := newIndexBuilder()

	if r == nil {
		r = &Repository{}
	}
	if err := b.setRepository(r); err != nil {
		return nil, err
	}
	return b, nil
}

func newIndexBuilder() *IndexBuilder {
	return &IndexBuilder{
		indexFormatVersion: IndexFormatVersion,
		featureVersion:     FeatureVersion,

		contentPostings: newPostingsBuilder(),
		namePostings:    newPostingsBuilder(),
		fileEndSymbol:   []uint32{0},
		symIndex:        make(map[string]uint32),
		symKindIndex:    make(map[string]uint32),
		languageMap:     make(map[string]uint16),
	}
}

func (b *IndexBuilder) setRepository(desc *Repository) error {
	if err := desc.verify(); err != nil {
		return err
	}

	if len(desc.Branches) > 64 {
		return fmt.Errorf("too many branches")
	}

	repo := *desc

	// copy subrepomap without root
	repo.SubRepoMap = map[string]*Repository{}
	for k, v := range desc.SubRepoMap {
		if k != "" {
			repo.SubRepoMap[k] = v
		}
	}

	b.repoList = append(b.repoList, repo)

	return b.populateSubRepoIndices()
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
	Symbols         []DocumentSection
	SymbolsMetaData []*Symbol

	// Ranks is a vector of ranks for a document as provided by a DocumentRanksFile
	// file in the git repo.
	//
	// Two documents can be ordered by comparing the components of their rank
	// vectors. Bigger entries are better, as are longer vectors.
	//
	// This field is experimental and may change at any time without warning.
	Ranks []float64
}

type symbolSlice struct {
	symbols  []DocumentSection
	metaData []*Symbol
}

func (s symbolSlice) Len() int { return len(s.symbols) }

func (s symbolSlice) Swap(i, j int) {
	s.symbols[i], s.symbols[j] = s.symbols[j], s.symbols[i]
	s.metaData[i], s.metaData[j] = s.metaData[j], s.metaData[i]
}

func (s symbolSlice) Less(i, j int) bool {
	return s.symbols[i].Start < s.symbols[j].Start
}

// AddFile is a convenience wrapper for Add
func (b *IndexBuilder) AddFile(name string, content []byte) error {
	return b.Add(Document{Name: name, Content: content})
}

func (b *IndexBuilder) populateSubRepoIndices() error {
	if len(b.subRepoIndices) == len(b.repoList) {
		return nil
	}
	if len(b.subRepoIndices) != len(b.repoList)-1 {
		return fmt.Errorf("populateSubRepoIndices not called for a repo: %d != %d - 1", len(b.subRepoIndices), len(b.repoList))
	}
	repo := b.repoList[len(b.repoList)-1]
	b.subRepoIndices = append(b.subRepoIndices, mkSubRepoIndices(repo))
	return nil
}

func mkSubRepoIndices(repo Repository) map[string]uint32 {
	paths := []string{""}
	for k := range repo.SubRepoMap {
		paths = append(paths, k)
	}
	sort.Strings(paths)
	subRepoIndices := make(map[string]uint32, len(paths))
	for i, p := range paths {
		subRepoIndices[p] = uint32(i)
	}
	return subRepoIndices
}

const notIndexedMarker = "NOT-INDEXED: "

func (b *IndexBuilder) symbolID(sym string) uint32 {
	if _, ok := b.symIndex[sym]; !ok {
		b.symIndex[sym] = b.symID
		b.symID++
	}
	return b.symIndex[sym]
}

func (b *IndexBuilder) symbolKindID(t string) uint32 {
	if _, ok := b.symKindIndex[t]; !ok {
		b.symKindIndex[t] = b.symKindID
		b.symKindID++
	}
	return b.symKindIndex[t]
}

func (b *IndexBuilder) addSymbols(symbols []*Symbol) {
	for _, sym := range symbols {
		b.symMetaData = append(b.symMetaData,
			// This field was removed due to redundancy. To avoid
			// needing to reindex, it is set to zero for now. In the
			// future, this field will be completely removed. It
			// will require incrementing the feature version.
			0,
			b.symbolKindID(sym.Kind),
			b.symbolID(sym.Parent),
			b.symbolKindID(sym.ParentKind))
	}
}

func DetermineLanguageIfUnknown(doc *Document) {
	if doc.Language == "" {
		doc.Language = languages.GetLanguage(doc.Name, doc.Content)
	}
}

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
		doc.SymbolsMetaData = nil
		if doc.Language == "" {
			doc.Language = "skipped"
		}
	}

	DetermineLanguageIfUnknown(&doc)

	sort.Sort(symbolSlice{doc.Symbols, doc.SymbolsMetaData})
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
	b.addSymbols(doc.SymbolsMetaData)

	repoIdx := len(b.repoList) - 1
	subRepoIdx, ok := b.subRepoIndices[repoIdx][doc.SubRepositoryPath]
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

	if repoIdx > 1<<16 {
		return fmt.Errorf("too many repos in shard: max is %d", 1<<16)
	}

	b.subRepos = append(b.subRepos, subRepoIdx)
	b.repos = append(b.repos, uint16(repoIdx))

	// doc.Ranks might be nil. In case we don't use offline ranking, doc.Ranks is
	// always nil.
	b.ranks = append(b.ranks, doc.Ranks)

	hasher.Write(doc.Content)

	b.contentStrings = append(b.contentStrings, docStr)
	b.runeDocSections = append(b.runeDocSections, runeSecs...)

	b.nameStrings = append(b.nameStrings, nameStr)
	b.docSections = append(b.docSections, doc.Symbols)
	b.fileEndSymbol = append(b.fileEndSymbol, uint32(len(b.runeDocSections)))
	b.branchMasks = append(b.branchMasks, mask)
	b.checksums = append(b.checksums, hasher.Sum(nil)...)

	langCode, ok := b.languageMap[doc.Language]
	if !ok {
		if len(b.languageMap) >= 65535 {
			return fmt.Errorf("too many languages")
		}
		langCode = uint16(len(b.languageMap))
		b.languageMap[doc.Language] = langCode
	}
	b.languages = append(b.languages, uint8(langCode), uint8(langCode>>8))

	return nil
}

func (b *IndexBuilder) branchMask(br string) uint64 {
	for i, b := range b.repoList[len(b.repoList)-1].Branches {
		if b.Name == br {
			return uint64(1) << uint(i)
		}
	}
	return 0
}

type DocChecker struct {
	// A map to count the unique trigrams in a doc. Reused across docs to cut down on allocations.
	trigrams map[ngram]struct{}
}

// Check returns a reason why the given contents are probably not source texts.
func (t *DocChecker) Check(content []byte, maxTrigramCount int, allowLargeFile bool) error {
	if len(content) == 0 {
		return nil
	}

	if len(content) < ngramSize {
		return fmt.Errorf("file size smaller than %d", ngramSize)
	}

	if index := bytes.IndexByte(content, 0); index > 0 {
		return fmt.Errorf("binary data at byte offset %d", index)
	}

	// PERF: we only need to do the trigram check if the upperbound on content is greater than
	// our threshold. Also skip the trigram check if the file is explicitly marked as allowed.
	if trigramsUpperBound := len(content) - ngramSize + 1; trigramsUpperBound <= maxTrigramCount || allowLargeFile {
		return nil
	}

	var cur [3]rune
	byteCount := 0
	t.clearTrigrams(maxTrigramCount)

	for len(content) > 0 {
		r, sz := utf8.DecodeRune(content)
		content = content[sz:]
		byteCount += sz

		cur[0], cur[1], cur[2] = cur[1], cur[2], r
		if cur[0] == 0 {
			// start of file.
			continue
		}

		t.trigrams[runesToNGram(cur)] = struct{}{}
		if len(t.trigrams) > maxTrigramCount {
			// probably not text.
			return fmt.Errorf("number of trigrams exceeds %d", maxTrigramCount)
		}
	}
	return nil
}

func (t *DocChecker) clearTrigrams(maxTrigramCount int) {
	if t.trigrams == nil {
		t.trigrams = make(map[ngram]struct{}, maxTrigramCount)
	}
	for key := range t.trigrams {
		delete(t.trigrams, key)
	}
}
