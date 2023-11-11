// Package assets contains static assets for the front-end Web app.
//
// It exports a Provider global variable, that should be used by all code
// seeking to provide access to assets, regardless of their type (dev, oss
// or enterprise).

// You must also import the embedded assets:
//
//	import _ "github.com/sourcegraph/sourcegraph/client/web/dist"
//
// And to support working with dev assets, with the web builder process handling them for you, you can use:
//
//	 func main() {
//		if os.Getenv("WEB_BUILDER_DEV_SERVER") == "1" {
//			assets.UseDevAssetsProvider()
//		}
//		// ...
//	 }
//
// If this step isn't done, the default assets provider implementation, FailingAssetsProvider will ensure
// the binary panics when launched and will explicitly tell you about the problem.
//
// This enables to express which bundle type is needed at compile time, expressed through package dependency,
// which in turn enables Bazel to build the right bundle and embed it through go embeds without relying on
// external configuration or flags, keeping the analysis cache intact regardless of which bundle is being built.
package assets
