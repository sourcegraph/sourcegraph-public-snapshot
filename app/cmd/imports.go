// Package cmd imports sub-packages of app that are imported for side
// effects but that would cause import cycles if imported in package
// app.
//
// Any executable built containing the app should import this package
// as well.
package cmd

import (
	// Import these packages for their side effects of registering
	// route handlers.
	_ "src.sourcegraph.com/sourcegraph/app/internal/localauth"
	_ "src.sourcegraph.com/sourcegraph/app/internal/oauth2client"
	_ "src.sourcegraph.com/sourcegraph/app/internal/oauth2server"
	_ "src.sourcegraph.com/sourcegraph/app/internal/static"
)
