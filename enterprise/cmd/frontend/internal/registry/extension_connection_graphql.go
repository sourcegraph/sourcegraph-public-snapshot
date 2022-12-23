package registry

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry/stores"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func init() {
	registry.ListLocalRegistryExtensions = listLocalRegistryExtensions
}

func listLocalRegistryExtensions(ctx context.Context, db database.DB, args graphqlbackend.RegistryExtensionConnectionArgs) ([]graphqlbackend.RegistryExtension, error) {
	if args.ExtensionIDs != nil {
		extids := filterStripLocalExtensionIDs(*args.ExtensionIDs)
		args.ExtensionIDs = &extids
	}

	vs, err := stores.Extensions(db).List(ctx, toDBExtensionsListOptions(args))
	if err != nil {
		return nil, err
	}
	prefixLocalExtensionID(vs...)

	releasesByExtensionID, err := getLatestForBatch(ctx, db, vs)
	if err != nil {
		return nil, err
	}
	var ys []graphqlbackend.RegistryExtension
	for _, v := range vs {
		ys = append(ys, &extensionDBResolver{db: db, v: v, r: releasesByExtensionID[v.ID]})
	}
	return ys, nil
}

func toDBExtensionsListOptions(args graphqlbackend.RegistryExtensionConnectionArgs) stores.ExtensionsListOptions {
	var opt stores.ExtensionsListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)
	if args.ExtensionIDs != nil {
		opt.ExtensionIDs = *args.ExtensionIDs
	}
	return opt
}

// filterStripLocalExtensionIDs filters to local extension IDs and strips the
// host prefix.
func filterStripLocalExtensionIDs(extensionIDs []string) []string {
	prefix := registry.GetLocalRegistryExtensionIDPrefix()
	local := []string{}
	for _, id := range extensionIDs {
		parts := strings.SplitN(id, "/", 3)
		if prefix != nil && len(parts) == 3 && parts[0] == *prefix {
			local = append(local, parts[1]+"/"+parts[2])
		} else if (prefix == nil || *prefix == "") && len(parts) == 2 {
			local = append(local, id)
		}
	}
	return local
}
