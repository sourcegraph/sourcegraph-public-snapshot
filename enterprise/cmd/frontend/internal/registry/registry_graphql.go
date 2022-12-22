package registry

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry/stores"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func init() {
	frontendregistry.RegistryExtensionByIDInt32 = registryExtensionByIDInt32
}

func registryExtensionByIDInt32(ctx context.Context, db database.DB, id int32) (graphqlbackend.RegistryExtension, error) {
	if conf.Extensions() == nil {
		return nil, graphqlbackend.ErrExtensionsDisabled
	}
	if err := frontendregistry.ExtensionRegistryReadEnabled(); err != nil {
		return nil, err
	}
	x, err := stores.Extensions(db).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	prefixLocalExtensionID(x)
	return &extensionDBResolver{db: db, v: x}, nil
}
