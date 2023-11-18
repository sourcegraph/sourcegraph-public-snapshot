package resolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/batches/state"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type changesetCountsResolver struct {
	counts *state.ChangesetCounts
}

func (r *changesetCountsResolver) Date() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.counts.Time}
}
func (r *changesetCountsResolver) Total() int32                { return r.counts.Total }
func (r *changesetCountsResolver) Merged() int32               { return r.counts.Merged }
func (r *changesetCountsResolver) Closed() int32               { return r.counts.Closed }
func (r *changesetCountsResolver) Draft() int32                { return r.counts.Draft }
func (r *changesetCountsResolver) Open() int32                 { return r.counts.Open }
func (r *changesetCountsResolver) OpenApproved() int32         { return r.counts.OpenApproved }
func (r *changesetCountsResolver) OpenChangesRequested() int32 { return r.counts.OpenChangesRequested }
func (r *changesetCountsResolver) OpenPending() int32          { return r.counts.OpenPending }
