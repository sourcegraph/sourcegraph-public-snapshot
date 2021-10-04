package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestServeVerifyEmail(t *testing.T) {
	db := new(dbtesting.MockDB)

	t.Run("primary email is already set", func(t *testing.T) {
		database.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
			return &types.User{ID: 1}, nil
		}
		database.Mocks.UserEmails.Get = func(userID int32, email string) (emailCanonicalCase string, verified bool, err error) {
			return "alice@example.com", false, nil
		}
		database.Mocks.UserEmails.Verify = func(ctx context.Context, userID int32, email, code string) (bool, error) {
			return true, nil
		}
		database.Mocks.UserEmails.GetPrimaryEmail = func(ctx context.Context, id int32) (email string, verified bool, err error) {
			return "alice@example.com", true, nil
		}
		database.Mocks.UserEmails.SetPrimaryEmail = func(ctx context.Context, userID int32, email string) error {
			t.Error("SetPrimaryEmail should not be called")
			return nil
		}
		database.Mocks.Authz.GrantPendingPermissions = func(ctx context.Context, args *database.GrantPendingPermissionsArgs) error {
			return nil
		}
		defer func() {
			database.Mocks.Users = database.MockUsers{}
			database.Mocks.UserEmails = database.MockUserEmails{}
			database.Mocks.Authz = database.MockAuthz{}
		}()

		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		handler := serveVerifyEmail(db)
		handler(resp, req)
	})

	t.Run("primary email is not set", func(t *testing.T) {
		calledSetPrimaryEmail := false
		database.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
			return &types.User{ID: 1}, nil
		}
		database.Mocks.UserEmails.Get = func(userID int32, email string) (emailCanonicalCase string, verified bool, err error) {
			return "alice@example.com", false, nil
		}
		database.Mocks.UserEmails.Verify = func(ctx context.Context, userID int32, email, code string) (bool, error) {
			return true, nil
		}
		database.Mocks.UserEmails.GetPrimaryEmail = func(ctx context.Context, id int32) (email string, verified bool, err error) {
			return "", false, errors.New("primary email not found")
		}
		database.Mocks.UserEmails.SetPrimaryEmail = func(ctx context.Context, userID int32, email string) error {
			calledSetPrimaryEmail = true
			return nil
		}
		database.Mocks.Authz.GrantPendingPermissions = func(ctx context.Context, args *database.GrantPendingPermissionsArgs) error {
			return nil
		}
		defer func() {
			database.Mocks.Users = database.MockUsers{}
			database.Mocks.UserEmails = database.MockUserEmails{}
			database.Mocks.Authz = database.MockAuthz{}
		}()

		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		handler := serveVerifyEmail(db)
		handler(resp, req)

		assert.True(t, calledSetPrimaryEmail, "SetPrimaryEmail should be called")
	})
}
