package graphqlbackend

import (
	"context"
	"sort"
	"strings"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/router"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
)

type registryExtensionConnectionArgs struct {
	ConnectionArgs
	Query *string
}

func (r *extensionRegistryResolver) Extensions(ctx context.Context, args *struct {
	registryExtensionConnectionArgs
	Publisher              *graphql.ID
	Local                  bool
	Remote                 bool
	PrioritizeExtensionIDs *[]string
}) (*registryExtensionConnectionResolver, error) {
	var opt db.RegistryExtensionsListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)

	if args.Publisher != nil {
		p, err := unmarshalRegistryPublisherID(*args.Publisher)
		if err != nil {
			return nil, err
		}
		opt.Publisher.UserID = p.userID
		opt.Publisher.OrgID = p.orgID
	}
	if args.Query != nil {
		opt.Query = *args.Query
	}

	var prioritizeExtensionIDs map[string]struct{}
	if args.PrioritizeExtensionIDs != nil {
		prioritizeExtensionIDs = make(map[string]struct{}, len(*args.PrioritizeExtensionIDs))
		for _, id := range *args.PrioritizeExtensionIDs {
			prioritizeExtensionIDs[id] = struct{}{}
			opt.PrioritizeExtensionIDs = append(opt.PrioritizeExtensionIDs, id)
		}
	}

	return &registryExtensionConnectionResolver{
		opt:                    opt,
		includeLocal:           args.Local,
		includeRemote:          args.Remote,
		prioritizeExtensionIDs: prioritizeExtensionIDs,
	}, nil
}

func (r *UserResolver) RegistryExtensions(ctx context.Context, args *struct {
	registryExtensionConnectionArgs
}) (*registryExtensionConnectionResolver, error) {
	if conf.Extensions() == nil {
		return nil, errExtensionsDisabled
	}

	opt := db.RegistryExtensionsListOptions{Publisher: db.RegistryPublisher{UserID: r.user.ID}}
	if args.Query != nil {
		opt.Query = *args.Query
	}
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &registryExtensionConnectionResolver{opt: opt, includeLocal: true}, nil
}

func (r *orgResolver) RegistryExtensions(ctx context.Context, args *struct {
	registryExtensionConnectionArgs
}) (*registryExtensionConnectionResolver, error) {
	if conf.Extensions() == nil {
		return nil, errExtensionsDisabled
	}

	opt := db.RegistryExtensionsListOptions{Publisher: db.RegistryPublisher{OrgID: r.org.ID}}
	if args.Query != nil {
		opt.Query = *args.Query
	}
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &registryExtensionConnectionResolver{opt: opt, includeLocal: true}, nil
}

// registryExtensionConnectionResolver resolves a list of registry extensions.
type registryExtensionConnectionResolver struct {
	opt db.RegistryExtensionsListOptions

	includeLocal, includeRemote bool
	prioritizeExtensionIDs      map[string]struct{}

	// cache results because they are used by multiple fields
	once               sync.Once
	registryExtensions []*registryExtensionMultiResolver
	err                error
}

// filterStripLocalExtensionIDs filters to local extension IDs and strips the
// host prefix.
func filterStripLocalExtensionIDs(extensionIDs []string) []string {
	prefix := backend.GetLocalRegistryExtensionIDPrefix()
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

func (r *registryExtensionConnectionResolver) compute(ctx context.Context) ([]*registryExtensionMultiResolver, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		// Query local registry extensions.
		var local []*db.RegistryExtension
		if r.includeLocal {
			opt3 := opt2
			opt3.PrioritizeExtensionIDs = filterStripLocalExtensionIDs(opt3.PrioritizeExtensionIDs)
			local, r.err = db.RegistryExtensions.List(ctx, opt3)
			if r.err != nil {
				return
			}
			r.err = backend.PrefixLocalExtensionID(local...)
			if r.err != nil {
				return
			}
		}

		var remote []*registry.Extension

		// BACKCOMPAT: Include synthesized extensions for known language servers.
		if r.includeLocal {
			remote = append(remote, backend.ListSynthesizedRegistryExtensions(ctx, opt2)...)
		}

		// Query remote registry extensions, if filters would match any.
		if opt2.Publisher.IsZero() && r.includeRemote {
			xs, err := backend.ListRemoteRegistryExtensions(ctx, opt2.Query)
			if err != nil {
				// Continue execution even if r.err != nil so that partial (local) results are returned
				// even when the remote registry is inaccessible.
				r.err = err
			}
			remote = append(remote, xs...)
		}

		r.registryExtensions = make([]*registryExtensionMultiResolver, len(local)+len(remote))
		for i, x := range local {
			r.registryExtensions[i] = &registryExtensionMultiResolver{local: &registryExtensionDBResolver{v: x}}
		}
		for i, x := range remote {
			r.registryExtensions[len(local)+i] = &registryExtensionMultiResolver{remote: &registryExtensionRemoteResolver{v: x}}
		}

		if r.prioritizeExtensionIDs != nil {
			// Sort prioritized extension IDs first.
			sort.SliceStable(r.registryExtensions, func(i, j int) bool {
				_, pi := r.prioritizeExtensionIDs[r.registryExtensions[i].ExtensionID()]
				_, pj := r.prioritizeExtensionIDs[r.registryExtensions[j].ExtensionID()]
				return pi && !pj
			})
		}
	})
	return r.registryExtensions, r.err
}

func (r *registryExtensionConnectionResolver) Nodes(ctx context.Context) ([]*registryExtensionMultiResolver, error) {
	// See (*registryExtensionConnectionResolver).Error for why we ignore the error.
	xs, _ := r.compute(ctx)
	if r.opt.LimitOffset != nil && len(xs) > r.opt.Limit {
		xs = xs[:r.opt.Limit]
	}
	return xs, nil
}

func (r *registryExtensionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	var total int

	if r.includeLocal {
		dbCount, err := db.RegistryExtensions.Count(ctx, r.opt)
		if err != nil {
			return 0, err
		}
		total += dbCount
	}

	// Count remote extensions. Performing an actual fetch is necessary.
	//
	// See (*registryExtensionConnectionResolver).Error for why we ignore the error.
	xs, _ := r.compute(ctx)
	for _, x := range xs {
		if x.remote != nil {
			total++
		}
	}

	return int32(total), nil
}

func (r *registryExtensionConnectionResolver) PageInfo(ctx context.Context) (*PageInfo, error) {
	// See (*registryExtensionConnectionResolver).Error for why we ignore the error.
	registryExtensions, _ := r.compute(ctx)
	return &PageInfo{hasNextPage: r.opt.LimitOffset != nil && len(registryExtensions) > r.opt.Limit}, nil
}

func (r *registryExtensionConnectionResolver) URL(ctx context.Context) (*string, error) {
	if r.opt.Publisher.IsZero() {
		return nil, nil
	}

	publisher, err := getRegistryPublisher(ctx, r.opt.Publisher)
	if err != nil {
		return nil, err
	}
	p := publisher.toDBRegistryPublisher()
	url := router.RegistryPublisherExtensions(p.UserID != 0, p.OrgID != 0, p.NonCanonicalName)
	if url == "" {
		return nil, errRegistryUnknownPublisher
	}
	return &url, nil
}

func (r *registryExtensionConnectionResolver) Error(ctx context.Context) *string {
	// See the GraphQL API schema documentation for this field for an explanation of why we return
	// errors in this way.
	//
	// TODO(sqs): When https://github.com/graph-gophers/graphql-go/pull/219 or similar is merged, we
	// can make the other fields return data *and* an error, instead of using this separate error
	// field.
	_, err := r.compute(ctx)
	if err == nil {
		return nil
	}
	return strptr(err.Error())
}
