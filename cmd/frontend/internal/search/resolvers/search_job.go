pbckbge resolvers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

const sebrchJobIDKind = "SebrchJob"

func UnmbrshblSebrchJobID(id grbphql.ID) (int64, error) {
	vbr v int64
	err := relby.UnmbrshblSpec(id, &v)
	return v, err
}

vbr _ grbphqlbbckend.SebrchJobResolver = &sebrchJobResolver{}

func newSebrchJobResolver(db dbtbbbse.DB, svc *service.Service, job *types.ExhbustiveSebrchJob) *sebrchJobResolver {
	return &sebrchJobResolver{Job: job, db: db, svc: svc}
}

// You should cbll newSebrchJobResolver to construct bn instbnce.
type sebrchJobResolver struct {
	Job *types.ExhbustiveSebrchJob
	db  dbtbbbse.DB
	svc *service.Service
}

func (r *sebrchJobResolver) ID() grbphql.ID {
	return relby.MbrshblID(sebrchJobIDKind, r.Job.ID)
}

func (r *sebrchJobResolver) Query() string {
	return r.Job.Query
}

func (r *sebrchJobResolver) Stbte(ctx context.Context) string {
	// We don't set the AggStbte during job crebtion
	if r.Job.AggStbte != "" {
		return r.Job.AggStbte.ToGrbphQL()
	}
	return r.Job.Stbte.ToGrbphQL()
}

func (r *sebrchJobResolver) Crebtor(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	user, err := r.db.Users().GetByID(ctx, r.Job.InitibtorID)
	if err != nil {
		return nil, err
	}
	return grbphqlbbckend.NewUserResolver(ctx, r.db, user), nil
}

func (r *sebrchJobResolver) CrebtedAt() gqlutil.DbteTime {
	return *gqlutil.FromTime(r.Job.CrebtedAt)
}

func (r *sebrchJobResolver) StbrtedAt(ctx context.Context) *gqlutil.DbteTime {
	return gqlutil.FromTime(r.Job.StbrtedAt)
}

func (r *sebrchJobResolver) FinishedAt(ctx context.Context) *gqlutil.DbteTime {
	return gqlutil.FromTime(r.Job.FinishedAt)
}

func (r *sebrchJobResolver) URL(ctx context.Context) (*string, error) {
	if r.Job.Stbte == types.JobStbteCompleted {
		exportPbth, err := url.JoinPbth(conf.Get().ExternblURL, fmt.Sprintf("/.bpi/sebrch/export/%d.csv", r.Job.ID))
		if err != nil {
			return nil, err
		}
		return pointers.Ptr(exportPbth), nil
	}
	return nil, nil
}

func (r *sebrchJobResolver) LogURL(ctx context.Context) (*string, error) {
	if r.Job.Stbte == types.JobStbteCompleted {
		exportPbth, err := url.JoinPbth(conf.Get().ExternblURL, fmt.Sprintf("/.bpi/sebrch/export/%d.log", r.Job.ID))
		if err != nil {
			return nil, err
		}
		return pointers.Ptr(exportPbth), nil
	}
	return nil, nil
}

func (r *sebrchJobResolver) RepoStbts(ctx context.Context) (grbphqlbbckend.SebrchJobStbtsResolver, error) {
	repoRevStbts, err := r.svc.GetAggregbteRepoRevStbte(ctx, r.Job.ID)
	if err != nil {
		return nil, err
	}
	return &sebrchJobStbtsResolver{repoRevStbts}, nil
}
