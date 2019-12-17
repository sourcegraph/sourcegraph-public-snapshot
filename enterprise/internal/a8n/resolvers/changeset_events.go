package resolvers

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
)

type changesetEventsConnectionResolver struct {
	store     *ee.Store
	changeset *a8n.Changeset
	opts      ee.ListChangesetEventsOpts

	// cache results because they are used by multiple fields
	once            sync.Once
	changesetEvents []*a8n.ChangesetEvent
	next            int64
	err             error
}

func (r *changesetEventsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetEventResolver, error) {
	changesetEvents, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.ChangesetEventResolver, 0, len(changesetEvents))
	for _, c := range changesetEvents {
		resolvers = append(resolvers, &changesetEventResolver{store: r.store, changeset: r.changeset, ChangesetEvent: c})
	}
	return resolvers, nil
}

func (r *changesetEventsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountChangesetEventsOpts{ChangesetID: r.opts.ChangesetIDs[0]}
	count, err := r.store.CountChangesetEvents(ctx, opts)
	return int32(count), err
}

func (r *changesetEventsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(next != 0), nil
}

func (r *changesetEventsConnectionResolver) compute(ctx context.Context) ([]*a8n.ChangesetEvent, int64, error) {
	r.once.Do(func() {
		r.changesetEvents, r.next, r.err = r.store.ListChangesetEvents(ctx, r.opts)
	})
	return r.changesetEvents, r.next, r.err
}

type changesetEventResolver struct {
	store     *ee.Store
	changeset *a8n.Changeset
	*a8n.ChangesetEvent
}

const changesetEventIDKind = "ChangesetEvent"

func marshalchangesetEventID(id int64) graphql.ID {
	return relay.MarshalID(changesetEventIDKind, id)
}

func (r *changesetEventResolver) ID() graphql.ID {
	return marshalchangesetEventID(r.ChangesetEvent.ID)
}

func (r *changesetEventResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.ChangesetEvent.CreatedAt}
}

func (r *changesetEventResolver) Changeset(ctx context.Context) (graphqlbackend.ExternalChangesetResolver, error) {
	return &changesetResolver{store: r.store, Changeset: r.changeset}, nil
}

type changesetCountsResolver struct {
	counts *ee.ChangesetCounts
}

func (r *changesetCountsResolver) Date() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.counts.Time}
}
func (r *changesetCountsResolver) Total() int32                { return r.counts.Total }
func (r *changesetCountsResolver) Merged() int32               { return r.counts.Merged }
func (r *changesetCountsResolver) Closed() int32               { return r.counts.Closed }
func (r *changesetCountsResolver) Open() int32                 { return r.counts.Open }
func (r *changesetCountsResolver) OpenApproved() int32         { return r.counts.OpenApproved }
func (r *changesetCountsResolver) OpenChangesRequested() int32 { return r.counts.OpenChangesRequested }
func (r *changesetCountsResolver) OpenPending() int32          { return r.counts.OpenPending }
