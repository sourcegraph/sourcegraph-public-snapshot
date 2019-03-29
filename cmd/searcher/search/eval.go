// NOTICE: This file contains code derived from
// https://github.com/google/zoekt (license follows).
//
// Copyright 2018 Google Inc. All rights reserved.
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

package search

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"sort"

	"github.com/pkg/errors"
	srcapi "github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	api "github.com/sourcegraph/sourcegraph/pkg/search/zoekt"
	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt/matchtree"
	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt/query"
	"golang.org/x/net/trace"
)

// This file adapts zoekt's matchtree to work on content without any indexes.

// The costs we have when evaluating. We start at 0 and work our way
// up.
const (
	costConst          = 0
	costFileNameSubstr = 1
	costFileNameRegexp = 2
	costContentSubstr  = 3
	costContentRegexp  = 4

	costMin = costConst
	costMax = costContentRegexp
)

const source = api.Source("textjit")

// StoreSearcher provides a pkg/search.Searcher which searches over a search
// store.
type StoreSearcher struct {
	Store *Store
}

// Search implements pkg/search.Search
func (s *StoreSearcher) Search(ctx context.Context, q query.Q, opts *api.Options) (sr *api.Result, err error) {
	if len(opts.Repositories) != 1 {
		return nil, errors.New("searcher requires only one repository specified")
	}

	var res api.Result
	repo := api.Repository{Name: opts.Repositories[0]}
	status := api.RepositoryStatusSearched

	tr := trace.New("search", repo.String())
	tr.LazyPrintf("query: %s", q)
	tr.LazyPrintf("opts:  %+v", opts)
	defer func() {
		if sr != nil {
			tr.LazyPrintf("num files: %d", len(sr.Files))
		}
		if err != nil {
			if status == api.RepositoryStatusSearched {
				status = api.RepositoryStatusError
			}
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.LazyPrintf("status: %s", status)
		tr.Finish()
		if err != nil {
			err = errors.Wrapf(err, "failed to search %s for %s", repo, q)
		}
	}()

	// Read and remove the commit from the query. We don't need to keep it
	// since our matchtree logic doesn't want it.
	q = query.Simplify(query.Map(q, func(q query.Q) query.Q {
		if s, ok := q.(*query.Ref); ok {
			repo.Commit = srcapi.CommitID(s.Pattern)
			return &query.Const{Value: true}
		}
		return q
	}, nil))
	if len(repo.Commit) != 40 {
		return nil, errors.Errorf("Commit must be resolved (Commit=%q)", repo.Commit)
	}

	emptyResultWithStatus := func(s api.RepositoryStatusType) *api.Result {
		status = s
		return &api.Result{Stats: api.Stats{Status: []api.RepositoryStatus{{Repository: repo, Source: source, Status: s}}}}
	}

	if c, ok := q.(*query.Const); ok && !c.Value {
		return &res, nil
	}

	q = query.Map(q, nil, query.ExpandFileContent)

	mt, err := newMatchTree(q)
	if err != nil {
		return nil, err
	}
	tr.LazyPrintf("matchtree: %s", mt)

	// Fetch/Open the zip file
	prepareCtx := ctx
	if opts.FetchTimeout > 0 {
		var cancel context.CancelFunc
		prepareCtx, cancel = context.WithTimeout(ctx, opts.FetchTimeout)
		defer cancel()
	}
	path, err := s.Store.prepareZip(prepareCtx, gitserver.Repo{Name: repo.Name}, repo.Commit)
	if err != nil {
		if errcode.IsTimeout(err) {
			return emptyResultWithStatus(api.RepositoryStatusTimedOut), nil
		} else if errcode.IsNotFound(err) {
			return emptyResultWithStatus(api.RepositoryStatusMissing), nil
		}
		return nil, err
	}
	zf, err := s.Store.zipCache.get(path)
	if err != nil {
		return nil, err
	}
	defer zf.Close()

	cp := &contentProvider{zf: zf}
	minDoc := uint32(0)
	maxDoc := cp.docCount()

nextFileMatch:
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				status = api.RepositoryStatusTimedOut
			}
			break nextFileMatch
		default:
		}

		nextDoc := mt.NextDoc()
		if nextDoc < minDoc {
			nextDoc = minDoc
		}
		if nextDoc >= maxDoc {
			break
		}
		minDoc = nextDoc + 1

		mt.Prepare(nextDoc)
		cp.setDocument(nextDoc)

		known := make(map[matchtree.MatchTree]bool)
		for cost := costMin; cost <= costMax; cost++ {
			v, ok := mt.Matches(cp, cost, known)
			if ok && !v {
				continue nextFileMatch
			}

			if cost == costMax && !ok {
				return nil, errors.Errorf("did not decide. doc %d, known %v", nextDoc, known)
			}
		}

		candidates := gatherMatches(mt, known)
		matches := cp.fillMatches(candidates)

		res.Files = append(res.Files, api.FileMatch{
			Path:        cp.file.Name,
			Repository:  repo,
			LineMatches: matches,
		})

		if opts.TotalMaxMatchCount > 0 && len(res.Files) > opts.TotalMaxMatchCount {
			status = api.RepositoryStatusLimitHit
			break
		}
	}

	res.Stats = api.Stats{
		MatchCount: matchCount(res.Files),
		Status:     []api.RepositoryStatus{{Repository: repo, Source: source, Status: status}},
	}

	return &res, nil
}

// Close implements pkg/search.Close
func (s *StoreSearcher) Close() {
	// We don't have a way to close Store yet, but storeSearcher lives the
	// same lifetime as the process (except in tests which is fine).
}

func (s *StoreSearcher) String() string {
	return "StoreSearcher(" + s.Store.String() + ")"
}

func matchCount(files []api.FileMatch) int {
	count := 0
	for _, f := range files {
		if f.IsPathMatch() {
			count++
		}
		count += len(f.LineMatches)
	}
	return count
}

type regexpMatchTree struct {
	regexp *regexp.Regexp

	fileName bool

	// mutable
	done  bool
	found []*candidateMatch

	// nextDoc, prepare.
	matchtree.All
}

type substrMatchTree struct {
	needle        []byte
	caseSensitive bool
	fileName      bool

	// mutable
	done  bool
	found []*candidateMatch

	// nextDoc, prepare.
	matchtree.All
}

func (t *regexpMatchTree) Prepare(doc uint32) {
	t.done = false
	t.found = t.found[:0]
	t.All.Prepare(doc)
}

func (t *substrMatchTree) Prepare(doc uint32) {
	t.done = false
	t.found = t.found[:0]
	t.All.Prepare(doc)
}

func (t *regexpMatchTree) String() string {
	return fmt.Sprintf("re(%s)", t.regexp)
}

func (t *substrMatchTree) String() string {
	f := ""
	if t.fileName {
		f = "f"
	}

	return fmt.Sprintf("%ssubstr(%q)", f, t.needle)
}

func (t *regexpMatchTree) Matches(cp matchtree.ContentProvider, cost int, known map[matchtree.MatchTree]bool) (bool, bool) {
	if t.done {
		return len(t.found) > 0, true
	}
	if t.fileName && cost < costFileNameRegexp {
		return false, false
	}
	if !t.fileName && cost < costContentRegexp {
		return false, false
	}

	idxs := t.regexp.FindAllIndex(cp.Data(t.fileName), -1)
	found := t.found
	if cap(found) < len(idxs) {
		found = make([]*candidateMatch, 0, len(idxs))
	}
	found = found[:len(idxs)]
	for i, idx := range idxs {
		found[i] = &candidateMatch{
			byteOffset:  idx[0],
			byteMatchSz: idx[1] - idx[0],
			fileName:    t.fileName,
		}
	}
	t.found = found

	t.done = true
	return len(t.found) > 0, true
}

func (t *substrMatchTree) Matches(cp matchtree.ContentProvider, cost int, known map[matchtree.MatchTree]bool) (bool, bool) {
	if t.done {
		return len(t.found) > 0, true
	}
	if t.fileName && cost < costFileNameSubstr {
		return false, false
	}
	if !t.fileName && cost < costContentSubstr {
		return false, false
	}

	dataOrig := cp.Data(t.fileName)
	data := dataOrig
	if !t.caseSensitive {
		// TODO cache
		data = bytes.ToLower(dataOrig)
	}

	var (
		found  = t.found[:0]
		offset = 0
	)
	for len(data) >= len(t.needle) {
		i := bytes.Index(data, t.needle)
		if i == -1 {
			break
		}
		found = append(found, &candidateMatch{
			byteOffset:  offset + i,
			byteMatchSz: len(t.needle),
			fileName:    t.fileName,
		})
		offset += i + len(t.needle)
		data = data[i+len(t.needle):]
	}
	t.found = found

	t.done = true
	return len(t.found) > 0, true
}

func newMatchTree(q query.Q) (matchtree.MatchTree, error) {
	atom := func(q query.Q) (matchtree.MatchTree, error) {
		switch s := q.(type) {
		case *query.Regexp:
			subQ := query.RegexpToQuery(s.Regexp, 3)
			subQ = query.Map(subQ, nil, func(q query.Q) query.Q {
				if sub, ok := q.(*query.Substring); ok {
					sub.FileName = s.FileName
					sub.CaseSensitive = s.CaseSensitive
				}
				return q
			})

			subMT, err := newMatchTree(subQ)
			if err != nil {
				return nil, err
			}

			prefix := ""
			if !s.CaseSensitive {
				prefix = "(?i)"
			}

			tr := &regexpMatchTree{
				regexp:   regexp.MustCompile(prefix + s.Regexp.String()),
				fileName: s.FileName,
			}

			return matchtree.And(tr, &matchtree.NoVisit{MatchTree: subMT}), nil

		case *query.Substring:
			b := []byte(s.Pattern)
			if !s.CaseSensitive {
				b = bytes.ToLower(b)
			}
			if len(b) == 0 {
				return nil, errors.Errorf("expected non-empty substring query")
			}
			return &substrMatchTree{
				needle:        b,
				caseSensitive: s.CaseSensitive,
				fileName:      s.FileName,
			}, nil

		}
		return nil, errors.Errorf("unexpected query atom %T: %v", q, q)
	}

	return matchtree.NewMatchTree(q, atom)
}

// candidateMatch is a candidate match for a substring.
type candidateMatch struct {
	// Offsets are relative to the start of the filename or file contents.
	byteOffset  int
	byteMatchSz int

	fileName bool
}

type contentProvider struct {
	zf   *zipFile
	file *srcFile

	// Cache
	fileName []byte
}

func (c *contentProvider) Data(fileName bool) []byte {
	if fileName {
		if c.fileName == nil {
			c.fileName = []byte(c.file.Name)
		}
		return c.fileName
	}
	return c.zf.DataFor(c.file)
}

func (c *contentProvider) docCount() uint32 {
	return uint32(len(c.zf.Files))
}

func (c *contentProvider) setDocument(i uint32) {
	c.file = &c.zf.Files[i]
	c.fileName = nil
}

func (c *contentProvider) fillMatches(candidates []*candidateMatch) []api.LineMatch {
	data := c.Data(false)
	var result []api.LineMatch
	// We assume candidates is sorted by byteOffset and has already had
	// overlapping matches merged.
	for len(candidates) > 0 {
		m := candidates[0]
		byteEnd := m.byteOffset + m.byteMatchSz
		lineStart := bytes.LastIndexByte(data[:m.byteOffset+1], '\n') + 1
		lineEnd := bytes.IndexByte(data[byteEnd:], '\n')
		if lineEnd < 0 {
			lineEnd = len(data)
		} else {
			lineEnd += byteEnd
		}

		fragments := []api.LineFragmentMatch{}
		for len(candidates) > 0 {
			m := candidates[0]
			if m.byteOffset+m.byteMatchSz > lineEnd {
				break
			}
			candidates = candidates[1:]

			fragments = append(fragments, api.LineFragmentMatch{
				LineOffset:  m.byteOffset - lineStart,
				MatchLength: m.byteMatchSz,
			})
		}

		result = append(result, api.LineMatch{
			// Intentionally create a copy since we can't hold onto data
			Line:          append([]byte{}, data[lineStart:lineEnd]...),
			LineNumber:    bytes.Count(data[:lineStart], []byte{'\n'}) + 1,
			LineFragments: fragments,
		})
	}

	return result
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
func gatherMatches(mt matchtree.MatchTree, known map[matchtree.MatchTree]bool) []*candidateMatch {
	var cands []*candidateMatch
	matchtree.VisitMatches(mt, known, func(mt matchtree.MatchTree) {
		if smt, ok := mt.(*substrMatchTree); ok {
			cands = append(cands, smt.found...)
		}
		if rmt, ok := mt.(*regexpMatchTree); ok {
			cands = append(cands, rmt.found...)
		}
	})

	// We only want content matches. Empty list means we match the filepath.
	res := cands[:0]
	for _, c := range cands {
		if !c.fileName {
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
