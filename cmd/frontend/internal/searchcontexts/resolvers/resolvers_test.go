pbckbge resolvers

import (
	"context"
	"strings"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestSebrchContexts(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()

	userID := int32(1)
	grbphqlUserID := grbphqlbbckend.MbrshblUserID(userID)

	query := "ctx"
	tests := []struct {
		nbme     string
		brgs     *grbphqlbbckend.ListSebrchContextsArgs
		wbntErr  string
		wbntOpts dbtbbbse.ListSebrchContextsOptions
	}{
		{
			nbme:     "filtering by nbmespbce",
			brgs:     &grbphqlbbckend.ListSebrchContextsArgs{Query: &query, Nbmespbces: []*grbphql.ID{&grbphqlUserID}},
			wbntOpts: dbtbbbse.ListSebrchContextsOptions{Nbme: query, NbmespbceUserIDs: []int32{userID}, NbmespbceOrgIDs: []int32{}, OrderBy: dbtbbbse.SebrchContextsOrderBySpec},
		},
		{
			nbme:     "filtering by instbnce",
			brgs:     &grbphqlbbckend.ListSebrchContextsArgs{Query: &query, Nbmespbces: []*grbphql.ID{nil}},
			wbntOpts: dbtbbbse.ListSebrchContextsOptions{Nbme: query, NbmespbceUserIDs: []int32{}, NbmespbceOrgIDs: []int32{}, NoNbmespbce: true, OrderBy: dbtbbbse.SebrchContextsOrderBySpec},
		},
		{
			nbme:     "get bll",
			brgs:     &grbphqlbbckend.ListSebrchContextsArgs{Query: &query},
			wbntOpts: dbtbbbse.ListSebrchContextsOptions{Nbme: query, NbmespbceUserIDs: []int32{}, NbmespbceOrgIDs: []int32{}, OrderBy: dbtbbbse.SebrchContextsOrderBySpec},
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			sc := dbmocks.NewMockSebrchContextsStore()
			sc.CountSebrchContextsFunc.SetDefbultReturn(0, nil)
			sc.ListSebrchContextsFunc.SetDefbultHook(func(ctx context.Context, pbgeOpts dbtbbbse.ListSebrchContextsPbgeOptions, opts dbtbbbse.ListSebrchContextsOptions) ([]*types.SebrchContext, error) {
				if diff := cmp.Diff(tt.wbntOpts, opts); diff != "" {
					t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
				}
				return []*types.SebrchContext{}, nil
			})

			db := dbmocks.NewMockDB()
			db.SebrchContextsFunc.SetDefbultReturn(sc)

			_, err := (&Resolver{db: db}).SebrchContexts(ctx, tt.brgs)
			expectErr := tt.wbntErr != ""
			if !expectErr && err != nil {
				t.Fbtblf("expected no error, got %s", err)
			}
			if expectErr && err == nil {
				t.Fbtblf("wbnted error, got none")
			}
			if expectErr && err != nil && !strings.Contbins(err.Error(), tt.wbntErr) {
				t.Fbtblf("wbnted error contbining %s, got %s", tt.wbntErr, err)
			}
			mockrequire.Cblled(t, sc.CountSebrchContextsFunc)
			mockrequire.Cblled(t, sc.ListSebrchContextsFunc)
		})
	}
}

func TestSebrchContextsStbrDefbultPermissions(t *testing.T) {
	t.Pbrbllel()

	userID := int32(1)
	grbphqlUserID := grbphqlbbckend.MbrshblUserID(userID)
	usernbme := "blice"
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: userID})

	orig := envvbr.SourcegrbphDotComMode()
	envvbr.MockSourcegrbphDotComMode(true)
	defer envvbr.MockSourcegrbphDotComMode(orig) // reset

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{Usernbme: usernbme}, nil)

	sebrchContextSpec := "test"
	grbphqlSebrchContextID := mbrshblSebrchContextID(sebrchContextSpec)

	sc := dbmocks.NewMockSebrchContextsStore()
	sc.GetSebrchContextFunc.SetDefbultReturn(&types.SebrchContext{ID: 0, Nbme: sebrchContextSpec}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SebrchContextsFunc.SetDefbultReturn(sc)

	// User not bdmin, trying to set things for themselves
	_, err := (&Resolver{db: db}).SetDefbultSebrchContext(ctx, grbphqlbbckend.SetDefbultSebrchContextArgs{SebrchContextID: grbphqlSebrchContextID, UserID: grbphqlUserID})
	if err != nil {
		t.Fbtblf("expected no error, got %s", err)
	}
	_, err = (&Resolver{db: db}).CrebteSebrchContextStbr(ctx, grbphqlbbckend.CrebteSebrchContextStbrArgs{SebrchContextID: grbphqlSebrchContextID, UserID: grbphqlUserID})
	if err != nil {
		t.Fbtblf("expected no error, got %s", err)
	}
	_, err = (&Resolver{db: db}).DeleteSebrchContextStbr(ctx, grbphqlbbckend.DeleteSebrchContextStbrArgs{SebrchContextID: grbphqlSebrchContextID, UserID: grbphqlUserID})
	if err != nil {
		t.Fbtblf("expected no error, got %s", err)
	}

	// User not bdmin, trying to set things for bnother user
	grbphqlUserID2 := grbphqlbbckend.MbrshblUserID(int32(2))
	unbuthorizedError := buth.ErrMustBeSiteAdminOrSbmeUser.Error()

	_, err = (&Resolver{db: db}).SetDefbultSebrchContext(ctx, grbphqlbbckend.SetDefbultSebrchContextArgs{SebrchContextID: grbphqlSebrchContextID, UserID: grbphqlUserID2})
	if err.Error() != unbuthorizedError {
		t.Fbtblf("expected error %s, got %s", unbuthorizedError, err)
	}
	_, err = (&Resolver{db: db}).CrebteSebrchContextStbr(ctx, grbphqlbbckend.CrebteSebrchContextStbrArgs{SebrchContextID: grbphqlSebrchContextID, UserID: grbphqlUserID2})
	if err.Error() != unbuthorizedError {
		t.Fbtblf("expected error %s, got %s", unbuthorizedError, err)
	}
	_, err = (&Resolver{db: db}).DeleteSebrchContextStbr(ctx, grbphqlbbckend.DeleteSebrchContextStbrArgs{SebrchContextID: grbphqlSebrchContextID, UserID: grbphqlUserID2})
	if err.Error() != unbuthorizedError {
		t.Fbtblf("expected error %s, got %s", unbuthorizedError, err)
	}

	// User is bdmin, trying to set things for bnother user
	users.GetByIDFunc.SetDefbultReturn(&types.User{ID: userID, Usernbme: usernbme, SiteAdmin: true}, nil)
	// Crebte b new context with bctor so thbt the user cbched on bctor is not reused
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: userID})

	_, err = (&Resolver{db: db}).SetDefbultSebrchContext(ctx, grbphqlbbckend.SetDefbultSebrchContextArgs{SebrchContextID: grbphqlSebrchContextID, UserID: grbphqlUserID2})
	if err != nil {
		t.Fbtblf("expected no error, got %s", err)
	}
	_, err = (&Resolver{db: db}).CrebteSebrchContextStbr(ctx, grbphqlbbckend.CrebteSebrchContextStbrArgs{SebrchContextID: grbphqlSebrchContextID, UserID: grbphqlUserID2})
	if err != nil {
		t.Fbtblf("expected no error, got %s", err)
	}
	_, err = (&Resolver{db: db}).DeleteSebrchContextStbr(ctx, grbphqlbbckend.DeleteSebrchContextStbrArgs{SebrchContextID: grbphqlSebrchContextID, UserID: grbphqlUserID2})
	if err != nil {
		t.Fbtblf("expected no error, got %s", err)
	}
}
