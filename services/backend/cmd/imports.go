// Package cmd imports sub-packages of server that are imported for
// side effects but that would cause import cycles if imported in
// package server.
//
// Any executable built containing the server should import this
// package as well.
package cmd

import (
	// Import this packages for the side effects of registering stores.
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)
