package resolvers

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
	bulkJob *btypes.BulkJob
}

var _ graphqlbackend.BulkJobResolver = &bulkJobResolver{}

func (r *bulkJobResolver) ID() graphql.ID {
	return marshalBulkJobID(r.bulkJob.ID)
}

func (r *bulkJobResolver) Type() (string, error) {
	return changesetJobTypeToBulkJobType(r.bulkJob.Type)
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

	changesetIDs := uniqueChangesetIDsForBulkJobErrors(errors)

	changesetsByID := map[int64]*btypes.Changeset{}
	reposByID := map[api.RepoID]*types.Repo{}
	if len(changesetIDs) > 0 {
		// Load all changesets and repos at once, to avoid N+1 queries.
		changesets, _, err := r.store.ListChangesets(ctx, store.ListChangesetsOpts{IDs: changesetIDs})
		if err != nil {
			return nil, err
		}
		for _, ch := range changesets {
			changesetsByID[ch.ID] = ch
		}
		// 🚨 SECURITY: database.Repos.GetReposSetByIDs uses the authzFilter under the hood and
		// filters out repositories that the user doesn't have access to.
		reposByID, err = r.store.Repos().GetReposSetByIDs(ctx, changesets.RepoIDs()...)
		if err != nil {
			return nil, err
		}
	}

	res := make([]graphqlbackend.ChangesetJobErrorResolver, 0, len(errors))
	for _, e := range errors {
		ch := changesetsByID[e.ChangesetID]
		repo, accessible := reposByID[ch.RepoID]
		resolver := &changesetJobErrorResolver{store: r.store, changeset: ch, repo: repo}
		if accessible {
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

func changesetJobTypeToBulkJobType(t btypes.ChangesetJobType) (string, error) {
	switch t {
	case btypes.ChangesetJobTypeComment:
		return "COMMENT", nil
	default:
		return "", fmt.Errorf("invalid job type %q", t)
	}
}

func uniqueChangesetIDsForBulkJobErrors(errors []*btypes.BulkJobError) []int64 {
	changesetIDsMap := map[int64]struct{}{}
	changesetIDs := []int64{}
	for _, e := range errors {
		if _, ok := changesetIDsMap[e.ChangesetID]; ok {
			continue
		}
		changesetIDs = append(changesetIDs, e.ChangesetID)
		changesetIDsMap[e.ChangesetID] = struct{}{}
	}
	return changesetIDs
}
