package registry

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry/stores"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func init() {
	conf.DefaultRemoteRegistry = "https://sourcegraph.com/.api/registry"
	registry.GetLocalExtensionByExtensionID = func(ctx context.Context, db database.DB, extensionIDWithoutPrefix string) (graphqlbackend.RegistryExtension, error) {
		x, err := stores.Extensions(db).GetByExtensionID(ctx, extensionIDWithoutPrefix)
		if err != nil {
			return nil, err
		}
		prefixLocalExtensionID(x)
		return &extensionDBResolver{db: db, v: x}, nil
	}
}

// prefixLocalExtensionID adds the local registry's extension ID prefix (from
// GetLocalRegistryExtensionIDPrefix) to all extensions' extension IDs in the list.
func prefixLocalExtensionID(xs ...*stores.Extension) {
	prefix := registry.GetLocalRegistryExtensionIDPrefix()
	if prefix == nil {
		return
	}
	for _, x := range xs {
		x.NonCanonicalExtensionID = *prefix + "/" + x.NonCanonicalExtensionID
		x.NonCanonicalRegistry = *prefix
	}
}
