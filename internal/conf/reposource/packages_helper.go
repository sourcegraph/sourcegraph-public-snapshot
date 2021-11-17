package reposource

type PackageDependency interface {
	// Give the name of the dependency as used by the package manager, including version information.
	PackageSyntax() string
}

var _ PackageDependency = MavenDependency{}
