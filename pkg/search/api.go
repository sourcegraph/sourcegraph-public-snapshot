package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
)

// FileMatch contains all the matches within a file.
type FileMatch struct {
	Path string

	Repository Repository

	// LineMatches contains the lines that match. If empty then we matched the
	// Path only.
	LineMatches []LineMatch
}

// IsPathMatch returns true if the FileMatch is a match for Path.
func (f *FileMatch) IsPathMatch() bool {
	return len(f.LineMatches) == 0
}

// LineMatch holds the matches within a single line in a file.
type LineMatch struct {
	// The line in which a match was found.
	Line       []byte
	LineNumber int

	LineFragments []LineFragmentMatch
}

// LineFragmentMatch a segment of matching text within a line.
type LineFragmentMatch struct {
	// Offset within the line, in bytes.
	LineOffset int

	// Number bytes that match.
	MatchLength int
}

// Repository is a repository at a commit.
type Repository struct {
	Name       api.RepoName
	Commit     api.CommitID
	RefPattern string
}

func (r *Repository) String() string {
	var b strings.Builder
	b.Grow(len(r.Name) + len(r.Commit) + len(r.RefPattern) + 2)
	b.WriteString(string(r.Name))
	if r.RefPattern != "" {
		b.WriteByte('@')
		b.WriteString(r.RefPattern)
	}
	if r.Commit != "" {
		b.WriteByte('@')
		b.WriteString(string(r.Commit))
	}
	return b.String()
}

// RepositoryStatus is the status of searching a repository with an
// attribution source.
type RepositoryStatus struct {
	// Repository is the repository searched.
	Repository Repository

	// Source is a short description of the searcher.
	Source Source

	// Status is the status of searching Repository for Source.
	Status RepositoryStatusType
}

func (r *RepositoryStatus) String() string {
	return fmt.Sprintf("RepositoryStatus(%s,%s)=%s", r.Source, r.Repository, r.Status)
}

// Source is a short description of a Searcher. eg: textjit, textindexed,
// symbol, ...
type Source string

// RepositoryStatusType is the status of searching a repository.
type RepositoryStatusType string

const (
	// RepositoryStatusSearched indicates the repository was successfully
	// searched. Note: If we hit match limits or timeouts while searching the
	// repository, its state will be limithit or timedout. So this can occur
	// even when we have results from the repository.
	RepositoryStatusSearched RepositoryStatusType = "searched"

	// RepositoryStatusLimitHit indicates the repository was not searched
	// since a match limit was reached.
	RepositoryStatusLimitHit RepositoryStatusType = "limithit"

	// RepositoryStatusTimedOut indicates that the search on the repository
	// timed out before finishing searching it.
	RepositoryStatusTimedOut RepositoryStatusType = "timedout"

	// RepositoryStatusCloning indicates the search failed for the repository
	// due to it still being cloned.
	RepositoryStatusCloning RepositoryStatusType = "cloning"

	// RepositoryStatusMissing indicates the search failed for the repository
	// since the repository does not exist.
	RepositoryStatusMissing RepositoryStatusType = "missing"

	// RepositoryStatusCommitMissing indicates the search failed for the
	// repository since the commit does not exist.
	RepositoryStatusCommitMissing RepositoryStatusType = "commitmissing"

	// RepositoryStatusError indicates the search failed for the repository
	// due to an unexpected error. Implementations of Searcher should not use
	// this type. Rather this type can be used in aggregators to indicate
	// partial failure.
	RepositoryStatusError RepositoryStatusType = "error"
)

// Stats contains statistics aggregated during searching.
type Stats struct {
	// MatchCount is the number of non-overlapping matches found. The match
	// count is decided by each Searcher. For example in text search each
	// LineMatch contributes to this count, but if its a match just on the
	// FileName, then that also contributes one.
	MatchCount int

	// RepositoryStatus explains the status of searching each repository
	// listed in Options.Repositories.
	Status []RepositoryStatus

	// Unavailable is a list of search sources which we wanted to use, but
	// were not available.
	Unavailable []Source
}

func (s *Stats) Add(o *Stats) {
	s.MatchCount += o.MatchCount
	s.Status = append(s.Status, o.Status...)
	s.Unavailable = append(s.Unavailable, o.Unavailable...)
}

// Result contains search matches and extra data
type Result struct {
	Stats

	Files []FileMatch

	// TODO this can probably be generalised to support result types other
	// than Files.
}

// Add combines the results from o into r.
func (r *Result) Add(o *Result) {
	r.Stats.Add(&o.Stats)
	r.Files = append(r.Files, o.Files...)
}

// Searcher provides an interface to Searching.
type Searcher interface {
	Search(ctx context.Context, q query.Q, opts *Options) (*Result, error)
	Close()
	String() string
}

// Options for Search.
type Options struct {
	// Repositories limits search to the named repositories.
	Repositories []api.RepoName

	// TotalMaxMatchCount if non-zero stops looking for more matches once we
	// have this many matches across shards.
	TotalMaxMatchCount int

	// MaxWallTime if non-zero will abort the search after this much time has
	// passed. It will return results found so far.
	MaxWallTime time.Duration

	// FetchTimeout if non-zero will abort fetching files to search after this
	// much time has passed. This is a subset of MaxWallTime. For example
	// searcher fetches file from gitserver which can be slow.
	FetchTimeout time.Duration

	// MaxDocDisplayCount if non-zero trims the number of results after
	// collating and sorting the results.
	MaxDocDisplayCount int
}

func (s *Options) String() string {
	return fmt.Sprintf("%#v", s)
}

// ShallowCopy returns a shallow copy of Options. Note: That means
// Repositories slice is the same underlying array.
func (s *Options) ShallowCopy() *Options {
	o := *s
	return &o
}
