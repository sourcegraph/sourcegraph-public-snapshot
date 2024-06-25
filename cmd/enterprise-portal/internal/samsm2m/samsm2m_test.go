package samsm2m

import (
	"context"
	"net/http"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type mockSAMSClient struct {
	result *sams.IntrospectTokenResponse
	error  error
}

func (m mockSAMSClient) IntrospectSAMSToken(context.Context, string) (*sams.IntrospectTokenResponse, error) {
	return m.result, m.error
}

type request map[string]string

func (r request) Header() http.Header {
	h := make(http.Header)
	for k, v := range r {
		h.Add(k, v)
	}
	return h
}

func TestRequireScope(t *testing.T) {
	requiredScope := EnterprisePortalScope("codyaccess", scopes.ActionRead)

	for _, tc := range []struct {
		name       string
		metadata   map[string]string
		samsClient TokenIntrospector
		wantErr    autogold.Value
	}{
		{
			name:       "no metadata",
			metadata:   nil,
			samsClient: nil, // will not be used
			wantErr:    autogold.Expect("unauthenticated: no authorization header"),
		},
		{
			name:       "no authorization header",
			metadata:   map[string]string{"somethingelse": "foobar"},
			samsClient: nil, // will not be used
			wantErr:    autogold.Expect("unauthenticated: no authorization header"),
		},
		{
			name:       "malformed authorization header",
			metadata:   map[string]string{"authorization": "bearer"},
			samsClient: nil, // will not be used
			wantErr:    autogold.Expect("unauthenticated: invalid authorization header: token type missing in Authorization header"),
		},
		{
			name:       "token ok, introspect failed",
			metadata:   map[string]string{"authorization": "bearer foobar"},
			samsClient: mockSAMSClient{error: errors.New("introspection failed")},
			wantErr:    autogold.Expect("internal: unable to validate token"),
		},
		{
			name:       "token ok, but inactive",
			metadata:   map[string]string{"authorization": "bearer foobar"},
			samsClient: mockSAMSClient{result: &sams.IntrospectTokenResponse{Active: false}},
			wantErr:    autogold.Expect("permission_denied: permission denied"),
		},
		{
			name:     "token ok and active, but invalid scope",
			metadata: map[string]string{"authorization": "bearer foobar"},
			samsClient: mockSAMSClient{result: &sams.IntrospectTokenResponse{
				Active: true,
				Scopes: scopes.ToScopes([]string{"foo", "bar"}),
			}},
			wantErr: autogold.Expect("permission_denied: insufficient scope"),
		},
		{
			name:     "token ok and active and valid scope",
			metadata: map[string]string{"authorization": "bearer foobar"},
			samsClient: mockSAMSClient{
				result: &sams.IntrospectTokenResponse{
					Active: true,
					Scopes: append(scopes.ToScopes([]string{"foo", "bar"}), requiredScope),
				},
			},
			wantErr: nil, // success
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := RequireScope(ctx, logtest.Scoped(t), tc.samsClient, requiredScope, request(tc.metadata))
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				tc.wantErr.Equal(t, err.Error())
			}
		})
	}
}
