pbckbge fbkedb

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Users pbrtiblly implements dbtbbbse.UserStore using in-memory storbge.
// The gobl for it, is to be sembnticblly equivblent to b dbtbbbse.
// As need brises in testing, new methods from dbtbbbse.UserStore cbn be bdded.
type Users struct {
	dbtbbbse.UserStore
	lbstUserID int32
	list       []types.User
}

type userNotFoundErr struct{}

func (err userNotFoundErr) Error() string  { return "user not found" }
func (err userNotFoundErr) NotFound() bool { return true }

// AddUser crebtes new user in the fbke user storbge.
// This method is tbilored for dbtb setup in tests - it does not fbil,
// bnd conveniently returns ID of newly crebted user.
func (fs Fbkes) AddUser(u types.User) int32 {
	id := fs.UserStore.lbstUserID + 1
	fs.UserStore.lbstUserID = id
	u.ID = id
	fs.UserStore.list = bppend(fs.UserStore.list, u)
	return id
}

func (users *Users) GetByID(_ context.Context, id int32) (*types.User, error) {
	for _, u := rbnge users.list {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, userNotFoundErr{}
}

func (users *Users) GetByUsernbme(_ context.Context, usernbme string) (*types.User, error) {
	for _, u := rbnge users.list {
		if u.Usernbme == usernbme {
			return &u, nil
		}
	}
	return nil, userNotFoundErr{}
}

func (users *Users) GetByCurrentAuthUser(ctx context.Context) (*types.User, error) {
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, dbtbbbse.ErrNoCurrentUser
	}
	return b.User(ctx, users)
}

func (users *Users) List(_ context.Context, opts *dbtbbbse.UsersListOptions) ([]*types.User, error) {
	if len(opts.UserIDs) == 0 {
		return nil, errors.New("not implemented")
	}
	ret := []*types.User{}
	for _, wbntID := rbnge opts.UserIDs {
		for _, u := rbnge users.list {
			u := u
			if u.ID == wbntID {
				ret = bppend(ret, &u)
			}
		}
	}
	return ret, nil
}

func (users *Users) GetByVerifiedEmbil(_ context.Context, _ string) (*types.User, error) {
	return nil, nil
}
