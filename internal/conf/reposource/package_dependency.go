package reposource

import "github.com/sourcegraph/sourcegraph/internal/api"

// [NOTE: Dependency-terminology]
// In a dependency graph of packages, such as when doing package resolution,
// you have the notion of dependencies as a pair of a package + a version range
// (potentially having only one version in the limiting case) + potentially
// some other information. After resolution, dependencies are pinned to a
// specific version.
//
// From the point of view of package repository support in Sourcegraph, the
// notion of a dependency corresponds to the latter, since package resolution
// is handled by the appropriate package manager, and we only deal with the frozen
// dependency versions when generating LSIF.
//
// This is why a dependency can be represented as a package + version pair.
//
// One edge case here is the root of the graph, which is technically not a
// dependency. However, we still use the same type to represent the root for
// practical purposes. For naming values, prefer "VersionedPackage" for
// situations where there is no connotation of a dependency edge.
type PackageDependency interface {
	// The scheme of the dependency (semanticdb, npm)
	Scheme() string

	// Returns the name of the dependency as used by the package manager,
	// excluding version information.
	PackageSyntax() string
	// Returns the name of the dependency as used by the package manager,
	// including version information.
	PackageManagerSyntax() string
	// The version of the package.
	PackageVersion() string

	// RepoName provides a name that is "globally unique" for a Sourcegraph instance.
	// The returned value is used for repo:... in queries.
	RepoName() api.RepoName

	// Description provides a human readable description of the package's purpose.
	// May be empty.
	Description() string

	// Returns the git tag associated with the given dependency version, used
	// rev: or repo:foo@rev
	GitTagFromVersion() string

	// Less implements a comparison method with another PackageDependency for sorting.
	Less(PackageDependency) bool
}

var (
	_ PackageDependency = (*MavenDependency)(nil)
	_ PackageDependency = (*NpmDependency)(nil)
	_ PackageDependency = (*GoDependency)(nil)
	_ PackageDependency = (*PythonDependency)(nil)
)
