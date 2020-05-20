package version

const devVersion = "0.0.0+dev" // version string for unreleased development builds

// version is configured at build time via ldflags like this:
// -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=1.2.3"
var version = devVersion

// Version returns the version string configured at build time.
func Version() string {
	return version
}

// IsDev reports whether the version string is an unreleased development build.
func IsDev(version string) bool {
	return version == devVersion
}

// Mock is used by tests to mock the result of Version and IsDev.
func Mock(mockVersion string) {
	version = mockVersion
}
