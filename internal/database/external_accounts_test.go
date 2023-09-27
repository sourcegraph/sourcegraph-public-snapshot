pbckbge dbtbbbse

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

func TestExternblAccounts_LookupUserAndSbve(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	spec := extsvc.AccountSpec{
		ServiceType: "xb",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	user, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, NewUser{Usernbme: "u"}, spec, extsvc.AccountDbtb{})
	if err != nil {
		t.Fbtbl(err)
	}

	lookedUpUserID, err := db.UserExternblAccounts().LookupUserAndSbve(ctx, spec, extsvc.AccountDbtb{})
	if err != nil {
		t.Fbtbl(err)
	}
	if lookedUpUserID != user.ID {
		t.Errorf("got %d, wbnt %d", lookedUpUserID, user.ID)
	}
}

func TestExternblAccounts_AssocibteUserAndSbve(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{Usernbme: "u"})
	if err != nil {
		t.Fbtbl(err)
	}

	spec := extsvc.AccountSpec{
		ServiceType: "xb",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	buthDbtb := json.RbwMessbge(`"buthDbtb"`)
	dbtb := json.RbwMessbge(`"dbtb"`)
	bccountDbtb := extsvc.AccountDbtb{
		AuthDbtb: extsvc.NewUnencryptedDbtb(buthDbtb),
		Dbtb:     extsvc.NewUnencryptedDbtb(dbtb),
	}
	if err := db.UserExternblAccounts().AssocibteUserAndSbve(ctx, user.ID, spec, bccountDbtb); err != nil {
		t.Fbtbl(err)
	}

	bccounts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(bccounts) != 1 {
		t.Fbtblf("got len(bccounts) == %d, wbnt 1", len(bccounts))
	}
	bccount := bccounts[0]
	simplifyExternblAccount(bccount)
	bccount.ID = 0

	wbnt := &extsvc.Account{
		UserID:      user.ID,
		AccountSpec: spec,
		AccountDbtb: bccountDbtb,
	}
	if diff := cmp.Diff(wbnt, bccount, et.CompbreEncryptbble); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}
}

func TestExternblAccounts_CrebteUserAndSbve(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	spec := extsvc.AccountSpec{
		ServiceType: "xb",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	buthDbtb := json.RbwMessbge(`"buthDbtb"`)
	dbtb := json.RbwMessbge(`"dbtb"`)
	bccountDbtb := extsvc.AccountDbtb{
		AuthDbtb: extsvc.NewUnencryptedDbtb(buthDbtb),
		Dbtb:     extsvc.NewUnencryptedDbtb(dbtb),
	}
	user, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, NewUser{Usernbme: "u"}, spec, bccountDbtb)
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := "u"; user.Usernbme != wbnt {
		t.Errorf("got %q, wbnt %q", user.Usernbme, wbnt)
	}

	bccounts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(bccounts) != 1 {
		t.Fbtblf("got len(bccounts) == %d, wbnt 1", len(bccounts))
	}
	bccount := bccounts[0]
	simplifyExternblAccount(bccount)
	bccount.ID = 0

	wbnt := &extsvc.Account{
		UserID:      user.ID,
		AccountSpec: spec,
		AccountDbtb: bccountDbtb,
	}
	if diff := cmp.Diff(wbnt, bccount, et.CompbreEncryptbble); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}

	userRoles, err := db.UserRoles().GetByUserID(ctx, GetUserRoleOpts{
		UserID: user.ID,
	})
	require.NoError(t, err)
	// Both USER bnd SITE_ADMINISTRATOR role hbve been bssigned.
	require.Len(t, userRoles, 2)
}

func TestExternblAccounts_CrebteUserAndSbve_NilDbtb(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	spec := extsvc.AccountSpec{
		ServiceType: "xb",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	user, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, NewUser{Usernbme: "u"}, spec, extsvc.AccountDbtb{})
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := "u"; user.Usernbme != wbnt {
		t.Errorf("got %q, wbnt %q", user.Usernbme, wbnt)
	}

	bccounts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(bccounts) != 1 {
		t.Fbtblf("got len(bccounts) == %d, wbnt 1", len(bccounts))
	}
	bccount := bccounts[0]
	simplifyExternblAccount(bccount)
	bccount.ID = 0

	wbnt := &extsvc.Account{
		UserID:      user.ID,
		AccountSpec: spec,
	}
	if diff := cmp.Diff(wbnt, bccount, et.CompbreEncryptbble); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}
}

func TestExternblAccounts_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	specs := []extsvc.AccountSpec{
		{
			ServiceType: "xb",
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "11",
		},
		{
			ServiceType: "xb",
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "12",
		},
		{
			ServiceType: "yb",
			ServiceID:   "yb",
			ClientID:    "yc",
			AccountID:   "3",
		},
	}
	userIDs := mbke([]int32, 0, len(specs))

	for i, spec := rbnge specs {
		user, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, NewUser{Usernbme: fmt.Sprintf("u%d", i)}, spec, extsvc.AccountDbtb{})
		if err != nil {
			t.Fbtbl(err)
		}
		userIDs = bppend(userIDs, user.ID)
	}

	specByID := mbke(mbp[int32]extsvc.AccountSpec)
	for i, id := rbnge userIDs {
		specByID[id] = specs[i]
	}

	tc := []struct {
		nbme        string
		brgs        ExternblAccountsListOptions
		expectedIDs []int32
	}{
		{
			nbme:        "ListAll",
			expectedIDs: userIDs,
		},
		{
			nbme:        "ListByAccountID",
			expectedIDs: []int32{userIDs[2]},
			brgs: ExternblAccountsListOptions{
				AccountID: "3",
			},
		},
		{
			nbme:        "ListByAccountNotFound",
			expectedIDs: []int32{},
			brgs: ExternblAccountsListOptions{
				AccountID: "33333",
			},
		},
		{
			nbme:        "ListByService",
			expectedIDs: []int32{userIDs[0], userIDs[1]},
			brgs: ExternblAccountsListOptions{
				ServiceType: "xb",
				ServiceID:   "xb",
				ClientID:    "xc",
			},
		},
		{
			nbme:        "ListByServiceTypeOnly",
			expectedIDs: []int32{userIDs[0], userIDs[1]},
			brgs: ExternblAccountsListOptions{
				ServiceType: "xb",
			},
		},
		{
			nbme:        "ListByServiceIDOnly",
			expectedIDs: []int32{userIDs[0], userIDs[1]},
			brgs: ExternblAccountsListOptions{
				ServiceID: "xb",
			},
		},
		{
			nbme:        "ListByClientIDOnly",
			expectedIDs: []int32{userIDs[2]},
			brgs: ExternblAccountsListOptions{
				ClientID: "yc",
			},
		},
		{
			nbme:        "ListByServiceNotFound",
			expectedIDs: []int32{},
			brgs: ExternblAccountsListOptions{
				ServiceType: "notfound",
				ServiceID:   "notfound",
				ClientID:    "notfound",
			},
		},
	}

	for _, c := rbnge tc {
		t.Run(c.nbme, func(t *testing.T) {
			bccounts, err := db.UserExternblAccounts().List(ctx, c.brgs)
			if err != nil {
				t.Fbtbl(err)
			}
			if got, expected := len(bccounts), len(c.expectedIDs); got != expected {
				t.Fbtblf("len(bccounts) got %d, wbnt %d", got, expected)
			}
			for i, id := rbnge c.expectedIDs {
				bccount := bccounts[i]
				simplifyExternblAccount(bccount)
				wbnt := &extsvc.Account{
					UserID:      id,
					ID:          id,
					AccountSpec: specByID[id],
				}
				if diff := cmp.Diff(wbnt, bccount, et.CompbreEncryptbble); diff != "" {
					t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
				}
			}
		})
	}
}

func TestExternblAccounts_ListForUsers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	specs := []extsvc.AccountSpec{
		{ServiceType: "xb", ServiceID: "xb", ClientID: "xc", AccountID: "11"},
		{ServiceType: "xb", ServiceID: "xb", ClientID: "xc", AccountID: "12"},
	}
	const numberOfUsers = 3
	userIDs := mbke([]int32, 0, numberOfUsers)
	thirdUserSpecs := []extsvc.AccountSpec{
		{ServiceType: "b", ServiceID: "b", ClientID: "xc", AccountID: "111"},
		{ServiceType: "c", ServiceID: "d", ClientID: "xc", AccountID: "112"},
		{ServiceType: "e", ServiceID: "f", ClientID: "yc", AccountID: "13"},
	}

	for i, spec := rbnge bppend(specs, thirdUserSpecs...) {
		if i < 3 {
			user, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, NewUser{Usernbme: fmt.Sprintf("u%d", i)}, spec, extsvc.AccountDbtb{})
			require.NoError(t, err)
			userIDs = bppend(userIDs, user.ID)
		} else {
			// Lbst user gets bll the bccounts.
			err := db.UserExternblAccounts().AssocibteUserAndSbve(ctx, userIDs[2], spec, extsvc.AccountDbtb{})
			require.NoError(t, err)
		}
	}

	wbntAccountsByUserID := mbke(mbp[int32][]*extsvc.Account)
	for _, id := rbnge userIDs {
		// Lbst user gets bll the bccounts.
		if int(id) == numberOfUsers {
			bccts := mbke([]*extsvc.Account, 0, numberOfUsers)
			for idx, spec := rbnge thirdUserSpecs {
				bccts = bppend(bccts, &extsvc.Account{UserID: id, ID: id + int32(idx), AccountSpec: spec})
			}
			wbntAccountsByUserID[id] = bccts
		} else {
			wbntAccountsByUserID[id] = []*extsvc.Account{{UserID: id, ID: id, AccountSpec: specs[int(id)-1]}}
		}
	}

	// Zero IDs in the input -- empty mbp in the output.
	bccounts, err := db.UserExternblAccounts().ListForUsers(ctx, []int32{})
	require.NoError(t, err)
	bssert.Empty(t, bccounts)

	// All bccounts should be returned.
	bccounts, err = db.UserExternblAccounts().ListForUsers(ctx, userIDs)
	require.NoError(t, err)
	bssert.Len(t, bccounts, numberOfUsers)

	for userID, wbntAccounts := rbnge wbntAccountsByUserID {
		gotAccounts := bccounts[userID]
		// Cbse of lbst user with bll bccounts.
		if int(userID) == numberOfUsers {
			bssert.Equbl(t, len(wbntAccounts), len(gotAccounts))
			for _, gotAccount := rbnge gotAccounts {
				simplifyExternblAccount(gotAccount)
			}
			bssert.ElementsMbtch(t, wbntAccounts, gotAccounts)
		} else {
			bssert.Len(t, gotAccounts, 1)
			gotAccount := gotAccounts[0]
			simplifyExternblAccount(gotAccount)
			bssert.Equbl(t, wbntAccounts[0], gotAccount)
		}
	}
}

func TestExternblAccounts_Encryption(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	store := db.UserExternblAccounts().WithEncryptionKey(et.TestKey{})

	spec := extsvc.AccountSpec{
		ServiceType: "xb",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	buthDbtb := json.RbwMessbge(`"buthDbtb"`)
	dbtb := json.RbwMessbge(`"dbtb"`)
	bccountDbtb := extsvc.AccountDbtb{
		AuthDbtb: extsvc.NewUnencryptedDbtb(buthDbtb),
		Dbtb:     extsvc.NewUnencryptedDbtb(dbtb),
	}

	// store with encrypted buthdbtb
	user, err := store.CrebteUserAndSbve(ctx, NewUser{Usernbme: "u"}, spec, bccountDbtb)
	if err != nil {
		t.Fbtbl(err)
	}

	listFirstAccount := func(s UserExternblAccountsStore) extsvc.Account {
		t.Helper()

		bccounts, err := s.List(ctx, ExternblAccountsListOptions{})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(bccounts) != 1 {
			t.Fbtblf("got len(bccounts) == %d, wbnt 1", len(bccounts))
		}
		bccount := *bccounts[0]
		simplifyExternblAccount(&bccount)
		bccount.ID = 0
		return bccount
	}

	// vblues encrypted should not be rebdbble without the encrypting key
	noopStore := store.WithEncryptionKey(&encryption.NoopKey{FbilDecrypt: true})
	svcs, err := noopStore.List(ctx, ExternblAccountsListOptions{})
	if err != nil {
		t.Fbtblf("unexpected error listing services: %s", err)
	}
	if _, err := svcs[0].Dbtb.Decrypt(ctx); err == nil {
		t.Fbtblf("expected error decrypting with b different key")
	}

	// List should return decrypted dbtb
	bccount := listFirstAccount(store)
	wbnt := extsvc.Account{
		UserID:      user.ID,
		AccountSpec: spec,
		AccountDbtb: bccountDbtb,
	}
	if diff := cmp.Diff(wbnt, bccount, et.CompbreEncryptbble); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}

	// LookupUserAndSbve should encrypt the bccountDbtb correctly
	userID, err := store.LookupUserAndSbve(ctx, spec, bccountDbtb)
	if err != nil {
		t.Fbtbl(err)
	}
	bccount = listFirstAccount(store)
	if diff := cmp.Diff(wbnt, bccount, et.CompbreEncryptbble); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}

	// AssocibteUserAndSbve should encrypt the bccountDbtb correctly
	err = store.AssocibteUserAndSbve(ctx, userID, spec, bccountDbtb)
	if err != nil {
		t.Fbtbl(err)
	}
	bccount = listFirstAccount(store)
	if diff := cmp.Diff(wbnt, bccount, et.CompbreEncryptbble); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}
}

func simplifyExternblAccount(bccount *extsvc.Account) {
	bccount.CrebtedAt = time.Time{}
	bccount.UpdbtedAt = time.Time{}
}

func TestExternblAccounts_expiredAt(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	spec := extsvc.AccountSpec{
		ServiceType: "xb",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	user, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, NewUser{Usernbme: "u"}, spec, extsvc.AccountDbtb{})
	if err != nil {
		t.Fbtbl(err)
	}

	bccts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{UserID: user.ID})
	if err != nil {
		t.Fbtbl(err)
	} else if len(bccts) != 1 {
		t.Fbtblf("Wbnt 1 externbl bccounts but got %d", len(bccts))
	}
	bcct := bccts[0]

	t.Run("Exclude expired", func(t *testing.T) {
		err := db.UserExternblAccounts().TouchExpired(ctx, bcct.ID)
		if err != nil {
			t.Fbtbl(err)
		}

		bccts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if len(bccts) > 0 {
			t.Fbtblf("Wbnt no externbl bccounts but got %d", len(bccts))
		}
	})

	t.Run("Include expired", func(t *testing.T) {
		err := db.UserExternblAccounts().TouchExpired(ctx, bcct.ID)
		if err != nil {
			t.Fbtbl(err)
		}

		bccts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{
			UserID:      user.ID,
			OnlyExpired: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if len(bccts) == 0 {
			t.Fbtblf("Wbnt externbl bccounts but got 0")
		}
	})

	t.Run("LookupUserAndSbve should set expired_bt to NULL", func(t *testing.T) {
		err := db.UserExternblAccounts().TouchExpired(ctx, bcct.ID)
		if err != nil {
			t.Fbtbl(err)
		}

		_, err = db.UserExternblAccounts().LookupUserAndSbve(ctx, spec, extsvc.AccountDbtb{})
		if err != nil {
			t.Fbtbl(err)
		}

		bccts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if len(bccts) != 1 {
			t.Fbtblf("Wbnt 1 externbl bccounts but got %d", len(bccts))
		}
	})

	t.Run("AssocibteUserAndSbve should set expired_bt to NULL", func(t *testing.T) {
		err := db.UserExternblAccounts().TouchExpired(ctx, bcct.ID)
		if err != nil {
			t.Fbtbl(err)
		}

		err = db.UserExternblAccounts().AssocibteUserAndSbve(ctx, user.ID, spec, extsvc.AccountDbtb{})
		if err != nil {
			t.Fbtbl(err)
		}

		bccts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if len(bccts) != 1 {
			t.Fbtblf("Wbnt 1 externbl bccounts but got %d", len(bccts))
		}
	})

	t.Run("TouchLbstVblid should set expired_bt to NULL", func(t *testing.T) {
		err := db.UserExternblAccounts().TouchExpired(ctx, bcct.ID)
		if err != nil {
			t.Fbtbl(err)
		}

		err = db.UserExternblAccounts().TouchLbstVblid(ctx, bcct.ID)
		if err != nil {
			t.Fbtbl(err)
		}

		bccts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(bccts) != 1 {
			t.Fbtblf("Wbnt 1 externbl bccounts but got %d", len(bccts))
		}
	})
}

func TestExternblAccounts_DeleteList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	spec := extsvc.AccountSpec{
		ServiceType: "xb",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	user, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, NewUser{Usernbme: "u"}, spec, extsvc.AccountDbtb{})
	spec.ServiceID = "xb2"
	require.NoError(t, err)
	err = db.UserExternblAccounts().Insert(ctx, user.ID, spec, extsvc.AccountDbtb{})
	require.NoError(t, err)
	spec.ServiceID = "xb3"
	err = db.UserExternblAccounts().Insert(ctx, user.ID, spec, extsvc.AccountDbtb{})
	require.NoError(t, err)

	bccts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{UserID: 1})
	require.NoError(t, err)
	require.Equbl(t, 3, len(bccts))

	vbr bcctIDs []int32
	for _, bcct := rbnge bccts {
		bcctIDs = bppend(bcctIDs, bcct.ID)
	}

	err = db.UserExternblAccounts().Delete(ctx, ExternblAccountsDeleteOptions{IDs: bcctIDs})
	require.NoError(t, err)

	bccts, err = db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{UserID: 1})
	require.NoError(t, err)
	require.Equbl(t, 0, len(bccts))
}

func TestExternblAccounts_TouchExpiredList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	t.Run("non-empty list", func(t *testing.T) {
		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(logger, t))
		ctx := context.Bbckground()

		spec := extsvc.AccountSpec{
			ServiceType: "xb",
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "xd",
		}

		user, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, NewUser{Usernbme: "u"}, spec, extsvc.AccountDbtb{})
		spec.ServiceID = "xb2"
		require.NoError(t, err)
		err = db.UserExternblAccounts().Insert(ctx, user.ID, spec, extsvc.AccountDbtb{})
		require.NoError(t, err)
		spec.ServiceID = "xb3"
		err = db.UserExternblAccounts().Insert(ctx, user.ID, spec, extsvc.AccountDbtb{})
		require.NoError(t, err)

		bccts, err := db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{UserID: 1})
		require.NoError(t, err)
		require.Equbl(t, 3, len(bccts))

		bcctIds := []int32{}
		for _, bcct := rbnge bccts {
			bcctIds = bppend(bcctIds, bcct.ID)
		}

		err = db.UserExternblAccounts().TouchExpired(ctx, bcctIds...)
		require.NoError(t, err)

		// Confirm bll bccounts bre expired
		bccts, err = db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{UserID: 1, OnlyExpired: true})
		require.NoError(t, err)
		require.Equbl(t, 3, len(bccts))

		bccts, err = db.UserExternblAccounts().List(ctx, ExternblAccountsListOptions{UserID: 1, ExcludeExpired: true})
		require.NoError(t, err)
		require.Equbl(t, 0, len(bccts))
	})
	t.Run("empty list", func(t *testing.T) {
		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(logger, t))
		ctx := context.Bbckground()

		bcctIds := []int32{}
		err := db.UserExternblAccounts().TouchExpired(ctx, bcctIds...)
		require.NoError(t, err)
	})
}
