package graphqlbackend

import (
	"fmt"
	"reflect"

	"github.com/sourcegraph/sourcegraph/internal/database"
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
	return NewGitTreeEntryResolver(fm.db, fm.Commit(), CreateFileInfo(fm.Path, false))
}

func (fm *FileMatchResolver) Commit() *GitCommitResolver {
	commit := NewGitCommitResolver(fm.db, fm.RepoResolver, fm.CommitID, nil)
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
