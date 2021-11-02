package api

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type featuredExtensionsResolver struct {
	// cache results because they are used by multiple fields
	once sync.Once

	featuredExtensions []graphqlbackend.RegistryExtension
	err                error
	db                 dbutil.DB
}

func (r *extensionRegistryResolver) FeaturedExtensions(ctx context.Context) (graphqlbackend.FeaturedExtensionsConnection, error) {
	return &featuredExtensionsResolver{db: r.db}, nil
}

func (r *featuredExtensionsResolver) compute(ctx context.Context) ([]graphqlbackend.RegistryExtension, error) {
	r.once.Do(func() {
		r.featuredExtensions, r.err = GetFeaturedExtensions(ctx, r.db)
	})
	return r.featuredExtensions, r.err
}

func (r *featuredExtensionsResolver) Nodes(ctx context.Context) ([]graphqlbackend.RegistryExtension, error) {
	// See (*featuredExtensionsResolver).Error for why we ignore the error.
	xs, _ := r.compute(ctx)
	return xs, nil
}

func (r *featuredExtensionsResolver) Error(ctx context.Context) *string {
	// See the GraphQL API schema documentation for this field for an explanation of why we return
	// errors in this way.
	_, err := r.compute(ctx)
	if err != nil {
		return strptr(err.Error())
	}
	return nil
}
