package graphqlbackend

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestSiteConfiguration(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		t.Run("ReturnSafeConfigsOnly is false", func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)
			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			_, err := newSchemaResolver(db, gitserver.NewTestClient(t), nil).Site().Configuration(ctx, &SiteConfigurationArgs{
				ReturnSafeConfigsOnly: pointers.Ptr(false),
			})

			if err == nil || !errors.Is(err, auth.ErrMustBeSiteAdmin) {
				t.Fatalf("err: want %q but got %v", auth.ErrMustBeSiteAdmin, err)
			}
		})

		t.Run("ReturnSafeConfigsOnly is true", func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)
			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			r, err := newSchemaResolver(db, gitserver.NewTestClient(t), nil).Site().Configuration(ctx, &SiteConfigurationArgs{
				ReturnSafeConfigsOnly: pointers.Ptr(true),
			})
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}

			// all other fields except `EffectiveContents` should not be visible
			_, err = r.ID(ctx)
			if err == nil || !errors.Is(err, auth.ErrMustBeSiteAdmin) {
				t.Fatalf("err: want %q but got %v", auth.ErrMustBeSiteAdmin, err)
			}

			_, err = r.History(ctx, nil)
			if err == nil || !errors.Is(err, auth.ErrMustBeSiteAdmin) {
				t.Fatalf("err: want %q but got %v", auth.ErrMustBeSiteAdmin, err)
			}

			_, err = r.ValidationMessages(ctx)
			if err == nil || !errors.Is(err, auth.ErrMustBeSiteAdmin) {
				t.Fatalf("err: want %q but got %v", auth.ErrMustBeSiteAdmin, err)
			}

			_, err = r.EffectiveContents(ctx)
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}
		})
	})

	t.Run("authenticated as admin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{
			ID:        1,
			SiteAdmin: true,
		}, nil)

		siteConfig := &database.SiteConfig{
			ID:               1,
			AuthorUserID:     1,
			Contents:         `{"batchChanges.rolloutWindows": [{"rate":"unlimited"}]}`,
			RedactedContents: `{"batchChanges.rolloutWindows": [{"rate":"unlimited"}]}`,
		}
		conf := dbmocks.NewMockConfStore()
		conf.SiteGetLatestFunc.SetDefaultReturn(siteConfig, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.ConfFunc.SetDefaultReturn(conf)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		t.Run("ReturnSafeConfigsOnly is false", func(t *testing.T) {
			r, err := newSchemaResolver(db, gitserver.NewTestClient(t), nil).Site().Configuration(ctx, &SiteConfigurationArgs{
				ReturnSafeConfigsOnly: pointers.Ptr(false),
			})
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}

			sID, err := r.ID(ctx)
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}
			if sID != int32(siteConfig.ID) {
				t.Fatalf("expected config ID to be %d, got %d", sID, int32(siteConfig.ID))
			}

			_, err = r.History(ctx, nil)
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}

			_, err = r.ValidationMessages(ctx)
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}

			_, err = r.EffectiveContents(ctx)
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}
		})

		t.Run("ReturnSafeConfigsOnly is true", func(t *testing.T) {
			r, err := newSchemaResolver(db, gitserver.NewTestClient(t), nil).Site().Configuration(ctx, &SiteConfigurationArgs{
				ReturnSafeConfigsOnly: pointers.Ptr(true),
			})
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}

			_, err = r.ID(ctx)
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}

			_, err = r.History(ctx, nil)
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}

			_, err = r.ValidationMessages(ctx)
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}

			_, err = r.EffectiveContents(ctx)
			if err != nil {
				t.Fatalf("err: want nil but got %v", err)
			}
		})
	})
}

func TestSiteConfigurationHistory(t *testing.T) {
	stubs := setupSiteConfigStubs(t)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: stubs.users[0].ID})
	schemaResolver, err := newSchemaResolver(stubs.db, gitserver.NewTestClient(t), nil).Site().Configuration(ctx, &SiteConfigurationArgs{})
	if err != nil {
		t.Fatalf("failed to create schemaResolver: %v", err)
	}

	testCases := []struct {
		name                  string
		args                  *gqlutil.ConnectionResolverArgs
		expectedSiteConfigIDs []int32
	}{
		{
			name:                  "first: 2",
			args:                  &gqlutil.ConnectionResolverArgs{First: pointers.Ptr(int32(2))},
			expectedSiteConfigIDs: []int32{6, 4},
		},
		{
			name:                  "first: 6 (exact number of items that exist in the database)",
			args:                  &gqlutil.ConnectionResolverArgs{First: pointers.Ptr(int32(6))},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			name:                  "first: 20 (more items than what exists in the database)",
			args:                  &gqlutil.ConnectionResolverArgs{First: pointers.Ptr(int32(20))},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			name:                  "last: 2",
			args:                  &gqlutil.ConnectionResolverArgs{Last: pointers.Ptr(int32(2))},
			expectedSiteConfigIDs: []int32{2, 1},
		},
		{
			name:                  "last: 6 (exact number of items that exist in the database)",
			args:                  &gqlutil.ConnectionResolverArgs{Last: pointers.Ptr(int32(6))},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			name:                  "last: 20 (more items than what exists in the database)",
			args:                  &gqlutil.ConnectionResolverArgs{Last: pointers.Ptr(int32(20))},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			name: "first: 2, after: 4",
			args: &gqlutil.ConnectionResolverArgs{
				First: pointers.Ptr(int32(2)),
				After: pointers.Ptr(string(marshalSiteConfigurationChangeID(4))),
			},
			expectedSiteConfigIDs: []int32{3, 2},
		},
		{
			name: "first: 10, after: 4 (overflow)",
			args: &gqlutil.ConnectionResolverArgs{
				First: pointers.Ptr(int32(10)),
				After: pointers.Ptr(string(marshalSiteConfigurationChangeID(4))),
			},
			expectedSiteConfigIDs: []int32{3, 2, 1},
		},
		{
			name: "first: 10, after: 7 (same as get all items, but latest ID in DB is 6)",
			args: &gqlutil.ConnectionResolverArgs{
				First: pointers.Ptr(int32(10)),
				After: pointers.Ptr(string(marshalSiteConfigurationChangeID(7))),
			},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			name: "first: 10, after: 1 (beyond the last cursor in DB which is 1)",
			args: &gqlutil.ConnectionResolverArgs{
				First: pointers.Ptr(int32(10)),
				After: pointers.Ptr(string(marshalSiteConfigurationChangeID(1))),
			},
			expectedSiteConfigIDs: []int32{},
		},
		{
			name: "last: 2, before: 1",
			args: &gqlutil.ConnectionResolverArgs{
				Last:   pointers.Ptr(int32(2)),
				Before: pointers.Ptr(string(marshalSiteConfigurationChangeID(1))),
			},
			expectedSiteConfigIDs: []int32{3, 2},
		},
		{
			name: "last: 10, before: 1 (overflow)",
			args: &gqlutil.ConnectionResolverArgs{
				Last:   pointers.Ptr(int32(10)),
				Before: pointers.Ptr(string(marshalSiteConfigurationChangeID(1))),
			},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2},
		},
		{
			name: "last: 10, before: 0 (same as get all items, but oldest ID in DB is 1)",
			args: &gqlutil.ConnectionResolverArgs{
				Last:   pointers.Ptr(int32(10)),
				Before: pointers.Ptr(string(marshalSiteConfigurationChangeID(0))),
			},
			expectedSiteConfigIDs: []int32{6, 4, 3, 2, 1},
		},
		{
			name: "last: 10, before: 7 (beyond the latest cursor in DB which is 6)",
			args: &gqlutil.ConnectionResolverArgs{
				Last:   pointers.Ptr(int32(10)),
				Before: pointers.Ptr(string(marshalSiteConfigurationChangeID(7))),
			},
			expectedSiteConfigIDs: []int32{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			connectionResolver, err := schemaResolver.History(ctx, tc.args)
			if err != nil {
				t.Fatalf("failed to get history: %v", err)
			}

			siteConfigChangeResolvers, err := connectionResolver.Nodes(ctx)
			if err != nil {
				t.Fatalf("failed to get nodes: %v", err)
			}

			siteConfigChangeResolverIDs := make([]int32, len(siteConfigChangeResolvers))
			for i, s := range siteConfigChangeResolvers {
				siteConfigChangeResolverIDs[i] = s.siteConfig.ID
			}

			if diff := cmp.Diff(tc.expectedSiteConfigIDs, siteConfigChangeResolverIDs, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("unexpected site config ids (-want +got):%s\n", diff)
			}
		})
	}

}

func TestIsRequiredOutOfBandMigration(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion oobmigration.Version
		latestVersion  oobmigration.Version
		migration      oobmigration.Migration
		want           bool
	}{
		{
			name:          "not deprecated",
			latestVersion: oobmigration.Version{Major: 4, Minor: 3},
			migration:     oobmigration.Migration{},
			want:          false,
		},
		{
			name:          "deprecated but finished",
			latestVersion: oobmigration.Version{Major: 4, Minor: 3},
			migration: oobmigration.Migration{
				Deprecated: &oobmigration.Version{Major: 3, Minor: 43},
				Progress:   1,
			},
			want: false,
		},
		{
			name:          "deprecated after the latest",
			latestVersion: oobmigration.Version{Major: 4, Minor: 3},
			migration: oobmigration.Migration{
				Deprecated: &oobmigration.Version{Major: 4, Minor: 4},
			},
			want: false,
		},
		{
			name:          "deprecated at latest and unfinished",
			latestVersion: oobmigration.Version{Major: 4, Minor: 3},
			migration: oobmigration.Migration{
				Deprecated: &oobmigration.Version{Major: 4, Minor: 3},
			},
			want: true,
		},
		{
			name:          "deprecated prior to latest and unfinished",
			latestVersion: oobmigration.Version{Major: 4, Minor: 3},
			migration: oobmigration.Migration{
				Deprecated: &oobmigration.Version{Major: 4, Minor: 3},
				Progress:   1,
			},
			want: false,
		},
		{
			name:           "deprecated prior to current and unfinished",
			currentVersion: oobmigration.Version{Major: 4, Minor: 3},
			migration: oobmigration.Migration{
				Deprecated: &oobmigration.Version{Major: 3, Minor: 43},
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isRequiredOutOfBandMigration(test.currentVersion, test.latestVersion, test.migration)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_allowEdit(t *testing.T) {
	tests := []struct {
		name      string
		before    string
		after     string
		allowlist []string
		want      autogold.Value
		ok        bool
	}{
		{
			name:      "allowed",
			before:    `{}`,
			after:     `{"externalURL": "https://sg.local.com"}`,
			allowlist: []string{"externalURL"},
			ok:        true,
		},
		{
			name:   "not allowed",
			before: `{}`,
			after: `
{
  "observability.alerts": [
    {
      "level": "critical",
      "notifier": {
        "type": "slack",
        "url": "some-url",
        "username": "username"
      }
    },
  ]
}`,
			allowlist: []string{"externalURL"},
			want:      autogold.Expect([]string{"observability.alerts"}),
			ok:        false,
		},
		{
			name:   "nested and mixed",
			before: `{}`,
			after: `
{
  "experimentalFeatures": {
    "searchJobs": true,
  },
  "auth.providers": [
    {
      "type": "builtin"
    },
  ],
  "email.smtp": {
    "authentication": "PLAIN",
    "host": "smtp.company.local",
    "password": "password",
    "port": 587,
    "username": "username",
  },
}`,
			allowlist: []string{"auth.providers"},
			want:      autogold.Expect([]string{"email.smtp", "experimentalFeatures::searchJobs"}),
			ok:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := allowEdit(tt.before, tt.after, tt.allowlist)
			require.Equal(t, tt.ok, ok)
			if !ok {
				sort.Strings(got)
				tt.want.Equal(t, got)
			}
		})
	}
}
