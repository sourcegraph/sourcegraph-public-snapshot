package reposource

// VersionedPackage is a Package that additionally includes a concrete version.
// The version must be a concrete version, it cannot be a version range.
type VersionedPackage interface {
	Package

	// PackageVersion returns the version of the package.
	PackageVersion() string

	// GitTagFromVersion returns the git tag associated with the given dependency version, used rev: or repo:foo@rev
	GitTagFromVersion() string

	// VersionedPackageSyntax is the string-formatted encoding of this VersionedPackage (including the version),
	// as accepted by the ecosystem's package manager.
	VersionedPackageSyntax() string

	// Less implements a comparison method with another VersionedPackage for sorting.
	Less(VersionedPackage) bool
}

var (
	_ VersionedPackage = (*MavenVersionedPackage)(nil)
	_ VersionedPackage = (*NpmVersionedPackage)(nil)
	_ VersionedPackage = (*GoVersionedPackage)(nil)
	_ VersionedPackage = (*PythonVersionedPackage)(nil)
	_ VersionedPackage = (*RustVersionedPackage)(nil)
)
