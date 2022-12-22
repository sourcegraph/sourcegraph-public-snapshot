package api

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
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

func (r *extensionRegistryResolver) FilterRemoteExtensions(ids []string) []string {
	extensions := make([]*registry.Extension, len(ids))
	for i, id := range ids {
		extensions[i] = &registry.Extension{ExtensionID: id}
	}
	keepExtensions := FilterRemoteExtensions(extensions)
	keep := make([]string, len(keepExtensions))
	for i, extension := range keepExtensions {
		keep[i] = extension.ExtensionID
	}
	return keep
}
