// Package types defines types used by the frontend.
package types

import (
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// Repo represents a source code repository.
type Repo struct {
	// ID is the unique numeric ID for this repository.
	ID api.RepoID

	// ExternalRepo identifies this repository by its ID on the external service where it resides (and the external
	// service itself).
	ExternalRepo *api.ExternalRepoSpec

	// URI is a normalized identifier for this repository based on its primary clone
	// URL. E.g., "github.com/user/repo".
	URI api.RepoURI
	// Description is a brief description of the repository.
	Description string
	// Language is the primary programming language used in this repository.
	Language string
	// Enabled is whether the repository is enabled. Disabled repositories are
	// not accessible by users (except site admins).
	Enabled bool
	// Fork is whether this repository is a fork of another repository.
	Fork bool
	// CreatedAt is when this repository was created on Sourcegraph.
	CreatedAt time.Time
	// UpdatedAt is when this repository's metadata was last updated on Sourcegraph.
	UpdatedAt *time.Time
	// IndexedRevision is the revision that the global index is currently based on. It is only used by the indexer
	// to determine if reindexing is necessary. Setting it to nil/null will cause the indexer to reindex the next
	// time it gets triggered for this repository.
	IndexedRevision *api.CommitID
	// FreezeIndexedRevision, when true, tells the indexer not to update the indexed revision if it is already set.
	// This is a kludge that lets us freeze the indexed repository revision for specific deployments
	FreezeIndexedRevision bool
}

// DependencyReferencesOptions specifies options for querying dependency references.
type DependencyReferencesOptions struct {
	Language   string // e.g. "go"
	api.RepoID        // repository whose file:line:character describe the symbol of interest
	api.CommitID
	File            string
	Line, Character int

	// Limit specifies the number of dependency references to return.
	Limit int // e.g. 20
}

type SiteConfig struct {
	SiteID      string
	Email       string
	UpdatedAt   string
	Initialized bool // whether the initial site admin account has been created
}

// User represents a registered user.
type User struct {
	ID               int32
	ExternalID       *string
	Username         string
	ExternalProvider string
	DisplayName      string
	AvatarURL        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	SiteAdmin        bool
}

// OrgRepo represents a repo that exists on a native client's filesystem, but
// does not necessarily have its contents cloned to a remote Sourcegraph server.
type OrgRepo struct {
	ID                api.RepoID
	CanonicalRemoteID api.RepoURI
	CloneURL          string
	OrgID             int32
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ThreadLines struct {
	// HTMLBefore is unsanitized HTML lines before the user selection.
	HTMLBefore string

	// HTML is unsanitized HTML lines of the user selection.
	HTML string

	// HTMLAfter is unsanitized HTML lines after the user selection.
	HTMLAfter                string
	TextBefore               string
	Text                     string
	TextAfter                string
	TextSelectionRangeStart  int32
	TextSelectionRangeLength int32
}

type Thread struct {
	ID                int32
	OrgRepoID         api.RepoID
	RepoRevisionPath  string
	LinesRevisionPath string
	RepoRevision      string
	LinesRevision     string
	Branch            *string
	StartLine         int32
	EndLine           int32
	StartCharacter    int32
	EndCharacter      int32
	RangeLength       int32
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ArchivedAt        *time.Time
	AuthorUserID      int32
	Lines             *ThreadLines
}

type Comment struct {
	ID           int32
	ThreadID     int32
	Contents     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	AuthorUserID int32
}

// SharedItem represents a shared thread or comment. Note that a code snippet
// is also just a thread.
type SharedItem struct {
	ULID         string
	Public       bool
	AuthorUserID int32
	ThreadID     *int32
	CommentID    *int32 // optional
}

type Org struct {
	ID          int32
	Name        string
	DisplayName *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrgMembership struct {
	ID        int32
	OrgID     int32
	UserID    int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserTag struct {
	ID     int32
	UserID int32
	Name   string
}

type OrgTag struct {
	ID    int32
	OrgID int32
	Name  string
}

type PhabricatorRepo struct {
	ID       int32
	URI      api.RepoURI
	URL      string
	Callsign string
}

type UserActivity struct {
	UserID           int32
	PageViews        int32
	SearchQueries    int32
	LastPageViewTime *time.Time
}
