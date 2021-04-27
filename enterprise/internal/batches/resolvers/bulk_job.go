package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

const bulkJobIDKind = "BulkJob"

func marshalBulkJobID(id string) graphql.ID {
	return relay.MarshalID(bulkJobIDKind, id)
}

func unmarshalBulkJobID(id graphql.ID) (bulkJobID string, err error) {
	err = relay.UnmarshalSpec(id, &bulkJobID)
	return
}

type bulkJobResolver struct {
	store   *store.Store
	bulkJob *store.BulkJob
}

var _ graphqlbackend.BulkJobResolver = &bulkJobResolver{}

func (r *bulkJobResolver) ID() graphql.ID {
	return marshalBulkJobID(r.bulkJob.ID)
}

func (r *bulkJobResolver) Type() string {
	return translateBulkJobType(r.bulkJob.Type)
}

func translateBulkJobType(t btypes.ChangesetJobType) string {
	switch t {
	case btypes.ChangesetJobTypeComment:
		return "COMMENT"
	default:
		return "INVALID"
	}
}

func (r *bulkJobResolver) State() string {
	return string(r.bulkJob.State)
}

func (r *bulkJobResolver) Progress() float64 {
	return r.bulkJob.Progress
}

func (r *bulkJobResolver) Errors(ctx context.Context) ([]graphqlbackend.ChangesetJobErrorResolver, error) {
	errors, err := r.store.ListBulkJobErrors(ctx, store.ListBulkJobErrorsOpts{BulkJobID: r.bulkJob.ID})
	if err != nil {
		return nil, err
	}
	res := make([]graphqlbackend.ChangesetJobErrorResolver, 0, len(errors))
	for _, e := range errors {
		changeset, err := r.store.GetChangeset(ctx, store.GetChangesetOpts{ID: e.ChangesetID})
		if err != nil {
			return nil, err
		}

		// 🚨 SECURITY: database.Repos.Get uses the authzFilter under the hood and
		// filters out repositories that the user doesn't have access to.
		repo, err := r.store.Repos().Get(ctx, changeset.RepoID)
		if err != nil && !errcode.IsNotFound(err) {
			return nil, err
		}
		resolver := &changesetJobErrorResolver{store: r.store, changeset: changeset, repo: repo}
		if !errcode.IsNotFound(err) {
			resolver.error = e.Error
		}
		res = append(res, resolver)
	}
	return res, nil
}

func (r *bulkJobResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.bulkJob.CreatedAt}
}

func (r *bulkJobResolver) FinishedAt() *graphqlbackend.DateTime {
	if r.bulkJob.FinishedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.bulkJob.FinishedAt}
}
