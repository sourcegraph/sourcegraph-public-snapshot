package auth

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

func TestCreateOrUpdateUser(t *testing.T) {
	const (
		wantUsername     = "u"
		wantUserID       = 1
		wantAuthedUserID = 2
	)

	mockUsersGetByID := func() {
		// This should always pass, because in the impl we just were able to retrieve the user.
		db.Mocks.Users.GetByID = func(_ context.Context, userID int32) (*types.User, error) {
			return &types.User{ID: userID, Username: wantUsername}, nil
		}
	}
	mockUsersUpdate := func() {
		db.Mocks.Users.Update = func(int32, db.UserUpdate) error { return nil }
	}
	mockNewUser := func(t *testing.T) {
		db.Mocks.ExternalAccounts.LookupUserAndSave = func(a db.ExternalAccountSpec, d db.ExternalAccountData) (userID int32, err error) {
			return 0, &errcode.Mock{IsNotFound: true}
		}
	}
	mockExistingUser := func(t *testing.T) {
		db.Mocks.ExternalAccounts.LookupUserAndSave = func(a db.ExternalAccountSpec, d db.ExternalAccountData) (userID int32, err error) {
			return wantUserID, nil
		}
	}

	t.Run("new user, new external account", func(t *testing.T) {
		var calledCreateUserAndSave bool
		mockUsersUpdate()
		mockNewUser(t)
		db.Mocks.ExternalAccounts.CreateUserAndSave = func(u db.NewUser, a db.ExternalAccountSpec, d db.ExternalAccountData) (createdUserID int32, err error) {
			calledCreateUserAndSave = true
			return wantUserID, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		userID, _, err := CreateOrUpdateUser(context.Background(), db.NewUser{Username: wantUsername}, db.ExternalAccountSpec{}, db.ExternalAccountData{})
		if err != nil {
			t.Fatal(err)
		}
		if userID != wantUserID {
			t.Errorf("got %d, want %d", userID, wantUserID)
		}
		if !calledCreateUserAndSave {
			t.Error("!calledCreateUserAndSave")
		}
	})

	t.Run("new user, existing external account", func(t *testing.T) {
		var calledCreateUserAndSave bool
		mockUsersUpdate()
		mockNewUser(t)
		wantErr := errors.New("x")
		db.Mocks.ExternalAccounts.CreateUserAndSave = func(u db.NewUser, a db.ExternalAccountSpec, d db.ExternalAccountData) (createdUserID int32, err error) {
			calledCreateUserAndSave = true
			return 0, wantErr
		}
		defer func() { db.Mocks = db.MockStores{} }()
		if _, _, err := CreateOrUpdateUser(context.Background(), db.NewUser{Username: wantUsername}, db.ExternalAccountSpec{}, db.ExternalAccountData{}); err != wantErr {
			t.Fatalf("got err %q, want %q", err, wantErr)
		}
		if !calledCreateUserAndSave {
			t.Error("!calledCreateUserAndSave")
		}
	})

	t.Run("authed user, new external account", func(t *testing.T) {
		var calledAssociateUserAndSave bool
		mockUsersGetByID()
		mockUsersUpdate()
		mockExistingUser(t)
		db.Mocks.ExternalAccounts.AssociateUserAndSave = func(userID int32, a db.ExternalAccountSpec, d db.ExternalAccountData) error {
			if userID != wantAuthedUserID {
				t.Errorf("got %d, want %d", userID, wantAuthedUserID)
			}
			calledAssociateUserAndSave = true
			return nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: wantAuthedUserID})
		userID, _, err := CreateOrUpdateUser(ctx, db.NewUser{Username: wantUsername}, db.ExternalAccountSpec{}, db.ExternalAccountData{})
		if err != nil {
			t.Fatal(err)
		}
		if userID != wantAuthedUserID {
			t.Errorf("got %d, want %d", userID, wantUserID)
		}
		if !calledAssociateUserAndSave {
			t.Error("!calledAssociateUserAndSave")
		}
	})

	t.Run("authed user, existing (conflicting) external account", func(t *testing.T) {
		var calledAssociateUserAndSave bool
		mockUsersGetByID()
		mockUsersUpdate()
		mockExistingUser(t)
		wantErr := errors.New("x")
		db.Mocks.ExternalAccounts.AssociateUserAndSave = func(userID int32, a db.ExternalAccountSpec, d db.ExternalAccountData) error {
			calledAssociateUserAndSave = true
			return wantErr
		}
		defer func() { db.Mocks = db.MockStores{} }()
		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: wantAuthedUserID})
		if _, _, err := CreateOrUpdateUser(ctx, db.NewUser{Username: wantUsername}, db.ExternalAccountSpec{}, db.ExternalAccountData{}); err != wantErr {
			t.Fatalf("got err %q, want %q", err, wantErr)
		}
		if !calledAssociateUserAndSave {
			t.Error("!calledAssociateUserAndSave")
		}
	})
}
