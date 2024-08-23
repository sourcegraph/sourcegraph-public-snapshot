// Copyright 2017 Google Inc. All rights reserved.
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

// IndexFormatVersion is a version number. It is increased every time the
// on-disk index format is changed.
// 5: subrepositories.
// 6: remove size prefix for posting varint list.
// 7: move subrepos into Repository struct.
// 8: move repoMetaData out of indexMetadata
// 9: use bigendian uint64 for trigrams.
// 10: sections for rune offsets.
// 11: file ends in rune offsets.
// 12: 64-bit branchmasks.
// 13: content checksums
// 14: languages
// 15: rune based symbol sections
// 16: ctags metadata
const IndexFormatVersion = 16

// FeatureVersion is increased if a feature is added that requires reindexing data
// without changing the format version
// 2: Rank field for shards.
// 3: Rank documents within shards
// 4: Dedup file bugfix
// 5: Remove max line size limit
// 6: Include '#' into the LineFragment template
// 7: Record skip reasons in the index.
// 8: Record source path in the index.
// 9: Store ctags metadata & bump default max file size
// 10: Compound shards; more flexible TOC format.
// 11: Bloom filters for file names & contents
// 12: go-enry for identifying file languages
const FeatureVersion = 12

// WriteMinFeatureVersion and ReadMinFeatureVersion constrain forwards and backwards
// compatibility. For example, if a new way to encode filenameNgrams on disk is
// added using a new section but the old one is retained, this would only bump
// FeatureVersion, since the previous version can read the file and ignore the
// new section, but the index files should be regenerated.
// When the new encoding is fully rolled out and stable, the section with the old
// encoding and the associated reader can be removed, and WriteMinFeatureVersion and
// ReadMinFeatureVersion can be set to the current FeatureVersion, indicating
// that the reader must handle the new version and that older versions are no
// longer valid.
// In this way, compatibility with arbitrary version offsets can be indicated.

// WriteMinFeatureVersion constrains forwards compatibility by emitting files
// that won't load in zoekt with a FeatureVersion below it.
const WriteMinFeatureVersion = 10

// ReadMinFeatureVersion constrains backwards compatibility by refusing to
// load a file with a FeatureVersion below it.
const ReadMinFeatureVersion = 8

// 17: compound shard (multi repo)
const NextIndexFormatVersion = 17

type indexTOC struct {
	fileContents compoundSection
	fileNames    compoundSection
	fileSections compoundSection
	postings     compoundSection
	newlines     compoundSection
	ngramText    simpleSection
	runeOffsets  simpleSection
	fileEndRunes simpleSection
	languages    simpleSection

	fileEndSymbol  simpleSection
	symbolMap      lazyCompoundSection
	symbolKindMap  compoundSection
	symbolMetaData simpleSection

	branchMasks simpleSection
	subRepos    simpleSection

	nameNgramText    simpleSection
	namePostings     compoundSection
	nameRuneOffsets  simpleSection
	metaData         simpleSection
	repoMetaData     simpleSection
	nameEndRunes     simpleSection
	contentChecksums simpleSection
	runeDocSections  simpleSection

	repos simpleSection

	ranks simpleSection
}

func (t *indexTOC) sections() []section {
	// This old sections list is only needed to maintain backwards compatibility,
	// and can be removed when a migration to tagged sections is complete.
	return []section{
		// This must be first, so it can be reliably read across
		// file format versions.
		&t.metaData,
		&t.repoMetaData,
		&t.fileContents,
		&t.fileNames,
		&t.fileSections,
		&t.fileEndSymbol,
		&t.symbolMap,
		&t.symbolKindMap,
		&t.symbolMetaData,
		&t.newlines,
		&t.ngramText,
		&t.postings,
		&t.nameNgramText,
		&t.namePostings,
		&t.branchMasks,
		&t.subRepos,
		&t.runeOffsets,
		&t.nameRuneOffsets,
		&t.fileEndRunes,
		&t.nameEndRunes,
		&t.contentChecksums,
		&t.languages,
		&t.runeDocSections,
	}
}

func (t *indexTOC) sectionsNext() []section {
	return append(t.sections(), &t.repos)
}

type taggedSection struct {
	tag string
	sec section
}

func (t *indexTOC) sectionsTagged() map[string]section {
	out := map[string]section{}
	for _, ent := range t.sectionsTaggedList() {
		out[ent.tag] = ent.sec
	}
	for _, ent := range t.sectionsTaggedCompatibilityList() {
		out[ent.tag] = ent.sec
	}
	return out
}

func (t *indexTOC) sectionsTaggedList() []taggedSection {
	var unusedSimple simpleSection

	return []taggedSection{
		{"metaData", &t.metaData},
		{"repoMetaData", &t.repoMetaData},
		{"fileContents", &t.fileContents},
		{"fileNames", &t.fileNames},
		{"fileSections", &t.fileSections},
		{"fileEndSymbol", &t.fileEndSymbol},
		{"symbolMap", &t.symbolMap},
		{"symbolKindMap", &t.symbolKindMap},
		{"symbolMetaData", &t.symbolMetaData},
		{"newlines", &t.newlines},
		{"ngramText", &t.ngramText},
		{"postings", &t.postings},
		{"nameNgramText", &t.nameNgramText},
		{"namePostings", &t.namePostings},
		{"branchMasks", &t.branchMasks},
		{"subRepos", &t.subRepos},
		{"runeOffsets", &t.runeOffsets},
		{"nameRuneOffsets", &t.nameRuneOffsets},
		{"fileEndRunes", &t.fileEndRunes},
		{"nameEndRunes", &t.nameEndRunes},
		{"contentChecksums", &t.contentChecksums},
		{"languages", &t.languages},
		{"runeDocSections", &t.runeDocSections},
		{"repos", &t.repos},

		// We no longer write bloom sections, but we still return them here to
		// avoid warnings about unknown sections.
		{"nameBloom", &unusedSimple},
		{"contentBloom", &unusedSimple},

		{"ranks", &t.ranks},
	}
}

// sectionsTaggedCompatibilityList returns a list of sections that will be
// handled or converted for backwards compatiblity, but aren't written by
// the current iteration of the indexer.
func (t *indexTOC) sectionsTaggedCompatibilityList() []taggedSection {
	return []taggedSection{}
}
