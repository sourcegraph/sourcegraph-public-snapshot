package graphql

import (
	"context"

	"github.com/sourcegraph/go-lsp"

	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
)

type LocationResolver interface {
	Resource() *sharedresolvers.GitTreeEntryResolver
	Range() *rangeResolver
	URL(ctx context.Context) (string, error)
	CanonicalURL() string
}

type locationResolver struct {
	resource *sharedresolvers.GitTreeEntryResolver
	lspRange *lsp.Range
}

var _ LocationResolver = &locationResolver{}

func NewLocationResolver(resource *sharedresolvers.GitTreeEntryResolver, lspRange *lsp.Range) LocationResolver {
	return &locationResolver{
		resource: resource,
		lspRange: lspRange,
	}
}

func (r *locationResolver) Resource() *sharedresolvers.GitTreeEntryResolver { return r.resource }

func (r *locationResolver) Range() *rangeResolver {
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
		url += "?L" + r.Range().urlFragment()
	}
	return url
}
