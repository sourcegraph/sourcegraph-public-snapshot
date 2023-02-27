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

func (r *registryExtensionConnectionResolver) compute(ctx context.Context) ([]graphqlbackend.RegistryExtension, error) {
	r.once.Do(func() {
		args2 := r.args
		if args2.First != nil {
			tmp := *args2.First
			tmp++ // so we can detect if there is a next page
			args2.First = &tmp
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
			// The ExtensionIDs arg ("only include extensions specified by these IDs") is not
			// applied at query time for remote extensions, so apply it here.
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

		r.registryExtensions = make([]graphqlbackend.RegistryExtension, len(remote))
		for i, x := range remote {
			r.registryExtensions[i] = &registryExtensionRemoteResolver{v: x}
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
