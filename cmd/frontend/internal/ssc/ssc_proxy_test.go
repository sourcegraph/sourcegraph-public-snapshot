package ssc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSSCAPIProxy(t *testing.T) {
	// Fixed configuration values of the testHandler for easy reference.
	var (
		testURLPrefix     = "/.api/ssc/proxy"
		testCodyProConfig = schema.CodyProConfig{
			SamsBackendOrigin: "https://sams.sourcegraph.com:1234",
			SscBackendOrigin:  "https://ssc.sourcegraph.com:1234",
		}
	)

	testHandler := APIProxyHandler{
		CodyProConfig: &testCodyProConfig,
		DB:            nil,
		Logger:        logtest.NoOp(t),
		URLPrefix:     testURLPrefix,
	}

	const testUserID = int32(12345)

	// Confirm that we pull the Sourcegraph user details from the
	// incomming request context correctly, and return the expected
	// errors.
	t.Run("getUserIDFromContext", func(t *testing.T) {
		// Call getUserIDFromContext with the supplied actor in the request context.
		runTest := func(a *actor.Actor) (int32, error) {
			ctx := actor.WithActor(context.Background(), a)
			return testHandler.getUserIDFromContext(ctx)
		}

		t.Run("Success", func(t *testing.T) {
			userID, err := runTest(&actor.Actor{
				UID: testUserID,
			})
			assert.EqualValues(t, testUserID, userID)
			assert.NoError(t, err)
		})
		t.Run("ErrorNoActor", func(t *testing.T) {
			_, err := runTest(nil)
			assert.ErrorContains(t, err, "no credentials available")
		})
		t.Run("ErrorNonNuman", func(t *testing.T) {
			_, err := runTest(&actor.Actor{
				UID:                 testUserID,
				SourcegraphOperator: true,
			})
			assert.ErrorContains(t, err, "request not made on behalf of a user")
		})
	})

	t.Run("buildProxyRequest", func(t *testing.T) {
		t.Run("Basics", func(t *testing.T) {
			reqBody := `
			{
				"addMember": {
					"accountID": "...",
					"role": "member"
				}
			}
			`
			req := httptest.NewRequest(
				http.MethodPatch,
				testURLPrefix+"teams/current/members?api-version=2",
				strings.NewReader(reqBody))
			req.Header.Add("Authorization", "Bearer original-access-token")
			req.Header.Add("X-Sourcegraph", "Another random header")

			proxyReq, err := testHandler.buildProxyRequest(req, "sams-access-token")
			require.NoError(t, err)

			// HTTP method and URL.
			assert.Equal(t, http.MethodPatch, proxyReq.Method)
			wantURL := testCodyProConfig.SscBackendOrigin + "/cody/api/v1/teams/current/members?api-version=2"
			assert.Equal(t, wantURL, proxyReq.URL.String())

			// We do NOT pass along all HTTP request headers, and only pass the SAMS auth token.
			assert.Equal(t, 1, len(proxyReq.Header))
			assert.Equal(t, "Bearer sams-access-token", proxyReq.Header.Get("Authorization"))

			// The request body was copied as well.
			require.NotNil(t, proxyReq.Body)
			proxyReqBodyBytes, err := io.ReadAll(proxyReq.Body)
			require.NoError(t, err)
			assert.Equal(t, reqBody, string(proxyReqBodyBytes))
		})
		t.Run("URLRewriting", func(t *testing.T) {
			tests := []struct {
				ReqURLPath      string
				WantURLPath     string
				WantQueryParams string
			}{
				{
					testURLPrefix + "source/url/path",
					"/cody/api/v1/source/url/path", "",
				},
				// Confirm URL query parameters are handled correctly.
				{
					testURLPrefix + "endpoint/alpha?param=1234&continuationToken='xxxxxxx'",
					// The query parameters will be canonicalized by URL encoding and
					// sorting by key.
					"/cody/api/v1/endpoint/alpha", "continuationToken=%27xxxxxxx%27&param=1234",
				},

				// Pathalogical case where the URL prefix is different from
				// where the HTTP handler is registered.
				{
					"/users/settings",
					"/cody/api/v1/users/settings", "",
				},
				{
					"users/settings",
					// BUG? A quirk of how this test runs. We prefix the input string with
					// "https://sourcegraph.com", so "sourcegraph.comuser/settings" comes
					// out looking weird...
					"/cody/api/v1/settings", "",
				},

				// URL path schenanigans. We escape the incomming URL so that the proxied request
				// always has the required route prefix.
				{
					// The source URL path is cleaned, so the "folder1/folder2" gets removed.
					testURLPrefix + "folder1/folder2/../../some-other-folder",
					"/cody/api/v1/some-other-folder", "",
				},
				{
					// But the cleaning is only applied to the user-controlled URL path.
					// So we will always have the SSC-side URL prefix.
					testURLPrefix + "../../../../some-unknown-parent-folder",
					"/cody/api/v1/some-unknown-parent-folder", "",
				},
				{
					// Another instance of URL canonicalization, and it not
					// escaping the "/cody/api/v1" prefix.
					testURLPrefix + "legit/../../../not-legit",
					"/cody/api/v1/not-legit", "",
				},
			}
			for i, test := range tests {
				inURL := "https://sourcegraph.com" + test.ReqURLPath
				sourceReq := httptest.NewRequest(http.MethodGet, inURL, nil /* body */)
				proxyReq, err := testHandler.buildProxyRequest(sourceReq, "")
				assert.NoError(t, err, "URL rewriting scenario %d", i)

				assert.True(t, strings.HasPrefix(proxyReq.URL.String(), testCodyProConfig.SscBackendOrigin))
				assert.Equal(t, test.WantURLPath, proxyReq.URL.Path)
				assert.Equal(t, test.WantQueryParams, proxyReq.URL.RawQuery)
			}
		})
	})

	t.Run("getSAMSCredentialsForUser", func(t *testing.T) {
		testToken := oauth2.Token{
			RefreshToken: "refresh-token",
			AccessToken:  "access-token",
			TokenType:    "token-type",
			Expiry:       time.Now(),
		}
		testTokenJSON, err := json.Marshal(testToken)
		require.NoError(t, err)

		const testAccountID = int32(1)
		validSAMSIdentity := &extsvc.Account{
			ID:     testAccountID,
			UserID: testUserID,
			AccountData: extsvc.AccountData{
				AuthData: extsvc.NewUnencryptedData(testTokenJSON),
			},
		}

		t.Run("Success", func(t *testing.T) {
			// Setup the database mocks to return the test token.
			mockDB := dbmocks.NewMockDB()
			testHandler.DB = mockDB

			mockUEA := dbmocks.NewMockUserExternalAccountsStore()
			mockDB.UserExternalAccountsFunc.SetDefaultReturn(mockUEA)

			// Run a quick test before we mock the List call. Confirm error that the user
			// has no available SAMS identity.
			t.Run("ErrorNoSAMSIdentity", func(t *testing.T) {
				ctx := context.Background()
				_, _, err := testHandler.getSAMSCredentialsForUser(ctx, testUserID)
				assert.ErrorContains(t, err, "user does not have a SAMS identity")
			})

			// List user accounts
			mockUEA.ListFunc.PushHook(func(_ context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
				assert.Equal(t, testUserID, opts.UserID)
				assert.Equal(t, "openidconnect", opts.ServiceType)
				assert.Equal(t, testCodyProConfig.SamsBackendOrigin, opts.ServiceID)

				return []*extsvc.Account{
					validSAMSIdentity,
					// Bogus user identities that we are implicity confirming won't
					// cause problems. (We only get the first.)
					nil,
					nil,
				}, nil
			})
			// Get the first SAMS identity / external user account.
			mockUEA.GetFunc.PushHook(func(_ context.Context, id int32) (*extsvc.Account, error) {
				assert.Equal(t, testAccountID, id)
				return validSAMSIdentity, nil
			})

			// Try to get the user's SAMS creds given this.
			ctx := context.Background()
			ident, token, err := testHandler.getSAMSCredentialsForUser(ctx, testUserID)
			require.NoError(t, err)

			assert.Equal(t, validSAMSIdentity.ID, ident.ID)
			assert.Equal(t, validSAMSIdentity.ServiceID, ident.ServiceID)

			assert.Equal(t, testToken.AccessToken, token.AccessToken)
			assert.Equal(t, testToken.RefreshToken, token.RefreshToken)
			assert.WithinDuration(t, testToken.Expiry, token.Expiry, time.Second)
		})
	})
}
