package vcs

import (
	"errors"

	"golang.org/x/tools/godoc/vfs"
)

// A Repository is a VCS repository.
type Repository interface {
	// ResolveRevision returns the revision that the given revision
	// specifier resolves to, or a non-nil error if there is no such
	// revision.
	//
	// Implementations may choose to return ErrRevisionNotFound in all
	// cases where the revision is not found, or more specific errors
	// (such as ErrCommitNotFound) if spec can be partially resolved
	// or determined to be a certain kind of revision specifier.
	ResolveRevision(spec string) (CommitID, error)

	// ResolveTag returns the tag with the given name, or
	// ErrTagNotFound if no such tag exists.
	ResolveTag(name string) (CommitID, error)

	// ResolveBranch returns the branch with the given name, or
	// ErrBranchNotFound if no such branch exists.
	ResolveBranch(name string) (CommitID, error)

	// Branches returns a list of all branches in the repository.
	Branches(BranchesOptions) ([]*Branch, error)

	// Tags returns a list of all tags in the repository.
	Tags() ([]*Tag, error)

	// GetCommit returns the commit with the given commit ID, or
	// ErrCommitNotFound if no such commit exists.
	GetCommit(CommitID) (*Commit, error)

	// Commits returns all commits matching the options, as well as
	// the total number of commits (the count of which is not subject
	// to the N/Skip options).
	//
	// Optionally, the caller can request the total not to be computed,
	// as this can be expensive for large branches.
	Commits(CommitsOptions) (commits []*Commit, total uint, err error)

	// Committers returns the per-author commit statistics of the repo.
	Committers(CommittersOptions) ([]*Committer, error)

	// FileSystem opens the repository file tree at a given commit ID.
	//
	// Implementations may choose to check that the commit exists
	// before FileSystem returns or to defer the check until
	// operations are performed on the filesystem. (For example, an
	// implementation proxying a remote filesystem may not want to
	// incur the round-trip to check that the commit exists.)
	FileSystem(at CommitID) (vfs.FileSystem, error)
}

// A Blamer is a repository that can blame portions of a file.
type Blamer interface {
	BlameFile(path string, opt *BlameOptions) ([]*Hunk, error)
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

// A Differ is a repository that can compute diffs between two
// commits.
type Differ interface {
	// Diff shows changes between two commits. If base or head do not
	// exist, an error is returned.
	Diff(base, head CommitID, opt *DiffOptions) (*Diff, error)
}

// A CrossRepoDiffer is a repository that can compute diffs with
// respect to a commit in a different repository.
type CrossRepoDiffer interface {
	// CrossRepoDiff shows changes between two commits in different
	// repositories. If base or head do not exist, an error is
	// returned.
	CrossRepoDiff(base CommitID, headRepo Repository, head CommitID, opt *DiffOptions) (*Diff, error)
}

var (
	ErrRefNotFound      = errors.New("ref not found")
	ErrBranchNotFound   = errors.New("branch not found")
	ErrCommitNotFound   = errors.New("commit not found")
	ErrRevisionNotFound = errors.New("revision not found")
	ErrTagNotFound      = errors.New("tag not found")
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
	return p[i].Commit.Author.Date.Time().Before(p[j].Commit.Author.Date.Time())
}
func (p ByAuthorDate) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type Tags []*Tag

func (p Tags) Len() int           { return len(p) }
func (p Tags) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p Tags) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// A FileLister is a repository that can perform actions related to
// listing the entire file tree.
type FileLister interface {
	// ListFiles returns list of all file names in the repo at the
	// given commit. Returned file paths are forward slash separated,
	// relative to the base directory of the repository, and sorted
	// alphabetically. E.g., returned paths have the form "path/to/file.txt".
	ListFiles(CommitID) ([]string, error)
}
