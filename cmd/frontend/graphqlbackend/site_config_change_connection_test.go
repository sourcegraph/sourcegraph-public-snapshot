package graphqlbackend

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type siteConfigStubs struct {
	db    *database.MockDB
	users *database.MockUserStore
	conf  *database.MockConfStore

	now         time.Time
	siteConfigs []*database.SiteConfig
}

func setupSiteConfigStubs() *siteConfigStubs {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		switch id {
		case 1:
			return &types.User{
				ID:          1,
				Username:    "foo",
				DisplayName: "foo user",
			}, nil
		case 2:
			return &types.User{
				ID:          2,
				Username:    "bar",
				DisplayName: "bar user",
			}, nil
		default:
			return nil, errors.Newf("user ID %d not mocked in this test", id)
		}
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)

	now := time.Now()
	// FIXME: Don't use the same now for all the site configs
	siteConfigs := []*database.SiteConfig{
		{
			ID:        1,
			Contents:  ``,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           2,
			AuthorUserID: 1,
			Contents: `{
  foo: 1
}`,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           3,
			AuthorUserID: 2,
			// Newline added.
			Contents: `{
  foo: 1,
  bar: 2
}`,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           4,
			AuthorUserID: 1,
			// Existing line removed.
			Contents: `{
  bar: 2
}`,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           5,
			AuthorUserID: 1,
			// Existing line changed.
			Contents: `{
  bar: 3
}`,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	conf := database.NewMockConfStore()
	db.ConfFunc.SetDefaultReturn(conf)

	conf.ListSiteConfigsFunc.SetDefaultHook(
		func(ctx context.Context, opt database.SiteConfigListOptions) ([]*database.SiteConfig, error) {
			var input []*database.SiteConfig
			if opt.OrderByDirection == database.DescendingOrderByDirection {
				// Reverse the result set if ORDER BY DESC is being used.
				for i := len(siteConfigs) - 1; i >= 0; i-- {
					input = append(input, siteConfigs[i])
				}
			} else {
				input = siteConfigs
			}

			if opt.LimitOffset != nil {
				limit := math.Min(float64(len(siteConfigs)), float64(opt.LimitOffset.Limit))
				input = input[:int(limit)]
			}

			return input, nil
		})

	return &siteConfigStubs{
		db:          db,
		users:       users,
		conf:        conf,
		now:         now,
		siteConfigs: siteConfigs,
	}
}

func TestSiteConfigConnection(t *testing.T) {
	stubs := setupSiteConfigStubs()

	siteConfigs := stubs.siteConfigs

	stubs.conf.SiteGetLatestFunc.SetDefaultReturn(siteConfigs[len(siteConfigs)-1], nil)
	stubs.conf.GetSiteConfigCountFunc.SetDefaultReturn(len(siteConfigs), nil)
	// stubs.conf.ListSiteConfigsFunc.SetDefaultReturn(siteConfigs, nil)

	expectedNow := stubs.now.Format("2006-01-02T15:04:05Z")

	// FIXME: Test for hasNextPage, hasPreviousPage
	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, stubs.db),
			Label:  "Get first 2 site configuration history",
			Query: `
			{
			  site {
				  id
					configuration {
					  id
						history(first: 2){
							totalCount
							nodes{
								id
								previousID
								author{
									id,
									username,
									displayName
								}
								createdAt
								updatedAt
							}
						}
					}
			  }
			}
		`,
			ExpectedResult: fmt.Sprintf(`
			{
					"site": {
						"id": "U2l0ZToic2l0ZSI=",
						"configuration": {
							"id": 5,
							"history": {
								"totalCount": 5,
								"nodes": [
									{
										"id": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6MQ==",
										"previousID": null,
										"author": null,
										"createdAt": %[1]q,
										"updatedAt": %[1]q
									},
									{
										"id": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6Mg==",
										"previousID": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6MQ==",
										"author": {
											"id": "VXNlcjox",
											"username": "foo",
											"displayName": "foo user"
										},
										"createdAt": %[1]q,
										"updatedAt": %[1]q
									}
								]
							}
						}
					}
			}
		`, expectedNow),
		},
		{
			Schema: mustParseGraphQLSchema(t, stubs.db),
			Label:  "Get last 2 site configuration history",
			Query: `
			{
			  site {
				  id
					configuration {
					  id
						history(last: 2){
							totalCount
							nodes{
								id
								previousID
								author{
									id,
									username,
									displayName
								}
								createdAt
								updatedAt
							}
						}
					}
			  }
			}
		`,
			ExpectedResult: fmt.Sprintf(`
			{
					"site": {
						"id": "U2l0ZToic2l0ZSI=",
						"configuration": {
							"id": 5,
							"history": {
								"totalCount": 5,
								"nodes": [
									{
										"id": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6NQ==",
										"previousID": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6NA==",
										"author": {
											"id": "VXNlcjox",
											"username": "foo",
											"displayName": "foo user"
										},
										"createdAt": %[1]q,
										"updatedAt": %[1]q
									},
									{
										"id": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6NA==",
										"previousID": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6Mw==",
										"author": {
											"id": "VXNlcjox",
											"username": "foo",
											"displayName": "foo user"
										},
										"createdAt": %[1]q,
										"updatedAt": %[1]q
									}
								]
							}
						}
					}
			}
		`, expectedNow),
		},
	})
}

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

func TestSiteConfigurationChangeConnectionStoreComputeNodes(t *testing.T) {
	stubs := setupSiteConfigStubs()

	ctx := context.Background()
	store := SiteConfigurationChangeConnectionStore{db: stubs.db}

	testCases := []struct {
		name                          string
		paginationArgs                *database.PaginationArgs
		expectedSiteConfigIDs         []int32
		expectedPreviousSiteConfigIDs []*int32
	}{
		{
			name:                          "nil paginationArgs",
			paginationArgs:                nil,
			expectedSiteConfigIDs:         []int32{1, 2, 3, 4, 5},
			expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{0, 1, 2, 3, 4}),
		},
		{
			name: "first: 2",
			paginationArgs: &database.PaginationArgs{
				First: toIntPtr(2),
			},
			expectedSiteConfigIDs:         []int32{1, 2},
			expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{0, 1}),
		},
		{
			name: "first: 5 (exactly what exists in the database)",
			paginationArgs: &database.PaginationArgs{
				First: toIntPtr(5),
			},
			expectedSiteConfigIDs:         []int32{1, 2, 3, 4, 5},
			expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{0, 1, 2, 3, 4}),
		},
		{
			name: "first: 20 (more than what exists in the database)",
			paginationArgs: &database.PaginationArgs{
				First: toIntPtr(20),
			},
			expectedSiteConfigIDs:         []int32{1, 2, 3, 4, 5},
			expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{0, 1, 2, 3, 4}),
		},
		{
			name: "last: 2",
			paginationArgs: &database.PaginationArgs{
				Last: toIntPtr(2),
			},
			expectedSiteConfigIDs:         []int32{5, 4},
			expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{4, 3}),
		},
		{
			name: "last: 5 (exactly what exists in the database)",
			paginationArgs: &database.PaginationArgs{
				Last: toIntPtr(5),
			},
			expectedSiteConfigIDs:         []int32{5, 4, 3, 2, 1},
			expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{4, 3, 2, 1, 0}),
		},
		{
			name: "last: 20 (more than what exists in the database)",
			paginationArgs: &database.PaginationArgs{
				Last: toIntPtr(20),
			},
			expectedSiteConfigIDs:         []int32{5, 4, 3, 2, 1},
			expectedPreviousSiteConfigIDs: toListOfIntPtrs([]int32{4, 3, 2, 1, 0}),
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
				if tc.expectedPreviousSiteConfigIDs[i] == nil && got.previousSiteConfig == nil {
					continue
				}

				// If we expect no previousSiteConfig, but got got one, error out.
				if tc.expectedPreviousSiteConfigIDs[i] == nil && got.previousSiteConfig != nil {
					t.Fatalf("position %d: expected previousSiteConfig to be nil, but got %v", i, got.previousSiteConfig)
				}

				// If we expect previousSiteConfig, but got got nil, error out.
				if tc.expectedPreviousSiteConfigIDs[i] != nil && got.previousSiteConfig == nil {
					t.Fatalf("position %d: expected previousSiteConfig to be non-nil, but got nil", i)
				}

				// If we have a mismatched ID of expected previousSiteConfig vs what we got, error out.
				if got.previousSiteConfig.ID != *tc.expectedPreviousSiteConfigIDs[i] {
					t.Fatalf(
						"position %d: expected previousSiteConfig.ID %d, but got %d",
						i, *tc.expectedPreviousSiteConfigIDs[i], got.previousSiteConfig.ID,
					)
				}
			}
		})
	}
}
