package accessrequest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/schema"
)
func TestRequestAccess(t *testing.T) {
	logger := logtest.NoOp(t)
	users := database.NewMockUserStore()
	accessRequests := database.NewMockAccessRequestStore()
	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.AccessRequestsFunc.SetDefaultReturn(accessRequests)
	db.EventLogsFunc.SetDefaultReturn(database.NewMockEventLogStore())
	db.SecurityEventLogsFunc.SetDefaultReturn(database.NewMockSecurityEventLogsStore())
	db.UserEmailsFunc.SetDefaultReturn(database.NewMockUserEmailsStore())

	handler := HandleRequestAccess(logger, db)

	t.Run("experimental feature disabled", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					AccessRequests: &schema.AccessRequests {
						Enabled: false,
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
		assert.Equal(t, "Request access is not enabled.\n", res.Body.String())
	})

	t.Run("builtin signup enabled", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{
						Builtin: &schema.BuiltinAuthProvider{
							Type: "builtin",
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
		assert.Equal(t, http.StatusBadRequest, res.Code)

		// test empty email
		req, err = http.NewRequest(http.MethodPost, "/-/request-access", strings.NewReader(`{"name": "a1", "additionalInfo": "a1"}}`))
		require.NoError(t, err)
		res = httptest.NewRecorder()
		handler(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("correct inputs", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/-/request-access", strings.NewReader(`{"email": "a1@example.com", "name": "a1", "additionalInfo": "a1"}`))
		req.WithContext(context.Background())
		require.NoError(t, err)

		res := httptest.NewRecorder()
		handler(res, req)
		assert.Equal(t, http.StatusCreated, res.Code)
	})
}

