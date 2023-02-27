package reposource

import "github.com/sourcegraph/sourcegraph/internal/api"

type PackageName string

// Package encodes the abstract notion of a publishable artifact from different languages ecosystems.
// For example, Package refers to:
// - an npm package in the JS/TS ecosystem.
// - a go module in the Go ecosystem.
// - a PyPi package in the Python ecosystem.
// - a Maven artifact (groupID + artifactID) for Java/JVM ecosystem.
// Notably, Package does not include a version.
// See VersionedPackage for a Package that includes a version.
type Package interface {
	// Scheme is the LSIF moniker scheme that's used by the primary LSIF indexer for
	// this ecosystem. For example, "semanticdb" for scip-java and "npm" for scip-typescript.
	Scheme() string

	// PackageSyntax is the string-formatted encoding of this Package, as accepted by the ecosystem's package manager.
	// Notably, the version is not included.
	PackageSyntax() PackageName

	// RepoName provides a name that is "globally unique" for a Sourcegraph instance.
	// The returned value is used for repo:... in queries.
	RepoName() api.RepoName

	// Description provides a human-readable description of the package's purpose.
	// May be empty.
	Description() string
}
