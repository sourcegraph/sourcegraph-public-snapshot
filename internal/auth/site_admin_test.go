package auth

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/require"
)

var usersMap = map[int32]*types.User{1: {ID: 1, SiteAdmin: true}, 100: {ID: 100, SiteAdmin: false}}

func TestCheckCurrentUserIsSiteAdmin(t *testing.T) {
	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
		userID := actor.FromContext(ctx).UID
		if user, ok := usersMap[userID]; ok {
			return user, nil
		} else {
			return nil, database.NewUserNotFoundError(userID)
		}
	})
	db.UsersFunc.SetDefaultReturn(users)

	tests := map[string]struct {
		userID  int32
		wantErr bool
		err     error
	}{
		"internal user": {
			userID:  0,
			wantErr: false,
		},
		"site admin": {
			userID:  1,
			wantErr: false,
		},
		"non site admin": {
			userID:  100,
			wantErr: true,
			err:     ErrMustBeSiteAdmin,
		},
		"non authenticated": {
			userID:  99,
			wantErr: true,
			err:     ErrNotAuthenticated,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var ctx context.Context
			if test.userID == 0 {
				ctx = actor.WithInternalActor(context.Background())
			} else {
				ctx = actor.WithActor(context.Background(), &actor.Actor{UID: test.userID})
			}

			err := CheckCurrentUserIsSiteAdmin(ctx, db)

			if test.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckUserIsSiteAdmin(t *testing.T) {
	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		if user, ok := usersMap[id]; ok {
			return user, nil
		} else {
			return nil, database.NewUserNotFoundError(id)
		}
	})
	db.UsersFunc.SetDefaultReturn(users)

	tests := map[string]struct {
		userID  int32
		wantErr bool
		err     error
	}{
		"internal user": {
			userID:  0,
			wantErr: false,
		},
		"site admin": {
			userID:  1,
			wantErr: false,
		},
		"non site admin": {
			userID:  100,
			wantErr: true,
			err:     ErrMustBeSiteAdmin,
		},
		"non authenticated": {
			userID:  99,
			wantErr: true,
			err:     ErrNotAuthenticated,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var ctx context.Context
			if test.userID == 0 {
				ctx = actor.WithInternalActor(context.Background())
			} else {
				ctx = actor.WithActor(context.Background(), &actor.Actor{UID: test.userID})
			}

			err := CheckUserIsSiteAdmin(ctx, db, test.userID)

			if test.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckSiteAdminOrSameUser(t *testing.T) {
	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		if user, ok := usersMap[id]; ok {
			return user, nil
		} else {
			return nil, database.NewUserNotFoundError(id)
		}
	})
	users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
		userID := actor.FromContext(ctx).UID
		if user, ok := usersMap[userID]; ok {
			return user, nil
		} else {
			return nil, database.NewUserNotFoundError(userID)
		}
	})
	db.UsersFunc.SetDefaultReturn(users)

	tests := map[string]struct {
		ctxUserID     int32
		subjectUserID int32
		wantErr       bool
		err           error
	}{
		"internal user": {
			ctxUserID: 0,
			wantErr:   false,
		},
		"site admin checking for self": {
			ctxUserID:     1,
			subjectUserID: 1,
			wantErr:       false,
		},
		"site admin checking for other user": {
			ctxUserID:     1,
			subjectUserID: 100,
			wantErr:       false,
		},
		"same user": {
			ctxUserID:     100,
			subjectUserID: 100,
			wantErr:       false,
		},
		"non site admin checking for other": {
			ctxUserID:     100,
			subjectUserID: 99,
			wantErr:       true,
			err:           ErrMustBeSiteAdminOrSameUser,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var ctx context.Context
			if test.ctxUserID == 0 {
				ctx = actor.WithInternalActor(context.Background())
			} else {
				ctx = actor.WithActor(context.Background(), &actor.Actor{UID: test.ctxUserID})
			}

			err := CheckSiteAdminOrSameUser(ctx, db, test.subjectUserID)

			if test.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckSameUser(t *testing.T) {
	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
		userID := actor.FromContext(ctx).UID
		if user, ok := usersMap[userID]; ok {
			return user, nil
		} else {
			return nil, database.NewUserNotFoundError(userID)
		}
	})
	db.UsersFunc.SetDefaultReturn(users)

	tests := map[string]struct {
		userID  int32
		wantErr bool
		err     error
	}{
		"internal user": {
			userID:  0,
			wantErr: false,
		},
		"same user": {
			userID:  1,
			wantErr: false,
		},
		"some other user": {
			userID:  100,
			wantErr: true,
			err:     &InsufficientAuthorizationError{Message: "must be authenticated as user with id 100"},
		},
	}

	// Current user is always either internal or with ID=1 in this test.
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var ctx context.Context
			if test.userID == 0 {
				ctx = actor.WithInternalActor(context.Background())
			} else {
				ctx = actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			}

			err := CheckSameUser(ctx, test.userID)

			if test.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCurrentUser(t *testing.T) {
	db := dbmocks.NewMockDB()
	users := dbmocks.NewMockUserStore()
	sampleError := errors.New("oops")
	users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
		userID := actor.FromContext(ctx).UID
		if userID == 1337 {
			return nil, sampleError
		}
		if user, ok := usersMap[userID]; ok {
			return user, nil
		} else {
			return nil, database.NewUserNotFoundError(userID)
		}
	})
	db.UsersFunc.SetDefaultReturn(users)

	tests := map[string]struct {
		userID  int32
		wantErr bool
		err     error
	}{
		"found user": {
			userID:  1,
			wantErr: false,
		},
		"not found user": {
			userID:  0,
			wantErr: false,
		},
		"db error": {
			userID:  1337,
			wantErr: true,
			err:     sampleError,
		},
	}

	// Current user is always either internal or with ID=1 in this test.
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: test.userID})

			haveUser, err := CurrentUser(ctx, db)

			if test.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, test.err.Error())
			} else if test.userID == 0 {
				require.NoError(t, err)
				require.Nil(t, haveUser)
			} else {
				require.NoError(t, err)
				require.Equal(t, haveUser.ID, test.userID)
			}
		})
	}
}
