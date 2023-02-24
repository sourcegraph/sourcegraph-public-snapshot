package api

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func init() {
	graphqlbackend.ExtensionRegistry = func(db database.DB) graphqlbackend.ExtensionRegistryResolver {
		ExtensionRegistry.db = db
		return &ExtensionRegistry
	}
}

// ExtensionRegistry is the implementation of the GraphQL type ExtensionRegistry.
//
// To supply implementations of extension registry functionality, use the fields on this value of
// extensionRegistryResolver.
var ExtensionRegistry extensionRegistryResolver

// extensionRegistryResolver implements the GraphQL type ExtensionRegistry.
type extensionRegistryResolver struct {
	db database.DB
}
