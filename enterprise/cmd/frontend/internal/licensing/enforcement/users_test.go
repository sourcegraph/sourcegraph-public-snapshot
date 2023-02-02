package enforcement

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/cloud"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestEnforcement_PreCreateUser(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	tests := []struct {
		name            string
		license         *license.Info
		activeUserCount int
		mockSetup       func(*testing.T)
		spec            *extsvc.AccountSpec
		wantErr         bool
	}{
		// See the impl for why we treat UserCount == 0 as unlimited.
		{
			name:            "unlimited",
			license:         &license.Info{UserCount: 0, ExpiresAt: expiresAt},
			activeUserCount: 5,
			wantErr:         false,
		},

		{
			name:            "no true-up",
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 0,
			wantErr:         false,
		},
		{
			name:            "no true-up and not exceeded user count",
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 5,
			wantErr:         false,
		},
		{
			name:            "no true-up and exceeding user count",
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 10,
			wantErr:         true,
		},
		{
			name:            "no true-up and exceeded user count",
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 11,
			wantErr:         true,
		},

		{
			name:            "true-up and not exceeded user count",
			license:         &license.Info{Tags: []string{licensing.TrueUpUserCountTag}, UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 5,
			wantErr:         false,
		},
		{
			name:            "true-up and exceeded user count",
			license:         &license.Info{Tags: []string{licensing.TrueUpUserCountTag}, UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 15,
			wantErr:         false,
		},

		{
			name:    "license expired",
			license: &license.Info{ExpiresAt: time.Now().Add(-1 * time.Minute)},
			wantErr: true,
		},

		{
			name:            "exempt SOAP users",
			license:         &license.Info{UserCount: 10, ExpiresAt: time.Now().Add(-1 * time.Minute)}, // An expired license
			activeUserCount: 15,                                                                        // Exceeded free plan user count
			mockSetup: func(t *testing.T) {
				cloud.MockSiteConfig(
					t,
					&cloud.SchemaSiteConfig{
						AuthProviders: &cloud.SchemaAuthProviders{
							SourcegraphOperator: &cloud.SchemaAuthProviderSourcegraphOperator{},
						},
					},
				)
			},
			spec: &extsvc.AccountSpec{
				ServiceType: auth.SourcegraphOperatorProviderType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

			users := database.NewStrictMockUserStore()
			users.CountFunc.SetDefaultReturn(test.activeUserCount, nil)

			db := database.NewStrictMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			if test.mockSetup != nil {
				test.mockSetup(t)
			}

			err := NewBeforeCreateUserHook()(context.Background(), db, test.spec)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnforcement_AfterCreateUser(t *testing.T) {
	logger := logtest.Scoped(t)

	tests := []struct {
		name                  string
		setup                 func(t *testing.T)
		license               *license.Info
		wantCalledExecContext bool
		wantSiteAdmin         bool
	}{
		{
			name:                  "with a valid license",
			license:               &license.Info{UserCount: 10},
			wantCalledExecContext: false,
			wantSiteAdmin:         false,
		},
		{
			name: "dotcom mode should always do nothing",
			setup: func(t *testing.T) {
				orig := envvar.SourcegraphDotComMode()
				envvar.MockSourcegraphDotComMode(true)
				t.Cleanup(func() {
					envvar.MockSourcegraphDotComMode(orig)
				})
			},
			wantCalledExecContext: false,
			wantSiteAdmin:         false,
		},
		{
			name:                  "free license sets new user to be site admin",
			license:               &licensing.GetFreeLicenseInfo().Info,
			wantCalledExecContext: true,
			wantSiteAdmin:         true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(t)
			}

			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

			calledExecContext := false
			db := &fakeDB{
				execContext: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
					calledExecContext = true
					return nil, nil
				},
			}
			user := new(types.User)

			hook := NewAfterCreateUserHook()
			if hook != nil {
				err := NewAfterCreateUserHook()(context.Background(), database.NewDBWith(logger, basestore.NewWithHandle(db)), user)
				if err != nil {
					t.Fatal(err)
				}
			}

			if test.wantCalledExecContext != calledExecContext {
				t.Errorf("calledExecContext: want %v but got %v", test.wantCalledExecContext, calledExecContext)
			}
			if test.wantSiteAdmin != user.SiteAdmin {
				t.Errorf("siteAdmin: want %v but got %v", test.wantSiteAdmin, user.SiteAdmin)
			}
		})
	}
}

type fakeDB struct {
	basestore.TransactableHandle
	execContext func(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (db *fakeDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.execContext(ctx, query, args...)
}

func TestEnforcement_PreSetUserIsSiteAdmin(t *testing.T) {
	tests := []struct {
		name        string
		license     *license.Info
		isSiteAdmin bool
		wantErr     bool
	}{
		{
			name:        "promote to site admin with a valid license is OK",
			license:     &license.Info{ExpiresAt: time.Now().Add(1 * time.Hour)},
			isSiteAdmin: true,
			wantErr:     false,
		},
		{
			name:        "revoke site admin with a valid license is OK",
			license:     &license.Info{UserCount: 10, ExpiresAt: time.Now().Add(1 * time.Hour)},
			isSiteAdmin: false,
			wantErr:     false,
		},
		{
			name:        "revoke site admin without a license is not OK",
			isSiteAdmin: false,
			wantErr:     true,
		},
		{
			name:        "promote to site admin with expired license is not OK",
			license:     &license.Info{UserCount: 10, ExpiresAt: time.Now().Add(-1 * time.Hour)},
			isSiteAdmin: true,
			wantErr:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()
			err := NewBeforeSetUserIsSiteAdmin()(test.isSiteAdmin)
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Errorf("got error %v, want %v", gotErr, test.wantErr)
			}
		})
	}
}
