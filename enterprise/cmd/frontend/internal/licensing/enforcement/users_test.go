package enforcement

import (
	"context"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/cloud"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		license      *license.Info
		setSiteAdmin bool
	}{
		{
			name:         "with a valid license",
			license:      &license.Info{UserCount: 10},
			setSiteAdmin: false,
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
			setSiteAdmin: false,
		},
		{
			name:         "free license sets new user to be site admin",
			license:      &licensing.GetFreeLicenseInfo().Info,
			setSiteAdmin: true,
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

			db, usersStore, roleStore, userRoleStore := mockDBAndStores(t)
			user := new(types.User)

			hook := NewAfterCreateUserHook()
			if hook != nil {
				err := NewAfterCreateUserHook()(context.Background(), db, user)
				if err != nil {
					t.Fatal(err)
				}
			}

			if test.setSiteAdmin {
				mockrequire.CalledOnce(t, usersStore.SetIsSiteAdminFunc)
				mockrequire.CalledOnce(t, roleStore.GetFunc)
				mockrequire.CalledOnce(t, userRoleStore.AssignFunc)
			}
		})
	}
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

func mockDBAndStores(t *testing.T) (*database.MockDB, *database.MockUserStore, *database.MockRoleStore, *database.MockUserRoleStore) {
	t.Helper()

	usersStore := database.NewMockUserStore()
	usersStore.SetIsSiteAdminFunc.SetDefaultReturn(nil)

	sr := &types.Role{ID: 1}

	roleStore := database.NewMockRoleStore()
	roleStore.GetFunc.SetDefaultHook(func(ctx context.Context, gro database.GetRoleOpts) (*types.Role, error) {
		if gro.Name != string(types.SiteAdministratorSystemRole) {
			t.Fatalf("expected Roles.Get to be called with SITE_ADMINISTRATOR name, got %s", gro.Name)
		}

		return sr, nil
	})

	userRoleStore := database.NewMockUserRoleStore()
	userRoleStore.AssignFunc.SetDefaultHook(func(ctx context.Context, curo database.AssignUserRoleOpts) (*types.UserRole, error) {
		if curo.RoleID != sr.ID {
			t.Fatalf("expected UserRoles.Create() to be called with roleID: %d, got %d", sr.ID, curo.RoleID)
		}
		return &types.UserRole{}, nil
	})

	db := database.NewMockDB()
	db.WithTransactFunc.SetDefaultHook(func(ctx context.Context, f func(database.DB) error) error {
		return f(db)
	})
	db.UsersFunc.SetDefaultReturn(usersStore)
	db.RolesFunc.SetDefaultReturn(roleStore)
	db.UserRolesFunc.SetDefaultReturn(userRoleStore)

	return db, usersStore, roleStore, userRoleStore
}
