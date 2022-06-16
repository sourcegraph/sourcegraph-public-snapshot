package enforcement

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestEnforcement_PreCreateUser(t *testing.T) {
	if !licensing.EnforceTiers {
		licensing.EnforceTiers = true
		defer func() { licensing.EnforceTiers = false }()
	}

	expiresAt := time.Now().Add(time.Hour)
	tests := []struct {
		license         *license.Info
		activeUserCount int
		wantErr         bool
	}{
		// See the impl for why we treat UserCount == 0 as unlimited.
		{
			license:         &license.Info{UserCount: 0, ExpiresAt: expiresAt},
			activeUserCount: 5,
			wantErr:         false,
		},

		// Non-true-up licenses.
		{
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 0,
			wantErr:         false,
		},
		{
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 5,
			wantErr:         false,
		},
		{
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 9,
			wantErr:         false,
		},
		{
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 10,
			wantErr:         true,
		},
		{
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 11,
			wantErr:         true,
		},
		{
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 12,
			wantErr:         true,
		},

		// True-up licenses.
		{
			license:         &license.Info{Tags: []string{licensing.TrueUpUserCountTag}, UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 5,
			wantErr:         false,
		},
		{
			license:         &license.Info{Tags: []string{licensing.TrueUpUserCountTag}, UserCount: 10, ExpiresAt: expiresAt},
			activeUserCount: 15,
			wantErr:         false,
		},

		// License expired
		{
			license: &license.Info{ExpiresAt: time.Now().Add(-1 * time.Minute)},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("license %s with %d active users", test.license, test.activeUserCount), func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

			users := database.NewStrictMockUserStore()
			users.CountFunc.SetDefaultReturn(test.activeUserCount, nil)

			db := database.NewStrictMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			err := NewBeforeCreateUserHook()(context.Background(), db)
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Errorf("got error %v, want %v", gotErr, test.wantErr)
			}
		})
	}
}

func TestEnforcement_AfterCreateUser(t *testing.T) {
	if !licensing.EnforceTiers {
		licensing.EnforceTiers = true
		defer func() { licensing.EnforceTiers = false }()
	}

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
			name:                  "no license sets new user to be site admin",
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
				err := NewAfterCreateUserHook()(context.Background(), database.NewDBWith(basestore.NewWithHandle(db)), user)
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
	if !licensing.EnforceTiers {
		licensing.EnforceTiers = true
		defer func() { licensing.EnforceTiers = false }()
	}

	tests := []struct {
		name        string
		license     *license.Info
		isSiteAdmin bool
		wantErr     bool
	}{
		{
			name:        "promote to site admin is OK",
			isSiteAdmin: true,
			wantErr:     false,
		},
		{
			name:        "revoke site admin with a valid license is OK",
			license:     &license.Info{UserCount: 10},
			isSiteAdmin: false,
			wantErr:     false,
		},
		{
			name:        "revoke site admin without a license is not OK",
			isSiteAdmin: false,
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
