package database

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func TestExternalAccounts_LookupUserAndSave(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	user, err := db.Users().CreateWithExternalAccount(ctx, NewUser{Username: "u"}, &extsvc.Account{AccountSpec: spec})
	if err != nil {
		t.Fatal(err)
	}

	acct, err := db.UserExternalAccounts().Update(ctx,
		&extsvc.Account{
			AccountSpec: spec,
		})
	if err != nil {
		t.Fatal(err)
	}
	if acct.UserID != user.ID {
		t.Errorf("got %d, want %d", acct.UserID, user.ID)
	}
}

func TestExternalAccounts_AssociateUserAndSave(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	user, err := db.Users().Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	authData := json.RawMessage(`"authData"`)
	data := json.RawMessage(`"data"`)
	accountData := extsvc.AccountData{
		AuthData: extsvc.NewUnencryptedData(authData),
		Data:     extsvc.NewUnencryptedData(data),
	}
	if _, err := db.UserExternalAccounts().Upsert(ctx,
		&extsvc.Account{
			UserID:      user.ID,
			AccountSpec: spec,
			AccountData: accountData,
		}); err != nil {
		t.Fatal(err)
	}

	accounts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
	}
	account := accounts[0]
	simplifyExternalAccount(account)
	account.ID = 0

	want := &extsvc.Account{
		UserID:      user.ID,
		AccountSpec: spec,
		AccountData: accountData,
	}
	if diff := cmp.Diff(want, account, et.CompareEncryptable); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestExternalAccounts_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	specs := []extsvc.AccountSpec{
		{
			ServiceType: "xa",
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "11",
		},
		{
			ServiceType: "xa",
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "12",
		},
		{
			ServiceType: "ya",
			ServiceID:   "yb",
			ClientID:    "yc",
			AccountID:   "3",
		},
	}
	userIDs := make([]int32, 0, len(specs))

	for i, spec := range specs {
		user, err := db.Users().CreateWithExternalAccount(ctx, NewUser{Username: fmt.Sprintf("u%d", i)}, &extsvc.Account{AccountSpec: spec})
		if err != nil {
			t.Fatal(err)
		}
		userIDs = append(userIDs, user.ID)
	}

	specByID := make(map[int32]extsvc.AccountSpec)
	for i, id := range userIDs {
		specByID[id] = specs[i]
	}

	tc := []struct {
		name        string
		args        ExternalAccountsListOptions
		expectedIDs []int32
	}{
		{
			name:        "ListAll",
			expectedIDs: userIDs,
		},
		{
			name:        "ListByAccountID",
			expectedIDs: []int32{userIDs[2]},
			args: ExternalAccountsListOptions{
				AccountID: "3",
			},
		},
		{
			name:        "ListByAccountNotFound",
			expectedIDs: []int32{},
			args: ExternalAccountsListOptions{
				AccountID: "33333",
			},
		},
		{
			name:        "ListByService",
			expectedIDs: []int32{userIDs[0], userIDs[1]},
			args: ExternalAccountsListOptions{
				ServiceType: "xa",
				ServiceID:   "xb",
				ClientID:    "xc",
			},
		},
		{
			name:        "ListByServiceTypeOnly",
			expectedIDs: []int32{userIDs[0], userIDs[1]},
			args: ExternalAccountsListOptions{
				ServiceType: "xa",
			},
		},
		{
			name:        "ListByServiceIDOnly",
			expectedIDs: []int32{userIDs[0], userIDs[1]},
			args: ExternalAccountsListOptions{
				ServiceID: "xb",
			},
		},
		{
			name:        "ListByClientIDOnly",
			expectedIDs: []int32{userIDs[2]},
			args: ExternalAccountsListOptions{
				ClientID: "yc",
			},
		},
		{
			name:        "ListByServiceNotFound",
			expectedIDs: []int32{},
			args: ExternalAccountsListOptions{
				ServiceType: "notfound",
				ServiceID:   "notfound",
				ClientID:    "notfound",
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			accounts, err := db.UserExternalAccounts().List(ctx, c.args)
			if err != nil {
				t.Fatal(err)
			}
			if got, expected := len(accounts), len(c.expectedIDs); got != expected {
				t.Fatalf("len(accounts) got %d, want %d", got, expected)
			}
			for i, id := range c.expectedIDs {
				account := accounts[i]
				simplifyExternalAccount(account)
				want := &extsvc.Account{
					UserID:      id,
					ID:          id,
					AccountSpec: specByID[id],
				}
				if diff := cmp.Diff(want, account, et.CompareEncryptable); diff != "" {
					t.Fatalf("Mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestExternalAccounts_ListForUsers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	specs := []extsvc.AccountSpec{
		{ServiceType: "xa", ServiceID: "xb", ClientID: "xc", AccountID: "11"},
		{ServiceType: "xa", ServiceID: "xb", ClientID: "xc", AccountID: "12"},
	}
	const numberOfUsers = 3
	userIDs := make([]int32, 0, numberOfUsers)
	thirdUserSpecs := []extsvc.AccountSpec{
		{ServiceType: "a", ServiceID: "b", ClientID: "xc", AccountID: "111"},
		{ServiceType: "c", ServiceID: "d", ClientID: "xc", AccountID: "112"},
		{ServiceType: "e", ServiceID: "f", ClientID: "yc", AccountID: "13"},
	}

	for i, spec := range append(specs, thirdUserSpecs...) {
		if i < 3 {
			user, err := db.Users().CreateWithExternalAccount(ctx, NewUser{Username: fmt.Sprintf("u%d", i)}, &extsvc.Account{AccountSpec: spec})
			require.NoError(t, err)
			userIDs = append(userIDs, user.ID)
		} else {
			// Last user gets all the accounts.
			_, err := db.UserExternalAccounts().Upsert(ctx,
				&extsvc.Account{
					UserID:      userIDs[2],
					AccountSpec: spec,
				})
			require.NoError(t, err)
		}
	}

	wantAccountsByUserID := make(map[int32][]*extsvc.Account)
	for _, id := range userIDs {
		// Last user gets all the accounts.
		if int(id) == numberOfUsers {
			accts := make([]*extsvc.Account, 0, numberOfUsers)
			for idx, spec := range thirdUserSpecs {
				accts = append(accts, &extsvc.Account{UserID: id, ID: id + int32(idx), AccountSpec: spec})
			}
			wantAccountsByUserID[id] = accts
		} else {
			wantAccountsByUserID[id] = []*extsvc.Account{{UserID: id, ID: id, AccountSpec: specs[int(id)-1]}}
		}
	}

	// Zero IDs in the input -- empty map in the output.
	accounts, err := db.UserExternalAccounts().ListForUsers(ctx, []int32{})
	require.NoError(t, err)
	assert.Empty(t, accounts)

	// All accounts should be returned.
	accounts, err = db.UserExternalAccounts().ListForUsers(ctx, userIDs)
	require.NoError(t, err)
	assert.Len(t, accounts, numberOfUsers)

	for userID, wantAccounts := range wantAccountsByUserID {
		gotAccounts := accounts[userID]
		// Case of last user with all accounts.
		if int(userID) == numberOfUsers {
			assert.Equal(t, len(wantAccounts), len(gotAccounts))
			for _, gotAccount := range gotAccounts {
				simplifyExternalAccount(gotAccount)
			}
			assert.ElementsMatch(t, wantAccounts, gotAccounts)
		} else {
			assert.Len(t, gotAccounts, 1)
			gotAccount := gotAccounts[0]
			simplifyExternalAccount(gotAccount)
			assert.Equal(t, wantAccounts[0], gotAccount)
		}
	}
}

func TestExternalAccounts_Encryption(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	defaultKeyring := keyring.Default()
	keyring.MockDefault(keyring.Ring{UserExternalAccountKey: et.TestKey{}})
	t.Cleanup(func() {
		keyring.MockDefault(defaultKeyring)
	})
	store := db.UserExternalAccounts()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	authData := json.RawMessage(`"authData"`)
	data := json.RawMessage(`"data"`)
	accountData := extsvc.AccountData{
		AuthData: extsvc.NewUnencryptedData(authData),
		Data:     extsvc.NewUnencryptedData(data),
	}

	// store with encrypted authdata
	user, err := db.Users().CreateWithExternalAccount(ctx, NewUser{Username: "u"}, &extsvc.Account{AccountSpec: spec, AccountData: accountData})
	if err != nil {
		t.Fatal(err)
	}

	listFirstAccount := func(s UserExternalAccountsStore) extsvc.Account {
		t.Helper()

		accounts, err := s.List(ctx, ExternalAccountsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(accounts) != 1 {
			t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
		}
		account := *accounts[0]
		simplifyExternalAccount(&account)
		account.ID = 0
		return account
	}

	// values encrypted should not be readable without the encrypting key
	noopStore := store.WithEncryptionKey(&encryption.NoopKey{FailDecrypt: true})
	svcs, err := noopStore.List(ctx, ExternalAccountsListOptions{})
	if err != nil {
		t.Fatalf("unexpected error listing services: %s", err)
	}
	if _, err := svcs[0].Data.Decrypt(ctx); err == nil {
		t.Fatalf("expected error decrypting with a different key")
	}

	// List should return decrypted data
	account := listFirstAccount(store)
	want := extsvc.Account{
		UserID:      user.ID,
		AccountSpec: spec,
		AccountData: accountData,
	}
	if diff := cmp.Diff(want, account, et.CompareEncryptable); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// LookupUserAndSave should encrypt the accountData correctly
	account.AccountSpec = spec
	account.AccountData = accountData
	acct, err := store.Update(ctx, &account)
	if err != nil {
		t.Fatal(err)
	}
	account = listFirstAccount(store)
	if diff := cmp.Diff(want, account, et.CompareEncryptable); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// AssociateUserAndSave should encrypt the accountData correctly
	_, err = store.Upsert(ctx,
		&extsvc.Account{
			UserID:      acct.UserID,
			AccountSpec: spec,
			AccountData: accountData,
		})
	if err != nil {
		t.Fatal(err)
	}
	account = listFirstAccount(store)
	if diff := cmp.Diff(want, account, et.CompareEncryptable); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func simplifyExternalAccount(account *extsvc.Account) {
	account.CreatedAt = time.Time{}
	account.UpdatedAt = time.Time{}
}

func TestExternalAccounts_expiredAt(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	user, err := db.Users().CreateWithExternalAccount(ctx, NewUser{Username: "u"}, &extsvc.Account{AccountSpec: spec})
	if err != nil {
		t.Fatal(err)
	}

	accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	} else if len(accts) != 1 {
		t.Fatalf("Want 1 external accounts but got %d", len(accts))
	}
	acct := accts[0]

	t.Run("Exclude expired", func(t *testing.T) {
		err := db.UserExternalAccounts().TouchExpired(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(accts) > 0 {
			t.Fatalf("Want no external accounts but got %d", len(accts))
		}
	})

	t.Run("Include expired", func(t *testing.T) {
		err := db.UserExternalAccounts().TouchExpired(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{
			UserID:      user.ID,
			OnlyExpired: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(accts) == 0 {
			t.Fatalf("Want external accounts but got 0")
		}
	})

	t.Run("LookupUserAndSave should set expired_at to NULL", func(t *testing.T) {
		err := db.UserExternalAccounts().TouchExpired(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.UserExternalAccounts().Update(ctx, &extsvc.Account{AccountSpec: spec})
		if err != nil {
			t.Fatal(err)
		}

		accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(accts) != 1 {
			t.Fatalf("Want 1 external accounts but got %d", len(accts))
		}
	})

	t.Run("AssociateUserAndSave should set expired_at to NULL", func(t *testing.T) {
		err := db.UserExternalAccounts().TouchExpired(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.UserExternalAccounts().Upsert(ctx,
			&extsvc.Account{
				UserID:      user.ID,
				AccountSpec: spec,
			})
		if err != nil {
			t.Fatal(err)
		}

		accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(accts) != 1 {
			t.Fatalf("Want 1 external accounts but got %d", len(accts))
		}
	})

	t.Run("TouchLastValid should set expired_at to NULL", func(t *testing.T) {
		err := db.UserExternalAccounts().TouchExpired(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		err = db.UserExternalAccounts().TouchLastValid(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(accts) != 1 {
			t.Fatalf("Want 1 external accounts but got %d", len(accts))
		}
	})
}

func TestExternalAccounts_DeleteList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	user, err := db.Users().CreateWithExternalAccount(ctx, NewUser{Username: "u"}, &extsvc.Account{AccountSpec: spec})
	spec.ServiceID = "xb2"
	require.NoError(t, err)
	_, err = db.UserExternalAccounts().Insert(ctx,
		&extsvc.Account{
			UserID:      user.ID,
			AccountSpec: spec,
		})
	require.NoError(t, err)
	spec.ServiceID = "xb3"
	_, err = db.UserExternalAccounts().Insert(ctx,
		&extsvc.Account{
			UserID:      user.ID,
			AccountSpec: spec,
		})
	require.NoError(t, err)

	accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{UserID: 1})
	require.NoError(t, err)
	require.Equal(t, 3, len(accts))

	var acctIDs []int32
	for _, acct := range accts {
		acctIDs = append(acctIDs, acct.ID)
	}

	err = db.UserExternalAccounts().Delete(ctx, ExternalAccountsDeleteOptions{IDs: acctIDs})
	require.NoError(t, err)

	accts, err = db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{UserID: 1})
	require.NoError(t, err)
	require.Equal(t, 0, len(accts))
}

func TestExternalAccounts_TouchExpiredList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	t.Run("non-empty list", func(t *testing.T) {
		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(t))
		ctx := context.Background()

		spec := extsvc.AccountSpec{
			ServiceType: "xa",
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "xd",
		}

		user, err := db.Users().CreateWithExternalAccount(ctx, NewUser{Username: "u"}, &extsvc.Account{AccountSpec: spec})
		spec.ServiceID = "xb2"
		require.NoError(t, err)
		_, err = db.UserExternalAccounts().Insert(ctx,
			&extsvc.Account{
				UserID:      user.ID,
				AccountSpec: spec,
			})
		require.NoError(t, err)
		spec.ServiceID = "xb3"
		_, err = db.UserExternalAccounts().Insert(ctx,
			&extsvc.Account{
				UserID:      user.ID,
				AccountSpec: spec,
			})
		require.NoError(t, err)

		accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{UserID: 1})
		require.NoError(t, err)
		require.Equal(t, 3, len(accts))

		acctIds := []int32{}
		for _, acct := range accts {
			acctIds = append(acctIds, acct.ID)
		}

		err = db.UserExternalAccounts().TouchExpired(ctx, acctIds...)
		require.NoError(t, err)

		// Confirm all accounts are expired
		accts, err = db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{UserID: 1, OnlyExpired: true})
		require.NoError(t, err)
		require.Equal(t, 3, len(accts))

		accts, err = db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{UserID: 1, ExcludeExpired: true})
		require.NoError(t, err)
		require.Equal(t, 0, len(accts))
	})
	t.Run("empty list", func(t *testing.T) {
		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(t))
		ctx := context.Background()

		acctIds := []int32{}
		err := db.UserExternalAccounts().TouchExpired(ctx, acctIds...)
		require.NoError(t, err)
	})
}
