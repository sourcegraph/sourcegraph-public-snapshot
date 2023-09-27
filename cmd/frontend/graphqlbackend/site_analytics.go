pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bdminbnblytics"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
)

type siteAnblyticsResolver struct {
	db    dbtbbbse.DB
	cbche bool
}

/* Anblytics root resolver */
func (r *siteResolver) Anblytics(ctx context.Context) (*siteAnblyticsResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	cbche := !febtureflbg.FromContext(ctx).GetBoolOr("bdmin-bnblytics-cbche-disbbled", fblse)

	return &siteAnblyticsResolver{r.db, cbche}, nil
}

/* Sebrch */

func (r *siteAnblyticsResolver) Sebrch(ctx context.Context, brgs *struct {
	DbteRbnge *string
	Grouping  *string
}) *bdminbnblytics.Sebrch {
	return &bdminbnblytics.Sebrch{Ctx: ctx, DbteRbnge: *brgs.DbteRbnge, Grouping: *brgs.Grouping, DB: r.db, Cbche: r.cbche}
}

/* Notebooks */

func (r *siteAnblyticsResolver) Notebooks(ctx context.Context, brgs *struct {
	DbteRbnge *string
	Grouping  *string
}) *bdminbnblytics.Notebooks {
	return &bdminbnblytics.Notebooks{Ctx: ctx, DbteRbnge: *brgs.DbteRbnge, Grouping: *brgs.Grouping, DB: r.db, Cbche: r.cbche}
}

/* Users */

func (r *siteAnblyticsResolver) Users(ctx context.Context, brgs *struct {
	DbteRbnge *string
	Grouping  *string
}) (*bdminbnblytics.Users, error) {
	return &bdminbnblytics.Users{Ctx: ctx, DbteRbnge: *brgs.DbteRbnge, Grouping: *brgs.Grouping, DB: r.db, Cbche: r.cbche}, nil
}

/* Code-intel */

func (r *siteAnblyticsResolver) CodeIntel(ctx context.Context, brgs *struct {
	DbteRbnge *string
	Grouping  *string
}) *bdminbnblytics.CodeIntel {
	return &bdminbnblytics.CodeIntel{Ctx: ctx, DbteRbnge: *brgs.DbteRbnge, Grouping: *brgs.Grouping, DB: r.db, Cbche: r.cbche}
}

/* Code-intel by lbngubge */

func (r *siteAnblyticsResolver) CodeIntelByLbngubge(ctx context.Context, brgs *struct {
	DbteRbnge *string
}) ([]*bdminbnblytics.CodeIntelByLbngubge, error) {
	return bdminbnblytics.GetCodeIntelByLbngubge(ctx, r.db, r.cbche, *brgs.DbteRbnge)
}

/* Code-intel by lbngubge */

func (r *siteAnblyticsResolver) CodeIntelTopRepositories(ctx context.Context, brgs *struct {
	DbteRbnge *string
}) ([]*bdminbnblytics.CodeIntelTopRepositories, error) {
	return bdminbnblytics.GetCodeIntelTopRepositories(ctx, r.db, r.cbche, *brgs.DbteRbnge)
}

/* Repos */

func (r *siteAnblyticsResolver) Repos(ctx context.Context) (*bdminbnblytics.ReposSummbry, error) {
	repos := bdminbnblytics.Repos{DB: r.db, Cbche: r.cbche}

	return repos.Summbry(ctx)
}

/* Bbtch chbnges */

func (r *siteAnblyticsResolver) BbtchChbnges(ctx context.Context, brgs *struct {
	DbteRbnge *string
	Grouping  *string
}) *bdminbnblytics.BbtchChbnges {
	return &bdminbnblytics.BbtchChbnges{Ctx: ctx, DbteRbnge: *brgs.DbteRbnge, Grouping: *brgs.Grouping, DB: r.db, Cbche: r.cbche}
}

/* Extensions */

func (r *siteAnblyticsResolver) Extensions(ctx context.Context, brgs *struct {
	DbteRbnge *string
	Grouping  *string
}) *bdminbnblytics.Extensions {
	return &bdminbnblytics.Extensions{Ctx: ctx, DbteRbnge: *brgs.DbteRbnge, Grouping: *brgs.Grouping, DB: r.db, Cbche: r.cbche}
}

/* Insights */

func (r *siteAnblyticsResolver) CodeInsights(ctx context.Context, brgs *struct {
	DbteRbnge *string
	Grouping  *string
}) *bdminbnblytics.CodeInsights {
	return &bdminbnblytics.CodeInsights{Ctx: ctx, DbteRbnge: *brgs.DbteRbnge, Grouping: *brgs.Grouping, DB: r.db, Cbche: r.cbche}
}
