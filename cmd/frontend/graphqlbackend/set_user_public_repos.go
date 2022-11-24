package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) SetUserPublicRepos(ctx context.Context, args struct {
	UserID   graphql.ID
	RepoURIs []string
}) (*EmptyResponse, error) {
	return nil, errors.Errorf("SetUserPublicRepos has been deprecated")
}
