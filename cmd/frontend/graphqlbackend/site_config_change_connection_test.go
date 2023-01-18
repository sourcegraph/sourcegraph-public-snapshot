package graphqlbackend

import (
	"context"
	"testing"

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
		siteConfig, err := conf.SiteCreateIfUpToDate(ctx, toInt32Ptr(lastID), input.AuthorUserID, input.Contents, false)
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

func toIntPtr(n int) *int {
	return &n
}

func toInt32Ptr(n int32) *int32 {
	return &n
}

func TestSiteConfigurationChangeConnectionStoreComputeNodes(t *testing.T) {
	stubs := setupSiteConfigStubs(t)

	ctx := context.Background()
	store := SiteConfigurationChangeConnectionStore{db: stubs.db}

	testCases := []struct {
		name                  string
		paginationArgs        *database.PaginationArgs
		expectedSiteConfigIDs []int32
		// expectedPreviousSiteConfigIDs []*int32
	}{
		{
			name:                  "nil paginationArgs (return everything in insertion order)",
			paginationArgs:        nil,
			expectedSiteConfigIDs: []int32{1, 2, 3, 4, 5},
			// expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{0, 1, 2, 3, 4}),
		},
		{
			name: "first: 2",
			paginationArgs: &database.PaginationArgs{
				First: toIntPtr(2),
			},
			expectedSiteConfigIDs: []int32{5, 4},
			// expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{0, 1}),
		},
		{
			name: "first: 5 (exact number of items that exist in the database)",
			paginationArgs: &database.PaginationArgs{
				First: toIntPtr(5),
			},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
			// expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{0, 1, 2, 3, 4}),
		},
		{
			name: "first: 20 (more items than what exists in the database)",
			paginationArgs: &database.PaginationArgs{
				First: toIntPtr(20),
			},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
			// expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{0, 1, 2, 3, 4}),
		},
		{
			name: "last: 2",
			paginationArgs: &database.PaginationArgs{
				Last: toIntPtr(2),
			},
			expectedSiteConfigIDs: []int32{1, 2},
			// expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{4, 3}),
		},
		{
			name: "last: 5 (exact number of items that exist in the database)",
			paginationArgs: &database.PaginationArgs{
				Last: toIntPtr(5),
			},
			expectedSiteConfigIDs: []int32{1, 2, 3, 4, 5},
			// expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{4, 3, 2, 1, 0}),
		},
		{
			name: "last: 20 (more items than what exists in the database)",
			paginationArgs: &database.PaginationArgs{
				Last: toIntPtr(20),
			},
			expectedSiteConfigIDs: []int32{1, 2, 3, 4, 5},
			// 	// expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{4, 3, 2, 1, 0}),
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

			for i, got := range siteConfigChangeResolvers {
				if got.siteConfig.ID != tc.expectedSiteConfigIDs[i] {
					t.Errorf("position %d: expected siteConfig.ID %d, but got %d", i, tc.expectedSiteConfigIDs[i], got.siteConfig.ID)
				}

				// Expected nil previousSiteConfig and got nil previousSiteConfig? Test passes, so move on to
				// the next item.
				// if tc.expectedPreviousSiteConfigIDs[i] == nil && got.previousSiteConfig == nil {
				// 	continue
				// }

				// If we expect no previousSiteConfig, but got got one, error out.
				// if tc.expectedPreviousSiteConfigIDs[i] == nil && got.previousSiteConfig != nil {
				// 	t.Fatalf("position %d: expected previousSiteConfig to be nil, but got %v", i, got.previousSiteConfig)
				// }

				// If we expect previousSiteConfig, but got got nil, error out.
				// if tc.expectedPreviousSiteConfigIDs[i] != nil && got.previousSiteConfig == nil {
				// 	t.Fatalf("position %d: expected previousSiteConfig to be non-nil, but got nil", i)
				// }

				// If we have a mismatched ID of expected previousSiteConfig vs what we got, error out.
				// 	if got.previousSiteConfig.ID != *tc.expectedPreviousSiteConfigIDs[i] {
				// 		t.Fatalf(
				// 			"position %d: expected previousSiteConfig.ID %d, but got %d",
				// 			i, *tc.expectedPreviousSiteConfigIDs[i], got.previousSiteConfig.ID,
				// 		)
				// 	}
			}
		})
	}
}
