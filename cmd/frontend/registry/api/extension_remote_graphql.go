package api

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
)

// registryExtensionRemoteResolver implements the GraphQL type RegistryExtension with data from a
// remote registry.
type registryExtensionRemoteResolver struct {
	v *registry.Extension
}

var _ graphqlbackend.RegistryExtension = &registryExtensionRemoteResolver{}

func (r *registryExtensionRemoteResolver) ID() graphql.ID {
	return MarshalRegistryExtensionID(RegistryExtensionID{
		RemoteID: &registryExtensionRemoteID{Registry: r.v.RegistryURL, UUID: r.v.UUID},
	})
}

// registryExtensionRemoteID identifies a registry extension on a remote registry. It is encoded in
// RegistryExtensionID.
type registryExtensionRemoteID struct {
	Registry string `json:"r"`
	UUID     string `json:"u"`
}

func (r *registryExtensionRemoteResolver) ExtensionID() string { return r.v.ExtensionID }

func (r *registryExtensionRemoteResolver) Manifest(context.Context) (graphqlbackend.ExtensionManifest, error) {
	return NewExtensionManifest(r.v.Manifest), nil
}
