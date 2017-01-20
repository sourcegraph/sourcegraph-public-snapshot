package sourcegraph

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

	// PkgQuery is the JSON containment query. It matches all packages
	// that have the same values for keys defined in PkgQuery.
	PkgQuery map[string]interface{}

	// Limit is the maximum size of the returned package list.
	Limit int
}
