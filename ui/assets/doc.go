// Package assets contains static assets for the front-end Web app.
//
// It exports a Provider global variable, that should be used by all code
// seeking to provide access to assets, regardless of their type (dev, oss
// or enterprise).
//
// To select a particular bundle variant, use _one_ of the following imports in
// the main.go:
//
//   - If you want the oss bundle:
//     import _ "github.com/sourcegraph/sourcegraph/ui/assets/oss" // Select oss assets
//   - If you want the enterprise bundle:
//     import _ "github.com/sourcegraph/sourcegraph/ui/assets/enterprise" // Select enterprise assets
//
// And to support working with dev assets, with the webpack process handling them for you, you can use:
//
//	 func main() {
//		if os.Getenv("WEBPACK_DEV_SERVER") == "1" {
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
