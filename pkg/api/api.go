// Package api contains an API client and types for cross-service communication.
package api

import (
	"fmt"
	"time"

	"github.com/sourcegraph/go-lsp/lspext"
)

// RepoID is the unique identifier for a repository.
type RepoID int32

// RepoName is the name of a repository, consisting of one or more "/"-separated path components.
//
// Previously, this was called RepoURI.
type RepoName string

// CommitID is the 40-character SHA-1 hash for a Git commit.
type CommitID string

// Repo represents a source code repository.
type Repo struct {
	// ID is the unique numeric ID for this repository on Sourcegraph.
	ID RepoID

	// ExternalRepo identifies this repository by its ID on the external service where it resides (and the external
	// service itself).
	ExternalRepo *ExternalRepoSpec

	// Name is the name of the repository (such as "github.com/user/repo").
	Name RepoName
	// Enabled is whether the repository is enabled. Disabled repositories are
	// not accessible by users (except site admins).
	Enabled bool
	// IndexedRevision is the revision that the cross-repo code intelligence index is currently
	// based on. It is only used by the indexer to determine if reindexing is necessary. Setting it
	// to nil/null will cause the indexer to reindex the next time it gets triggered for this
	// repository.
	IndexedRevision *CommitID
	// FreezeIndexedRevision, when true, tells the indexer not to
	// update the indexed revision if it is already set. This is a
	// kludge that lets us freeze the indexed repository revision for
	// specific deployments
	FreezeIndexedRevision bool
}

func (Repo) Fork() bool {
	// TODO(sqs): update callers
	return false
}

// InsertRepoOp represents an operation to insert a repository.
type InsertRepoOp struct {
	Name         RepoName
	Description  string
	Fork         bool
	Archived     bool
	Enabled      bool
	ExternalRepo *ExternalRepoSpec
}

// ExternalRepoSpec specifies a repository on an external service (such as GitHub or GitLab).
type ExternalRepoSpec struct {
	// ID is the repository's ID on the external service. Its value is opaque except to the repo-updater.
	//
	// For GitHub, this is the GitHub GraphQL API's node ID for the repository.
	ID string

	// ServiceType is the type of external service. Its value is opaque except to the repo-updater.
	//
	// Example: "github", "gitlab", etc.
	ServiceType string

	// ServiceID is the particular instance of the external service where this repository resides. Its value is
	// opaque but typically consists of the canonical base URL to the service.
	//
	// Implementations must take care to normalize this URL. For example, if different GitHub.com repository code
	// paths used slightly different values here (such as "https://github.com/" and "https://github.com", note the
	// lack of trailing slash), then the same logical repository would be incorrectly treated as multiple distinct
	// repositories depending on the code path that provided its ServiceID value.
	//
	// Example: "https://github.com/", "https://github-enterprise.example.com/"
	ServiceID string
}

// Equal returns true if r is equal to s.
func (r *ExternalRepoSpec) Equal(s *ExternalRepoSpec) bool {
	if r == s { // handles the case when r and s are both nil
		return true
	}
	if s == nil || r == nil {
		return false
	}
	return r.ID == s.ID && r.ServiceType == s.ServiceType && r.ServiceID == s.ServiceID
}

func (r *ExternalRepoSpec) String() string {
	return fmt.Sprintf("ExternalRepoSpec{%s %s %s}", r.ServiceID, r.ServiceType, r.ID)
}

type DependencyReferences struct {
	References []*DependencyReference
	Location   lspext.SymbolLocationInformation
}

// DependencyReference effectively says that RepoID has made a reference to a
// dependency.
type DependencyReference struct {
	Language string                 // the programming language of the dependency
	DepData  map[string]interface{} // includes additional information about the dependency, e.g. whether or not it is vendored for Go
	RepoID                          // the repository who made the reference to the dependency.
	Hints    map[string]interface{} // hints which should be passed to workspace/xreferences in order to more quickly find the definition.
}

func (d *DependencyReference) String() string {
	return fmt.Sprintf("DependencyReference{DepData: %v, RepoID: %v, Hints: %v}", d.DepData, d.RepoID, d.Hints)
}

// PackageInfo is the metadata of a build-system- or
// package-manager-level package that is defined by the repository
// identified by the value of the RepoID field.
type PackageInfo struct {
	// RepoID is the id of the repository that defines the package
	RepoID

	// Lang is the programming language of the package
	Lang string

	// Pkg is the package metadata
	Pkg map[string]interface{}

	// Dependencies describes the package's dependencies.
	//
	// NOTE: This field is only set when listing packages directly from the language
	// server. It may not be set when retrieving persisted package information; in that
	// case, you need to separately query for the dependencies.
	Dependencies []lspext.DependencyReference
}

// ListPackagesOp specifies a Pkgs.ListPackages operation
type ListPackagesOp struct {
	// Lang, if non-empty, is the language to which to restrict the list operation.
	Lang string

	// RepoID, if non-zero, is the repository to which the set of
	// returned packages should be restricted.
	RepoID

	// PkgQuery is the JSON containment query. It matches all packages
	// that have the same values for keys defined in PkgQuery.
	PkgQuery map[string]interface{}

	// Limit is the maximum size of the returned package list.
	Limit int
}

// A SettingsSubject is something that can have settings. Exactly 1 field must be nonzero.
type SettingsSubject struct {
	Site bool   // whether this is for global settings
	Org  *int32 // the org's ID
	User *int32 // the user's ID
}

func (s SettingsSubject) String() string {
	switch {
	case s.Site:
		return "site"
	case s.Org != nil:
		return fmt.Sprintf("org %d", *s.Org)
	case s.User != nil:
		return fmt.Sprintf("user %d", *s.User)
	default:
		return "unknown settings subject"
	}
}

// Settings contains settings for a subject.
type Settings struct {
	ID           int32           // the unique ID of this settings value
	Subject      SettingsSubject // the subject of these settings
	AuthorUserID *int32          // the ID of the user who authored this settings value
	Contents     string          // the raw JSON (with comments and trailing commas allowed)
	CreatedAt    time.Time       // the date when this settings value was created
}
