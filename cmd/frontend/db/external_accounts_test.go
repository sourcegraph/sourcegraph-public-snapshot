package db

import (
	"reflect"
	"testing"
	"time"

	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
)

func TestExternalAccounts_LookupUserAndSave(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	spec := ExternalAccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	userID, err := ExternalAccounts.CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, ExternalAccountData{})
	if err != nil {
		t.Fatal(err)
	}

	lookedUpUserID, err := ExternalAccounts.LookupUserAndSave(ctx, spec, ExternalAccountData{})
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
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	spec := ExternalAccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	if err := ExternalAccounts.AssociateUserAndSave(ctx, user.ID, spec, ExternalAccountData{}); err != nil {
		t.Fatal(err)
	}

	accounts, err := ExternalAccounts.List(ctx, ExternalAccountsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
	}
	account := *accounts[0]
	simplifyExternalAccount(&account)
	account.ID = 0
	if want := (ExternalAccount{UserID: user.ID, ExternalAccountSpec: spec}); !reflect.DeepEqual(account, want) {
		t.Errorf("got %+v, want %+v", account, want)
	}
}

func TestExternalAccounts_CreateUserAndSave(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	spec := ExternalAccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	userID, err := ExternalAccounts.CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, ExternalAccountData{})
	if err != nil {
		t.Fatal(err)
	}

	user, err := Users.GetByID(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "u"; user.Username != want {
		t.Errorf("got %q, want %q", user.Username, want)
	}

	accounts, err := ExternalAccounts.List(ctx, ExternalAccountsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
	}
	account := *accounts[0]
	simplifyExternalAccount(&account)
	account.ID = 0
	if want := (ExternalAccount{UserID: userID, ExternalAccountSpec: spec}); !reflect.DeepEqual(account, want) {
		t.Errorf("got %+v, want %+v", account, want)
	}
}

func simplifyExternalAccount(account *ExternalAccount) {
	account.CreatedAt = time.Time{}
	account.UpdatedAt = time.Time{}
}
