// Package assets contains static assets for the front-end Web app.
//
// It exports a Provider global variable, that should be used by all code
// seeking to provide access to assets, regardless of their type (dev, oss
// or enterprise).

// Before using assets you have to call the package Init function, which sets up the global Provider
// with the path to the assets, which by default is /assets-dist
//
// import "github.com/sourcegraph/sourcegraph/ui/assets"
//
//	func main() {
//		assets.Init()
//	}
//
// And to support working with dev assets, with the web builder process handling them for you, you can use:
//
//	 func main() {
//		assets.Init()
//		if os.Getenv("WEB_BUILDER_DEV_SERVER") == "1" {
//			assets.UseDevAssetsProvider()
//		}
//		// ...
//	 }
//
// If `assets.Init()` isn't called, the default assets provider implementation, FailingAssetsProvider will ensure
// the binary panics when launched and will explicitly tell you about the problem.
//
// This enables to express which bundle type is needed at compile time, expressed through package dependency,
// which in turn enables Bazel to build the right bundle and embed it through go embeds without relying on
// external configuration or flags, keeping the analysis cache intact regardless of which bundle is being built.
package assets
