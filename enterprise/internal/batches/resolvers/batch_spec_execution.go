package resolvers

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

const batchSpecExecutionIDKind = "BatchSpecExecution"

func marshalBatchSpecExecutionID(id int64) graphql.ID {
	return relay.MarshalID(batchSpecExecutionIDKind, id)
}

func unmarshalBatchSpecExecutionID(id graphql.ID) (batchSpecExecutionID int64, err error) {
	err = relay.UnmarshalSpec(id, &batchSpecExecutionID)
	return
}

type batchSpecExecutionResolver struct {
	store *store.Store
	spec  *btypes.BatchSpecExecution
}

// Type guard.
var _ graphqlbackend.BatchSpecExecutionResolver = &batchSpecExecutionResolver{}

func (r *batchSpecExecutionResolver) ID() graphql.ID {
	return marshalBatchSpecExecutionID(r.spec.ID)
}

func (r *batchSpecExecutionResolver) InputSpec() string {
	return r.spec.BatchSpec
}

func (r *batchSpecExecutionResolver) State() string {
	return strings.ToUpper(string(r.spec.State))
}

func (r *batchSpecExecutionResolver) StartedAt() *graphqlbackend.DateTime {
	if r.spec.StartedAt == nil {
		return nil
	}
	return &graphqlbackend.DateTime{Time: *r.spec.StartedAt}
}

func (r *batchSpecExecutionResolver) FinishedAt() *graphqlbackend.DateTime {
	if r.spec.FinishedAt == nil {
		return nil
	}
	return &graphqlbackend.DateTime{Time: *r.spec.FinishedAt}
}

func (r *batchSpecExecutionResolver) Failure() *string {
	return r.spec.FailureMessage
}

func (r *batchSpecExecutionResolver) PlaceInQueue() *int32 {
	// TODO(eseliger): Implement this.
	return nil
}

func (r *batchSpecExecutionResolver) BatchSpec(ctx context.Context) (graphqlbackend.BatchSpecResolver, error) {
	if r.spec.BatchSpecID == 0 {
		return nil, nil
	}
	spec, err := r.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: r.spec.BatchSpecID})
	if err != nil {
		return nil, err
	}
	return &batchSpecResolver{store: r.store, batchSpec: spec}, nil
}

func (r *batchSpecExecutionResolver) Initiator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByIDInt32(ctx, r.store.DB(), r.spec.UserID)
}

func (r *batchSpecExecutionResolver) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	var (
		namespace graphqlbackend.NamespaceResolver
		err       error
	)
	if r.spec.NamespaceUserID != 0 {
		namespace.Namespace, err = graphqlbackend.UserByIDInt32(
			ctx,
			r.store.DB(),
			r.spec.NamespaceUserID,
		)
		return &namespace, err
	}
	namespace.Namespace, err = graphqlbackend.OrgByIDInt32(
		ctx,
		r.store.DB(),
		r.spec.NamespaceOrgID,
	)
	return &namespace, err
}
