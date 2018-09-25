package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func registryExtensionByID(ctx context.Context, id graphql.ID) (*registryExtensionMultiResolver, error) {
	registryExtensionID, err := unmarshalRegistryExtensionID(id)
	if err != nil {
		return nil, err
	}
	switch {
	case registryExtensionID.LocalID != 0:
		x, err := registryExtensionByIDInt32(ctx, registryExtensionID.LocalID)
		if err != nil {
			return nil, err
		}
		return &registryExtensionMultiResolver{local: x}, nil
	case registryExtensionID.RemoteID != nil:
		x, err := backend.GetRemoteRegistryExtension(ctx, "uuid", registryExtensionID.RemoteID.UUID)
		if err != nil {
			return nil, err
		}
		return &registryExtensionMultiResolver{remote: &registryExtensionRemoteResolver{v: x}}, nil
	default:
		return nil, errors.New("invalid registry extension ID")
	}
}

func registryExtensionByIDInt32(ctx context.Context, id int32) (*registryExtensionDBResolver, error) {
	if conf.Extensions() == nil {
		return nil, errExtensionsDisabled
	}
	x, err := db.RegistryExtensions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := backend.PrefixLocalExtensionID(x); err != nil {
		return nil, err
	}
	return &registryExtensionDBResolver{v: x}, nil
}

// registryExtensionID identifies a registry extension, either locally or on a remote
// registry. Exactly 1 field must be set.
type registryExtensionID struct {
	LocalID  int32                      `json:"l,omitempty"`
	RemoteID *registryExtensionRemoteID `json:"r,omitempty"`
}

func marshalRegistryExtensionID(id registryExtensionID) graphql.ID {
	return relay.MarshalID("RegistryExtension", id)
}

func unmarshalRegistryExtensionID(id graphql.ID) (registryExtensionID registryExtensionID, err error) {
	err = relay.UnmarshalSpec(id, &registryExtensionID)
	return
}
