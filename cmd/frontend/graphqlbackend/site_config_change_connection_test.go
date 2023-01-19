package graphqlbackend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type siteConfigStubs struct {
	db          database.DB
	users       []*types.User
	siteConfigs []*database.SiteConfig
}

func setupSiteConfigStubs(t *testing.T) *siteConfigStubs {
	logger := log.NoOp()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	usersToCreate := []database.NewUser{
		{Username: "foo", DisplayName: "foo user"},
		{Username: "bar", DisplayName: "bar user"},
	}

	var users []*types.User
	for _, input := range usersToCreate {
		user, err := db.Users().Create(ctx, input)
		if err != nil {
			t.Fatal(err)
		}

		if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
			t.Fatal(err)
		}

		users = append(users, user)
	}

	conf := db.Conf()
	siteConfigsToCreate := []*database.SiteConfig{
		{
			Contents: `
{
  "auth.Providers": []
}`,
		},
		{
			AuthorUserID: 2,
			// A new line is added.
			Contents: `
{
  "disableAutoGitUpdates": true,
  "auth.Providers": []
}`,
		},
		{
			AuthorUserID: 1,
			// Existing line is changed.
			Contents: `
{
  "disableAutoGitUpdates": false,
  "auth.Providers": []
}`,
		},
		{
			AuthorUserID: 1,
			// Existing line is removed.
			Contents: `
{
  "auth.Providers": []
}`,
		},
	}

	lastID := int32(0)
	// This will create 5 entries, because the first time conf.SiteCreateIfupToDate is called it
	// will create two entries in the DB.
	for _, input := range siteConfigsToCreate {
		siteConfig, err := conf.SiteCreateIfUpToDate(ctx, int32Ptr(lastID), input.AuthorUserID, input.Contents, false)
		if err != nil {
			t.Fatal(err)
		}

		lastID = siteConfig.ID
	}

	return &siteConfigStubs{
		db:    db,
		users: users,
		// siteConfigs: siteConfigs,
	}
}

// func TestSiteConfigConnection(t *testing.T) {
// 	stubs := setupSiteConfigStubs(t)

// 	// siteConfigs := stubs.siteConfigs

// 	// stubs.conf.SiteGetLatestFunc.SetDefaultReturn(siteConfigs[len(siteConfigs)-1], nil)
// 	// stubs.conf.GetSiteConfigCountFunc.SetDefaultReturn(len(siteConfigs), nil)
// 	// stubs.conf.ListSiteConfigsFunc.SetDefaultReturn(siteConfigs, nil)

// 	// expectedNow := stubs.now.Format("2006-01-02T15:04:05Z")

// 	// FIXME: Test for hasNextPage, hasPreviousPage
// 	RunTests(t, []*Test{
// 		{
// 			Schema: mustParseGraphQLSchema(t, stubs.db),
// 			Label:  "Get first 2 site configuration history",
// 			Query: `
// 			{
// 			  site {
// 				  id
// 					configuration {
// 					  id
// 						history(first: 2){
// 							totalCount
// 							nodes{
// 								id
// 								previousID
// 								author{
// 									id,
// 									username,
// 									displayName
// 								}
// 								createdAt
// 								updatedAt
// 							}
// 						}
// 					}
// 			  }
// 			}
// 		`,
// 			ExpectedResult: fmt.Sprintf(`
// 			{
// 					"site": {
// 						"id": "U2l0ZToic2l0ZSI=",
// 						"configuration": {
// 							"id": 5,
// 							"history": {
// 								"totalCount": 5,
// 								"nodes": [
// 									{
// 										"id": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6NQ==",
// 										"author": null,
// 										"createdAt": %[1]q,
// 										"updatedAt": %[1]q
// 									},
// 									{
// 										"id": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6NA==",
// 										"author": {
// 											"id": "VXNlcjox",
// 											"username": "foo",
// 											"displayName": "foo user"
// 										},
// 										"createdAt": %[1]q,
// 										"updatedAt": %[1]q
// 									}
// 								]
// 							}
// 						}
// 					}
// 			}
// 		`, expectedNow),
// 		},
// 		{
// 			Schema: mustParseGraphQLSchema(t, stubs.db),
// 			Label:  "Get last 2 site configuration history",
// 			Query: `
// 			{
// 			  site {
// 				  id
// 					configuration {
// 					  id
// 						history(last: 2){
// 							totalCount
// 							nodes{
// 								id
// 								previousID
// 								author{
// 									id,
// 									username,
// 									displayName
// 								}
// 								createdAt
// 								updatedAt
// 							}
// 						}
// 					}
// 			  }
// 			}
// 		`,
// 			ExpectedResult: fmt.Sprintf(`
// 			{
// 					"site": {
// 						"id": "U2l0ZToic2l0ZSI=",
// 						"configuration": {
// 							"id": 5,
// 							"history": {
// 								"totalCount": 5,
// 								"nodes": [
// 									{
// 										"id": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6Mg==",
// 										"author": {
// 											"id": "VXNlcjox",
// 											"username": "foo",
// 											"displayName": "foo user"
// 										},
// 										"createdAt": %[1]q,
// 										"updatedAt": %[1]q
// 									},
// 									{
// 										"id": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6MQ==",
// 										"author": {
// 											"id": "VXNlcjox",
// 											"username": "foo",
// 											"displayName": "foo user"
// 										},
// 										"createdAt": %[1]q,
// 										"updatedAt": %[1]q
// 									}
// 								]
// 							}
// 						}
// 					}
// 			}
// 		`, expectedNow),
// 		},
// 	})
// }

func toListOfIntPtrs(input []int32) []*int32 {
	list := make([]*int32, len(input))
	for i, item := range input {
		// Reassign to update the address of item.
		item := item
		// ID 0 is not possible. Use this to indicate that this should be 0.
		if item == 0 {
			list[i] = nil
		} else {
			list[i] = &item
		}
	}

	return list
}

func TestSiteConfigurationChangeConnectionStoreComputeNodes(t *testing.T) {
	stubs := setupSiteConfigStubs(t)

	ctx := context.Background()
	store := SiteConfigurationChangeConnectionStore{db: stubs.db}

	if _, err := store.ComputeNodes(ctx, nil); err == nil {
		t.Fatalf("expected error but got nil")
	}

	testCases := []struct {
		name                  string
		paginationArgs        *database.PaginationArgs
		expectedSiteConfigIDs []int32
		// value of 0 in expectedPreviousSIteConfigIDs means nil in the test assertion.
		expectedPreviousSiteConfigIDs []int32
	}{
		{
			name: "first: 2",
			paginationArgs: &database.PaginationArgs{
				First: intPtr(2),
			},
			expectedSiteConfigIDs:         []int32{5, 4},
			expectedPreviousSiteConfigIDs: []int32{4, 3},
		},
		{
			name: "first: 5 (exact number of items that exist in the database)",
			paginationArgs: &database.PaginationArgs{
				First: intPtr(5),
			},
			expectedSiteConfigIDs:         []int32{5, 4, 3, 2, 1},
			expectedPreviousSiteConfigIDs: []int32{4, 3, 2, 1, 0},
		},
		{
			name: "first: 20 (more items than what exists in the database)",
			paginationArgs: &database.PaginationArgs{
				First: intPtr(20),
			},
			expectedSiteConfigIDs:         []int32{5, 4, 3, 2, 1},
			expectedPreviousSiteConfigIDs: []int32{4, 3, 2, 1, 0},
		},
		{
			name: "last: 2",
			paginationArgs: &database.PaginationArgs{
				Last: intPtr(2),
			},
			expectedSiteConfigIDs:         []int32{1, 2},
			expectedPreviousSiteConfigIDs: []int32{0, 1},
		},
		{
			name: "last: 5 (exact number of items that exist in the database)",
			paginationArgs: &database.PaginationArgs{
				Last: intPtr(5),
			},
			expectedSiteConfigIDs:         []int32{1, 2, 3, 4, 5},
			expectedPreviousSiteConfigIDs: []int32{0, 1, 2, 3, 4},
		},
		{
			name: "last: 20 (more items than what exists in the database)",
			paginationArgs: &database.PaginationArgs{
				Last: intPtr(20),
			},
			expectedSiteConfigIDs:         []int32{1, 2, 3, 4, 5},
			expectedPreviousSiteConfigIDs: []int32{0, 1, 2, 3, 4},
		},
		{
			name: "first: 2, after: 4",
			paginationArgs: &database.PaginationArgs{
				First: intPtr(2),
				After: intPtr(4),
			},
			expectedSiteConfigIDs:         []int32{3, 2},
			expectedPreviousSiteConfigIDs: []int32{2, 1},
		},
		{
			name: "first: 10, after: 4",
			paginationArgs: &database.PaginationArgs{
				First: intPtr(10),
				After: intPtr(4),
			},
			expectedSiteConfigIDs:         []int32{3, 2, 1},
			expectedPreviousSiteConfigIDs: []int32{2, 1, 0},
		},
		{
			name: "first: 2, after: 1",
			paginationArgs: &database.PaginationArgs{
				First: intPtr(2),
				After: intPtr(1),
			},
			expectedSiteConfigIDs:         []int32{},
			expectedPreviousSiteConfigIDs: []int32{},
		},
		{
			name: "last: 2, before: 2",
			paginationArgs: &database.PaginationArgs{
				Last:   intPtr(2),
				Before: intPtr(2),
			},
			expectedSiteConfigIDs:         []int32{3, 4},
			expectedPreviousSiteConfigIDs: []int32{2, 3},
		},
		{
			name: "last: 10, before: 2",
			paginationArgs: &database.PaginationArgs{
				Last:   intPtr(10),
				Before: intPtr(2),
			},
			expectedSiteConfigIDs:         []int32{3, 4, 5},
			expectedPreviousSiteConfigIDs: []int32{2, 3, 4},
		},
		{
			name: "last: 2, before: 5",
			paginationArgs: &database.PaginationArgs{
				Last:   intPtr(2),
				Before: intPtr(5),
			},
			expectedSiteConfigIDs:         []int32{},
			expectedPreviousSiteConfigIDs: []int32{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			siteConfigChangeResolvers, err := store.ComputeNodes(ctx, tc.paginationArgs)
			if err != nil {
				t.Errorf("expected nil, but got error: %v", err)
			}

			gotLength := len(siteConfigChangeResolvers)
			expectedLength := len(tc.expectedSiteConfigIDs)
			if gotLength != expectedLength {
				t.Fatalf("mismatched number of SiteConfigurationChangeResolvers, expected %d, got %d", expectedLength, gotLength)
			}

			gotIDs := make([]int32, gotLength)
			for i, got := range siteConfigChangeResolvers {
				gotIDs[i] = got.siteConfig.ID
			}

			if diff := cmp.Diff(tc.expectedSiteConfigIDs, gotIDs); diff != "" {
				t.Errorf("mismatched siteConfig.ID, diff %v", diff)
			}

			if len(tc.expectedPreviousSiteConfigIDs) == 0 {
				return
			}

			gotPreviousSiteConfigIDs := make([]int32, gotLength)
			for i, got := range siteConfigChangeResolvers {
				if got.previousSiteConfig == nil {
					gotPreviousSiteConfigIDs[i] = 0
				} else {
					gotPreviousSiteConfigIDs[i] = got.previousSiteConfig.ID
				}
			}

			if diff := cmp.Diff(tc.expectedPreviousSiteConfigIDs, gotPreviousSiteConfigIDs); diff != "" {
				t.Errorf("mismatched siteConfig.ID, diff %v", diff)
			}
		})
	}
}

func TestModifyArgs(t *testing.T) {
	testCases := []struct {
		name             string
		args             *database.PaginationArgs
		expectedArgs     *database.PaginationArgs
		expectedModified bool
	}{
		{
			name:             "first: 5 (first page)",
			args:             &database.PaginationArgs{First: intPtr(5)},
			expectedArgs:     &database.PaginationArgs{First: intPtr(6)},
			expectedModified: true,
		},
		{
			name:             "first: 5, after: 10 (next page)",
			args:             &database.PaginationArgs{First: intPtr(5), After: intPtr(10)},
			expectedArgs:     &database.PaginationArgs{First: intPtr(6), After: intPtr(10)},
			expectedModified: true,
		},
		{
			name:             "last: 5 (last page)",
			args:             &database.PaginationArgs{Last: intPtr(5)},
			expectedArgs:     &database.PaginationArgs{Last: intPtr(5)},
			expectedModified: false,
		},
		{
			name:             "last: 5, before: 10 (previous page)",
			args:             &database.PaginationArgs{Last: intPtr(5), Before: intPtr(10)},
			expectedArgs:     &database.PaginationArgs{Last: intPtr(6), Before: intPtr(9)},
			expectedModified: true,
		},
		{
			name:             "last: 5, before: 1 (edge case)",
			args:             &database.PaginationArgs{Last: intPtr(5), Before: intPtr(1)},
			expectedArgs:     &database.PaginationArgs{Last: intPtr(6), Before: intPtr(0)},
			expectedModified: true,
		},
		{
			name:             "last: 5, before: 0 (same as last page but a mathematical  edge case)",
			args:             &database.PaginationArgs{Last: intPtr(5), Before: intPtr(0)},
			expectedArgs:     &database.PaginationArgs{Last: intPtr(5), Before: intPtr(0)},
			expectedModified: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			modified := modifyArgs(tc.args)
			if modified != tc.expectedModified {
				t.Errorf("Expected modified to be %v, but got %v", modified, tc.expectedModified)
			}

			if diff := cmp.Diff(tc.args, tc.expectedArgs); diff != "" {
				t.Errorf("Mismatch in modified args: %v", diff)
			}
		})
	}
}
