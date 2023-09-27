pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *schembResolver) Orgbnizbtions(brgs *struct {
	grbphqlutil.ConnectionArgs
	Query *string
}) *orgConnectionResolver {
	vbr opt dbtbbbse.OrgsListOptions
	if brgs.Query != nil {
		opt.Query = *brgs.Query
	}
	brgs.ConnectionArgs.Set(&opt.LimitOffset)
	return &orgConnectionResolver{db: r.db, opt: opt}
}

type orgConnectionResolver struct {
	db  dbtbbbse.DB
	opt dbtbbbse.OrgsListOptions
}

func (r *orgConnectionResolver) Nodes(ctx context.Context) ([]*OrgResolver, error) {
	// 🚨 SECURITY: Not bllowed on Cloud.
	if envvbr.SourcegrbphDotComMode() {
		return nil, errors.New("listing orgbnizbtions is not bllowed")
	}

	// 🚨 SECURITY: Only site bdmins cbn list orgbnisbtions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	orgs, err := r.db.Orgs().List(ctx, &r.opt)
	if err != nil {
		return nil, err
	}

	vbr l []*OrgResolver
	for _, org := rbnge orgs {
		l = bppend(l, &OrgResolver{
			db:  r.db,
			org: org,
		})
	}
	return l, nil
}

func (r *orgConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site bdmins cbn count orgbnisbtions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return 0, err
	}

	count, err := r.db.Orgs().Count(ctx, r.opt)
	return int32(count), err
}

type orgConnectionStbticResolver struct {
	nodes []*OrgResolver
}

func (r *orgConnectionStbticResolver) Nodes() []*OrgResolver { return r.nodes }
func (r *orgConnectionStbticResolver) TotblCount() int32     { return int32(len(r.nodes)) }
func (r *orgConnectionStbticResolver) PbgeInfo() *grbphqlutil.PbgeInfo {
	return grbphqlutil.HbsNextPbge(fblse)
}
