package resolvers

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.ChangesetSpecResolver = &changesetSpecResolver{}

type changesetSpecResolver struct {
}

func (r *changesetSpecResolver) ID() (graphql.ID, error) {
	return "", errors.New("TODO: not implemented")
}

func (r *changesetSpecResolver) ExpiresAt() *graphqlbackend.DateTime {
	return &graphqlbackend.DateTime{Time: time.Now()}
}
