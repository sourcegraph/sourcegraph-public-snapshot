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
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"golang.org/x/net/trace"

	"github.com/google/zoekt/query"
)

const maxUInt16 = 0xffff

// DebugScore controls whether we collect data on match scores are
// constructed. Intended for use in tests.
var DebugScore = false

func (m *FileMatch) addScore(what string, s float64) {
	if DebugScore {
		m.Debug += fmt.Sprintf("%s:%f, ", what, s)
	}
	m.Score += s
}

func (d *indexData) simplify(in query.Q) query.Q {
	eval := query.Map(in, func(q query.Q) query.Q {
		switch r := q.(type) {
		case *query.Repo:
			return &query.Const{Value: strings.Contains(d.repoMetaData.Name, r.Pattern)}
		case *query.RepoSet:
			return &query.Const{Value: r.Set[d.repoMetaData.Name]}
		case *query.Language:
			_, has := d.metaData.LanguageMap[r.Language]
			if !has {
				return &query.Const{Value: false}
			}
		}
		return q
	})
	return query.Simplify(eval)
}

func (o *SearchOptions) SetDefaults() {
	if o.ShardMaxMatchCount == 0 {
		// We cap the total number of matches, so overly broad
		// searches don't crash the machine.
		o.ShardMaxMatchCount = 100000
	}
	if o.TotalMaxMatchCount == 0 {
		o.TotalMaxMatchCount = 10 * o.ShardMaxMatchCount
	}
	if o.ShardMaxImportantMatch == 0 {
		o.ShardMaxImportantMatch = 10
	}
	if o.TotalMaxImportantMatch == 0 {
		o.TotalMaxImportantMatch = 10 * o.ShardMaxImportantMatch
	}
}

func (d *indexData) Search(ctx context.Context, q query.Q, opts *SearchOptions) (sr *SearchResult, err error) {
	copyOpts := *opts
	opts = &copyOpts
	opts.SetDefaults()
	importantMatchCount := 0

	var res SearchResult
	if len(d.fileNameIndex) == 0 {
		return &res, nil
	}

	select {
	case <-ctx.Done():
		res.Stats.ShardsSkipped++
		return &res, nil
	default:
	}

	tr := trace.New("indexData.Search", d.file.Name())
	tr.LazyPrintf("opts: %+v", opts)
	defer func() {
		if sr != nil {
			tr.LazyPrintf("num files: %d", len(sr.Files))
			tr.LazyPrintf("stats: %+v", sr.Stats)
		}
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	q = d.simplify(q)
	tr.LazyLog(q, true)
	if c, ok := q.(*query.Const); ok && !c.Value {
		return &res, nil
	}

	if opts.EstimateDocCount {
		res.Stats.ShardFilesConsidered = len(d.fileBranchMasks)
		return &res, nil
	}

	q = query.Map(q, query.ExpandFileContent)

	mt, err := d.newMatchTree(q, &res.Stats)
	if err != nil {
		return nil, err
	}

	totalAtomCount := 0
	visitMatchTree(mt, func(t matchTree) {
		totalAtomCount++
	})

	cp := &contentProvider{
		id:    d,
		stats: &res.Stats,
	}

	docCount := uint32(len(d.fileBranchMasks))
	lastDoc := int(-1)

nextFileMatch:
	for {
		canceled := false
		select {
		case <-ctx.Done():
			canceled = true
		default:
		}

		nextDoc := mt.nextDoc()
		if int(nextDoc) <= lastDoc {
			nextDoc = uint32(lastDoc + 1)
		}
		if nextDoc >= docCount {
			break
		}
		lastDoc = int(nextDoc)

		if canceled || res.Stats.MatchCount >= opts.ShardMaxMatchCount ||
			importantMatchCount >= opts.ShardMaxImportantMatch {
			res.Stats.FilesSkipped += d.repoListEntry.Stats.Documents - lastDoc
			break
		}

		res.Stats.FilesConsidered++
		mt.prepare(nextDoc)

		cp.setDocument(nextDoc)

		known := make(map[matchTree]bool)
		for cost := costMin; cost <= costMax; cost++ {
			v, ok := mt.matches(cp, cost, known)
			if ok && !v {
				continue nextFileMatch
			}

			if cost == costMax && !ok {
				log.Panicf("did not decide. Repo %s, doc %d, known %v",
					d.repoMetaData.Name, nextDoc, known)
			}
		}

		fileMatch := FileMatch{
			Repository: d.repoMetaData.Name,
			FileName:   string(d.fileName(nextDoc)),
			Checksum:   d.getChecksum(nextDoc),
			Language:   d.languageMap[d.languages[nextDoc]],
		}

		if s := d.subRepos[nextDoc]; s > 0 {
			if s >= uint32(len(d.subRepoPaths)) {
				log.Panicf("corrupt index: subrepo %d beyond %v", s, d.subRepoPaths)
			}
			path := d.subRepoPaths[s]
			fileMatch.SubRepositoryPath = path
			sr := d.repoMetaData.SubRepoMap[path]
			fileMatch.SubRepositoryName = sr.Name
			if idx := d.branchIndex(nextDoc); idx >= 0 {
				fileMatch.Version = sr.Branches[idx].Version
			}
		} else {
			idx := d.branchIndex(nextDoc)
			if idx >= 0 {
				fileMatch.Version = d.repoMetaData.Branches[idx].Version
			}
		}

		atomMatchCount := 0
		visitMatches(mt, known, func(mt matchTree) {
			atomMatchCount++
		})
		finalCands := gatherMatches(mt, known)

		if len(finalCands) == 0 {
			nm := d.fileName(nextDoc)
			finalCands = append(finalCands,
				&candidateMatch{
					caseSensitive: false,
					fileName:      true,
					substrBytes:   nm,
					substrLowered: nm,
					file:          nextDoc,
					runeOffset:    0,
					byteOffset:    0,
					byteMatchSz:   uint32(len(nm)),
				})
		}
		fileMatch.LineMatches = cp.fillMatches(finalCands)

		maxFileScore := 0.0
		for i := range fileMatch.LineMatches {
			if maxFileScore < fileMatch.LineMatches[i].Score {
				maxFileScore = fileMatch.LineMatches[i].Score

			}

			// Order by ordering in file.
			fileMatch.LineMatches[i].Score += scoreLineOrderFactor * (1.0 - (float64(i) / float64(len(fileMatch.LineMatches))))
		}

		// Maintain ordering of input files. This
		// strictly dominates the in-file ordering of
		// the matches.
		fileMatch.addScore("fragment", maxFileScore)
		fileMatch.addScore("atom", float64(atomMatchCount)/float64(totalAtomCount)*scoreFactorAtomMatch)

		// Prefer earlier docs.
		fileMatch.addScore("doc-order", scoreFileOrderFactor*(1.0-float64(nextDoc)/float64(len(d.boundaries))))
		fileMatch.addScore("shard-order", scoreShardRankFactor*float64(d.repoMetaData.Rank)/maxUInt16)

		if fileMatch.Score > scoreImportantThreshold {
			importantMatchCount++
		}
		fileMatch.Branches = d.gatherBranches(nextDoc, mt, known)
		sortMatchesByScore(fileMatch.LineMatches)
		if opts.Whole {
			fileMatch.Content = cp.data(false)
		}

		res.Files = append(res.Files, fileMatch)
		res.Stats.MatchCount += len(fileMatch.LineMatches)
		res.Stats.FileCount++
	}
	SortFilesByScore(res.Files)

	addRepo(&res, &d.repoMetaData)
	for _, v := range d.repoMetaData.SubRepoMap {
		addRepo(&res, v)
	}

	mt.updateStats(&res.Stats)
	return &res, nil
}

func addRepo(res *SearchResult, repo *Repository) {
	if res.RepoURLs == nil {
		res.RepoURLs = map[string]string{}
	}
	res.RepoURLs[repo.Name] = repo.FileURLTemplate

	if res.LineFragments == nil {
		res.LineFragments = map[string]string{}
	}
	res.LineFragments[repo.Name] = repo.LineFragmentTemplate
}

type sortByOffsetSlice []*candidateMatch

func (m sortByOffsetSlice) Len() int      { return len(m) }
func (m sortByOffsetSlice) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m sortByOffsetSlice) Less(i, j int) bool {
	return m[i].byteOffset < m[j].byteOffset
}

// Gather matches from this document. This never returns a mixture of
// filename/content matches: if there are content matches, all
// filename matches are trimmed from the result. The matches are
// returned in document order and are non-overlapping.
func gatherMatches(mt matchTree, known map[matchTree]bool) []*candidateMatch {
	var cands []*candidateMatch
	visitMatches(mt, known, func(mt matchTree) {
		if smt, ok := mt.(*substrMatchTree); ok {
			cands = append(cands, smt.current...)
		}
		if rmt, ok := mt.(*regexpMatchTree); ok {
			cands = append(cands, rmt.found...)
		}
	})

	foundContentMatch := false
	for _, c := range cands {
		if !c.fileName {
			foundContentMatch = true
			break
		}
	}

	res := cands[:0]
	for _, c := range cands {
		if !foundContentMatch || !c.fileName {
			res = append(res, c)
		}
	}
	cands = res

	// Merge adjacent candidates. This guarantees that the matches
	// are non-overlapping.
	sort.Sort((sortByOffsetSlice)(cands))
	res = cands[:0]
	for i, c := range cands {
		if i == 0 {
			res = append(res, c)
			continue
		}
		last := res[len(res)-1]
		lastEnd := last.byteOffset + last.byteMatchSz
		end := c.byteOffset + c.byteMatchSz
		if lastEnd >= c.byteOffset {
			if end > lastEnd {
				last.byteMatchSz = end - last.byteOffset
			}
			continue
		}

		res = append(res, c)
	}

	return res
}

func (d *indexData) branchIndex(docID uint32) int {
	mask := d.fileBranchMasks[docID]
	idx := 0
	for mask != 0 {
		if mask&0x1 != 0 {
			return idx
		}
		idx++
		mask >>= 1
	}
	return -1
}

// gatherBranches returns a list of branch names.
func (d *indexData) gatherBranches(docID uint32, mt matchTree, known map[matchTree]bool) []string {
	foundBranchQuery := false
	var branches []string

	visitMatches(mt, known, func(mt matchTree) {
		bq, ok := mt.(*branchQueryMatchTree)
		if ok {
			foundBranchQuery = true
			branches = append(branches,
				d.branchNames[uint(bq.mask)])
		}
	})

	if !foundBranchQuery {
		mask := d.fileBranchMasks[docID]
		id := uint32(1)
		for mask != 0 {
			if mask&0x1 != 0 {
				branches = append(branches, d.branchNames[uint(id)])
			}
			id <<= 1
			mask >>= 1
		}
	}
	return branches
}

func (d *indexData) List(ctx context.Context, q query.Q) (rl *RepoList, err error) {
	tr := trace.New("indexData.List", d.file.Name())
	defer func() {
		if rl != nil {
			tr.LazyPrintf("repos size: %d", len(rl.Repos))
			tr.LazyPrintf("crashes: %d", rl.Crashes)
		}
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	q = d.simplify(q)
	tr.LazyLog(q, true)
	c, ok := q.(*query.Const)

	if !ok {
		return nil, fmt.Errorf("List should receive Repo-only query.")
	}

	l := &RepoList{}
	if c.Value {
		l.Repos = append(l.Repos, &d.repoListEntry)
	}
	return l, nil
}
