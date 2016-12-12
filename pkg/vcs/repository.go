package vcs

import (
	"context"
	"errors"
	"os"
)

// RepoNotExistError is an error that reports a repository doesn't exist.
type RepoNotExistError struct {
	// CloneInProgress reports whether the repository is in process of being cloned.
	CloneInProgress bool
}

func (e RepoNotExistError) Error() string {
	if e.CloneInProgress {
		return "repository does not exist (clone in progress)"
	}
	return "repository does not exist"
}

// IsRepoNotExist reports if err is a RepoNotExistError.
func IsRepoNotExist(err error) bool {
	_, ok := err.(RepoNotExistError)
	return ok
}

var ErrRepoExist = errors.New("repository already exists")

// A Repository is a VCS repository.
type Repository interface {
	String() string

	// ResolveRevision returns the revision that the given revision
	// specifier resolves to, or a non-nil error if there is no such
	// revision.
	//
	// Implementations may choose to return ErrRevisionNotFound in all
	// cases where the revision is not found, or more specific errors
	// (such as ErrCommitNotFound) if spec can be partially resolved
	// or determined to be a certain kind of revision specifier.
	ResolveRevision(ctx context.Context, spec string) (CommitID, error)

	// Branches returns a list of all branches in the repository.
	Branches(context.Context, BranchesOptions) ([]*Branch, error)

	// Tags returns a list of all tags in the repository.
	Tags(context.Context) ([]*Tag, error)

	// GetCommit returns the commit with the given commit ID, or
	// ErrCommitNotFound if no such commit exists.
	GetCommit(context.Context, CommitID) (*Commit, error)

	// Commits returns all commits matching the options, as well as
	// the total number of commits (the count of which is not subject
	// to the N/Skip options).
	//
	// Optionally, the caller can request the total not to be computed,
	// as this can be expensive for large branches.
	Commits(context.Context, CommitsOptions) (commits []*Commit, total uint, err error)

	// Committers returns the per-author commit statistics of the repo.
	Committers(context.Context, CommittersOptions) ([]*Committer, error)

	// Stat returns a FileInfo describing the named file at commit. If the file
	// is a symbolic link, the returned FileInfo describes the symbolic link.
	// Lstat makes no attempt to follow the link.
	Lstat(ctx context.Context, commit CommitID, name string) (os.FileInfo, error)

	// Stat returns a FileInfo describing the named file at commit.
	Stat(ctx context.Context, commit CommitID, name string) (os.FileInfo, error)

	// ReadFile returns the content of the named file at commit.
	ReadFile(ctx context.Context, commit CommitID, name string) ([]byte, error)

	// Readdir reads the contents of the named directory at commit.
	ReadDir(ctx context.Context, commit CommitID, name string, recurse bool) ([]os.FileInfo, error)

	BlameFile(ctx context.Context, path string, opt *BlameOptions) ([]*Hunk, error)

	// Diff shows changes between two commits. If base or head do not
	// exist, an error is returned.
	Diff(ctx context.Context, base, head CommitID, opt *DiffOptions) (*Diff, error)

	// MergeBase returns the merge base commit for the specified
	// commits.
	MergeBase(context.Context, CommitID, CommitID) (CommitID, error)

	// UpdateEverything updates all branches, tags, etc., to match the
	// default remote repository.
	UpdateEverything(context.Context, RemoteOpts) (*UpdateResult, error)
}

// BlameOptions configures a blame.
type BlameOptions struct {
	NewestCommit CommitID `json:",omitempty" url:",omitempty"`
	OldestCommit CommitID `json:",omitempty" url:",omitempty"` // or "" for the root commit

	StartLine int `json:",omitempty" url:",omitempty"` // 1-indexed start byte (or 0 for beginning of file)
	EndLine   int `json:",omitempty" url:",omitempty"` // 1-indexed end byte (or 0 for end of file)
}

// A Hunk is a contiguous portion of a file associated with a commit.
type Hunk struct {
	StartLine int // 1-indexed start line number
	EndLine   int // 1-indexed end line number
	StartByte int // 0-indexed start byte position (inclusive)
	EndByte   int // 0-indexed end byte position (exclusive)
	CommitID
	Author Signature
}

var (
	ErrRevisionNotFound = errors.New("revision not found")
)

type CommitID string

// Marshal implements proto.Marshaler.
func (c CommitID) Marshal() ([]byte, error) {
	return []byte(c), nil
}

// Unmarshal implements proto.Unmarshaler.
func (c *CommitID) Unmarshal(data []byte) error {
	*c = CommitID(data)
	return nil
}

// CommitsOptions specifies limits on the list of commits returned by
// (Repository).Commits.
type CommitsOptions struct {
	Head CommitID // include all commits reachable from this commit (required)
	Base CommitID // exlude all commits reachable from this commit (optional, like `git log Base..Head`)

	N    uint // limit the number of returned commits to this many (0 means no limit)
	Skip uint // skip this many commits at the beginning

	Path string // only commits modifying the given path are selected (optional)

	NoTotal bool // avoid counting the total number of commits
}

// CommittersOptions specifies limits on the list of committers returned by
// (Repository).Committers.
type CommittersOptions struct {
	N int // limit the number of returned committers, ordered by decreasing number of commits (0 means no limit)

	Rev string // the rev for which committer stats will be fetched ("" means use the current revision)
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
