package enterpriseportal

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type mockClientCredentials struct{}

func (mockClientCredentials) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: "mockClientCredentialsToken",
		Expiry:      time.Now().Add(time.Hour),
	}, nil
}

func TestSiteAdminProxy(t *testing.T) {
	for _, tc := range []struct {
		name             string
		actor            *actor.Actor
		actorIsSiteAdmin bool
	}{
		{
			name: "site admin",
			actor: &actor.Actor{
				UID: 1,
			},
			actorIsSiteAdmin: true,
		},
		{
			name: "not site admin",
			actor: &actor.Actor{
				UID: 1,
			},
			actorIsSiteAdmin: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const (
				proxyPrefix = "/.api/prefix"
				requestPath = "/enterpriseportal.subscriptions.v1.SubscriptionsService/ListEnterpriseSubscriptions"
				requestBody = `{"filters":[{"filter":{"is_archived":false}}]}`
			)

			var wasProxied bool
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				wasProxied = true
				assert.Equal(t, "Bearer mockClientCredentialsToken", r.Header.Get("authorization"))
				assert.Empty(t, r.Cookies())
				assert.Equal(t, requestPath, r.URL.Path)

				data, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				defer r.Body.Close()
				assert.Equal(t, requestBody, string(data))

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			}))
			t.Cleanup(target.Close)

			users := dbmocks.NewMockUserStore()
			users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
				if id == tc.actor.UID {
					return &types.User{
						ID:        id,
						Username:  t.Name(),
						SiteAdmin: tc.actorIsSiteAdmin,
					}, nil
				}
				return nil, errors.Newf("user %d not found", id)
			})
			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			proxy := newSiteAdminProxy(
				logtest.Scoped(t),
				db,
				mockClientCredentials{},
				"/.api/prefix",
				mustParseURL(target.URL))

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet,
				proxyPrefix+requestPath,
				bytes.NewReader([]byte(requestBody)))
			req.Header.Set("authorization", "token sgp_local_not_what_we_want")
			req.Header.Set("cookie", "name=value; name2=value2; name3=value3")
			req = req.WithContext(actor.WithActor(context.Background(), tc.actor))

			proxy.ServeHTTP(recorder, req)

			assert.Equal(t, tc.actorIsSiteAdmin, wasProxied,
				"received proxied request")
			if wasProxied {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assert.Equal(t, "ok", recorder.Body.String())
			} else {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			}
		})
	}
}
