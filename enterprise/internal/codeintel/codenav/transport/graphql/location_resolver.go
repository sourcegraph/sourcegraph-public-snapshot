package graphql

import (
	"context"

	"github.com/sourcegraph/go-lsp"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type locationResolver struct {
	resource *sharedresolvers.GitTreeEntryResolver
	lspRange *lsp.Range
}

func NewLocationResolver(resource *sharedresolvers.GitTreeEntryResolver, lspRange *lsp.Range) resolverstubs.LocationResolver {
	return &locationResolver{
		resource: resource,
		lspRange: lspRange,
	}
}

func (r *locationResolver) Resource() resolverstubs.GitTreeEntryResolver { return r.resource }

func (r *locationResolver) Range() resolverstubs.RangeResolver {
	return r.rangeInternal()
}

func (r *locationResolver) rangeInternal() *rangeResolver {
	if r.lspRange == nil {
		return nil
	}
	return &rangeResolver{*r.lspRange}
}

func (r *locationResolver) URL(ctx context.Context) (string, error) {
	url, err := r.resource.URL(ctx)
	if err != nil {
		return "", err
	}
	return r.urlPath(url), nil
}

func (r *locationResolver) CanonicalURL() string {
	url := r.resource.CanonicalURL()
	return r.urlPath(url)
}

func (r *locationResolver) urlPath(prefix string) string {
	url := prefix
	if r.lspRange != nil {
		url += "?L" + r.rangeInternal().urlFragment()
	}
	return url
}
