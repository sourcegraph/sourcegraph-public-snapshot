package fakedb

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Users partially implements database.UserStore using in-memory storage.
// The goal for it, is to be semantically equivalent to a database.
// As need arises in testing, new methods from database.UserStore can be added.
type Users struct {
	database.UserStore
	lastUserID int32
	list       []types.User
}

type userNotFoundErr struct{}

func (err userNotFoundErr) Error() string  { return "user not found" }
func (err userNotFoundErr) NotFound() bool { return true }

// AddUser creates new user in the fake user storage.
// This method is tailored for data setup in tests - it does not fail,
// and conveniently returns ID of newly created user.
func (fs Fakes) AddUser(u types.User) int32 {
	id := fs.UserStore.lastUserID + 1
	fs.UserStore.lastUserID = id
	u.ID = id
	fs.UserStore.list = append(fs.UserStore.list, u)
	return id
}

func (users *Users) GetByID(_ context.Context, id int32) (*types.User, error) {
	for _, u := range users.list {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, userNotFoundErr{}
}

func (users *Users) GetByUsername(_ context.Context, username string) (*types.User, error) {
	for _, u := range users.list {
		if u.Username == username {
			return &u, nil
		}
	}
	return nil, userNotFoundErr{}
}

func (users *Users) GetByCurrentAuthUser(ctx context.Context) (*types.User, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, database.ErrNoCurrentUser
	}
	return a.User(ctx, users)
}

func (users *Users) List(_ context.Context, opts *database.UsersListOptions) ([]*types.User, error) {
	if len(opts.UserIDs) == 0 {
		return nil, errors.New("not implemented")
	}
	ret := []*types.User{}
	for _, wantID := range opts.UserIDs {
		for _, u := range users.list {
			u := u
			if u.ID == wantID {
				ret = append(ret, &u)
			}
		}
	}
	return ret, nil
}

func (users *Users) GetByVerifiedEmail(_ context.Context, _ string) (*types.User, error) {
	return nil, nil
}
