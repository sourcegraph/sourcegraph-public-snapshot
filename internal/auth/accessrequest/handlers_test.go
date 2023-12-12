package accessrequest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRequestAccess(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.NoOp(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	handler := HandleRequestAccess(logger, db)

	t.Run("accessRequest feature is disabled", func(t *testing.T) {
		falseVal := false
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthAccessRequest: &schema.AuthAccessRequest{
					Enabled: &falseVal,
				},
			},
		})
		t.Cleanup(func() { conf.Mock(nil) })

		req, err := http.NewRequest(http.MethodPost, "/-/request-access", strings.NewReader(`{}`))
		require.NoError(t, err)

		res := httptest.NewRecorder()
		handler(res, req)
		assert.Equal(t, http.StatusForbidden, res.Code)
		assert.Equal(t, "experimental feature accessRequests is disabled, but received request\n", res.Body.String())
	})

	t.Run("builtin signup enabled", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{
						Builtin: &schema.BuiltinAuthProvider{
							Type:        "builtin",
							AllowSignup: true,
						},
					},
				},
			},
		})
		t.Cleanup(func() { conf.Mock(nil) })

		req, err := http.NewRequest(http.MethodPost, "/-/request-access", strings.NewReader(`{}`))
		require.NoError(t, err)

		res := httptest.NewRecorder()
		handler(res, req)
		assert.Equal(t, http.StatusConflict, res.Code)
		assert.Equal(t, "Use sign up instead.\n", res.Body.String())
	})

	t.Run("invalid email", func(t *testing.T) {
		// test incorrect email
		req, err := http.NewRequest(http.MethodPost, "/-/request-access", strings.NewReader(`{"email": "a1-example.com", "name": "a1", "additionalInfo": "a1"}`))
		require.NoError(t, err)
		res := httptest.NewRecorder()
		handler(res, req)
		assert.Equal(t, http.StatusUnprocessableEntity, res.Code)

		// test empty email
		req, err = http.NewRequest(http.MethodPost, "/-/request-access", strings.NewReader(`{"name": "a1", "additionalInfo": "a1"}}`))
		require.NoError(t, err)
		res = httptest.NewRecorder()
		handler(res, req)
		assert.Equal(t, http.StatusUnprocessableEntity, res.Code)
	})

	t.Run("existing user's email", func(t *testing.T) {
		// test that no explicit error is returned if the email is already in the users table
		newUser := database.NewUser{
			Username:        "u1",
			Email:           "u1@example.com",
			EmailIsVerified: true,
		}
		db.Users().Create(context.Background(), newUser)
		req, err := http.NewRequest(http.MethodPost, "/-/request-access", strings.NewReader(fmt.Sprintf(`{"email": "%s", "name": "u1", "additionalInfo": "u1"}`, newUser.Email)))
		require.NoError(t, err)
		res := httptest.NewRecorder()
		handler(res, req)
		assert.Equal(t, http.StatusCreated, res.Code)

		_, err = db.AccessRequests().GetByEmail(context.Background(), newUser.Email)
		require.Error(t, err)
		require.Equal(t, errcode.IsNotFound(err), true)
	})

	t.Run("existing access requests's email", func(t *testing.T) {
		// test that no explicit error is returned if the email is already in the access requests table
		accessRequest := types.AccessRequest{
			Name:  "a1",
			Email: "a1@example.com",
		}
		db.AccessRequests().Create(context.Background(), &accessRequest)
		_, err := db.AccessRequests().GetByEmail(context.Background(), accessRequest.Email)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/-/request-access", strings.NewReader(fmt.Sprintf(`{"email": "%s", "name": "%s", "additionalInfo": "%s"}`, accessRequest.Email, accessRequest.Name, accessRequest.AdditionalInfo)))
		require.NoError(t, err)
		res := httptest.NewRecorder()
		handler(res, req)
		assert.Equal(t, http.StatusCreated, res.Code)
	})

	t.Run("correct inputs", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/-/request-access", strings.NewReader(`{"email": "a2@example.com", "name": "a2", "additionalInfo": "af2"}`))
		req = req.WithContext(context.Background())
		require.NoError(t, err)

		res := httptest.NewRecorder()
		handler(res, req)
		assert.Equal(t, http.StatusCreated, res.Code)

		accessRequest, err := db.AccessRequests().GetByEmail(context.Background(), "a2@example.com")
		require.NoError(t, err)
		assert.Equal(t, "a2", accessRequest.Name)
		assert.Equal(t, "a2@example.com", accessRequest.Email)
		assert.Equal(t, "af2", accessRequest.AdditionalInfo)
	})
}
