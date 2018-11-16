package registry

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/registry"
)

func init() {
	registry.ListLocalRegistryExtensions = listLocalRegistryExtensions
	registry.CountLocalRegistryExtensions = countLocalRegistryExtensions
}

func listLocalRegistryExtensions(ctx context.Context, args graphqlbackend.RegistryExtensionConnectionArgs) ([]graphqlbackend.RegistryExtension, error) {
	if args.PrioritizeExtensionIDs != nil {
		ids := filterStripLocalExtensionIDs(*args.PrioritizeExtensionIDs)
		args.PrioritizeExtensionIDs = &ids
	}
	opt, err := toDBExtensionsListOptions(args)
	if err != nil {
		return nil, err
	}
	xs, err := dbExtensions{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	if err := prefixLocalExtensionID(xs...); err != nil {
		return nil, err
	}
	xs2 := make([]graphqlbackend.RegistryExtension, len(xs))
	for i, x := range xs {
		xs2[i] = &extensionDBResolver{v: x}
	}
	return xs2, nil
}

func countLocalRegistryExtensions(ctx context.Context, args graphqlbackend.RegistryExtensionConnectionArgs) (int, error) {
	opt, err := toDBExtensionsListOptions(args)
	if err != nil {
		return 0, err
	}
	return dbExtensions{}.Count(ctx, opt)
}

func toDBExtensionsListOptions(args graphqlbackend.RegistryExtensionConnectionArgs) (dbExtensionsListOptions, error) {
	var opt dbExtensionsListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)
	if args.Publisher != nil {
		p, err := unmarshalRegistryPublisherID(*args.Publisher)
		if err != nil {
			return opt, err
		}
		opt.Publisher.UserID = p.userID
		opt.Publisher.OrgID = p.orgID
	}
	if args.Query != nil {
		opt.Query = *args.Query
	}
	if args.PrioritizeExtensionIDs != nil {
		opt.PrioritizeExtensionIDs = *args.PrioritizeExtensionIDs
	}
	opt.ExcludeWIP = !args.IncludeWIP
	return opt, nil
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
