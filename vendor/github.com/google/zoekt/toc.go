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

// FormatVersion is a version number. It is increased every time the
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
const IndexFormatVersion = 15

// FeatureVersion is increased if a feature is added that requires reindexing data
// without changing the format version
// 2: Rank field for shards.
// 3: Rank documents within shards
// 4: Dedup file bugfix
// 5: Remove max line size limit
// 6: Include '#' into the LineFragment template
// 7: Record skip reasons in the index.
const FeatureVersion = 7

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
}

func (t *indexTOC) sections() []section {
	return []section{
		// This must be first, so it can be reliably read across
		// file format versions.
		&t.metaData,
		&t.repoMetaData,
		&t.fileContents,
		&t.fileNames,
		&t.fileSections,
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
