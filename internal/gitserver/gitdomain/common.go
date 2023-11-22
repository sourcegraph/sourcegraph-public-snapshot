package gitdomain

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gobwas/glob"

	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// OID is a Git OID (40-char hex-encoded).
type OID [20]byte

func (oid OID) String() string { return hex.EncodeToString(oid[:]) }

// ObjectType is a valid Git object type (commit, tag, tree, and blob).
type ObjectType string

// Standard Git object types.
const (
	ObjectTypeCommit ObjectType = "commit"
	ObjectTypeTag    ObjectType = "tag"
	ObjectTypeTree   ObjectType = "tree"
	ObjectTypeBlob   ObjectType = "blob"
)

// ModeSubmodule is an os.FileMode mask indicating that the file is a Git submodule.
//
// To avoid being reported as a regular file mode by (os.FileMode).IsRegular, it sets other bits
// (os.ModeDevice) beyond the Git "160000" commit mode bits. The choice of os.ModeDevice is
// arbitrary.
const ModeSubmodule = 0160000 | os.ModeDevice

// Submodule holds information about a Git submodule and is
// returned in the FileInfo's Sys field by Stat/ReadDir calls.
type Submodule struct {
	// URL is the submodule repository clone URL.
	URL string

	// Path is the path of the submodule relative to the repository root.
	Path string

	// CommitID is the pinned commit ID of the submodule (in the
	// submodule repository's commit ID space).
	CommitID api.CommitID
}

// ObjectInfo holds information about a Git object and is returned in (fs.FileInfo).Sys for blobs
// and trees from Stat/ReadDir calls.
type ObjectInfo interface {
	OID() OID
}

// GitObject represents a GitObject
type GitObject struct {
	ID   OID
	Type ObjectType
}

func (o *GitObject) ToProto() *proto.GitObject {
	var id []byte
	if o.ID != (OID{}) {
		id = o.ID[:]
	}

	var t proto.GitObject_ObjectType
	switch o.Type {
	case ObjectTypeCommit:
		t = proto.GitObject_OBJECT_TYPE_COMMIT
	case ObjectTypeTag:
		t = proto.GitObject_OBJECT_TYPE_TAG
	case ObjectTypeTree:
		t = proto.GitObject_OBJECT_TYPE_TREE
	case ObjectTypeBlob:
		t = proto.GitObject_OBJECT_TYPE_BLOB

	default:
		t = proto.GitObject_OBJECT_TYPE_UNSPECIFIED
	}

	return &proto.GitObject{
		Id:   id,
		Type: t,
	}
}

func (o *GitObject) FromProto(p *proto.GitObject) {
	id := p.GetId()
	var oid OID
	if len(id) == 20 {
		copy(oid[:], id)
	}

	var t ObjectType

	switch p.GetType() {
	case proto.GitObject_OBJECT_TYPE_COMMIT:
		t = ObjectTypeCommit
	case proto.GitObject_OBJECT_TYPE_TAG:
		t = ObjectTypeTag
	case proto.GitObject_OBJECT_TYPE_TREE:
		t = ObjectTypeTree
	case proto.GitObject_OBJECT_TYPE_BLOB:
		t = ObjectTypeBlob

	}

	*o = GitObject{
		ID:   oid,
		Type: t,
	}

}

// IsAbsoluteRevision checks if the revision is a git OID SHA string.
//
// Note: This doesn't mean the SHA exists in a repository, nor does it mean it
// isn't a ref. Git allows 40-char hexadecimal strings to be references.
func IsAbsoluteRevision(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !(('0' <= r && r <= '9') ||
			('a' <= r && r <= 'f') ||
			('A' <= r && r <= 'F')) {
			return false
		}
	}
	return true
}

func EnsureAbsoluteCommit(commitID api.CommitID) error {
	// We don't want to even be running commands on non-absolute
	// commit IDs if we can avoid it, because we can't cache the
	// expensive part of those computations.
	if !IsAbsoluteRevision(string(commitID)) {
		return errors.Errorf("non-absolute commit ID: %q", commitID)
	}
	return nil
}

// Commit represents a git commit
type Commit struct {
	ID        api.CommitID `json:"ID,omitempty"`
	Author    Signature    `json:"Author"`
	Committer *Signature   `json:"Committer,omitempty"`
	Message   Message      `json:"Message,omitempty"`
	// Parents are the commit IDs of this commit's parent commits.
	Parents []api.CommitID `json:"Parents,omitempty"`
}

// Message represents a git commit message
type Message string

// Subject returns the first line of the commit message
func (m Message) Subject() string {
	message := string(m)
	i := strings.Index(message, "\n")
	if i == -1 {
		return strings.TrimSpace(message)
	}
	return strings.TrimSpace(message[:i])
}

// Body returns the contents of the Git commit message after the subject.
func (m Message) Body() string {
	message := string(m)
	i := strings.Index(message, "\n")
	if i == -1 {
		return ""
	}
	return strings.TrimSpace(message[i:])
}

// Signature represents a commit signature
type Signature struct {
	Name  string    `json:"Name,omitempty"`
	Email string    `json:"Email,omitempty"`
	Date  time.Time `json:"Date"`
}

type RefType int

const (
	RefTypeUnknown RefType = iota
	RefTypeBranch
	RefTypeTag
)

// RefDescription describes a commit at the head of a branch or tag.
type RefDescription struct {
	Name            string
	Type            RefType
	IsDefaultBranch bool
	CreatedDate     *time.Time
}

// A ContributorCount is a contributor to a repository.
type ContributorCount struct {
	Name  string
	Email string
	Count int32
}

func (p *ContributorCount) String() string {
	return fmt.Sprintf("%d %s <%s>", p.Count, p.Name, p.Email)
}

// A Tag is a VCS tag.
type Tag struct {
	Name         string `json:"Name,omitempty"`
	api.CommitID `json:"CommitID,omitempty"`
	CreatorDate  time.Time
}

type Tags []*Tag

func (p Tags) Len() int           { return len(p) }
func (p Tags) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p Tags) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Ref describes a Git ref.
type Ref struct {
	Name     string // the full name of the ref (e.g., "refs/heads/mybranch")
	CommitID api.CommitID
}

// BehindAhead is a set of behind/ahead counts.
type BehindAhead struct {
	Behind uint32 `json:"Behind,omitempty"`
	Ahead  uint32 `json:"Ahead,omitempty"`
}

// A Branch is a git branch.
type Branch struct {
	// Name is the name of this branch.
	Name string `json:"Name,omitempty"`
	// Head is the commit ID of this branch's head commit.
	Head api.CommitID `json:"Head,omitempty"`
	// Commit optionally contains commit information for this branch's head commit.
	// It is populated if IncludeCommit option is set.
	Commit *Commit `json:"Commit,omitempty"`
	// Counts optionally contains the commit counts relative to specified branch.
	Counts *BehindAhead `json:"Counts,omitempty"`
}

// EnsureRefPrefix checks whether the ref is a full ref and contains the
// "refs/heads" prefix (i.e. "refs/heads/master") or just an abbreviated ref
// (i.e. "master") and adds the "refs/heads/" prefix if the latter is the case.
func EnsureRefPrefix(ref string) string {
	return "refs/heads/" + strings.TrimPrefix(ref, "refs/heads/")
}

// AbbreviateRef removes the "refs/heads/" prefix from a given ref. If the ref
// doesn't have the prefix, it returns it unchanged.
func AbbreviateRef(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

// Branches is a sortable slice of type Branch
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

var invalidBranch = lazyregexp.New(`\.\.|/\.|\.lock$|[\000-\037\177 ~^:?*[]+|^/|/$|//|\.$|@{|^@$|\\`)

// ValidateBranchName returns false if the given string is not a valid branch name.
// It follows the rules here: https://git-scm.com/docs/git-check-ref-format
// NOTE: It does not require a slash as mentioned in point 2.
func ValidateBranchName(branch string) bool {
	return !(invalidBranch.MatchString(branch) || strings.EqualFold(branch, "head"))
}

// RefGlob describes a glob pattern that either includes or excludes refs. Exactly 1 of the fields
// must be set.
type RefGlob struct {
	// Include is a glob pattern for including refs interpreted as in `git log --glob`. See the
	// git-log(1) manual page for details.
	Include string

	// Exclude is a glob pattern for excluding refs interpreted as in `git log --exclude`. See the
	// git-log(1) manual page for details.
	Exclude string
}

// RefGlobs is a compiled matcher based on RefGlob patterns. Use CompileRefGlobs to create it.
type RefGlobs []compiledRefGlobPattern

type compiledRefGlobPattern struct {
	pattern glob.Glob
	include bool // true for include, false for exclude
}

// CompileRefGlobs compiles the ordered ref glob patterns (interpreted as in `git log --glob
// ... --exclude ...`; see the git-log(1) manual page) into a matcher. If the input patterns are
// invalid, an error is returned.
func CompileRefGlobs(globs []RefGlob) (RefGlobs, error) {
	c := make(RefGlobs, len(globs))
	for i, g := range globs {
		// Validate exclude globs according to `git log --exclude`'s specs: "The patterns
		// given...must begin with refs/... If a trailing /* is intended, it must be given
		// explicitly."
		if g.Exclude != "" {
			if !strings.HasPrefix(g.Exclude, "refs/") {
				return nil, errors.Errorf(`git ref exclude glob must begin with "refs/" (got %q)`, g.Exclude)
			}
		}

		// Add implicits (according to `git log --glob`'s specs).
		if g.Include != "" {
			// `git log --glob`: "Leading refs/ is automatically prepended if missing.".
			if !strings.HasPrefix(g.Include, "refs/") {
				g.Include = "refs/" + g.Include
			}

			// `git log --glob`: "If pattern lacks ?, *, or [, /* at the end is implied." Also
			// support an important undocumented case: support exact matches. For example, the
			// pattern refs/heads/a should match the ref refs/heads/a (i.e., just appending /* to
			// the pattern would yield refs/heads/a/*, which would *not* match refs/heads/a, so we
			// need to make the /* optional).
			if !strings.ContainsAny(g.Include, "?*[") {
				var suffix string
				if strings.HasSuffix(g.Include, "/") {
					suffix = "*"
				} else {
					suffix = "/*"
				}
				g.Include += "{," + suffix + "}"
			}
		}

		var pattern string
		if g.Include != "" {
			pattern = g.Include
			c[i].include = true
		} else {
			pattern = g.Exclude
		}
		var err error
		c[i].pattern, err = glob.Compile(pattern)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

// Match reports whether the named ref matches the ref globs.
func (gs RefGlobs) Match(ref string) bool {
	match := false
	for _, g := range gs {
		if g.include == match {
			// If the glob does not change the outcome, skip it. (For example, if the ref is already
			// matched, and the next glob is another include glob.)
			continue
		}
		if g.pattern.Match(ref) {
			match = g.include
		}
	}
	return match
}

// Pathspec is a git term for a pattern that matches paths using glob-like syntax.
// https://git-scm.com/docs/gitglossary#Documentation/gitglossary.txt-aiddefpathspecapathspec
type Pathspec string

// PathspecLiteral constructs a pathspec that matches a path without interpreting "*" or "?" as special
// characters.
//
// See: https://git-scm.com/docs/gitglossary#Documentation/gitglossary.txt-literal
func PathspecLiteral(s string) Pathspec { return Pathspec(":(literal)" + s) }
