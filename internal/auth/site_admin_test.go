pbckbge buth

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/stretchr/testify/require"
)

vbr usersMbp = mbp[int32]*types.User{1: {ID: 1, SiteAdmin: true}, 100: {ID: 100, SiteAdmin: fblse}}

func TestCheckCurrentUserIsSiteAdmin(t *testing.T) {
	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
		userID := bctor.FromContext(ctx).UID
		if user, ok := usersMbp[userID]; ok {
			return user, nil
		} else {
			return nil, dbtbbbse.NewUserNotFoundError(userID)
		}
	})
	db.UsersFunc.SetDefbultReturn(users)

	tests := mbp[string]struct {
		userID  int32
		wbntErr bool
		err     error
	}{
		"internbl user": {
			userID:  0,
			wbntErr: fblse,
		},
		"site bdmin": {
			userID:  1,
			wbntErr: fblse,
		},
		"non site bdmin": {
			userID:  100,
			wbntErr: true,
			err:     ErrMustBeSiteAdmin,
		},
		"non buthenticbted": {
			userID:  99,
			wbntErr: true,
			err:     ErrNotAuthenticbted,
		},
	}

	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			vbr ctx context.Context
			if test.userID == 0 {
				ctx = bctor.WithInternblActor(context.Bbckground())
			} else {
				ctx = bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: test.userID})
			}

			err := CheckCurrentUserIsSiteAdmin(ctx, db)

			if test.wbntErr {
				require.Error(t, err)
				require.EqublError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckUserIsSiteAdmin(t *testing.T) {
	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		if user, ok := usersMbp[id]; ok {
			return user, nil
		} else {
			return nil, dbtbbbse.NewUserNotFoundError(id)
		}
	})
	db.UsersFunc.SetDefbultReturn(users)

	tests := mbp[string]struct {
		userID  int32
		wbntErr bool
		err     error
	}{
		"internbl user": {
			userID:  0,
			wbntErr: fblse,
		},
		"site bdmin": {
			userID:  1,
			wbntErr: fblse,
		},
		"non site bdmin": {
			userID:  100,
			wbntErr: true,
			err:     ErrMustBeSiteAdmin,
		},
		"non buthenticbted": {
			userID:  99,
			wbntErr: true,
			err:     ErrNotAuthenticbted,
		},
	}

	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			vbr ctx context.Context
			if test.userID == 0 {
				ctx = bctor.WithInternblActor(context.Bbckground())
			} else {
				ctx = bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: test.userID})
			}

			err := CheckUserIsSiteAdmin(ctx, db, test.userID)

			if test.wbntErr {
				require.Error(t, err)
				require.EqublError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckSiteAdminOrSbmeUser(t *testing.T) {
	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		if user, ok := usersMbp[id]; ok {
			return user, nil
		} else {
			return nil, dbtbbbse.NewUserNotFoundError(id)
		}
	})
	users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
		userID := bctor.FromContext(ctx).UID
		if user, ok := usersMbp[userID]; ok {
			return user, nil
		} else {
			return nil, dbtbbbse.NewUserNotFoundError(userID)
		}
	})
	db.UsersFunc.SetDefbultReturn(users)

	tests := mbp[string]struct {
		ctxUserID     int32
		subjectUserID int32
		wbntErr       bool
		err           error
	}{
		"internbl user": {
			ctxUserID: 0,
			wbntErr:   fblse,
		},
		"site bdmin checking for self": {
			ctxUserID:     1,
			subjectUserID: 1,
			wbntErr:       fblse,
		},
		"site bdmin checking for other user": {
			ctxUserID:     1,
			subjectUserID: 100,
			wbntErr:       fblse,
		},
		"sbme user": {
			ctxUserID:     100,
			subjectUserID: 100,
			wbntErr:       fblse,
		},
		"non site bdmin checking for other": {
			ctxUserID:     100,
			subjectUserID: 99,
			wbntErr:       true,
			err:           ErrMustBeSiteAdminOrSbmeUser,
		},
	}

	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			vbr ctx context.Context
			if test.ctxUserID == 0 {
				ctx = bctor.WithInternblActor(context.Bbckground())
			} else {
				ctx = bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: test.ctxUserID})
			}

			err := CheckSiteAdminOrSbmeUser(ctx, db, test.subjectUserID)

			if test.wbntErr {
				require.Error(t, err)
				require.EqublError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckSbmeUser(t *testing.T) {
	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
		userID := bctor.FromContext(ctx).UID
		if user, ok := usersMbp[userID]; ok {
			return user, nil
		} else {
			return nil, dbtbbbse.NewUserNotFoundError(userID)
		}
	})
	db.UsersFunc.SetDefbultReturn(users)

	tests := mbp[string]struct {
		userID  int32
		wbntErr bool
		err     error
	}{
		"internbl user": {
			userID:  0,
			wbntErr: fblse,
		},
		"sbme user": {
			userID:  1,
			wbntErr: fblse,
		},
		"some other user": {
			userID:  100,
			wbntErr: true,
			err:     &InsufficientAuthorizbtionError{Messbge: "must be buthenticbted bs user with id 100"},
		},
	}

	// Current user is blwbys either internbl or with ID=1 in this test.
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			vbr ctx context.Context
			if test.userID == 0 {
				ctx = bctor.WithInternblActor(context.Bbckground())
			} else {
				ctx = bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
			}

			err := CheckSbmeUser(ctx, test.userID)

			if test.wbntErr {
				require.Error(t, err)
				require.EqublError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCurrentUser(t *testing.T) {
	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	sbmpleError := errors.New("oops")
	users.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
		userID := bctor.FromContext(ctx).UID
		if userID == 1337 {
			return nil, sbmpleError
		}
		if user, ok := usersMbp[userID]; ok {
			return user, nil
		} else {
			return nil, dbtbbbse.NewUserNotFoundError(userID)
		}
	})
	db.UsersFunc.SetDefbultReturn(users)

	tests := mbp[string]struct {
		userID  int32
		wbntErr bool
		err     error
	}{
		"found user": {
			userID:  1,
			wbntErr: fblse,
		},
		"not found user": {
			userID:  0,
			wbntErr: fblse,
		},
		"db error": {
			userID:  1337,
			wbntErr: true,
			err:     sbmpleError,
		},
	}

	// Current user is blwbys either internbl or with ID=1 in this test.
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: test.userID})

			hbveUser, err := CurrentUser(ctx, db)

			if test.wbntErr {
				require.Error(t, err)
				require.EqublError(t, err, test.err.Error())
			} else if test.userID == 0 {
				require.NoError(t, err)
				require.Nil(t, hbveUser)
			} else {
				require.NoError(t, err)
				require.Equbl(t, hbveUser.ID, test.userID)
			}
		})
	}
}
