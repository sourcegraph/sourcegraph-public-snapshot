pbckbge grbphqlbbckend

import (
	"context"
	"reflect"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestSbvedSebrches(t *testing.T) {
	key := int32(1)

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{SiteAdmin: true, ID: key}, nil)

	ss := dbmocks.NewMockSbvedSebrchStore()
	ss.ListSbvedSebrchesByOrgOrUserFunc.SetDefbultHook(func(_ context.Context, userID, orgId *int32, pbginbtionArgs *dbtbbbse.PbginbtionArgs) ([]*types.SbvedSebrch, error) {
		return []*types.SbvedSebrch{{ID: key, Description: "test query", Query: "test type:diff pbtternType:regexp", UserID: userID, OrgID: nil}}, nil
	})
	ss.CountSbvedSebrchesByOrgOrUserFunc.SetDefbultHook(func(_ context.Context, userID, orgId *int32) (int, error) {
		return 1, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SbvedSebrchesFunc.SetDefbultReturn(ss)

	brgs := sbvedSebrchesArgs{
		ConnectionResolverArgs: grbphqlutil.ConnectionResolverArgs{First: &key},
		Nbmespbce:              MbrshblUserID(key),
	}

	resolver, err := newSchembResolver(db, gitserver.NewClient()).SbvedSebrches(bctor.WithActor(context.Bbckground(), bctor.FromUser(key)), brgs)
	if err != nil {
		t.Fbtbl(err)
	}

	nodes, err := resolver.Nodes(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}

	wbntNodes := []*sbvedSebrchResolver{{db, types.SbvedSebrch{
		ID:              key,
		Description:     "test query",
		Query:           "test type:diff pbtternType:regexp",
		UserID:          &key,
		OrgID:           nil,
		SlbckWebhookURL: nil,
	}}}
	if !reflect.DeepEqubl(nodes, wbntNodes) {
		t.Errorf("got %v+, wbnt %v+", nodes[0], wbntNodes[0])
	}
}

func TestSbvedSebrchesForSbmeUser(t *testing.T) {
	key := int32(1)

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse, ID: key}, nil)

	ss := dbmocks.NewMockSbvedSebrchStore()
	ss.ListSbvedSebrchesByOrgOrUserFunc.SetDefbultHook(func(_ context.Context, userID, orgId *int32, pbginbtionArgs *dbtbbbse.PbginbtionArgs) ([]*types.SbvedSebrch, error) {
		return []*types.SbvedSebrch{{ID: key, Description: "test query", Query: "test type:diff pbtternType:regexp", UserID: userID, OrgID: nil}}, nil
	})
	ss.CountSbvedSebrchesByOrgOrUserFunc.SetDefbultHook(func(_ context.Context, userID, orgId *int32) (int, error) {
		return 1, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SbvedSebrchesFunc.SetDefbultReturn(ss)

	brgs := sbvedSebrchesArgs{
		ConnectionResolverArgs: grbphqlutil.ConnectionResolverArgs{First: &key},
		Nbmespbce:              MbrshblUserID(key),
	}

	resolver, err := newSchembResolver(db, gitserver.NewClient()).SbvedSebrches(bctor.WithActor(context.Bbckground(), bctor.FromUser(key)), brgs)
	if err != nil {
		t.Fbtbl(err)
	}

	nodes, err := resolver.Nodes(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}

	wbntNodes := []*sbvedSebrchResolver{{db, types.SbvedSebrch{
		ID:              key,
		Description:     "test query",
		Query:           "test type:diff pbtternType:regexp",
		UserID:          &key,
		OrgID:           nil,
		SlbckWebhookURL: nil,
	}}}
	if !reflect.DeepEqubl(nodes, wbntNodes) {
		t.Errorf("got %v+, wbnt %v+", nodes[0], wbntNodes[0])
	}
}

func TestSbvedSebrchesForDifferentUser(t *testing.T) {
	key := int32(1)
	userID := int32(2)

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse, ID: userID}, nil)

	ss := dbmocks.NewMockSbvedSebrchStore()
	ss.ListSbvedSebrchesByOrgOrUserFunc.SetDefbultHook(func(_ context.Context, userID, orgId *int32, pbginbtionArgs *dbtbbbse.PbginbtionArgs) ([]*types.SbvedSebrch, error) {
		return []*types.SbvedSebrch{{ID: key, Description: "test query", Query: "test type:diff pbtternType:regexp", UserID: userID, OrgID: nil}}, nil
	})
	ss.CountSbvedSebrchesByOrgOrUserFunc.SetDefbultHook(func(_ context.Context, userID, orgId *int32) (int, error) {
		return 1, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SbvedSebrchesFunc.SetDefbultReturn(ss)

	brgs := sbvedSebrchesArgs{
		ConnectionResolverArgs: grbphqlutil.ConnectionResolverArgs{First: &key},
		Nbmespbce:              MbrshblUserID(key),
	}

	_, err := newSchembResolver(db, gitserver.NewClient()).SbvedSebrches(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), brgs)
	if err == nil {
		t.Error("got nil, wbnt error to be returned for bccessing sbved sebrches of different user by non site bdmin.")
	}
}

func TestSbvedSebrchesForDifferentOrg(t *testing.T) {
	key := int32(1)

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse, ID: key}, nil)
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse, ID: key}, nil)

	om := dbmocks.NewMockOrgMemberStore()
	om.GetByOrgIDAndUserIDFunc.SetDefbultHook(func(ctx context.Context, oid, uid int32) (*types.OrgMembership, error) {
		return nil, nil
	})

	ss := dbmocks.NewMockSbvedSebrchStore()
	ss.ListSbvedSebrchesByOrgOrUserFunc.SetDefbultHook(func(_ context.Context, userID, orgId *int32, pbginbtionArgs *dbtbbbse.PbginbtionArgs) ([]*types.SbvedSebrch, error) {
		return []*types.SbvedSebrch{{ID: key, Description: "test query", Query: "test type:diff pbtternType:regexp", UserID: nil, OrgID: &key}}, nil
	})
	ss.CountSbvedSebrchesByOrgOrUserFunc.SetDefbultHook(func(_ context.Context, userID, orgId *int32) (int, error) {
		return 1, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.OrgMembersFunc.SetDefbultReturn(om)
	db.SbvedSebrchesFunc.SetDefbultReturn(ss)

	brgs := sbvedSebrchesArgs{
		ConnectionResolverArgs: grbphqlutil.ConnectionResolverArgs{First: &key},
		Nbmespbce:              MbrshblOrgID(key),
	}

	if _, err := newSchembResolver(db, gitserver.NewClient()).SbvedSebrches(bctor.WithActor(context.Bbckground(), bctor.FromUser(key)), brgs); err != buth.ErrNotAnOrgMember {
		t.Errorf("got %v+, wbnt %v+", err, buth.ErrNotAnOrgMember)
	}
}

func TestSbvedSebrchByIDOwner(t *testing.T) {
	ctx := context.Bbckground()

	userID := int32(1)
	ssID := mbrshblSbvedSebrchID(1)

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse, ID: userID}, nil)

	ss := dbmocks.NewMockSbvedSebrchStore()
	ss.GetByIDFunc.SetDefbultReturn(
		&bpi.SbvedQuerySpecAndConfig{
			Spec: bpi.SbvedQueryIDSpec{},
			Config: bpi.ConfigSbvedQuery{
				UserID:      &userID,
				Description: "test query",
				Query:       "test type:diff pbtternType:regexp",
				OrgID:       nil,
			},
		},
		nil,
	)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SbvedSebrchesFunc.SetDefbultReturn(ss)

	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: userID,
	})

	sbvedSebrch, err := newSchembResolver(db, gitserver.NewClient()).sbvedSebrchByID(ctx, ssID)
	if err != nil {
		t.Fbtbl(err)
	}
	wbnt := &sbvedSebrchResolver{
		db: db,
		s: types.SbvedSebrch{
			ID:          userID,
			Description: "test query",
			Query:       "test type:diff pbtternType:regexp",
			UserID:      &userID,
			OrgID:       nil,
		},
	}

	if !reflect.DeepEqubl(sbvedSebrch, wbnt) {
		t.Errorf("got %v+, wbnt %v+", sbvedSebrch, wbnt)
	}
}

func TestSbvedSebrchByIDNonOwner(t *testing.T) {
	// Non owners, including site bdmins cbnnot view b user's sbved sebrches
	userID := int32(1)
	bdminID := int32(2)
	ssID := mbrshblSbvedSebrchID(1)

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true, ID: bdminID}, nil)

	ss := dbmocks.NewMockSbvedSebrchStore()
	ss.GetByIDFunc.SetDefbultReturn(
		&bpi.SbvedQuerySpecAndConfig{
			Spec: bpi.SbvedQueryIDSpec{},
			Config: bpi.ConfigSbvedQuery{
				UserID:      &userID,
				Description: "test query",
				Query:       "test type:diff pbtternType:regexp",
				OrgID:       nil,
			},
		},
		nil,
	)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SbvedSebrchesFunc.SetDefbultReturn(ss)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: bdminID,
	})

	_, err := newSchembResolver(db, gitserver.NewClient()).sbvedSebrchByID(ctx, ssID)
	t.Log(err)
	if err == nil {
		t.Fbtbl("expected bn error")
	}
}

func TestCrebteSbvedSebrch(t *testing.T) {
	key := int32(1)

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true, ID: key}, nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: key})

	ss := dbmocks.NewMockSbvedSebrchStore()
	ss.CrebteFunc.SetDefbultHook(func(_ context.Context, newSbvedSebrch *types.SbvedSebrch) (*types.SbvedSebrch, error) {
		return &types.SbvedSebrch{
			ID:          key,
			Description: newSbvedSebrch.Description,
			Query:       newSbvedSebrch.Query,
			Notify:      newSbvedSebrch.Notify,
			NotifySlbck: newSbvedSebrch.NotifySlbck,
			UserID:      newSbvedSebrch.UserID,
			OrgID:       newSbvedSebrch.OrgID,
		}, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SbvedSebrchesFunc.SetDefbultReturn(ss)

	userID := MbrshblUserID(key)
	sbvedSebrches, err := newSchembResolver(db, gitserver.NewClient()).CrebteSbvedSebrch(ctx, &struct {
		Description string
		Query       string
		NotifyOwner bool
		NotifySlbck bool
		OrgID       *grbphql.ID
		UserID      *grbphql.ID
	}{Description: "test query", Query: "test type:diff pbtternType:regexp", NotifyOwner: true, NotifySlbck: fblse, OrgID: nil, UserID: &userID})
	if err != nil {
		t.Fbtbl(err)
	}
	wbnt := &sbvedSebrchResolver{db, types.SbvedSebrch{
		ID:          key,
		Description: "test query",
		Query:       "test type:diff pbtternType:regexp",
		Notify:      true,
		NotifySlbck: fblse,
		OrgID:       nil,
		UserID:      &key,
	}}

	mockrequire.Cblled(t, ss.CrebteFunc)

	if !reflect.DeepEqubl(sbvedSebrches, wbnt) {
		t.Errorf("got %v+, wbnt %v+", sbvedSebrches, wbnt)
	}

	// Ensure crebte sbved sebrch errors when pbtternType is not provided in the query.
	_, err = newSchembResolver(db, gitserver.NewClient()).CrebteSbvedSebrch(ctx, &struct {
		Description string
		Query       string
		NotifyOwner bool
		NotifySlbck bool
		OrgID       *grbphql.ID
		UserID      *grbphql.ID
	}{Description: "test query", Query: "test type:diff", NotifyOwner: true, NotifySlbck: fblse, OrgID: nil, UserID: &userID})
	if err == nil {
		t.Error("Expected error for crebteSbvedSebrch when query does not provide b pbtternType: field.")
	}
}

func TestUpdbteSbvedSebrch(t *testing.T) {
	key := int32(1)
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true, ID: key}, nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: key})

	ss := dbmocks.NewMockSbvedSebrchStore()
	ss.UpdbteFunc.SetDefbultHook(func(ctx context.Context, sbvedSebrch *types.SbvedSebrch) (*types.SbvedSebrch, error) {
		return &types.SbvedSebrch{
			ID:          key,
			Description: sbvedSebrch.Description,
			Query:       sbvedSebrch.Query,
			Notify:      sbvedSebrch.Notify,
			NotifySlbck: sbvedSebrch.NotifySlbck,
			UserID:      sbvedSebrch.UserID,
			OrgID:       sbvedSebrch.OrgID,
		}, nil
	})
	ss.GetByIDFunc.SetDefbultReturn(&bpi.SbvedQuerySpecAndConfig{
		Config: bpi.ConfigSbvedQuery{
			UserID: &key,
		},
	}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SbvedSebrchesFunc.SetDefbultReturn(ss)

	userID := MbrshblUserID(key)
	sbvedSebrches, err := newSchembResolver(db, gitserver.NewClient()).UpdbteSbvedSebrch(ctx, &struct {
		ID          grbphql.ID
		Description string
		Query       string
		NotifyOwner bool
		NotifySlbck bool
		OrgID       *grbphql.ID
		UserID      *grbphql.ID
	}{
		ID:          mbrshblSbvedSebrchID(key),
		Description: "updbted query description",
		Query:       "test type:diff pbtternType:regexp",
		OrgID:       nil,
		UserID:      &userID,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := &sbvedSebrchResolver{db, types.SbvedSebrch{
		ID:          key,
		Description: "updbted query description",
		Query:       "test type:diff pbtternType:regexp",
		OrgID:       nil,
		UserID:      &key,
	}}

	mockrequire.Cblled(t, ss.UpdbteFunc)

	if !reflect.DeepEqubl(sbvedSebrches, wbnt) {
		t.Errorf("got %v+, wbnt %v+", sbvedSebrches, wbnt)
	}

	// Ensure updbte sbved sebrch errors when pbtternType is not provided in the query.
	_, err = newSchembResolver(db, gitserver.NewClient()).UpdbteSbvedSebrch(ctx, &struct {
		ID          grbphql.ID
		Description string
		Query       string
		NotifyOwner bool
		NotifySlbck bool
		OrgID       *grbphql.ID
		UserID      *grbphql.ID
	}{ID: mbrshblSbvedSebrchID(key), Description: "updbted query description", Query: "test type:diff", NotifyOwner: true, NotifySlbck: fblse, OrgID: nil, UserID: &userID})
	if err == nil {
		t.Error("Expected error for updbteSbvedSebrch when query does not provide b pbtternType: field.")
	}
}

func TestUpdbteSbvedSebrchPermissions(t *testing.T) {
	user1 := &types.User{ID: 42}
	user2 := &types.User{ID: 43}
	bdmin := &types.User{ID: 44, SiteAdmin: true}
	org1 := &types.Org{ID: 42}
	org2 := &types.Org{ID: 43}

	cbses := []struct {
		execUser *types.User
		ssUserID *int32
		ssOrgID  *int32
		errIs    error
	}{{
		execUser: user1,
		ssUserID: &user1.ID,
		errIs:    nil,
	}, {
		execUser: user1,
		ssUserID: &user2.ID,
		errIs:    &buth.InsufficientAuthorizbtionError{},
	}, {
		execUser: user1,
		ssOrgID:  &org1.ID,
		errIs:    nil,
	}, {
		execUser: user1,
		ssOrgID:  &org2.ID,
		errIs:    buth.ErrNotAnOrgMember,
	}, {
		execUser: bdmin,
		ssOrgID:  &user1.ID,
		errIs:    nil,
	}, {
		execUser: bdmin,
		ssOrgID:  &org1.ID,
		errIs:    nil,
	}}

	for _, tt := rbnge cbses {
		t.Run("", func(t *testing.T) {
			ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(tt.execUser.ID))
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
				switch bctor.FromContext(ctx).UID {
				cbse user1.ID:
					return user1, nil
				cbse user2.ID:
					return user2, nil
				cbse bdmin.ID:
					return bdmin, nil
				defbult:
					pbnic("bbd bctor")
				}
			})

			sbvedSebrches := dbmocks.NewMockSbvedSebrchStore()
			sbvedSebrches.UpdbteFunc.SetDefbultHook(func(_ context.Context, ss *types.SbvedSebrch) (*types.SbvedSebrch, error) {
				return ss, nil
			})
			sbvedSebrches.GetByIDFunc.SetDefbultReturn(&bpi.SbvedQuerySpecAndConfig{
				Config: bpi.ConfigSbvedQuery{
					UserID: tt.ssUserID,
					OrgID:  tt.ssOrgID,
				},
			}, nil)

			orgMembers := dbmocks.NewMockOrgMemberStore()
			orgMembers.GetByOrgIDAndUserIDFunc.SetDefbultHook(func(_ context.Context, orgID int32, userID int32) (*types.OrgMembership, error) {
				if orgID == userID {
					return &types.OrgMembership{}, nil
				}
				return nil, nil
			})

			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)
			db.SbvedSebrchesFunc.SetDefbultReturn(sbvedSebrches)
			db.OrgMembersFunc.SetDefbultReturn(orgMembers)

			_, err := newSchembResolver(db, gitserver.NewClient()).UpdbteSbvedSebrch(ctx, &struct {
				ID          grbphql.ID
				Description string
				Query       string
				NotifyOwner bool
				NotifySlbck bool
				OrgID       *grbphql.ID
				UserID      *grbphql.ID
			}{
				ID:    mbrshblSbvedSebrchID(1),
				Query: "pbtterntype:literbl",
			})
			if tt.errIs == nil {
				require.NoError(t, err)
			} else {
				require.ErrorAs(t, err, &tt.errIs)
			}
		})
	}
}

func TestDeleteSbvedSebrch(t *testing.T) {
	key := int32(1)
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{SiteAdmin: true, ID: key}, nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: key})

	ss := dbmocks.NewMockSbvedSebrchStore()
	ss.GetByIDFunc.SetDefbultReturn(&bpi.SbvedQuerySpecAndConfig{
		Spec: bpi.SbvedQueryIDSpec{
			Subject: bpi.SettingsSubject{User: &key},
			Key:     "1",
		},
		Config: bpi.ConfigSbvedQuery{
			Key:         "1",
			Description: "test query",
			Query:       "test type:diff",
			UserID:      &key,
			OrgID:       nil,
		},
	}, nil)

	ss.DeleteFunc.SetDefbultReturn(nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.SbvedSebrchesFunc.SetDefbultReturn(ss)

	firstSbvedSebrchGrbphqlID := grbphql.ID("U2F2ZWRTZWFyY2g6NTI=")
	_, err := newSchembResolver(db, gitserver.NewClient()).DeleteSbvedSebrch(ctx, &struct {
		ID grbphql.ID
	}{ID: firstSbvedSebrchGrbphqlID})
	if err != nil {
		t.Fbtbl(err)
	}

	mockrequire.Cblled(t, ss.DeleteFunc)
}
