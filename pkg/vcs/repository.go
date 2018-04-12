package vcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/pathmatch"
)

// RepoNotExistError is an error that reports a repository doesn't exist.
type RepoNotExistError struct {
	Repo api.RepoURI

	// CloneInProgress reports whether the repository is in process of being cloned.
	CloneInProgress bool
}

func (e *RepoNotExistError) Error() string {
	if e.CloneInProgress {
		return "repository does not exist (clone in progress): " + string(e.Repo)
	}
	return "repository does not exist: " + string(e.Repo)
}

// IsRepoNotExist reports if err is a RepoNotExistError.
func IsRepoNotExist(err error) bool {
	_, ok := err.(*RepoNotExistError)
	return ok
}

// IsCloneInProgress reports if err is a RepoNotExistError which has a clone
// in progress.
func IsCloneInProgress(err error) bool {
	if e, ok := err.(*RepoNotExistError); ok {
		return e.CloneInProgress
	}
	return false
}

var ErrRepoExist = errors.New("repository already exists")

// RevisionNotFoundError is an error that reports a revision doesn't exist.
type RevisionNotFoundError struct {
	Repo api.RepoURI
	Spec string
}

func (e *RevisionNotFoundError) Error() string {
	return fmt.Sprintf("revision not found: %s@%s", e.Repo, e.Spec)
}

// IsRevisionNotFound reports if err is a RevisionNotFoundError.
func IsRevisionNotFound(err error) bool {
	_, ok := err.(*RevisionNotFoundError)
	return ok
}

// A Repository is a VCS repository.
type Repository interface {
	String() string

	// ResolveRevision returns the revision that the given revision
	// specifier resolves to, or a non-nil error if there is no such
	// revision.
	//
	// Implementations may choose to return RevisionNotFoundError in all
	// cases where the revision is not found, or more specific errors
	// (such as ErrCommitNotFound) if spec can be partially resolved
	// or determined to be a certain kind of revision specifier.
	ResolveRevision(ctx context.Context, spec string, opt *ResolveRevisionOptions) (api.CommitID, error)

	// Branches returns a list of all branches in the repository.
	Branches(context.Context, BranchesOptions) ([]*Branch, error)

	// Tags returns a list of all tags in the repository.
	Tags(context.Context) ([]*Tag, error)

	// GetCommit returns the commit with the given commit ID, or
	// ErrCommitNotFound if no such commit exists.
	GetCommit(context.Context, api.CommitID) (*Commit, error)

	// Commits returns all commits matching the options.
	Commits(context.Context, CommitsOptions) ([]*Commit, error)

	// CommitCount returns the number of commits that would be returned by Commits.
	CommitCount(context.Context, CommitsOptions) (total uint, err error)

	// ShortLog returns the per-author commit statistics of the repo.
	ShortLog(context.Context, ShortLogOptions) ([]*PersonCount, error)

	// Stat returns a FileInfo describing the named file at commit. If the file
	// is a symbolic link, the returned FileInfo describes the symbolic link.
	// Lstat makes no attempt to follow the link.
	Lstat(ctx context.Context, commit api.CommitID, name string) (os.FileInfo, error)

	// Stat returns a FileInfo describing the named file at commit.
	Stat(ctx context.Context, commit api.CommitID, name string) (os.FileInfo, error)

	// ReadFile returns the content of the named file at commit.
	ReadFile(ctx context.Context, commit api.CommitID, name string) ([]byte, error)

	// Readdir reads the contents of the named directory at commit.
	ReadDir(ctx context.Context, commit api.CommitID, name string, recurse bool) ([]os.FileInfo, error)

	BlameFile(ctx context.Context, path string, opt *BlameOptions) ([]*Hunk, error)

	// Blames a file for the extension API returning GitBlame data as a raw string.
	BlameFileRaw(ctx context.Context, path string, opt *BlameOptions) (string, error)

	// ExecReader executes an arbitrary `git` command (`git [args...]`) and returns a reader
	// connected to its stdout.
	ExecReader(ctx context.Context, args []string) (r io.ReadCloser, err error)

	// Allows VSCode extensions to execute whitelisted commands from gitserver.
	GitCmdRaw(ctx context.Context, params []string) (string, error)

	// Diff shows changes between two commits. If base or head do not
	// exist, an error is returned.
	Diff(ctx context.Context, base, head api.CommitID, opt *DiffOptions) (*Diff, error)

	// MergeBase returns the merge base commit for the specified
	// commits.
	MergeBase(context.Context, api.CommitID, api.CommitID) (api.CommitID, error)

	// RawLogDiffSearch runs a raw `git log` command that is expected to return
	// logs with patches. It returns a subset of the output, including only hunks
	// that actually match the given pattern.
	//
	// If complete is false, then the results may have been parsed from only
	// partial output from the underlying git command (because, e.g., it
	// timed out during execution and only returned partial output).
	RawLogDiffSearch(ctx context.Context, opt RawLogDiffSearchOptions) (results []*LogCommitSearchResult, complete bool, err error)

	// BehindAhead returns the behind/ahead commit counts information
	// for right vs. left (both Git revspecs).
	BehindAhead(ctx context.Context, left, right string) (*BehindAhead, error)
}

// BlameOptions configures a blame.
type BlameOptions struct {
	NewestCommit api.CommitID `json:",omitempty" url:",omitempty"`
	OldestCommit api.CommitID `json:",omitempty" url:",omitempty"` // or "" for the root commit

	StartLine int `json:",omitempty" url:",omitempty"` // 1-indexed start byte (or 0 for beginning of file)
	EndLine   int `json:",omitempty" url:",omitempty"` // 1-indexed end byte (or 0 for end of file)
}

// A Hunk is a contiguous portion of a file associated with a commit.
type Hunk struct {
	StartLine int // 1-indexed start line number
	EndLine   int // 1-indexed end line number
	StartByte int // 0-indexed start byte position (inclusive)
	EndByte   int // 0-indexed end byte position (exclusive)
	api.CommitID
	Author  Signature
	Message string
}

// CommitsOptions specifies options for (Repository).Commits (Repository).CommitCount.
type CommitsOptions struct {
	Range string // commit range (revspec, "A..B", "A...B", etc.)

	N    uint // limit the number of returned commits to this many (0 means no limit)
	Skip uint // skip this many commits at the beginning

	MessageQuery string // include only commits whose commit message contains this substring

	Author string // include only commits whose author matches this
	After  string // include only commits after this date

	Path string // only commits modifying the given path are selected (optional)
}

// ShortLogOptions contains options for (Repository).ShortLog.
type ShortLogOptions struct {
	Range string // the range for which stats will be fetched
	After string // the date after which to collect commits
	Path  string // compute stats for commits that touch this path
}

// DiffOptions configures a diff.
type DiffOptions struct {
	Paths                 []string // constrain diff to these pathspecs
	DetectRenames         bool
	OrigPrefix, NewPrefix string // prefixes for orig and new filenames (e.g., "a/", "b/")

	ExcludeReachableFromBoth bool // like "<rev1>...<rev2>" (see `git rev-parse --help`)
}

// A Diff represents changes between two commits.
type Diff struct {
	Raw string // the raw diff output
}

type Branches []*Branch

func (p Branches) Len() int           { return len(p) }
func (p Branches) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p Branches) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// ByAuthorDate sorts by author date. Requires full commit information to be included.
type ByAuthorDate []*Branch

func (p ByAuthorDate) Len() int { return len(p) }
func (p ByAuthorDate) Less(i, j int) bool {
	return p[i].Commit.Author.Date.Before(p[j].Commit.Author.Date)
}
func (p ByAuthorDate) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type Tags []*Tag

func (p Tags) Len() int           { return len(p) }
func (p Tags) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p Tags) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

const (
	// FixedQuery is a value for SearchOptions.QueryType that
	// indicates the query is a fixed string, not a regex.
	FixedQuery = "fixed"

	// TODO(sqs): allow regexp searches, extended regexp searches, etc.
)

// TextSearchOptions contains common options for text search commands.
type TextSearchOptions struct {
	Pattern         string // the pattern to look for
	IsRegExp        bool   // whether the pattern is a regexp (if false, treated as exact string)
	IsCaseSensitive bool   // whether the pattern should be matched case-sensitively
}

// PathOptions contains common options for commands that can be limited
// to only certain paths.
type PathOptions struct {
	IncludePatterns []string // include paths matching all of these patterns
	ExcludePattern  string   // exclude paths matching any of these patterns
	IsRegExp        bool     // whether the pattern is a regexp (if false, treated as exact string)
	IsCaseSensitive bool     // whether the pattern should be matched case-sensitively
}

// CompilePathMatcher compiles the path options into a PathMatcher.
func CompilePathMatcher(options PathOptions) (pathmatch.PathMatcher, error) {
	return pathmatch.CompilePathPatterns(
		options.IncludePatterns, options.ExcludePattern,
		pathmatch.CompileOptions{CaseSensitive: options.IsCaseSensitive, RegExp: options.IsRegExp},
	)
}

// RawLogDiffSearchOptions specifies options to (Repository).RawLogDiffSearch.
type RawLogDiffSearchOptions struct {
	// Query specifies the search query to find.
	Query TextSearchOptions

	// MatchChangedOccurrenceCount makes the operation run `git log -S` not `git log -G`.
	// See `git log --help` for more information.
	MatchChangedOccurrenceCount bool

	// Diff is whether the diff should be computed and returned.
	Diff bool

	// OnlyMatchingHunks makes the diff only include hunks that match the query. If false,
	// all hunks from files that match the query are included.
	OnlyMatchingHunks bool

	// Paths specifies the paths to include/exclude.
	Paths PathOptions

	// FormatArgs is a list of format args that are passed to the `git log` command.
	// Because the output is parsed, it is expected to be in a known format. If the
	// FormatArgs does not match one of the server's expected values, the operation
	// will fail.
	//
	// If nil, the default format args are used.
	FormatArgs []string

	// RawArgs is a list of non-format args that are passed to the `git log` command.
	// It should not contain any "--" elements; those should be passed using the Paths
	// field.
	//
	// No arguments that affect the format of the output should be present in this
	// slice.
	Args []string
}

// LogCommitSearchResult describes a matching diff from (Repository).RawLogDiffSearch.
type LogCommitSearchResult struct {
	Commit         Commit      // the commit whose diff was matched
	Diff           *Diff       // the diff, with non-matching/irrelevant portions deleted (respecting diff syntax)
	DiffHighlights []Highlight // highlighted query matches in the diff

	// Refs is the list of ref names of this commit (from `git log --decorate`).
	Refs []string

	// SourceRefs is the list of ref names by which this commit was reached. (See
	// `git log --help` documentation on the `--source` flag.)
	SourceRefs []string

	// Incomplete indicates that this result may represent a subset of the actual data.
	// This can occur when the underlying command returns early due to an impending
	// timeout.
	Incomplete bool
}

// Highlight represents a highlighted region in a string.
type Highlight struct {
	Line      int // the 1-indexed line number
	Character int // the 1-indexed character on the line
	Length    int // the length of the highlight, in characters (on the same line)
}
