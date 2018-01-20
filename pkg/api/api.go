// Package api contains an API client and types for cross-service communication.
package api

import (
	"fmt"
	"time"

	"github.com/sourcegraph/go-langserver/pkg/lspext"
)

// Repo represents a source code repository.
type Repo struct {
	// ID is the unique numeric ID for this repository.
	ID int32
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
	// Fork is whether this repository is a fork.
	Fork bool
	// StarsCount is the number of users who have starred this repository.
	// Not persisted in DB!
	StarsCount *uint
	// ForksCount is the number of forks of this repository that exist.
	// Not persisted in DB!
	ForksCount *uint
	// Private is whether this repository is private. Note: this field
	// is currently only used when the repository is hosted on GitHub.
	// All locally hosted repositories should be public. If Private is
	// true for a locally hosted repository, the repository might never
	// be returned.
	Private bool
	// CreatedAt is when this repository was created. If it represents an externally
	// hosted (e.g., GitHub) repository, the creation date is when it was created at
	// that origin.
	CreatedAt *time.Time
	// UpdatedAt is when this repository's metadata was last updated (on its origin if
	// it's an externally hosted repository).
	UpdatedAt *time.Time
	// PushedAt is when this repository's was last (VCS-)pushed to.
	PushedAt *time.Time
	// IndexedRevision is the revision that the global index is currently based on. It is only used
	// by the indexer to determine if reindexing is necessary. Setting it to nil/null will cause
	// the indexer to reindex the next time it gets triggered for this repository.
	IndexedRevision *string
	// FreezeIndexedRevision, when true, tells the indexer not to
	// update the indexed revision if it is already set. This is a
	// kludge that lets us freeze the indexed repository revision for
	// specific deployments
	FreezeIndexedRevision bool
}

type DependencyReferences struct {
	References []*DependencyReference
	Location   lspext.SymbolLocationInformation
}

// DependencyReference effectively says that RepoID has made a reference to a
// dependency.
type DependencyReference struct {
	DepData map[string]interface{} // includes additional information about the dependency, e.g. whether or not it is vendored for Go
	RepoID  int32                  // the repository who made the reference to the dependency.
	Hints   map[string]interface{} // hints which should be passed to workspace/xreferences in order to more quickly find the definition.
}

func (d *DependencyReference) String() string {
	return fmt.Sprintf("DependencyReference{DepData: %v, RepoID: %v, Hints: %v}", d.DepData, d.RepoID, d.Hints)
}

// PackageInfo is the metadata of a build-system- or
// package-manager-level package that is defined by the repository
// identified by the value of the RepoID field.
type PackageInfo struct {
	// RepoID is the id of the repository that defines the package
	RepoID int32

	// Lang is the programming language of the package
	Lang string

	// Pkg is the package metadata
	Pkg map[string]interface{}
}

// ListPackagesOp specifies a Pkgs.ListPackages operation
type ListPackagesOp struct {
	// Lang, if non-empty, is the language to which to restrict the list operation.
	Lang string

	// RepoID, if non-zero, is the repository to which the set of
	// returned packages should be restricted.
	RepoID int32

	// PkgQuery is the JSON containment query. It matches all packages
	// that have the same values for keys defined in PkgQuery.
	PkgQuery map[string]interface{}

	// Limit is the maximum size of the returned package list.
	Limit int
}
