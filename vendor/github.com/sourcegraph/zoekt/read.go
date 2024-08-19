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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"log"
	"os"
	"sort"

	"github.com/rs/xid"
)

// IndexFile is a file suitable for concurrent read access. For performance
// reasons, it allows a mmap'd implementation.
type IndexFile interface {
	Read(off uint32, sz uint32) ([]byte, error)
	Size() (uint32, error)
	Close()
	Name() string
}

// reader is a stateful file
type reader struct {
	r   IndexFile
	off uint32
}

func (r *reader) seek(off uint32) {
	r.off = off
}

func (r *reader) U32() (uint32, error) {
	b, err := r.r.Read(r.off, 4)
	r.off += 4
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

func (r *reader) U64() (uint64, error) {
	b, err := r.r.Read(r.off, 8)
	r.off += 8
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

func (r *reader) ReadByte() (byte, error) {
	b, err := r.r.Read(r.off, 1)
	r.off += 1
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *reader) Varint() (uint64, error) {
	v, err := binary.ReadUvarint(r)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (r *reader) Str() (string, error) {
	slen, err := r.Varint()
	if err != nil {
		return "", err
	}
	b, err := r.r.Read(r.off, uint32(slen))
	if err != nil {
		return "", err
	}
	r.off += uint32(slen)
	return string(b), nil
}

func (r *reader) readTOC(toc *indexTOC) error {
	sz, err := r.r.Size()
	if err != nil {
		return err
	}
	r.off = sz - 8

	var tocSection simpleSection
	if err := tocSection.read(r); err != nil {
		return err
	}

	r.seek(tocSection.off)

	sectionCount, err := r.U32()
	if err != nil {
		return err
	}

	if sectionCount == 0 {
		// tagged sections are indicated by a 0 sectionCount,
		// and then a list of string-tagged type-indicated sections.
		secs := toc.sectionsTagged()
		for r.off < tocSection.off+tocSection.sz {
			tag, err := r.Str()
			if err != nil {
				return err
			}
			kind, err := r.Varint()
			if err != nil {
				return err
			}
			sec := secs[tag]
			if sec != nil && sec.kind() == sectionKind(kind) {
				// happy path
				if err := sec.read(r); err != nil {
					return err
				}
				continue
			}
			// error case: skip over unknown section
			if sec == nil {
				log.Printf("file %s TOC has unknown section %q", r.r.Name(), tag)
			} else {
				return fmt.Errorf("file %s TOC section %q expects kind %d, got kind %d", r.r.Name(), tag,
					kind, sec.kind())
			}
			if kind == 0 {
				if err := (&simpleSection{}).read(r); err != nil {
					return err
				}
			} else if kind == 1 {
				if err := (&compoundSection{}).read(r); err != nil {
					return err
				}
			}
		}
	} else {
		// TODO: Remove this branch when ReaderMinFeatureVersion >= 10

		secs := toc.sections()

		if len(secs) != int(sectionCount) {
			secs = toc.sectionsNext()
		}

		if len(secs) != int(sectionCount) {
			return fmt.Errorf("section count mismatch: got %d want %d", sectionCount, len(secs))
		}

		for _, s := range secs {
			if err := s.read(r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *indexData) readSectionBlob(sec simpleSection) ([]byte, error) {
	return r.file.Read(sec.off, sec.sz)
}

func readSectionU32(f IndexFile, sec simpleSection) ([]uint32, error) {
	if sec.sz%4 != 0 {
		return nil, fmt.Errorf("barf: section size %% 4 != 0: sz %d ", sec.sz)
	}
	blob, err := f.Read(sec.off, sec.sz)
	if err != nil {
		return nil, err
	}
	arr := make([]uint32, 0, len(blob)/4)
	for len(blob) > 0 {
		arr = append(arr, binary.BigEndian.Uint32(blob))
		blob = blob[4:]
	}
	return arr, nil
}

func readSectionU64(f IndexFile, sec simpleSection) ([]uint64, error) {
	if sec.sz%8 != 0 {
		return nil, fmt.Errorf("barf: section size %% 8 != 0: sz %d ", sec.sz)
	}
	blob, err := f.Read(sec.off, sec.sz)
	if err != nil {
		return nil, err
	}
	arr := make([]uint64, 0, len(blob)/8)
	for len(blob) > 0 {
		arr = append(arr, binary.BigEndian.Uint64(blob))
		blob = blob[8:]
	}
	return arr, nil
}

func (r *reader) readJSON(data interface{}, sec *simpleSection) error {
	blob, err := r.r.Read(sec.off, sec.sz)
	if err != nil {
		return err
	}

	return json.Unmarshal(blob, data)
}

// canReadVersion returns checks if zoekt can read in md. If it can't a
// non-nil error is returned.
func canReadVersion(md *IndexMetadata) bool {
	// Backwards compatible with v16
	return md.IndexFormatVersion == IndexFormatVersion || md.IndexFormatVersion == NextIndexFormatVersion
}

func (r *reader) readIndexData(toc *indexTOC) (*indexData, error) {
	d := indexData{
		file:        r.r,
		branchIDs:   []map[string]uint{},
		branchNames: []map[uint]string{},
	}

	repos, md, err := r.readMetadata(toc)
	if md != nil && !canReadVersion(md) {
		return nil, fmt.Errorf("file is v%d, want v%d", md.IndexFormatVersion, IndexFormatVersion)
	} else if err != nil {
		return nil, err
	}

	d.metaData = *md
	d.repoMetaData = make([]Repository, 0, len(repos))
	for _, r := range repos {
		d.repoMetaData = append(d.repoMetaData, *r)
	}

	if d.metaData.IndexFeatureVersion < ReadMinFeatureVersion {
		return nil, fmt.Errorf("file is feature version %d, want feature version >= %d", d.metaData.IndexFeatureVersion, ReadMinFeatureVersion)
	}

	if d.metaData.IndexMinReaderVersion > FeatureVersion {
		return nil, fmt.Errorf("file needs read feature version >= %d, have read feature version %d", d.metaData.IndexMinReaderVersion, FeatureVersion)
	}

	d.boundariesStart = toc.fileContents.data.off
	d.boundaries = toc.fileContents.relativeIndex()
	d.newlinesStart = toc.newlines.data.off
	d.newlinesIndex = toc.newlines.relativeIndex()
	d.docSectionsStart = toc.fileSections.data.off
	d.docSectionsIndex = toc.fileSections.relativeIndex()

	d.symbols.symKindIndex = toc.symbolKindMap.relativeIndex()
	d.fileEndSymbol, err = readSectionU32(d.file, toc.fileEndSymbol)
	if err != nil {
		return nil, err
	}

	// Call readSectionBlob on each section key, and store the result in
	// the blob value.
	for sect, blob := range map[simpleSection]*[]byte{
		toc.symbolMap.index:    &d.symbols.symIndex,
		toc.symbolMap.data:     &d.symbols.symContent,
		toc.symbolKindMap.data: &d.symbols.symKindContent,
		toc.symbolMetaData:     &d.symbols.symMetaData,
	} {
		if *blob, err = d.readSectionBlob(sect); err != nil {
			return nil, err
		}
	}

	d.checksums, err = d.readSectionBlob(toc.contentChecksums)
	if err != nil {
		return nil, err
	}

	d.languages, err = d.readSectionBlob(toc.languages)
	if err != nil {
		return nil, err
	}

	d.contentNgrams, err = d.newBtreeIndex(toc.ngramText, toc.postings)
	if err != nil {
		return nil, err
	}

	d.fileBranchMasks, err = readSectionU64(d.file, toc.branchMasks)
	if err != nil {
		return nil, err
	}

	d.fileNameContent, err = d.readSectionBlob(toc.fileNames.data)
	if err != nil {
		return nil, err
	}

	d.fileNameIndex = toc.fileNames.relativeIndex()

	d.fileNameNgrams, err = d.newBtreeIndex(toc.nameNgramText, toc.namePostings)
	if err != nil {
		return nil, err
	}

	for _, md := range d.repoMetaData {
		repoBranchIDs := make(map[string]uint, len(md.Branches))
		repoBranchNames := make(map[uint]string, len(md.Branches))
		for j, br := range md.Branches {
			id := uint(1) << uint(j)
			repoBranchIDs[br.Name] = id
			repoBranchNames[id] = br.Name
		}
		d.branchIDs = append(d.branchIDs, repoBranchIDs)
		d.branchNames = append(d.branchNames, repoBranchNames)
		d.rawConfigMasks = append(d.rawConfigMasks, encodeRawConfig(md.RawConfig))
	}

	blob, err := d.readSectionBlob(toc.runeDocSections)
	if err != nil {
		return nil, err
	}

	d.runeDocSections = unmarshalDocSections(blob, nil)

	var runeOffsets, fileNameRuneOffsets []uint32

	for sect, dest := range map[simpleSection]*[]uint32{
		toc.subRepos:        &d.subRepos,
		toc.runeOffsets:     &runeOffsets,
		toc.nameRuneOffsets: &fileNameRuneOffsets,
		toc.nameEndRunes:    &d.fileNameEndRunes,
		toc.fileEndRunes:    &d.fileEndRunes,
	} {
		if blob, err := d.readSectionBlob(sect); err != nil {
			return nil, err
		} else {
			*dest = fromSizedDeltas(blob, nil)
		}
	}

	d.runeOffsets = makeRuneOffsetMap(runeOffsets)
	d.fileNameRuneOffsets = makeRuneOffsetMap(fileNameRuneOffsets)

	d.subRepoPaths = make([][]string, 0, len(d.repoMetaData))
	for i := 0; i < len(d.repoMetaData); i++ {
		keys := make([]string, 0, len(d.repoMetaData[i].SubRepoMap)+1)
		keys = append(keys, "")
		for k := range d.repoMetaData[i].SubRepoMap {
			if k != "" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		d.subRepoPaths = append(d.subRepoPaths, keys)
	}

	d.languageMap = map[uint16]string{}
	for k, v := range d.metaData.LanguageMap {
		d.languageMap[v] = k
	}

	if err := d.verify(); err != nil {
		return nil, err
	}

	// roc.ranks.sz = 0 indicates that we are reading a shard without ranks, in
	// which case we skip reading the section and leave d.ranks = nil
	if toc.ranks.sz > 0 {
		err = d.readRanks(toc)
		if err != nil {
			return nil, err
		}
	}

	if d.metaData.IndexFormatVersion >= 17 {
		blob, err := d.readSectionBlob(toc.repos)
		if err != nil {
			return nil, err
		}
		d.repos = fromSizedDeltas16(blob, nil)
	} else {
		// every document is for repo index 0 (default value of uint16)
		d.repos = make([]uint16, len(d.fileBranchMasks))
	}

	if err := d.calculateStats(); err != nil {
		return nil, err
	}

	return &d, nil
}

func (r *reader) readMetadata(toc *indexTOC) ([]*Repository, *IndexMetadata, error) {
	var md IndexMetadata
	if err := r.readJSON(&md, &toc.metaData); err != nil {
		return nil, nil, err
	}

	// Sourcegraph specific: we support mutating metadata via an additional
	// ".meta" file. This is to support tombstoning. An additional benefit is we
	// can update metadata (such as Rank and Name) without re-indexing content.
	blob, err := os.ReadFile(r.r.Name() + ".meta")
	if err != nil && !os.IsNotExist(err) {
		return nil, &md, fmt.Errorf("failed to read meta file: %w", err)
	}

	if len(blob) == 0 {
		blob, err = r.r.Read(toc.repoMetaData.off, toc.repoMetaData.sz)
		if err != nil {
			return nil, &md, err
		}
	}

	var repos []*Repository
	if md.IndexFormatVersion >= 17 {
		if err := json.Unmarshal(blob, &repos); err != nil {
			return nil, &md, err
		}
	} else {
		repos = make([]*Repository, 1)
		if err := json.Unmarshal(blob, &repos[0]); err != nil {
			return nil, &md, err
		}
	}

	if md.ID == "" {
		if len(repos) == 0 {
			return nil, nil, fmt.Errorf("len(repos)=0. Cannot backfill ID")
		}
		md.ID = backfillID(repos[0].Name)
	}

	return repos, &md, nil
}

const ngramEncoding = 8

func (d *indexData) newBtreeIndex(ngramSec simpleSection, postings compoundSection) (btreeIndex, error) {
	bi := btreeIndex{file: d.file}

	textContent, err := d.readSectionBlob(ngramSec)
	if err != nil {
		return btreeIndex{}, err
	}

	// For 500k trigams we can expect approx 1000 leaf nodes (500k divided by
	// half the bucketSize) and 20 nodes on level 2 (all but the rightmost
	// inner nodes will have exactly v=50 children) plus a root node.
	bt := newBtree(btreeOpts{bucketSize: btreeBucketSize, v: 50})
	for i := 0; i < len(textContent); i += ngramEncoding {
		ng := ngram(binary.BigEndian.Uint64(textContent[i : i+ngramEncoding]))
		bt.insert(ng)
	}
	bt.freeze()

	bi.bt = bt

	// hold on to simple sections (8 bytes each)
	bi.ngramSec = ngramSec
	bi.postingIndex = postings.index

	return bi, nil
}

func (d *indexData) verify() error {
	// This is not an exhaustive check: the postings can easily
	// generate OOB acccesses, and are expensive to check, but this lets us rule out
	// other sources of OOB access.
	n := len(d.fileNameIndex)
	if n == 0 {
		return nil
	}

	n--
	for what, got := range map[string]int{
		"boundaries":        len(d.boundaries) - 1,
		"branch masks":      len(d.fileBranchMasks),
		"doc section index": len(d.docSectionsIndex) - 1,
		"newlines index":    len(d.newlinesIndex) - 1,
	} {
		if got != n {
			return fmt.Errorf("got %s %d, want %d", what, got, n)
		}
	}
	return nil
}

func (d *indexData) readContents(i uint32) ([]byte, error) {
	return d.readSectionBlob(simpleSection{
		off: d.boundariesStart + d.boundaries[i],
		sz:  d.boundaries[i+1] - d.boundaries[i],
	})
}

func (d *indexData) readContentSlice(off uint32, sz uint32) ([]byte, error) {
	// TODO(hanwen): cap result if it is at the end of the content
	// section.
	return d.readSectionBlob(simpleSection{
		off: d.boundariesStart + off,
		sz:  sz,
	})
}

func (d *indexData) readNewlines(i uint32, buf []uint32) ([]uint32, uint32, error) {
	sec := simpleSection{
		off: d.newlinesStart + d.newlinesIndex[i],
		sz:  d.newlinesIndex[i+1] - d.newlinesIndex[i],
	}
	blob, err := d.readSectionBlob(sec)
	if err != nil {
		return nil, 0, err
	}

	return fromSizedDeltas(blob, buf), sec.sz, nil
}

func (d *indexData) readDocSections(i uint32, buf []DocumentSection) ([]DocumentSection, uint32, error) {
	sec := simpleSection{
		off: d.docSectionsStart + d.docSectionsIndex[i],
		sz:  d.docSectionsIndex[i+1] - d.docSectionsIndex[i],
	}
	blob, err := d.readSectionBlob(sec)
	if err != nil {
		return nil, 0, err
	}

	ds := unmarshalDocSections(blob, buf)

	// can be nil if buf is nil and there are no doc sections. However, we rely
	// on it being non-nil to cache the read.
	if ds == nil {
		ds = make([]DocumentSection, 0)
	}

	return ds, sec.sz, nil
}

func (d *indexData) readRanks(toc *indexTOC) error {
	blob, err := d.readSectionBlob(toc.ranks)
	if err != nil {
		return err
	}

	return decodeRanks(blob, &d.ranks)
}

// NewSearcher creates a Searcher for a single index file.  Search
// results coming from this searcher are valid only for the lifetime
// of the Searcher itself, ie. []byte members should be copied into
// fresh buffers if the result is to survive closing the shard.
func NewSearcher(r IndexFile) (Searcher, error) {
	rd := &reader{r: r}

	var toc indexTOC
	if err := rd.readTOC(&toc); err != nil {
		return nil, err
	}
	indexData, err := rd.readIndexData(&toc)
	if err != nil {
		return nil, err
	}
	indexData.file = r
	return indexData, nil
}

// ReadMetadata returns the metadata of index shard without reading
// the index data. The IndexFile is not closed.
func ReadMetadata(inf IndexFile) ([]*Repository, *IndexMetadata, error) {
	rd := &reader{r: inf}
	var toc indexTOC
	if err := rd.readTOC(&toc); err != nil {
		return nil, nil, err
	}

	return rd.readMetadata(&toc)
}

// ReadMetadataPathAlive is like ReadMetadataPath except that it only returns
// alive repositories.
func ReadMetadataPathAlive(p string) ([]*Repository, *IndexMetadata, error) {
	repos, id, err := ReadMetadataPath(p)
	if err != nil {
		return nil, nil, err
	}
	alive := repos[:0]
	for _, repo := range repos {
		if !repo.Tombstone {
			alive = append(alive, repo)
		}
	}
	return alive, id, nil
}

// ReadMetadataPath returns the metadata of index shard at p without reading
// the index data. ReadMetadataPath is a helper for ReadMetadata which opens
// the IndexFile at p.
func ReadMetadataPath(p string) ([]*Repository, *IndexMetadata, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	iFile, err := NewIndexFile(f)
	if err != nil {
		return nil, nil, err
	}
	defer iFile.Close()

	return ReadMetadata(iFile)
}

// IndexFilePaths returns all paths for the IndexFile at filepath p that
// exist. Note: if no files exist this will return an empty slice and nil
// error.
//
// This is p and the ".meta" file for p.
func IndexFilePaths(p string) ([]string, error) {
	paths := []string{p, p + ".meta"}
	exist := paths[:0]
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			exist = append(exist, p)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return exist, nil
}

func loadIndexData(r IndexFile) (*indexData, error) {
	rd := &reader{r: r}

	var toc indexTOC
	if err := rd.readTOC(&toc); err != nil {
		return nil, err
	}
	return rd.readIndexData(&toc)
}

// PrintNgramStats outputs a list of the form
//
//	n_1 trigram_1
//	n_2 trigram_2
//	...
//
// where n_i is the length of the postings list of trigram_i stored in r.
func PrintNgramStats(r IndexFile) error {
	id, err := loadIndexData(r)
	if err != nil {
		return err
	}

	var rNgram [3]rune
	for ngram, ss := range id.contentNgrams.DumpMap() {
		rNgram = ngramToRunes(ngram)
		fmt.Printf("%d\t%q\n", ss.sz, string(rNgram[:]))
	}
	return nil
}

var crc64Table = crc64.MakeTable(crc64.ECMA)

// backfillID returns a 20 char long sortable ID. The ID only depends on s. It
// should only be used to set the ID of simple v16 shards on read.
func backfillID(s string) string {
	var id xid.ID

	// Our timestamps are based on Unix time. Shards without IDs are assigned IDs
	// based on the 0 epoch.
	binary.BigEndian.PutUint32(id[:], 0)
	binary.BigEndian.PutUint64(id[4:], crc64.Checksum([]byte(s), crc64Table))
	return id.String()
}
