package version

// version is configured at build time via ldflags like this:
// -ldflags "-X github.com/sourcegraph/sourcegraph/pkg/version.version=1.2.3"
var version = "dev"

// Version returns the version string configured at build time.
func Version() string {
	return version
}
