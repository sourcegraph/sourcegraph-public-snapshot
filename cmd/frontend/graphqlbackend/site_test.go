package graphqlbackend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestSiteConfiguration(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		_, err := newSchemaResolver(db, gitserver.NewClient(), jobutil.NewUnimplementedEnterpriseJobs()).Site().Configuration(ctx)

		if err == nil || !errors.Is(err, auth.ErrMustBeSiteAdmin) {
			t.Fatalf("err: want %q but got %v", auth.ErrMustBeSiteAdmin, err)
		}
	})
}

func TestSiteConfigurationHistory(t *testing.T) {
	stubs := setupSiteConfigStubs(t)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: stubs.users[0].ID})
	schemaResolver, err := newSchemaResolver(stubs.db, gitserver.NewClient(), jobutil.NewUnimplementedEnterpriseJobs()).Site().Configuration(ctx)
	if err != nil {
		t.Fatalf("failed to create schemaResolver: %v", err)
	}

	testCases := []struct {
		name                  string
		args                  *graphqlutil.ConnectionResolverArgs
		expectedSiteConfigIDs []int32
	}{
		{
			name:                  "first: 2",
			args:                  &graphqlutil.ConnectionResolverArgs{First: int32Ptr(2)},
			expectedSiteConfigIDs: []int32{5, 4},
		},
		{
			name:                  "first: 5 (exact number of items that exist in the database)",
			args:                  &graphqlutil.ConnectionResolverArgs{First: int32Ptr(5)},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
		},
		{
			name:                  "first: 20 (more items than what exists in the database)",
			args:                  &graphqlutil.ConnectionResolverArgs{First: int32Ptr(20)},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
		},
		{
			name:                  "last: 2",
			args:                  &graphqlutil.ConnectionResolverArgs{Last: int32Ptr(2)},
			expectedSiteConfigIDs: []int32{2, 1},
		},
		{
			name:                  "last: 5 (exact number of items that exist in the database)",
			args:                  &graphqlutil.ConnectionResolverArgs{Last: int32Ptr(5)},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
		},
		{
			name:                  "last: 20 (more items than what exists in the database)",
			args:                  &graphqlutil.ConnectionResolverArgs{Last: int32Ptr(20)},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
		},
		{
			name: "first: 2, after: 4",
			args: &graphqlutil.ConnectionResolverArgs{
				First: int32Ptr(2),
				After: stringPtr(string(marshalSiteConfigurationChangeID(4))),
			},
			expectedSiteConfigIDs: []int32{3, 2},
		},
		{
			name: "first: 10, after: 4 (overflow)",
			args: &graphqlutil.ConnectionResolverArgs{
				First: int32Ptr(10),
				After: stringPtr(string(marshalSiteConfigurationChangeID(4))),
			},
			expectedSiteConfigIDs: []int32{3, 2, 1},
		},
		{
			name: "first: 10, after: 6 (same as get all items, but latest ID in DB is 5)",
			args: &graphqlutil.ConnectionResolverArgs{
				First: int32Ptr(10),
				After: stringPtr(string(marshalSiteConfigurationChangeID(6))),
			},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
		},
		{
			name: "first: 10, after: 1 (beyond the last cursor in DB which is 1)",
			args: &graphqlutil.ConnectionResolverArgs{
				First: int32Ptr(10),
				After: stringPtr(string(marshalSiteConfigurationChangeID(1))),
			},
			expectedSiteConfigIDs: []int32{},
		},
		{
			name: "last: 2, before: 1",
			args: &graphqlutil.ConnectionResolverArgs{
				Last:   int32Ptr(2),
				Before: stringPtr(string(marshalSiteConfigurationChangeID(1))),
			},
			expectedSiteConfigIDs: []int32{3, 2},
		},
		{
			name: "last: 10, before: 1 (overflow)",
			args: &graphqlutil.ConnectionResolverArgs{
				Last:   int32Ptr(10),
				Before: stringPtr(string(marshalSiteConfigurationChangeID(1))),
			},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2},
		},
		{
			name: "last: 10, before: 0 (same as get all items, but oldest ID in DB is 1)",
			args: &graphqlutil.ConnectionResolverArgs{
				Last:   int32Ptr(10),
				Before: stringPtr(string(marshalSiteConfigurationChangeID(0))),
			},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
		},
		{
			name: "last: 10, before: 6 (beyond the latest cursor in DB which is 5)",
			args: &graphqlutil.ConnectionResolverArgs{
				Last:   int32Ptr(10),
				Before: stringPtr(string(marshalSiteConfigurationChangeID(6))),
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

			if len(siteConfigChangeResolvers) != len(tc.expectedSiteConfigIDs) {
				diff := cmp.Diff(tc.expectedSiteConfigIDs, siteConfigChangeResolverIDs)
				t.Fatalf(`mismatched number of resolvers, expected %d, got %d\n
diff in IDs: %s,\n
`, len(tc.expectedSiteConfigIDs), len(siteConfigChangeResolvers), diff)
			}

			for i, resolver := range siteConfigChangeResolvers {
				if resolver.siteConfig.ID != tc.expectedSiteConfigIDs[i] {
					t.Errorf("position %d: expected siteConfig.ID %d, but got %d", i, tc.expectedSiteConfigIDs[i], resolver.siteConfig.ID)
				}
			}
		})
	}

}

func TestIsRequiredOutOfBandMigration(t *testing.T) {
	tests := []struct {
		name      string
		version   oobmigration.Version
		migration oobmigration.Migration
		want      bool
	}{
		{
			name:      "not deprecated",
			version:   oobmigration.Version{Major: 4, Minor: 3},
			migration: oobmigration.Migration{},
			want:      false,
		},
		{
			name:    "deprecated but finished",
			version: oobmigration.Version{Major: 4, Minor: 3},
			migration: oobmigration.Migration{
				Deprecated: &oobmigration.Version{Major: 3, Minor: 43},
				Progress:   1,
			},
			want: false,
		},
		{
			name:    "deprecated after the current",
			version: oobmigration.Version{Major: 4, Minor: 3},
			migration: oobmigration.Migration{
				Deprecated: &oobmigration.Version{Major: 4, Minor: 4},
			},
			want: false,
		},

		{
			name:    "deprecated at current and unfinished",
			version: oobmigration.Version{Major: 4, Minor: 3},
			migration: oobmigration.Migration{
				Deprecated: &oobmigration.Version{Major: 4, Minor: 3},
			},
			want: true,
		},
		{
			name:    "deprecated prior to current and unfinished",
			version: oobmigration.Version{Major: 4, Minor: 3},
			migration: oobmigration.Migration{
				Deprecated: &oobmigration.Version{Major: 3, Minor: 43},
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isRequiredOutOfBandMigration(test.version, test.migration)
			assert.Equal(t, test.want, got)
		})
	}
}
