package gitdomain

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
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

// A PersonCount is a contributor to a repository.
type PersonCount struct {
	Name  string
	Email string
	Count int32
}

func (p *PersonCount) String() string {
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
