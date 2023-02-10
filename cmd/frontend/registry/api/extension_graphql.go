package api

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	graphqlbackend.NodeToRegistryExtension = func(node any) (graphqlbackend.RegistryExtension, bool) {
		n, ok := node.(*registryExtensionRemoteResolver)
		return n, ok
	}

	graphqlbackend.RegistryExtensionByID = registryExtensionByID
}

// RegistryExtensionID identifies a registry extension.
type RegistryExtensionID struct {
	RemoteID *registryExtensionRemoteID `json:"r,omitempty"`
}

func MarshalRegistryExtensionID(id RegistryExtensionID) graphql.ID {
	return relay.MarshalID("RegistryExtension", id)
}

func UnmarshalRegistryExtensionID(id graphql.ID) (registryExtensionID RegistryExtensionID, err error) {
	err = relay.UnmarshalSpec(id, &registryExtensionID)
	return
}

func registryExtensionByID(ctx context.Context, db database.DB, id graphql.ID) (graphqlbackend.RegistryExtension, error) {
	registryExtensionID, err := UnmarshalRegistryExtensionID(id)
	if err != nil {
		return nil, err
	}
	switch {
	case registryExtensionID.RemoteID != nil:
		x, err := getRemoteRegistryExtension(ctx, "uuid", registryExtensionID.RemoteID.UUID)
		if err != nil {
			return nil, err
		}
		return &registryExtensionRemoteResolver{v: x}, nil
	default:
		return nil, errors.New("invalid registry extension ID")
	}
}
