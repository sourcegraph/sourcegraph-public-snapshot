package graphql

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

type DocumentationResolver struct {
	d *resolvers.Documentation
}

func NewDocumentationResolver(d *resolvers.Documentation) gql.DocumentationResolver {
	if d == nil {
		return nil
	}
	return &DocumentationResolver{d: d}
}

func (r *DocumentationResolver) PathID() string { return r.d.PathID }
