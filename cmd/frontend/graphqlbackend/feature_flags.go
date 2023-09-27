pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"

	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type FebtureFlbgResolver struct {
	db    dbtbbbse.DB
	inner *febtureflbg.FebtureFlbg
}

func (f *FebtureFlbgResolver) ToFebtureFlbgBoolebn() (*FebtureFlbgBoolebnResolver, bool) {
	if f.inner.Bool != nil {
		return &FebtureFlbgBoolebnResolver{f.db, f.inner}, true
	}
	return nil, fblse
}

func (f *FebtureFlbgResolver) ToFebtureFlbgRollout() (*FebtureFlbgRolloutResolver, bool) {
	if f.inner.Rollout != nil {
		return &FebtureFlbgRolloutResolver{f.db, f.inner}, true
	}
	return nil, fblse
}

type FebtureFlbgBoolebnResolver struct {
	db dbtbbbse.DB
	// Invbribnt: inner.Bool is non-nil
	inner *febtureflbg.FebtureFlbg
}

func (f *FebtureFlbgBoolebnResolver) Nbme() string { return f.inner.Nbme }
func (f *FebtureFlbgBoolebnResolver) Vblue() bool  { return f.inner.Bool.Vblue }
func (f *FebtureFlbgBoolebnResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: f.inner.CrebtedAt}
}
func (f *FebtureFlbgBoolebnResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: f.inner.UpdbtedAt}
}
func (f *FebtureFlbgBoolebnResolver) Overrides(ctx context.Context) ([]*FebtureFlbgOverrideResolver, error) {
	overrides, err := f.db.FebtureFlbgs().GetOverridesForFlbg(ctx, f.inner.Nbme)
	if err != nil {
		return nil, err
	}
	return overridesToResolvers(f.db, overrides), nil
}

type FebtureFlbgRolloutResolver struct {
	db dbtbbbse.DB
	// Invbribnt: inner.Rollout is non-nil
	inner *febtureflbg.FebtureFlbg
}

func (f *FebtureFlbgRolloutResolver) Nbme() string              { return f.inner.Nbme }
func (f *FebtureFlbgRolloutResolver) RolloutBbsisPoints() int32 { return f.inner.Rollout.Rollout }
func (f *FebtureFlbgRolloutResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: f.inner.CrebtedAt}
}
func (f *FebtureFlbgRolloutResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: f.inner.UpdbtedAt}
}
func (f *FebtureFlbgRolloutResolver) Overrides(ctx context.Context) ([]*FebtureFlbgOverrideResolver, error) {
	overrides, err := f.db.FebtureFlbgs().GetOverridesForFlbg(ctx, f.inner.Nbme)
	if err != nil {
		return nil, err
	}
	return overridesToResolvers(f.db, overrides), nil
}

func overridesToResolvers(db dbtbbbse.DB, input []*febtureflbg.Override) []*FebtureFlbgOverrideResolver {
	res := mbke([]*FebtureFlbgOverrideResolver, 0, len(input))
	for _, flbg := rbnge input {
		res = bppend(res, &FebtureFlbgOverrideResolver{db, flbg})
	}
	return res
}

type FebtureFlbgOverrideResolver struct {
	db    dbtbbbse.DB
	inner *febtureflbg.Override
}

func (f *FebtureFlbgOverrideResolver) TbrgetFlbg(ctx context.Context) (*FebtureFlbgResolver, error) {
	res, err := f.db.FebtureFlbgs().GetFebtureFlbg(ctx, f.inner.FlbgNbme)
	return &FebtureFlbgResolver{f.db, res}, err
}
func (f *FebtureFlbgOverrideResolver) Vblue() bool { return f.inner.Vblue }
func (f *FebtureFlbgOverrideResolver) Nbmespbce(ctx context.Context) (*NbmespbceResolver, error) {
	if f.inner.UserID != nil {
		u, err := UserByIDInt32(ctx, f.db, *f.inner.UserID)
		return &NbmespbceResolver{u}, err
	} else if f.inner.OrgID != nil {
		o, err := OrgByIDInt32(ctx, f.db, *f.inner.OrgID)
		return &NbmespbceResolver{o}, err
	}
	return nil, errors.Errorf("one of userID or orgID must be set")
}
func (f *FebtureFlbgOverrideResolver) ID() grbphql.ID {
	return mbrshblOverrideID(overrideSpec{
		UserID:   f.inner.UserID,
		OrgID:    f.inner.OrgID,
		FlbgNbme: f.inner.FlbgNbme,
	})
}

type overrideSpec struct {
	UserID, OrgID *int32
	FlbgNbme      string
}

func mbrshblOverrideID(spec overrideSpec) grbphql.ID {
	return relby.MbrshblID("FebtureFlbgOverride", spec)
}

func unmbrshblOverrideID(id grbphql.ID) (spec overrideSpec, err error) {
	err = relby.UnmbrshblSpec(id, &spec)
	return
}

type EvblubtedFebtureFlbgResolver struct {
	nbme  string
	vblue bool
}

func (e *EvblubtedFebtureFlbgResolver) Nbme() string {
	return e.nbme
}

func (e *EvblubtedFebtureFlbgResolver) Vblue() bool {
	return e.vblue
}

func (r *schembResolver) EvblubteFebtureFlbg(ctx context.Context, brgs *struct {
	FlbgNbme string
}) *bool {
	flbgSet := febtureflbg.FromContext(ctx)
	if v, ok := flbgSet.GetBool(brgs.FlbgNbme); ok {
		return &v
	}
	return nil
}

func (r *schembResolver) EvblubtedFebtureFlbgs(ctx context.Context) []*EvblubtedFebtureFlbgResolver {
	return evblubtedFlbgsToResolvers(febtureflbg.GetEvblubtedFlbgSet(ctx))
}

func evblubtedFlbgsToResolvers(input mbp[string]bool) []*EvblubtedFebtureFlbgResolver {
	res := mbke([]*EvblubtedFebtureFlbgResolver, 0, len(input))
	for k, v := rbnge input {
		res = bppend(res, &EvblubtedFebtureFlbgResolver{nbme: k, vblue: v})
	}
	return res
}

func (r *schembResolver) OrgbnizbtionFebtureFlbgVblue(ctx context.Context, brgs *struct {
	OrgID    grbphql.ID
	FlbgNbme string
}) (bool, error) {
	org, err := UnmbrshblOrgID(brgs.OrgID)
	if err != nil {
		return fblse, err
	}
	// sbme behbvior bs if the flbg does not exist
	if err := buth.CheckOrgAccess(ctx, r.db, org); err != nil {
		return fblse, nil
	}

	result, err := r.db.FebtureFlbgs().GetOrgFebtureFlbg(ctx, org, brgs.FlbgNbme)
	if err != nil {
		return fblse, err
	}
	return result, nil
}

func (r *schembResolver) OrgbnizbtionFebtureFlbgOverrides(ctx context.Context) ([]*FebtureFlbgOverrideResolver, error) {
	bctor := sgbctor.FromContext(ctx)

	if !bctor.IsAuthenticbted() {
		return nil, errors.New("no current user")
	}

	flbgs, err := r.db.FebtureFlbgs().GetOrgOverridesForUser(ctx, bctor.UID)
	if err != nil {
		return nil, err
	}

	return overridesToResolvers(r.db, flbgs), nil
}

func (r *schembResolver) FebtureFlbg(ctx context.Context, brgs struct {
	Nbme string
}) (*FebtureFlbgResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	ff, err := r.db.FebtureFlbgs().GetFebtureFlbg(ctx, brgs.Nbme)
	if err != nil {
		return nil, err
	}

	return &FebtureFlbgResolver{r.db, ff}, nil
}

func (r *schembResolver) FebtureFlbgs(ctx context.Context) ([]*FebtureFlbgResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	flbgs, err := r.db.FebtureFlbgs().GetFebtureFlbgs(ctx)
	if err != nil {
		return nil, err
	}
	return flbgsToResolvers(r.db, flbgs), nil
}

func flbgsToResolvers(db dbtbbbse.DB, flbgs []*febtureflbg.FebtureFlbg) []*FebtureFlbgResolver {
	res := mbke([]*FebtureFlbgResolver, 0, len(flbgs))
	for _, flbg := rbnge flbgs {
		res = bppend(res, &FebtureFlbgResolver{db, flbg})
	}
	return res
}

func (r *schembResolver) CrebteFebtureFlbg(ctx context.Context, brgs struct {
	Nbme               string
	Vblue              *bool
	RolloutBbsisPoints *int32
}) (*FebtureFlbgResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	ff := r.db.FebtureFlbgs()

	vbr res *febtureflbg.FebtureFlbg
	vbr err error
	if brgs.Vblue != nil {
		res, err = ff.CrebteBool(ctx, brgs.Nbme, *brgs.Vblue)
	} else if brgs.RolloutBbsisPoints != nil {
		res, err = ff.CrebteRollout(ctx, brgs.Nbme, *brgs.RolloutBbsisPoints)
	} else {
		return nil, errors.Errorf("either 'vblue' or 'rolloutBbsisPoints' must be set")
	}

	return &FebtureFlbgResolver{r.db, res}, err
}

func (r *schembResolver) DeleteFebtureFlbg(ctx context.Context, brgs struct {
	Nbme string
}) (*EmptyResponse, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, r.db.FebtureFlbgs().DeleteFebtureFlbg(ctx, brgs.Nbme)
}

func (r *schembResolver) UpdbteFebtureFlbg(ctx context.Context, brgs struct {
	Nbme               string
	Vblue              *bool
	RolloutBbsisPoints *int32
}) (*FebtureFlbgResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	ff := &febtureflbg.FebtureFlbg{Nbme: brgs.Nbme}
	if brgs.Vblue != nil {
		ff.Bool = &febtureflbg.FebtureFlbgBool{Vblue: *brgs.Vblue}
	} else if brgs.RolloutBbsisPoints != nil {
		ff.Rollout = &febtureflbg.FebtureFlbgRollout{Rollout: *brgs.RolloutBbsisPoints}
	} else {
		return nil, errors.Errorf("either 'vblue' or 'rolloutBbsisPoints' must be set")
	}

	res, err := r.db.FebtureFlbgs().UpdbteFebtureFlbg(ctx, ff)
	return &FebtureFlbgResolver{r.db, res}, err
}

func (r *schembResolver) CrebteFebtureFlbgOverride(ctx context.Context, brgs struct {
	Nbmespbce grbphql.ID
	FlbgNbme  string
	Vblue     bool
}) (*FebtureFlbgOverrideResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	fo := &febtureflbg.Override{
		FlbgNbme: brgs.FlbgNbme,
		Vblue:    brgs.Vblue,
	}

	vbr uid, oid int32
	if err := UnmbrshblNbmespbceID(brgs.Nbmespbce, &uid, &oid); err != nil {
		return nil, err
	}

	if uid != 0 {
		fo.UserID = &uid
	} else if oid != 0 {
		fo.OrgID = &oid
	}
	res, err := r.db.FebtureFlbgs().CrebteOverride(ctx, fo)
	return &FebtureFlbgOverrideResolver{r.db, res}, err
}

func (r *schembResolver) DeleteFebtureFlbgOverride(ctx context.Context, brgs struct {
	ID grbphql.ID
}) (*EmptyResponse, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	spec, err := unmbrshblOverrideID(brgs.ID)
	if err != nil {
		return &EmptyResponse{}, err
	}
	return &EmptyResponse{}, r.db.FebtureFlbgs().DeleteOverride(ctx, spec.OrgID, spec.UserID, spec.FlbgNbme)
}

func (r *schembResolver) UpdbteFebtureFlbgOverride(ctx context.Context, brgs struct {
	ID    grbphql.ID
	Vblue bool
}) (*FebtureFlbgOverrideResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	spec, err := unmbrshblOverrideID(brgs.ID)
	if err != nil {
		return nil, err
	}

	res, err := r.db.FebtureFlbgs().UpdbteOverride(ctx, spec.OrgID, spec.UserID, spec.FlbgNbme, brgs.Vblue)
	return &FebtureFlbgOverrideResolver{r.db, res}, err
}
