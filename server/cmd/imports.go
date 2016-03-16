// Package cmd imports sub-packages of server that are imported for
// side effects but that would cause import cycles if imported in
// package server.
//
// Any executable built containing the server should import this
// package as well.
package cmd

import (
	// Import this package for the side effect of registering cli flags.
	_ "sourcegraph.com/sourcegraph/sourcegraph/server/local/cli"

	// Import these packages for their side effects of registering
	// stores.
	_ "sourcegraph.com/sourcegraph/sourcegraph/server/internal/store/fs"
	_ "sourcegraph.com/sourcegraph/sourcegraph/server/internal/store/pgsql"
)
