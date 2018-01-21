// Package types defines types used by the frontend.
package types

import (
	"fmt"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// Repo represents a source code repository.
type Repo struct {
	// ID is the unique numeric ID for this repository.
	ID api.RepoID
	// URI is a normalized identifier for this repository based on its primary clone
	// URL. E.g., "github.com/user/repo".
	URI string
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
	IndexedRevision *string
	// FreezeIndexedRevision, when true, tells the indexer not to update the indexed revision if it is already set.
	// This is a kludge that lets us freeze the indexed repository revision for specific deployments
	FreezeIndexedRevision bool
}

// RepoRevSpec specifies a repository at a specific commit.
type RepoRevSpec struct {
	// Repo is the ID of the repository.
	Repo api.RepoID

	// CommitID is the 40-character SHA-1 of the Git commit ID.
	//
	// Revision specifiers are not allowed here. To resolve a revision
	// specifier (such as a branch name or "master~7"), call
	// Repos.GetCommit.
	CommitID string
}

// DependencyReferencesOptions specifies options for querying dependency references.
type DependencyReferencesOptions struct {
	Language        string // e.g. "go"
	api.RepoID             // repository whose file:line:character describe the symbol of interest
	CommitID        string
	File            string
	Line, Character int

	// Limit specifies the number of dependency references to return.
	Limit int // e.g. 20
}

// A ConfigurationSubject is something that can have settings. A subject with no
// fields set represents the global site settings subject.
type ConfigurationSubject struct {
	Site *string // the site's ID
	Org  *int32  // the org's ID
	User *int32  // the user's ID
}

func (s ConfigurationSubject) String() string {
	switch {
	case s.Site != nil:
		return fmt.Sprintf("site %q", *s.Site)
	case s.Org != nil:
		return fmt.Sprintf("org %d", *s.Org)
	case s.User != nil:
		return fmt.Sprintf("user %d", *s.User)
	default:
		return "unknown configuration subject"
	}
}

// Settings contains configuration settings for a subject.
type Settings struct {
	ID           int32
	Subject      ConfigurationSubject
	AuthorUserID int32
	Contents     string
	CreatedAt    time.Time
}

type SiteConfig struct {
	SiteID           string
	Email            string
	TelemetryEnabled bool
	UpdatedAt        string
}

// User represents a registered user.
type User struct {
	ID               int32
	ExternalID       *string
	Username         string
	ExternalProvider string
	DisplayName      string
	AvatarURL        *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	SiteAdmin        bool
}

// OrgRepo represents a repo that exists on a native client's filesystem, but
// does not necessarily have its contents cloned to a remote Sourcegraph server.
type OrgRepo struct {
	ID                api.RepoID
	CanonicalRemoteID string
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
	ID              int32
	Name            string
	DisplayName     *string
	SlackWebhookURL *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrgMember struct {
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
	URI      string
	URL      string
	Callsign string
}

type UserActivity struct {
	UserID        int32
	PageViews     int32
	SearchQueries int32
}
