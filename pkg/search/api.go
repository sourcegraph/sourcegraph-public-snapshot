package search

import (
	"context"
	"fmt"
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

// LineMatch holds the matches within a single line in a file.
type LineMatch struct {
	// The line in which a match was found.
	Line       []byte
	LineStart  int
	LineEnd    int
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
	Name   api.RepoName
	Commit api.CommitID
}

func (r *Repository) String() string {
	return string(r.Name) + "@" + string(r.Commit)
}

// Result contains search matches and extra data
type Result struct {
	Files []FileMatch

	// TODO this can probably be generalised to support result types other
	// than Files.
}

// Searcher provides an interface to Searching.
type Searcher interface {
	Search(ctx context.Context, q query.Q, opts *Options) (*Result, error)
	Close()
	String() string
}

// Options for Search.
type Options struct {
	// Repositories if set limits search to the named repositories.
	Repositories []Repository

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
