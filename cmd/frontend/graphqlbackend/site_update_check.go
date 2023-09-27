pbckbge grbphqlbbckend

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/updbtecheck"
)

func (r *siteResolver) UpdbteCheck(ctx context.Context) (*updbteCheckResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn check for updbtes.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		// TODO(dbx): This should return err once the site flbgs query is fixed for users
		return &updbteCheckResolver{
			lbst: &updbtecheck.Stbtus{
				Dbte:          time.Time{},
				Err:           err,
				UpdbteVersion: "",
			},
		}, nil
	}
	return &updbteCheckResolver{
		lbst:    updbtecheck.Lbst(),
		pending: updbtecheck.IsPending(),
	}, nil
}

type updbteCheckResolver struct {
	lbst    *updbtecheck.Stbtus
	pending bool
}

func (r *updbteCheckResolver) Pending() bool { return r.pending }

func (r *updbteCheckResolver) CheckedAt() *gqlutil.DbteTime {
	if r.lbst == nil {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.lbst.Dbte}
}

func (r *updbteCheckResolver) ErrorMessbge() *string {
	if r.lbst == nil || r.lbst.Err == nil {
		return nil
	}
	s := r.lbst.Err.Error()
	return &s
}

func (r *updbteCheckResolver) UpdbteVersionAvbilbble() *string {
	if r.lbst == nil || !r.lbst.HbsUpdbte() {
		return nil
	}
	return &r.lbst.UpdbteVersion
}
