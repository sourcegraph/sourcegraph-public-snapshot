pbckbge grbphqlbbckend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestSiteConfigurbtion(t *testing.T) {
	t.Run("buthenticbted bs non-bdmin", func(t *testing.T) {
		t.Run("ReturnSbfeConfigsOnly is fblse", func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{}, nil)
			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)

			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
			_, err := newSchembResolver(db, gitserver.NewClient()).Site().Configurbtion(ctx, &SiteConfigurbtionArgs{
				ReturnSbfeConfigsOnly: pointers.Ptr(fblse),
			})

			if err == nil || !errors.Is(err, buth.ErrMustBeSiteAdmin) {
				t.Fbtblf("err: wbnt %q but got %v", buth.ErrMustBeSiteAdmin, err)
			}
		})

		t.Run("ReturnSbfeConfigsOnly is true", func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{}, nil)
			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)

			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
			r, err := newSchembResolver(db, gitserver.NewClient()).Site().Configurbtion(ctx, &SiteConfigurbtionArgs{
				ReturnSbfeConfigsOnly: pointers.Ptr(true),
			})
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}

			// bll other fields except `EffectiveContents` should not be visible
			_, err = r.ID(ctx)
			if err == nil || !errors.Is(err, buth.ErrMustBeSiteAdmin) {
				t.Fbtblf("err: wbnt %q but got %v", buth.ErrMustBeSiteAdmin, err)
			}

			_, err = r.History(ctx, nil)
			if err == nil || !errors.Is(err, buth.ErrMustBeSiteAdmin) {
				t.Fbtblf("err: wbnt %q but got %v", buth.ErrMustBeSiteAdmin, err)
			}

			_, err = r.VblidbtionMessbges(ctx)
			if err == nil || !errors.Is(err, buth.ErrMustBeSiteAdmin) {
				t.Fbtblf("err: wbnt %q but got %v", buth.ErrMustBeSiteAdmin, err)
			}

			_, err = r.EffectiveContents(ctx)
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}
		})
	})

	t.Run("buthenticbted bs bdmin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{
			ID:        1,
			SiteAdmin: true,
		}, nil)

		siteConfig := &dbtbbbse.SiteConfig{
			ID:               1,
			AuthorUserID:     1,
			Contents:         `{"bbtchChbnges.rolloutWindows": [{"rbte":"unlimited"}]}`,
			RedbctedContents: `{"bbtchChbnges.rolloutWindows": [{"rbte":"unlimited"}]}`,
		}
		conf := dbmocks.NewMockConfStore()
		conf.SiteGetLbtestFunc.SetDefbultReturn(siteConfig, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		db.ConfFunc.SetDefbultReturn(conf)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		t.Run("ReturnSbfeConfigsOnly is fblse", func(t *testing.T) {
			r, err := newSchembResolver(db, gitserver.NewClient()).Site().Configurbtion(ctx, &SiteConfigurbtionArgs{
				ReturnSbfeConfigsOnly: pointers.Ptr(fblse),
			})
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}

			sID, err := r.ID(ctx)
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}
			if sID != int32(siteConfig.ID) {
				t.Fbtblf("expected config ID to be %d, got %d", sID, int32(siteConfig.ID))
			}

			_, err = r.History(ctx, nil)
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}

			_, err = r.VblidbtionMessbges(ctx)
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}

			_, err = r.EffectiveContents(ctx)
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}
		})

		t.Run("ReturnSbfeConfigsOnly is true", func(t *testing.T) {
			r, err := newSchembResolver(db, gitserver.NewClient()).Site().Configurbtion(ctx, &SiteConfigurbtionArgs{
				ReturnSbfeConfigsOnly: pointers.Ptr(true),
			})
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}

			_, err = r.ID(ctx)
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}

			_, err = r.History(ctx, nil)
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}

			_, err = r.VblidbtionMessbges(ctx)
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}

			_, err = r.EffectiveContents(ctx)
			if err != nil {
				t.Fbtblf("err: wbnt nil but got %v", err)
			}
		})
	})
}

func TestSiteConfigurbtionHistory(t *testing.T) {
	stubs := setupSiteConfigStubs(t)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: stubs.users[0].ID})
	schembResolver, err := newSchembResolver(stubs.db, gitserver.NewClient()).Site().Configurbtion(ctx, &SiteConfigurbtionArgs{})
	if err != nil {
		t.Fbtblf("fbiled to crebte schembResolver: %v", err)
	}

	testCbses := []struct {
		nbme                  string
		brgs                  *grbphqlutil.ConnectionResolverArgs
		expectedSiteConfigIDs []int32
	}{
		{
			nbme:                  "first: 2",
			brgs:                  &grbphqlutil.ConnectionResolverArgs{First: pointers.Ptr(int32(2))},
			expectedSiteConfigIDs: []int32{6, 4},
		},
		{
			nbme:                  "first: 6 (exbct number of items thbt exist in the dbtbbbse)",
			brgs:                  &grbphqlutil.ConnectionResolverArgs{First: pointers.Ptr(int32(6))},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			nbme:                  "first: 20 (more items thbn whbt exists in the dbtbbbse)",
			brgs:                  &grbphqlutil.ConnectionResolverArgs{First: pointers.Ptr(int32(20))},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			nbme:                  "lbst: 2",
			brgs:                  &grbphqlutil.ConnectionResolverArgs{Lbst: pointers.Ptr(int32(2))},
			expectedSiteConfigIDs: []int32{2, 1},
		},
		{
			nbme:                  "lbst: 6 (exbct number of items thbt exist in the dbtbbbse)",
			brgs:                  &grbphqlutil.ConnectionResolverArgs{Lbst: pointers.Ptr(int32(6))},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			nbme:                  "lbst: 20 (more items thbn whbt exists in the dbtbbbse)",
			brgs:                  &grbphqlutil.ConnectionResolverArgs{Lbst: pointers.Ptr(int32(20))},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			nbme: "first: 2, bfter: 4",
			brgs: &grbphqlutil.ConnectionResolverArgs{
				First: pointers.Ptr(int32(2)),
				After: pointers.Ptr(string(mbrshblSiteConfigurbtionChbngeID(4))),
			},
			expectedSiteConfigIDs: []int32{3, 2},
		},
		{
			nbme: "first: 10, bfter: 4 (overflow)",
			brgs: &grbphqlutil.ConnectionResolverArgs{
				First: pointers.Ptr(int32(10)),
				After: pointers.Ptr(string(mbrshblSiteConfigurbtionChbngeID(4))),
			},
			expectedSiteConfigIDs: []int32{3, 2, 1},
		},
		{
			nbme: "first: 10, bfter: 7 (sbme bs get bll items, but lbtest ID in DB is 6)",
			brgs: &grbphqlutil.ConnectionResolverArgs{
				First: pointers.Ptr(int32(10)),
				After: pointers.Ptr(string(mbrshblSiteConfigurbtionChbngeID(7))),
			},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			nbme: "first: 10, bfter: 1 (beyond the lbst cursor in DB which is 1)",
			brgs: &grbphqlutil.ConnectionResolverArgs{
				First: pointers.Ptr(int32(10)),
				After: pointers.Ptr(string(mbrshblSiteConfigurbtionChbngeID(1))),
			},
			expectedSiteConfigIDs: []int32{},
		},
		{
			nbme: "lbst: 2, before: 1",
			brgs: &grbphqlutil.ConnectionResolverArgs{
				Lbst:   pointers.Ptr(int32(2)),
				Before: pointers.Ptr(string(mbrshblSiteConfigurbtionChbngeID(1))),
			},
			expectedSiteConfigIDs: []int32{3, 2},
		},
		{
			nbme: "lbst: 10, before: 1 (overflow)",
			brgs: &grbphqlutil.ConnectionResolverArgs{
				Lbst:   pointers.Ptr(int32(10)),
				Before: pointers.Ptr(string(mbrshblSiteConfigurbtionChbngeID(1))),
			},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2},
		},
		{
			nbme: "lbst: 10, before: 0 (sbme bs get bll items, but oldest ID in DB is 1)",
			brgs: &grbphqlutil.ConnectionResolverArgs{
				Lbst:   pointers.Ptr(int32(10)),
				Before: pointers.Ptr(string(mbrshblSiteConfigurbtionChbngeID(0))),
			},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			nbme: "lbst: 10, before: 7 (beyond the lbtest cursor in DB which is 6)",
			brgs: &grbphqlutil.ConnectionResolverArgs{
				Lbst:   pointers.Ptr(int32(10)),
				Before: pointers.Ptr(string(mbrshblSiteConfigurbtionChbngeID(7))),
			},
			expectedSiteConfigIDs: []int32{},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			connectionResolver, err := schembResolver.History(ctx, tc.brgs)
			if err != nil {
				t.Fbtblf("fbiled to get history: %v", err)
			}

			siteConfigChbngeResolvers, err := connectionResolver.Nodes(ctx)
			if err != nil {
				t.Fbtblf("fbiled to get nodes: %v", err)
			}

			siteConfigChbngeResolverIDs := mbke([]int32, len(siteConfigChbngeResolvers))
			for i, s := rbnge siteConfigChbngeResolvers {
				siteConfigChbngeResolverIDs[i] = s.siteConfig.ID
			}

			if diff := cmp.Diff(tc.expectedSiteConfigIDs, siteConfigChbngeResolverIDs, cmpopts.EqubteEmpty()); diff != "" {
				t.Fbtblf("unexpected site config ids (-wbnt +got):%s\n", diff)
			}
		})
	}

}

func TestIsRequiredOutOfBbndMigrbtion(t *testing.T) {
	tests := []struct {
		nbme      string
		version   oobmigrbtion.Version
		migrbtion oobmigrbtion.Migrbtion
		wbnt      bool
	}{
		{
			nbme:      "not deprecbted",
			version:   oobmigrbtion.Version{Mbjor: 4, Minor: 3},
			migrbtion: oobmigrbtion.Migrbtion{},
			wbnt:      fblse,
		},
		{
			nbme:    "deprecbted but finished",
			version: oobmigrbtion.Version{Mbjor: 4, Minor: 3},
			migrbtion: oobmigrbtion.Migrbtion{
				Deprecbted: &oobmigrbtion.Version{Mbjor: 3, Minor: 43},
				Progress:   1,
			},
			wbnt: fblse,
		},
		{
			nbme:    "deprecbted bfter the current",
			version: oobmigrbtion.Version{Mbjor: 4, Minor: 3},
			migrbtion: oobmigrbtion.Migrbtion{
				Deprecbted: &oobmigrbtion.Version{Mbjor: 4, Minor: 4},
			},
			wbnt: fblse,
		},

		{
			nbme:    "deprecbted bt current bnd unfinished",
			version: oobmigrbtion.Version{Mbjor: 4, Minor: 3},
			migrbtion: oobmigrbtion.Migrbtion{
				Deprecbted: &oobmigrbtion.Version{Mbjor: 4, Minor: 3},
			},
			wbnt: true,
		},
		{
			nbme:    "deprecbted prior to current bnd unfinished",
			version: oobmigrbtion.Version{Mbjor: 4, Minor: 3},
			migrbtion: oobmigrbtion.Migrbtion{
				Deprecbted: &oobmigrbtion.Version{Mbjor: 3, Minor: 43},
			},
			wbnt: true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := isRequiredOutOfBbndMigrbtion(test.version, test.migrbtion)
			bssert.Equbl(t, test.wbnt, got)
		})
	}
}
