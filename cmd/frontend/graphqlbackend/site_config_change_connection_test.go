package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type siteConfigStubs struct {
	db            database.DB
	users         []*types.User
	expectedDiffs map[int32]string
}

func toStringPtr(n int) *string {
	str := strconv.Itoa(n)

	return &str
}

func setupSiteConfigStubs(t *testing.T) *siteConfigStubs {
	logger := log.NoOp()
	db := database.NewDB(logger, dbtest.NewDB(t))
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
		// ID: 2 (because first time we create a config an initial config will be created first)
		{
			Contents: `{
  "auth.Providers": []
}`,
		},
		// ID: 3
		{
			AuthorUserID: 2,
			// A new line is added.
			Contents: `{
  "disableAutoGitUpdates": true,
  "auth.Providers": []
}`,
		},
		// ID: 4
		{
			AuthorUserID: 1,
			// Existing line is changed.
			Contents: `{
  "disableAutoGitUpdates": false,
  "auth.Providers": []
}`,
		},
		// ID: 5
		{
			AuthorUserID: 1,
			// Nothing is changed.
			//
			// This is the same as the previous entry, and this should not show up in the output of
			// any query that lists the diffs.
			Contents: `{
  "disableAutoGitUpdates": false,
  "auth.Providers": []
}`,
		},
		// ID: 6
		{
			AuthorUserID: 3, // This user no longer exists
			// Existing line is removed.
			Contents: `{
  "auth.Providers": []
}`,
		},
	}

	lastID := int32(0)
	// This will create 5 entries, because the first time conf.SiteCreateIfupToDate is called it
	// will create two entries in the DB.
	for _, input := range siteConfigsToCreate {
		siteConfig, err := conf.SiteCreateIfUpToDate(ctx, pointers.Ptr(lastID), input.AuthorUserID, input.Contents, false)
		if err != nil {
			t.Fatal(err)
		}

		lastID = siteConfig.ID
	}

	expectedDiffs := map[int32]string{
		// This first diff is between 6 and 4 and not 5 and 4 because:
		// 4 and 5 are identical entries
		//
		// Also, the diff is not between 6 and 5 because:
		// 4 came first in the series and 5 is the redundant / duplicate config and not 4. And 6 is
		// the next item that is different, we want to calculate the diff between these two and not
		// 6 and 5.
		6: `--- ID: 4
+++ ID: 6
@@ -1,4 +1,3 @@
 {
-  "disableAutoGitUpdates": false,
   "auth.Providers": []
 }
\ No newline at end of file
`,

		4: `--- ID: 3
+++ ID: 4
@@ -1,4 +1,4 @@
 {
-  "disableAutoGitUpdates": true,
+  "disableAutoGitUpdates": false,
   "auth.Providers": []
 }
\ No newline at end of file
`,

		3: `--- ID: 2
+++ ID: 3
@@ -1,3 +1,4 @@
 {
+  "disableAutoGitUpdates": true,
   "auth.Providers": []
 }
\ No newline at end of file
`,

		2: `--- ID: 1
+++ ID: 2
@@ -1,17 +1,3 @@
 {
-  // The externally accessible URL for Sourcegraph (i.e., what you type into your browser)
-  // This is required to be configured for Sourcegraph to work correctly.
-  // "externalURL": "https://sourcegraph.example.com",
-  // The authentication provider to use for identifying and signing in users.
-  // Only one entry is supported.
-  //
-  // The builtin auth provider with signup disallowed (shown below) means that
-  // after the initial site admin signs in, all other users must be invited.
-  //
-  // Other providers are documented at https://docs.sourcegraph.com/admin/auth.
-  "auth.providers": [
-    {
-      "type": "builtin"
-    }
-  ],
+  "auth.Providers": []
 }
\ No newline at end of file
`,

		1: `--- ID: 0
+++ ID: 1
@@ -1 +1,17 @@
+{
+  // The externally accessible URL for Sourcegraph (i.e., what you type into your browser)
+  // This is required to be configured for Sourcegraph to work correctly.
+  // "externalURL": "https://sourcegraph.example.com",
+  // The authentication provider to use for identifying and signing in users.
+  // Only one entry is supported.
+  //
+  // The builtin auth provider with signup disallowed (shown below) means that
+  // after the initial site admin signs in, all other users must be invited.
+  //
+  // Other providers are documented at https://docs.sourcegraph.com/admin/auth.
+  "auth.providers": [
+    {
+      "type": "builtin"
+    }
+  ],
+}
\ No newline at end of file
`,
	}

	return &siteConfigStubs{
		db:            db,
		users:         users,
		expectedDiffs: expectedDiffs,
	}
}

func TestSiteConfigConnection(t *testing.T) {
	stubs := setupSiteConfigStubs(t)
	expectedDiffs := stubs.expectedDiffs

	// Create a context with an admin user as the actor.
	contextWithActor := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

	RunTests(t, []*Test{
		{
			Schema:  mustParseGraphQLSchema(t, stubs.db),
			Label:   "Get first 2 site configuration history",
			Context: contextWithActor,
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
							  author{
								  id,
								  username,
								  displayName
							  }
							  diff
						  }
						  pageInfo {
							hasNextPage
							hasPreviousPage
							endCursor
							startCursor
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
						"id": 6,
						"history": {
							"totalCount": 5,
							"nodes": [
								{
									"id": %[1]q,
									"author": null,
									"diff": %[3]q
								},
								{
									"id": %[2]q,
									"author": {
										"id": "VXNlcjox",
										"username": "foo",
										"displayName": "foo user"
									},
									"diff": %[4]q
								}
							],
							"pageInfo": {
							  "hasNextPage": true,
							  "hasPreviousPage": false,
							  "endCursor": %[2]q,
							  "startCursor": %[1]q
							}
						}
					}
				}
			}
		`, marshalSiteConfigurationChangeID(6), marshalSiteConfigurationChangeID(4), expectedDiffs[6], expectedDiffs[4]),
		},
		{
			Schema:  mustParseGraphQLSchema(t, stubs.db),
			Label:   "Get last 3 site configuration history",
			Context: contextWithActor,
			Query: `
					{
						site {
							id
							configuration {
								id
								history(last: 3){
									totalCount
									nodes{
										id
										author{
											id,
											username,
											displayName
										}
										diff
									}
									pageInfo {
									  hasNextPage
									  hasPreviousPage
									  endCursor
									  startCursor
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
								"id": 6,
								"history": {
									"totalCount": 5,
									"nodes": [
										{
											"id": %[1]q,
											"author": {
												"id": "VXNlcjoy",
												"username": "bar",
												"displayName": "bar user"
											},

											"diff": %[4]q
										},
										{
											"id": %[2]q,
											"author": null,

											"diff": %[5]q
										},
										{
											"id": %[3]q,
											"author": null,

											"diff": %[6]q
										}
									],
									"pageInfo": {
									  "hasNextPage": false,
									  "hasPreviousPage": true,
									  "endCursor": %[3]q,
									  "startCursor": %[1]q
									}
								}
							}
						}
					}
				`, marshalSiteConfigurationChangeID(3), marshalSiteConfigurationChangeID(2), marshalSiteConfigurationChangeID(1),
				expectedDiffs[3], expectedDiffs[2], expectedDiffs[1],
			),
		},
		{
			Schema:  mustParseGraphQLSchema(t, stubs.db),
			Label:   "Get first 2 site configuration history based on an offset",
			Context: contextWithActor,
			Query: fmt.Sprintf(`
			{
				site {
					id
					configuration {
						id
						history(first: 2, after: %q){
							totalCount
							nodes{
								id
								author{
									id,
									username,
									displayName
								}
								diff
							}
							pageInfo {
							  hasNextPage
							  hasPreviousPage
							  endCursor
							  startCursor
							}
						}
					}
				}
			}
		`, marshalSiteConfigurationChangeID(6)),
			ExpectedResult: fmt.Sprintf(`
			{
				"site": {
					"id": "U2l0ZToic2l0ZSI=",
					"configuration": {
						"id": 6,
						"history": {
							"totalCount": 5,
							"nodes": [
								{
									"id": %[1]q,
									"author": {
										"id": "VXNlcjox",
										"username": "foo",
										"displayName": "foo user"
									},
									"diff": %[3]q
								},
								{
									"id": %[2]q,
									"author": {
										"id": "VXNlcjoy",
										"username": "bar",
										"displayName": "bar user"
									},
									"diff": %[4]q
								}
							],
							"pageInfo": {
							  "hasNextPage": true,
							  "hasPreviousPage": true,
							  "endCursor": %[2]q,
							  "startCursor": %[1]q
							}
						}
					}
				}
			}
		`, marshalSiteConfigurationChangeID(4), marshalSiteConfigurationChangeID(3), expectedDiffs[4], expectedDiffs[3]),
		},
		{
			Schema:  mustParseGraphQLSchema(t, stubs.db),
			Label:   "Get last 2 site configuration history based on an offset",
			Context: contextWithActor,
			Query: fmt.Sprintf(`
			{
			  site {
				  id
					configuration {
					  id
						history(last: 2, before: %q){
							totalCount
							nodes{
								id
								author{
									id,
									username,
									displayName
								}
								diff
							}
							pageInfo {
							  hasNextPage
							  hasPreviousPage
							  endCursor
							  startCursor
							}
						}
					}
			  }
			}
		`, marshalSiteConfigurationChangeID(3)),
			ExpectedResult: fmt.Sprintf(`
			{
				"site": {
					"id": "U2l0ZToic2l0ZSI=",
					"configuration": {
						"id": 6,
						"history": {
							"totalCount": 5,
							"nodes": [
								{
									"id": %[1]q,
									"author": null,
									"diff": %[3]q
								},
								{
									"id": %[2]q,
									"author": {
										"id": "VXNlcjox",
										"username": "foo",
										"displayName": "foo user"
									},
									"diff": %[4]q
								}
							],
							"pageInfo": {
							  "hasNextPage": true,
							  "hasPreviousPage": false,
							  "endCursor": %[2]q,
							  "startCursor": %[1]q
							}
						}
					}
				}
			}
		`, marshalSiteConfigurationChangeID(6), marshalSiteConfigurationChangeID(4), expectedDiffs[6], expectedDiffs[4]),
		},
	})
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
				First: pointers.Ptr(2),
			},
			// 5 is skipped because it is the same as 4.
			expectedSiteConfigIDs:         []int32{6, 4},
			expectedPreviousSiteConfigIDs: []int32{4, 3},
		},
		{
			name: "first: 6 (exact number of items that exist in the database)",
			paginationArgs: &database.PaginationArgs{
				First: pointers.Ptr(6),
			},
			expectedSiteConfigIDs:         []int32{6, 4, 3, 2, 1},
			expectedPreviousSiteConfigIDs: []int32{4, 3, 2, 1, 0},
		},
		{
			name: "first: 20 (more items than what exists in the database)",
			paginationArgs: &database.PaginationArgs{
				First: pointers.Ptr(20),
			},
			expectedSiteConfigIDs:         []int32{6, 4, 3, 2, 1},
			expectedPreviousSiteConfigIDs: []int32{4, 3, 2, 1, 0},
		},
		{
			name: "last: 2",
			paginationArgs: &database.PaginationArgs{
				Last: pointers.Ptr(2),
			},
			expectedSiteConfigIDs:         []int32{1, 2},
			expectedPreviousSiteConfigIDs: []int32{0, 1},
		},
		{
			name: "last: 6 (exact number of items that exist in the database)",
			paginationArgs: &database.PaginationArgs{
				Last: pointers.Ptr(6),
			},
			expectedSiteConfigIDs:         []int32{1, 2, 3, 4, 6},
			expectedPreviousSiteConfigIDs: []int32{0, 1, 2, 3, 4},
		},
		{
			name: "last: 20 (more items than what exists in the database)",
			paginationArgs: &database.PaginationArgs{
				Last: pointers.Ptr(20),
			},
			expectedSiteConfigIDs:         []int32{1, 2, 3, 4, 6},
			expectedPreviousSiteConfigIDs: []int32{0, 1, 2, 3, 4},
		},
		{
			name: "first: 2, after: 6",
			paginationArgs: &database.PaginationArgs{
				First: pointers.Ptr(2),
				After: toStringPtr(6),
			},
			expectedSiteConfigIDs:         []int32{4, 3},
			expectedPreviousSiteConfigIDs: []int32{3, 2},
		},
		{
			name: "first: 10, after: 6",
			paginationArgs: &database.PaginationArgs{
				First: pointers.Ptr(10),
				After: toStringPtr(6),
			},
			expectedSiteConfigIDs:         []int32{4, 3, 2, 1},
			expectedPreviousSiteConfigIDs: []int32{3, 2, 1, 0},
		},
		{
			name: "first: 2, after: 1",
			paginationArgs: &database.PaginationArgs{
				First: pointers.Ptr(2),
				After: toStringPtr(1),
			},
			expectedSiteConfigIDs:         []int32{},
			expectedPreviousSiteConfigIDs: []int32{},
		},
		{
			name: "last: 2, before: 2",
			paginationArgs: &database.PaginationArgs{
				Last:   pointers.Ptr(2),
				Before: toStringPtr(2),
			},
			expectedSiteConfigIDs:         []int32{3, 4},
			expectedPreviousSiteConfigIDs: []int32{2, 3},
		},
		{
			name: "last: 10, before: 2",
			paginationArgs: &database.PaginationArgs{
				Last:   pointers.Ptr(10),
				Before: toStringPtr(2),
			},
			expectedSiteConfigIDs:         []int32{3, 4, 6},
			expectedPreviousSiteConfigIDs: []int32{2, 3, 4},
		},
		{
			name: "last: 2, before: 6",
			paginationArgs: &database.PaginationArgs{
				Last:   pointers.Ptr(2),
				Before: toStringPtr(6),
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
				t.Errorf("mismatched siteConfig.ID, diff (-want, +got)\n%s", diff)
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
			args:             &database.PaginationArgs{First: pointers.Ptr(5)},
			expectedArgs:     &database.PaginationArgs{First: pointers.Ptr(6)},
			expectedModified: true,
		},
		{
			name:             "first: 5, after: 10 (next page)",
			args:             &database.PaginationArgs{First: pointers.Ptr(5), After: toStringPtr(10)},
			expectedArgs:     &database.PaginationArgs{First: pointers.Ptr(6), After: toStringPtr(10)},
			expectedModified: true,
		},
		{
			name:             "last: 5 (last page)",
			args:             &database.PaginationArgs{Last: pointers.Ptr(5)},
			expectedArgs:     &database.PaginationArgs{Last: pointers.Ptr(5)},
			expectedModified: false,
		},
		{
			name:             "last: 5, before: 10 (previous page)",
			args:             &database.PaginationArgs{Last: pointers.Ptr(5), Before: toStringPtr(10)},
			expectedArgs:     &database.PaginationArgs{Last: pointers.Ptr(6), Before: toStringPtr(9)},
			expectedModified: true,
		},
		{
			name:             "last: 5, before: 1 (edge case)",
			args:             &database.PaginationArgs{Last: pointers.Ptr(5), Before: toStringPtr(1)},
			expectedArgs:     &database.PaginationArgs{Last: pointers.Ptr(6), Before: toStringPtr(0)},
			expectedModified: true,
		},
		{
			name:             "last: 5, before: 0 (same as last page but a mathematical  edge case)",
			args:             &database.PaginationArgs{Last: pointers.Ptr(5), Before: toStringPtr(0)},
			expectedArgs:     &database.PaginationArgs{Last: pointers.Ptr(5), Before: toStringPtr(0)},
			expectedModified: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			modified, err := modifyArgs(tc.args)
			if err != nil {
				t.Fatal(err)
			}

			if modified != tc.expectedModified {
				t.Errorf("Expected modified to be %v, but got %v", modified, tc.expectedModified)
			}

			if diff := cmp.Diff(tc.args, tc.expectedArgs); diff != "" {
				t.Errorf("Mismatch in modified args: %v", diff)
			}
		})
	}
}
