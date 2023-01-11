package api

import (
	"context"
	"sort"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *extensionRegistryResolver) Extensions(ctx context.Context, args *graphqlbackend.RegistryExtensionConnectionArgs) (graphqlbackend.RegistryExtensionConnection, error) {
	return &registryExtensionConnectionResolver{
		args:                         *args,
		db:                           r.db,
		listRemoteRegistryExtensions: listRemoteRegistryExtensions,
	}, nil
}

// registryExtensionConnectionResolver resolves a list of registry extensions.
type registryExtensionConnectionResolver struct {
	args graphqlbackend.RegistryExtensionConnectionArgs

	db                           database.DB
	listRemoteRegistryExtensions func(_ context.Context, query string) ([]*registry.Extension, error)

	// cache results because they are used by multiple fields
	once               sync.Once
	registryExtensions []graphqlbackend.RegistryExtension
	err                error
}

var (
	// ListLocalRegistryExtensions lists and returns local registry extensions according to the args. If
	// there is no local extension registry, it is not implemented.
	ListLocalRegistryExtensions func(context.Context, database.DB, graphqlbackend.RegistryExtensionConnectionArgs) ([]graphqlbackend.RegistryExtension, error)
)

func (r *registryExtensionConnectionResolver) compute(ctx context.Context) ([]graphqlbackend.RegistryExtension, error) {
	r.once.Do(func() {
		args2 := r.args
		if args2.First != nil {
			tmp := *args2.First
			tmp++ // so we can detect if there is a next page
			args2.First = &tmp
		}

		// Query local registry extensions.
		var local []graphqlbackend.RegistryExtension
		if ListLocalRegistryExtensions != nil {
			local, r.err = ListLocalRegistryExtensions(ctx, r.db, args2)
			if r.err != nil {
				return
			}
		}

		// Query remote registry extensions, if filters would match any.
		var remote []*registry.Extension
		xs, err := r.listRemoteRegistryExtensions(ctx, "")
		if err != nil {
			// Continue execution even if r.err != nil so that partial (local) results are returned
			// even when the remote registry is inaccessible.
			r.err = err
		}

		if r.args.ExtensionIDs == nil {
			remote = append(remote, xs...)
		} else {
			// The ExtensionIDs arg ("only include extensions specified by these IDs") is only
			// applied at query time for local extensions, not remote extensions, so apply it
			// here.
			include := map[string]struct{}{}
			for _, id := range *r.args.ExtensionIDs {
				include[id] = struct{}{}
			}
			for _, x := range xs {
				if _, ok := include[x.ExtensionID]; ok {
					remote = append(remote, x)
				}
			}
		}

		r.registryExtensions = make([]graphqlbackend.RegistryExtension, len(local)+len(remote))
		copy(r.registryExtensions, local)
		for i, x := range remote {
			r.registryExtensions[len(local)+i] = &registryExtensionRemoteResolver{v: x}
		}

		sort.SliceStable(r.registryExtensions, func(i, j int) bool {
			return r.registryExtensions[i].ExtensionID() < r.registryExtensions[j].ExtensionID()
		})

		allowedExtensions := ExtensionRegistryListAllowedExtensions()
		if allowedExtensions != nil {
			filteredExtensions := []graphqlbackend.RegistryExtension{}
			for i := range r.registryExtensions {
				ext := r.registryExtensions[i]
				if _, ok := allowedExtensions[ext.ExtensionID()]; ok {
					filteredExtensions = append(filteredExtensions, ext)
				}
			}
			r.registryExtensions = filteredExtensions
		}
	})
	return r.registryExtensions, r.err
}

func (r *registryExtensionConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.RegistryExtension, error) {
	xs, _ := r.compute(ctx)
	if r.args.First != nil && len(xs) > int(*r.args.First) {
		xs = xs[:int(*r.args.First)]
	}
	return xs, nil
}
