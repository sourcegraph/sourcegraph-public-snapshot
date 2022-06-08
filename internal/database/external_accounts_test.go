package database

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func TestExternalAccounts_LookupUserAndSave(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	userID, err := db.UserExternalAccounts().CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, extsvc.AccountData{})
	if err != nil {
		t.Fatal(err)
	}

	lookedUpUserID, err := db.UserExternalAccounts().LookupUserAndSave(ctx, spec, extsvc.AccountData{})
	if err != nil {
		t.Fatal(err)
	}
	if lookedUpUserID != userID {
		t.Errorf("got %d, want %d", lookedUpUserID, userID)
	}
}

func TestExternalAccounts_AssociateUserAndSave(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
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
		AuthData: &authData,
		Data:     &data,
	}
	if err := db.UserExternalAccounts().AssociateUserAndSave(ctx, user.ID, spec, accountData); err != nil {
		t.Fatal(err)
	}

	accounts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
	}
	account := *accounts[0]
	simplifyExternalAccount(&account)
	account.ID = 0

	want := extsvc.Account{
		UserID:      user.ID,
		AccountSpec: spec,
		AccountData: accountData,
	}
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestExternalAccounts_CreateUserAndSave(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	authData := json.RawMessage(`"authData"`)
	data := json.RawMessage(`"data"`)
	accountData := extsvc.AccountData{
		AuthData: &authData,
		Data:     &data,
	}
	userID, err := db.UserExternalAccounts().CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, accountData)
	if err != nil {
		t.Fatal(err)
	}

	user, err := db.Users().GetByID(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "u"; user.Username != want {
		t.Errorf("got %q, want %q", user.Username, want)
	}

	accounts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
	}
	account := *accounts[0]
	simplifyExternalAccount(&account)
	account.ID = 0

	want := extsvc.Account{
		UserID:      userID,
		AccountSpec: spec,
		AccountData: accountData,
	}
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestExternalAccounts_CreateUserAndSave_NilData(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	userID, err := db.UserExternalAccounts().CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, extsvc.AccountData{})
	if err != nil {
		t.Fatal(err)
	}

	user, err := db.Users().GetByID(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "u"; user.Username != want {
		t.Errorf("got %q, want %q", user.Username, want)
	}

	accounts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
	}
	account := *accounts[0]
	simplifyExternalAccount(&account)
	account.ID = 0

	want := extsvc.Account{
		UserID:      userID,
		AccountSpec: spec,
	}
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestExternalAccounts_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
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
		id, err := db.UserExternalAccounts().CreateUserAndSave(ctx, NewUser{Username: fmt.Sprintf("u%d", i)}, spec, extsvc.AccountData{})
		if err != nil {
			t.Fatal(err)
		}
		userIDs = append(userIDs, id)
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
				AccountID: 3,
			},
		},
		{
			name:        "ListByAccountNotFound",
			expectedIDs: []int32{},
			args: ExternalAccountsListOptions{
				AccountID: 33333,
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
				want := extsvc.Account{
					UserID:      id,
					ID:          id,
					AccountSpec: specByID[id],
				}
				if diff := cmp.Diff(want, *account); diff != "" {
					t.Fatalf("Mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestExternalAccounts_Encryption(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	store := db.UserExternalAccounts().WithEncryptionKey(et.TestKey{})

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	authData := json.RawMessage(`"authData"`)
	data := json.RawMessage(`"data"`)
	accountData := extsvc.AccountData{
		AuthData: &authData,
		Data:     &data,
	}

	// store with encrypted authdata
	userID, err := store.CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, accountData)
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

	// create a store with a NoopKey to read the raw encrypted value
	noopStore := store.WithEncryptionKey(&encryption.NoopKey{})

	account := listFirstAccount(noopStore)

	// if the testKey worked, the data should just be a base64 encoded version
	if string(*account.AuthData) != base64.StdEncoding.EncodeToString([]byte(*accountData.AuthData)) {
		t.Fatalf("expected base64 encoded auth data, got %s", string(*account.AuthData))
	}

	// List should return decrypted data
	account = listFirstAccount(store)
	want := extsvc.Account{
		UserID:      userID,
		AccountSpec: spec,
		AccountData: accountData,
	}
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// LookupUserAndSave should encrypt the accountData correctly
	userID, err = store.LookupUserAndSave(ctx, spec, accountData)
	if err != nil {
		t.Fatal(err)
	}
	account = listFirstAccount(store)
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// AssociateUserAndSave should encrypt the accountData correctly
	err = store.AssociateUserAndSave(ctx, userID, spec, accountData)
	if err != nil {
		t.Fatal(err)
	}
	account = listFirstAccount(store)
	if diff := cmp.Diff(want, account); diff != "" {
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
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	userID, err := db.UserExternalAccounts().CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, extsvc.AccountData{})
	if err != nil {
		t.Fatal(err)
	}

	accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{UserID: userID})
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
			UserID:         userID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(accts) > 0 {
			t.Fatalf("Want no external accounts but got %d", len(accts))
		}
	})

	t.Run("LookupUserAndSave should set expired_at to NULL", func(t *testing.T) {
		err := db.UserExternalAccounts().TouchExpired(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.UserExternalAccounts().LookupUserAndSave(ctx, spec, extsvc.AccountData{})
		if err != nil {
			t.Fatal(err)
		}

		accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{
			UserID:         userID,
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

		err = db.UserExternalAccounts().AssociateUserAndSave(ctx, userID, spec, extsvc.AccountData{})
		if err != nil {
			t.Fatal(err)
		}

		accts, err := db.UserExternalAccounts().List(ctx, ExternalAccountsListOptions{
			UserID:         userID,
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
			UserID:         userID,
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
