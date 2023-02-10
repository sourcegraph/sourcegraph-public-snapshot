package fakedb

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type Users struct {
	database.UserStore
	lastUserID int32
	list       []types.User
}

func (users *Users) GetByID(_ context.Context, id int32) (*types.User, error) {
	for _, u := range users.list {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, nil
}

func (users *Users) GetByCurrentAuthUser(ctx context.Context) (*types.User, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, database.ErrNoCurrentUser
	}
	return a.User(ctx, users)
}

// NewUser creates new user in the fake user storage.
// This method is tailored for data setup in tests - it does not fail,
// and conveniently returns ID of newly created user.
func (users *Users) NewUser(u types.User) int32 {
	id := users.lastUserID + 1
	users.lastUserID = id
	u.ID = id
	users.list = append(users.list, u)
	return id
}
