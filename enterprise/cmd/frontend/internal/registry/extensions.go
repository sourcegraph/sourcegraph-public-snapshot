package registry

import (
	"context"

	"sourcegraph.com/cmd/frontend/graphqlbackend"
	"sourcegraph.com/cmd/frontend/registry"
	"sourcegraph.com/pkg/conf"
)

func init() {
	conf.DefaultRemoteRegistry = "https://sourcegraph.com/.api/registry"
	registry.GetLocalExtensionByExtensionID = func(ctx context.Context, extensionIDWithoutPrefix string) (graphqlbackend.RegistryExtension, error) {
		x, err := dbExtensions{}.GetByExtensionID(ctx, extensionIDWithoutPrefix)
		if err != nil {
			return nil, err
		}
		if err := prefixLocalExtensionID(x); err != nil {
			return nil, err
		}
		return &extensionDBResolver{v: x}, nil
	}
}

// prefixLocalExtensionID adds the local registry's extension ID prefix (from
// GetLocalRegistryExtensionIDPrefix) to all extensions' extension IDs in the list.
func prefixLocalExtensionID(xs ...*dbExtension) error {
	prefix := registry.GetLocalRegistryExtensionIDPrefix()
	if prefix == nil {
		return nil
	}
	for _, x := range xs {
		x.NonCanonicalExtensionID = *prefix + "/" + x.NonCanonicalExtensionID
		x.NonCanonicalRegistry = *prefix
	}
	return nil
}
