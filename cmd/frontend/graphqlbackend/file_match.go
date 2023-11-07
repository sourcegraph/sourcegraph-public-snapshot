package graphqlbackend

import (
	"fmt"
	"reflect"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// FileMatchResolver is a resolver for the GraphQL type `FileMatch`
type FileMatchResolver struct {
	result.FileMatch

	RepoResolver *RepositoryResolver
	db           database.DB
}

// Equal provides custom comparison which is used by go-cmp
func (fm *FileMatchResolver) Equal(other *FileMatchResolver) bool {
	return reflect.DeepEqual(fm, other)
}

func (fm *FileMatchResolver) Key() string {
	return fmt.Sprintf("%#v", fm.FileMatch.Key())
}

func (fm *FileMatchResolver) File() *GitTreeEntryResolver {
	// NOTE(sqs): Omits other commit fields to avoid needing to fetch them
	// (which would make it slow). This GitCommitResolver will return empty
	// values for all other fields.
	opts := GitTreeEntryResolverOpts{
		Commit: fm.Commit(),
		Stat:   CreateFileInfo(fm.Path, false),
	}
	return NewGitTreeEntryResolver(fm.db, gitserver.NewClient("graphql.filematch.tree"), opts)
}

func (fm *FileMatchResolver) Commit() *GitCommitResolver {
	commit := NewGitCommitResolver(fm.db, gitserver.NewClient("graphql.filematch.commit"), fm.RepoResolver, fm.CommitID, nil)
	commit.inputRev = fm.InputRev
	return commit
}

func (fm *FileMatchResolver) Repository() *RepositoryResolver {
	return fm.RepoResolver
}

func (fm *FileMatchResolver) RevSpec() *gitRevSpec {
	if fm.InputRev == nil || *fm.InputRev == "" {
		return nil // default branch
	}
	return &gitRevSpec{
		expr: &gitRevSpecExpr{expr: *fm.InputRev, repo: fm.Repository()},
	}
}

func (fm *FileMatchResolver) Symbols() []symbolResolver {
	return symbolResultsToResolvers(fm.db, fm.Commit(), fm.FileMatch.Symbols)
}

func (fm *FileMatchResolver) LineMatches() []lineMatchResolver {
	lineMatches := fm.FileMatch.ChunkMatches.AsLineMatches()
	r := make([]lineMatchResolver, 0, len(lineMatches))
	for _, lm := range lineMatches {
		r = append(r, lineMatchResolver{lm})
	}
	return r
}

func (fm *FileMatchResolver) ChunkMatches() []chunkMatchResolver {
	r := make([]chunkMatchResolver, 0, len(fm.FileMatch.ChunkMatches))
	for _, cm := range fm.FileMatch.ChunkMatches {
		r = append(r, chunkMatchResolver{cm})
	}
	return r
}

func (fm *FileMatchResolver) LimitHit() bool {
	return fm.FileMatch.LimitHit
}

func (fm *FileMatchResolver) ToRepository() (*RepositoryResolver, bool) { return nil, false }
func (fm *FileMatchResolver) ToFileMatch() (*FileMatchResolver, bool)   { return fm, true }
func (fm *FileMatchResolver) ToCommitSearchResult() (*CommitSearchResultResolver, bool) {
	return nil, false
}

type lineMatchResolver struct {
	*result.LineMatch
}

func (lm lineMatchResolver) Preview() string {
	return lm.LineMatch.Preview
}

func (lm lineMatchResolver) LineNumber() int32 {
	return lm.LineMatch.LineNumber
}

func (lm lineMatchResolver) OffsetAndLengths() [][]int32 {
	r := make([][]int32, len(lm.LineMatch.OffsetAndLengths))
	for i := range lm.LineMatch.OffsetAndLengths {
		r[i] = lm.LineMatch.OffsetAndLengths[i][:]
	}
	return r
}

func (lm lineMatchResolver) LimitHit() bool {
	return false
}

type chunkMatchResolver struct {
	result.ChunkMatch
}

func (c chunkMatchResolver) Content() string {
	return c.ChunkMatch.Content
}

func (c chunkMatchResolver) ContentStart() searchPositionResolver {
	return searchPositionResolver{c.ChunkMatch.ContentStart}
}

func (c chunkMatchResolver) Ranges() []searchRangeResolver {
	res := make([]searchRangeResolver, 0, len(c.ChunkMatch.Ranges))
	for _, r := range c.ChunkMatch.Ranges {
		res = append(res, searchRangeResolver{r})
	}
	return res
}

type searchPositionResolver struct {
	result.Location
}

func (l searchPositionResolver) Line() int32 {
	return int32(l.Location.Line)
}

func (l searchPositionResolver) Character() int32 {
	return int32(l.Location.Column)
}

type searchRangeResolver struct {
	result.Range
}

func (r searchRangeResolver) Start() searchPositionResolver {
	return searchPositionResolver{r.Range.Start}
}

func (r searchRangeResolver) End() searchPositionResolver {
	return searchPositionResolver{r.Range.End}
}
