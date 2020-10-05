package graphs

import (
	"context"
	"errors"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/graphs"
)

var _ graphqlbackend.GraphResolver = &graphResolver{}

// graphResolver implements the GraphQL type Graph.
type graphResolver struct {
	*graphs.Graph

	// Cache the owner on the resolver, since it's accessed more than once.
	ownerOnce sync.Once
	owner     *graphqlbackend.GraphOwnerResolver
	ownerErr  error
}

const graphIDKind = "Graph"

func marshalGraphID(id int64) graphql.ID {
	return relay.MarshalID(graphIDKind, id)
}

func unmarshalGraphID(id graphql.ID) (graphID int64, err error) {
	err = relay.UnmarshalSpec(id, &graphID)
	return
}

func (r *graphResolver) ID() graphql.ID {
	return marshalGraphID(r.Graph.ID)
}

func (r *graphResolver) Owner(ctx context.Context) (*graphqlbackend.GraphOwnerResolver, error) {
	return r.computeOwner(ctx)
}

func (r *graphResolver) computeOwner(ctx context.Context) (*graphqlbackend.GraphOwnerResolver, error) {
	r.ownerOnce.Do(func() {
		r.owner = &graphqlbackend.GraphOwnerResolver{}
		if r.Graph.OwnerUserID != 0 {
			r.owner.GraphOwner, r.ownerErr = graphqlbackend.UserByIDInt32(
				ctx,
				r.Graph.OwnerUserID,
			)
		} else {
			r.owner.GraphOwner, r.ownerErr = graphqlbackend.OrgByIDInt32(
				ctx,
				r.Graph.OwnerOrgID,
			)
		}
		if errcode.IsNotFound(r.ownerErr) {
			r.owner.GraphOwner = nil
			r.ownerErr = errors.New("owner of graph has been deleted")
		}
	})

	return r.owner, r.ownerErr
}

func (r *graphResolver) Name() string {
	return r.Graph.Name
}

func (r *graphResolver) Description() *string {
	return r.Graph.Description
}

func (r *graphResolver) Spec() string {
	return r.Graph.Spec
}

func (r *graphResolver) URL(ctx context.Context) (string, error) {
	o, err := r.Owner(ctx)
	if err != nil {
		return "", err
	}
	return graphURL(o, r), nil
}

func (r *graphResolver) EditURL(ctx context.Context) (string, error) {
	// TODO(sqs)
	urlStr, err := r.URL(ctx)
	if err != nil {
		return "", err
	}
	return urlStr + "/edit", nil
}

func (r *graphResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Graph.CreatedAt}
}

func (r *graphResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Graph.UpdatedAt}
}

func (r *graphResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	// TODO(sqs)
	return true, nil
}
