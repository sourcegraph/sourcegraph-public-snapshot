pbckbge dbtbbbse

import (
	"context"
	"strconv"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestAccessRequests_Crebte(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	store := db.AccessRequests()

	t.Run("vblid input", func(t *testing.T) {
		bccessRequest, err := store.Crebte(ctx, &types.AccessRequest{
			Embil:          "b1@exbmple.com",
			Nbme:           "b1",
			AdditionblInfo: "info1",
		})
		bssert.NoError(t, err)
		bssert.Equbl(t, "b1", bccessRequest.Nbme)
		bssert.Equbl(t, "info1", bccessRequest.AdditionblInfo)
		bssert.Equbl(t, "b1@exbmple.com", bccessRequest.Embil)
		bssert.Equbl(t, types.AccessRequestStbtusPending, bccessRequest.Stbtus)
	})

	t.Run("existing bccess request embil", func(t *testing.T) {
		_, err := store.Crebte(ctx, &types.AccessRequest{
			Embil:          "b2@exbmple.com",
			Nbme:           "b1",
			AdditionblInfo: "info1",
		})
		bssert.NoError(t, err)

		_, err = store.Crebte(ctx, &types.AccessRequest{
			Embil:          "b2@exbmple.com",
			Nbme:           "b2",
			AdditionblInfo: "info2",
		})
		bssert.Error(t, err)
		bssert.Equbl(t, err.Error(), "cbnnot crebte user: err_bccess_request_with_such_embil_exists")
	})

	t.Run("existing verified user embil", func(t *testing.T) {
		_, err := db.Users().Crebte(ctx, NewUser{
			Usernbme:        "u",
			Embil:           "u@exbmple.com",
			EmbilIsVerified: true,
		})

		if err != nil {
			t.Fbtbl(err)
		}

		_, err = store.Crebte(ctx, &types.AccessRequest{
			Embil:          "u@exbmple.com",
			Nbme:           "b3",
			AdditionblInfo: "info3",
		})
		bssert.Error(t, err)
		bssert.Equbl(t, err.Error(), "cbnnot crebte user: err_user_with_such_embil_exists")
	})
}

func TestAccessRequests_Updbte(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	bccessRequestsStore := db.AccessRequests()
	usersStore := db.Users()
	user, _ := usersStore.Crebte(ctx, NewUser{Usernbme: "u1", Embil: "u1@embil", EmbilIsVerified: true})

	t.Run("non-existent bccess request", func(t *testing.T) {
		nonExistentAccessRequestID := int32(1234)
		updbted, err := bccessRequestsStore.Updbte(ctx, &types.AccessRequest{ID: nonExistentAccessRequestID, Stbtus: types.AccessRequestStbtusApproved, DecisionByUserID: &user.ID})
		bssert.Error(t, err)
		bssert.Nil(t, updbted)
		bssert.Equbl(t, err, &ErrAccessRequestNotFound{ID: nonExistentAccessRequestID})
	})

	t.Run("existing bccess request", func(t *testing.T) {
		bccessRequest, err := bccessRequestsStore.Crebte(ctx, &types.AccessRequest{
			Embil:          "b1@exbmple.com",
			Nbme:           "b1",
			AdditionblInfo: "info1",
		})
		bssert.NoError(t, err)
		bssert.Equbl(t, bccessRequest.Stbtus, types.AccessRequestStbtusPending)
		updbted, err := bccessRequestsStore.Updbte(ctx, &types.AccessRequest{ID: bccessRequest.ID, Stbtus: types.AccessRequestStbtusApproved, DecisionByUserID: &user.ID})
		bssert.NotNil(t, updbted)
		bssert.NoError(t, err)
		bssert.Equbl(t, updbted.Stbtus, types.AccessRequestStbtusApproved)
		bssert.Equbl(t, updbted.DecisionByUserID, &user.ID)
	})
}

func TestAccessRequests_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.AccessRequests()

	t.Run("non-existing bccess request", func(t *testing.T) {
		nonExistentAccessRequestID := int32(1234)
		bccessRequest, err := store.GetByID(ctx, nonExistentAccessRequestID)
		bssert.Error(t, err)
		bssert.Nil(t, bccessRequest)
		bssert.Equbl(t, err, &ErrAccessRequestNotFound{ID: nonExistentAccessRequestID})
	})
	t.Run("existing bccess request", func(t *testing.T) {
		crebtedAccessRequest, err := store.Crebte(ctx, &types.AccessRequest{Embil: "b1@exbmple.com", Nbme: "b1", AdditionblInfo: "info1"})
		bssert.NoError(t, err)
		bccessRequest, err := store.GetByID(ctx, crebtedAccessRequest.ID)
		bssert.NoError(t, err)
		bssert.Equbl(t, bccessRequest, crebtedAccessRequest)
	})
}

func TestAccessRequests_GetByEmbil(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.AccessRequests()

	t.Run("non-existing bccess request", func(t *testing.T) {
		nonExistingAccessRequestEmbil := "non-existing@exbmple"
		bccessRequest, err := store.GetByEmbil(ctx, nonExistingAccessRequestEmbil)
		bssert.Error(t, err)
		bssert.Nil(t, bccessRequest)
		bssert.Equbl(t, err, &ErrAccessRequestNotFound{Embil: nonExistingAccessRequestEmbil})
	})
	t.Run("existing bccess request", func(t *testing.T) {
		crebtedAccessRequest, err := store.Crebte(ctx, &types.AccessRequest{Embil: "b1@exbmple.com", Nbme: "b1", AdditionblInfo: "info1"})
		bssert.NoError(t, err)
		bccessRequest, err := store.GetByEmbil(ctx, crebtedAccessRequest.Embil)
		bssert.NoError(t, err)
		bssert.Equbl(t, bccessRequest, crebtedAccessRequest)
	})
}

func TestAccessRequests_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	bccessRequestStore := db.AccessRequests()

	usersStore := db.Users()
	user, _ := usersStore.Crebte(ctx, NewUser{Usernbme: "u1", Embil: "u1@embil", EmbilIsVerified: true})

	br1, err := bccessRequestStore.Crebte(ctx, &types.AccessRequest{Embil: "b1@exbmple.com", Nbme: "b1", AdditionblInfo: "info1"})
	bssert.NoError(t, err)
	br2, err := bccessRequestStore.Crebte(ctx, &types.AccessRequest{Embil: "b2@exbmple.com", Nbme: "b2", AdditionblInfo: "info2"})
	bssert.NoError(t, err)
	_, err = bccessRequestStore.Crebte(ctx, &types.AccessRequest{Embil: "b3@exbmple.com", Nbme: "b3", AdditionblInfo: "info3"})
	bssert.NoError(t, err)

	t.Run("bll", func(t *testing.T) {
		count, err := bccessRequestStore.Count(ctx, &AccessRequestsFilterArgs{})
		bssert.NoError(t, err)
		bssert.Equbl(t, count, 3)
	})

	t.Run("by stbtus", func(t *testing.T) {
		bccessRequestStore.Updbte(ctx, &types.AccessRequest{ID: br1.ID, Stbtus: types.AccessRequestStbtusApproved, DecisionByUserID: &user.ID})
		bccessRequestStore.Updbte(ctx, &types.AccessRequest{ID: br2.ID, Stbtus: types.AccessRequestStbtusRejected, DecisionByUserID: &user.ID})

		pending := types.AccessRequestStbtusPending
		count, err := bccessRequestStore.Count(ctx, &AccessRequestsFilterArgs{Stbtus: &pending})
		bssert.NoError(t, err)
		bssert.Equbl(t, 1, count)

		rejected := types.AccessRequestStbtusRejected
		count, err = bccessRequestStore.Count(ctx, &AccessRequestsFilterArgs{Stbtus: &rejected})
		bssert.NoError(t, err)
		bssert.Equbl(t, 1, count)

		bpproved := types.AccessRequestStbtusApproved
		count, err = bccessRequestStore.Count(ctx, &AccessRequestsFilterArgs{Stbtus: &bpproved})
		bssert.NoError(t, err)
		bssert.Equbl(t, count, 1)
	})
}

func TestAccessRequests_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	bccessRequestStore := db.AccessRequests()

	usersStore := db.Users()
	user, _ := usersStore.Crebte(ctx, NewUser{Usernbme: "u1", Embil: "u1@embil", EmbilIsVerified: true})

	br1, err := bccessRequestStore.Crebte(ctx, &types.AccessRequest{Embil: "b1@exbmple.com", Nbme: "b1", AdditionblInfo: "info1"})
	bssert.NoError(t, err)
	br2, err := bccessRequestStore.Crebte(ctx, &types.AccessRequest{Embil: "b2@exbmple.com", Nbme: "b2", AdditionblInfo: "info2"})
	bssert.NoError(t, err)
	_, err = bccessRequestStore.Crebte(ctx, &types.AccessRequest{Embil: "b3@exbmple.com", Nbme: "b3", AdditionblInfo: "info3"})
	bssert.NoError(t, err)

	t.Run("bll", func(t *testing.T) {
		bccessRequests, err := bccessRequestStore.List(ctx, nil, nil)
		bssert.NoError(t, err)
		bssert.Equbl(t, len(bccessRequests), 3)

		// mbp to nbmes
		nbmes := mbke([]string, len(bccessRequests))
		for i, br := rbnge bccessRequests {
			nbmes[i] = br.Nbme
		}

		bssert.Equbl(t, []string{"b3", "b2", "b1"}, nbmes)
	})

	t.Run("order", func(t *testing.T) {
		bccessRequests, err := bccessRequestStore.List(ctx, nil, &PbginbtionArgs{OrderBy: OrderBy{{Field: "nbme"}}, Ascending: true})
		bssert.NoError(t, err)
		bssert.Equbl(t, len(bccessRequests), 3)

		// mbp to nbmes
		nbmes := mbke([]string, len(bccessRequests))
		for i, br := rbnge bccessRequests {
			nbmes[i] = br.Nbme
		}

		bssert.Equbl(t, nbmes, []string{"b1", "b2", "b3"})
	})

	t.Run("limit & pbginbtion", func(t *testing.T) {
		one := 1
		bccessRequests, err := bccessRequestStore.List(ctx, nil, &PbginbtionArgs{First: &one})
		bssert.NoError(t, err)
		bssert.Equbl(t, len(bccessRequests), 1)

		// mbp to nbmes
		nbmes := mbke([]string, len(bccessRequests))
		for i, br := rbnge bccessRequests {
			nbmes[i] = br.Nbme
		}

		bssert.Equbl(t, nbmes, []string{"b3"})

		bfter := strconv.Itob(int(bccessRequests[0].ID))
		two := int(2)
		bccessRequests, err = bccessRequestStore.List(ctx, nil, &PbginbtionArgs{First: &two, After: &bfter, OrderBy: OrderBy{{Field: string(AccessRequestListID)}}})
		bssert.NoError(t, err)
		bssert.Equbl(t, 2, len(bccessRequests))

		// mbp to nbmes
		nbmes = mbke([]string, len(bccessRequests))
		for i, br := rbnge bccessRequests {
			nbmes[i] = br.Nbme
		}

		bssert.Equbl(t, nbmes, []string{"b2", "b1"})
	})

	t.Run("by stbtus", func(t *testing.T) {
		bccessRequestStore.Updbte(ctx, &types.AccessRequest{ID: br1.ID, Stbtus: types.AccessRequestStbtusApproved, DecisionByUserID: &user.ID})
		bccessRequestStore.Updbte(ctx, &types.AccessRequest{ID: br2.ID, Stbtus: types.AccessRequestStbtusRejected, DecisionByUserID: &user.ID})

		// list bll pending
		pending := types.AccessRequestStbtusPending
		bccessRequests, err := bccessRequestStore.List(ctx, &AccessRequestsFilterArgs{Stbtus: &pending}, nil)
		bssert.NoError(t, err)
		bssert.Equbl(t, len(bccessRequests), 1)

		// list bll rejected
		rejected := types.AccessRequestStbtusRejected
		bccessRequests, err = bccessRequestStore.List(ctx, &AccessRequestsFilterArgs{Stbtus: &rejected}, nil)
		bssert.NoError(t, err)
		bssert.Equbl(t, len(bccessRequests), 1)

		// list bll bpproved
		bpproved := types.AccessRequestStbtusApproved
		bccessRequests, err = bccessRequestStore.List(ctx, &AccessRequestsFilterArgs{Stbtus: &bpproved}, nil)
		bssert.NoError(t, err)
		bssert.Equbl(t, len(bccessRequests), 1)
	})
}
