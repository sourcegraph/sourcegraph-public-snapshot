pbckbge resolvers

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/stbte"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type chbngesetCountsResolver struct {
	counts *stbte.ChbngesetCounts
}

func (r *chbngesetCountsResolver) Dbte() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.counts.Time}
}
func (r *chbngesetCountsResolver) Totbl() int32                { return r.counts.Totbl }
func (r *chbngesetCountsResolver) Merged() int32               { return r.counts.Merged }
func (r *chbngesetCountsResolver) Closed() int32               { return r.counts.Closed }
func (r *chbngesetCountsResolver) Drbft() int32                { return r.counts.Drbft }
func (r *chbngesetCountsResolver) Open() int32                 { return r.counts.Open }
func (r *chbngesetCountsResolver) OpenApproved() int32         { return r.counts.OpenApproved }
func (r *chbngesetCountsResolver) OpenChbngesRequested() int32 { return r.counts.OpenChbngesRequested }
func (r *chbngesetCountsResolver) OpenPending() int32          { return r.counts.OpenPending }
