package graphqlbackend

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func (r *schemaResolver) ExtensionRegistry() (*extensionRegistryResolver, error) {
	if conf.Platform() == nil {
		return nil, errors.New("platform disabled")
	}
	return &extensionRegistryResolver{}, nil
}

type extensionRegistryResolver struct{}

func (*extensionRegistryResolver) Extension(ctx context.Context, args *struct {
	ExtensionID string
}) (*registryExtensionMultiResolver, error) {
	return getExtensionByExtensionID(ctx, args.ExtensionID)
}

func getExtensionByExtensionID(ctx context.Context, extensionID string) (*registryExtensionMultiResolver, error) {
	local, remote, err := backend.GetExtensionByExtensionID(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	if local != nil {
		return &registryExtensionMultiResolver{local: &registryExtensionDBResolver{local}}, nil
	}
	return &registryExtensionMultiResolver{remote: &registryExtensionRemoteResolver{remote}}, nil
}

func (*extensionRegistryResolver) LocalExtensionIDPrefix() (*string, error) {
	return backend.GetLocalRegistryExtensionIDPrefix()
}
