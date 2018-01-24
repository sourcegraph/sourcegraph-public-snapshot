// Package api contains an API client and types for cross-service communication.
package api

import (
	"fmt"

	"github.com/sourcegraph/go-langserver/pkg/lspext"
)

// RepoID is the unique identifier for a repository.
type RepoID int32

// RepoURI is the name of a repository, consisting of one or more "/"-separated path components. It is a misnomer;
// it's not a valid URI because it conventionally does not have a scheme.
type RepoURI string

// CommitID is the 40-character SHA-1 hash for a Git commit.
type CommitID string

// Repo represents a source code repository.
type Repo struct {
	// ID is the unique numeric ID for this repository.
	ID RepoID
	// URI is a normalized identifier for this repository based on its primary clone
	// URL. E.g., "github.com/user/repo".
	URI RepoURI
	// IndexedRevision is the revision that the global index is currently based on. It is only used
	// by the indexer to determine if reindexing is necessary. Setting it to nil/null will cause
	// the indexer to reindex the next time it gets triggered for this repository.
	IndexedRevision *CommitID
	// FreezeIndexedRevision, when true, tells the indexer not to
	// update the indexed revision if it is already set. This is a
	// kludge that lets us freeze the indexed repository revision for
	// specific deployments
	FreezeIndexedRevision bool
}

type InsertRepoOp struct {
	URI         RepoURI
	Description string
	Fork        bool
	Enabled     bool
}

func (Repo) Fork() bool {
	// TODO(sqs): update callers
	return false
}

type DependencyReferences struct {
	References []*DependencyReference
	Location   lspext.SymbolLocationInformation
}

// DependencyReference effectively says that RepoID has made a reference to a
// dependency.
type DependencyReference struct {
	DepData map[string]interface{} // includes additional information about the dependency, e.g. whether or not it is vendored for Go
	RepoID                         // the repository who made the reference to the dependency.
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
	RepoID

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
	RepoID

	// PkgQuery is the JSON containment query. It matches all packages
	// that have the same values for keys defined in PkgQuery.
	PkgQuery map[string]interface{}

	// Limit is the maximum size of the returned package list.
	Limit int
}
