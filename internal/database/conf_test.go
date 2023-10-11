package database

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestSiteGetLatestDefault(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	latest, err := db.Conf().SiteGetLatest(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if latest == nil {
		t.Errorf("expected non-nil latest config since default config should be created, got: %+v", latest)
	}
}

func TestSiteCreate_RejectInvalidJSON(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	malformedJSON := "[This is malformed.}"

	_, err := db.Conf().SiteCreateIfUpToDate(ctx, nil, 0, malformedJSON, false)

	if err == nil || !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Fatalf("expected parse error after creating configuration with malformed JSON, got: %+v", err)
	}
}

func TestSiteCreateIfUpToDate(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)

	type input struct {
		lastID       int32
		authorUserID int32
		contents     string
	}

	type output struct {
		ID               int32
		authorUserID     int32
		contents         string
		redactedContents string
		err              error
	}

	type pair struct {
		input    input
		expected output
	}

	type test struct {
		name     string
		sequence []pair
	}

	configRateLimitZero := `{"defaultRateLimit": 0,"auth.providers": []}`
	configRateLimitOne := `{"defaultRateLimit": 1,"auth.providers": []}`

	jsonConfigRateLimitZero := `{
  "defaultRateLimit": 0,
  "auth.providers": []
}`

	jsonConfigRateLimitOne := `{
  "defaultRateLimit": 1,
  "auth.providers": []
}`

	for _, test := range []test{
		{
			name: "create_with_author_user_id",
			sequence: []pair{
				{
					input{
						lastID:       0,
						authorUserID: 1,
						contents:     configRateLimitZero,
					},
					output{
						ID:               2,
						authorUserID:     1,
						contents:         configRateLimitZero,
						redactedContents: jsonConfigRateLimitZero,
					},
				},
			},
		},
		{
			name: "create_one",
			sequence: []pair{
				{
					input{
						lastID:   0,
						contents: configRateLimitZero,
					},
					output{
						ID:               2,
						contents:         configRateLimitZero,
						redactedContents: jsonConfigRateLimitZero,
					},
				},
			},
		},
		{
			name: "create_two",
			sequence: []pair{
				{
					input{
						lastID:   0,
						contents: configRateLimitZero,
					},
					output{
						ID:               2,
						contents:         configRateLimitZero,
						redactedContents: jsonConfigRateLimitZero,
					},
				},
				{
					input{
						lastID:   2,
						contents: configRateLimitOne,
					},
					output{
						ID:               3,
						contents:         configRateLimitOne,
						redactedContents: jsonConfigRateLimitOne,
					},
				},
			},
		},
		{
			name: "do_not_update_if_outdated",
			sequence: []pair{
				{
					input{
						lastID:   0,
						contents: configRateLimitZero,
					},
					output{
						ID:               2,
						contents:         configRateLimitZero,
						redactedContents: jsonConfigRateLimitZero,
					},
				},
				{
					input{
						lastID: 0,
						// This configuration is now behind the first one, so it shouldn't be saved
						contents: configRateLimitOne,
					},
					output{
						ID:               2,
						contents:         configRateLimitOne,
						redactedContents: jsonConfigRateLimitOne,
						err:              errors.Append(ErrNewerEdit),
					},
				},
			},
		},
		{
			name: "maintain_commments_and_whitespace",
			sequence: []pair{
				{
					input{
						lastID: 0,
						contents: `{"disableAutoGitUpdates": true,

// This is a comment.
             "defaultRateLimit": 42,
             "auth.providers": [],
						}`,
					},
					output{
						ID: 2,
						contents: `{"disableAutoGitUpdates": true,

// This is a comment.
             "defaultRateLimit": 42,
             "auth.providers": [],
						}`,
						redactedContents: `{
  "disableAutoGitUpdates": true,
  // This is a comment.
  "defaultRateLimit": 42,
  "auth.providers": [],
}`,
					},
				},
			},
		},
		{
			name: "redact_sensitive_data",
			sequence: []pair{
				{
					input{
						lastID: 0,
						contents: `{"disableAutoGitUpdates": true,

		// This is a comment.
		             "defaultRateLimit": 42,
					 "auth.providers": [
					   {
						 "clientID": "sourcegraph-client-openid",
						 "clientSecret": "strongsecret",
						 "displayName": "Keycloak local OpenID Connect #1 (dev)",
						 "issuer": "http://localhost:3220/auth/realms/master",
						 "type": "openidconnect"
					   }
					 ]
								}`,
					},
					output{
						ID: 2,
						contents: `{"disableAutoGitUpdates": true,

		// This is a comment.
		             "defaultRateLimit": 42,
					 "auth.providers": [
					   {
						 "clientID": "sourcegraph-client-openid",
						 "clientSecret": "strongsecret",
						 "displayName": "Keycloak local OpenID Connect #1 (dev)",
						 "issuer": "http://localhost:3220/auth/realms/master",
						 "type": "openidconnect"
					   }
					 ]
								}`,
						redactedContents: `{
  "disableAutoGitUpdates": true,
  // This is a comment.
  "defaultRateLimit": 42,
  "auth.providers": [
    {
      "clientID": "sourcegraph-client-openid",
      "clientSecret": "REDACTED-DATA-CHUNK-f434ecc765",
      "displayName": "Keycloak local OpenID Connect #1 (dev)",
      "issuer": "http://localhost:3220/auth/realms/master",
      "type": "openidconnect"
    }
  ]
}`,
					},
				},
			},
		},
	} {
		// we were running the same test all the time, see this gist for more information
		// https://gist.github.com/posener/92a55c4cd441fc5e5e85f27bca008721
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			db := NewDB(logger, dbtest.NewDB(t))
			ctx := context.Background()
			for _, p := range test.sequence {
				output, err := db.Conf().SiteCreateIfUpToDate(ctx, &p.input.lastID, 0, p.input.contents, false)
				if err != nil {
					if errors.Is(err, p.expected.err) {
						continue
					}
					t.Fatal(err)
				}

				if output == nil {
					t.Fatal("got unexpected nil configuration after creation")
				}

				if diff := cmp.Diff(p.expected.contents, output.Contents); diff != "" {
					t.Fatalf("mismatched configuration contents after creation, (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(p.expected.redactedContents, output.RedactedContents); diff != "" {
					t.Fatalf("mismatched redacted_contents after creation, %v", diff)
				}

				if output.ID != p.expected.ID {
					t.Fatalf("returned configuration ID after creation - expected: %v, got:%v", p.expected.ID, output.ID)
				}

				latest, err := db.Conf().SiteGetLatest(ctx)
				if err != nil {
					t.Fatal(err)
				}

				if latest == nil {
					t.Fatalf("got unexpected nil configuration after GetLatest")
				}

				if latest.Contents != p.expected.contents {
					t.Fatalf("returned configuration contents after GetLatest - expected: %q, got:%q", p.expected.contents, latest.Contents)
				}
				if latest.ID != p.expected.ID {
					t.Fatalf("returned configuration ID after GetLatest - expected: %v, got:%v", p.expected.ID, latest.ID)
				}
			}
		})
	}
}

func createDummySiteConfigs(t *testing.T, ctx context.Context, s ConfStore) {
	config := `{"disableAutoGitUpdates": true, "auth.Providers": []}`

	siteConfig, err := s.SiteCreateIfUpToDate(ctx, nil, 0, config, false)
	require.NoError(t, err, "failed to create site config")

	// The first call to SiteCreatedIfUpToDate will always create a default entry if there are no
	// rows in the table yet and then eventually create another entry.
	//
	// lastID will be 2 here.
	lastID := siteConfig.ID

	// Change config so that we have a new entry in the DB - ID: 3
	config = `{"auth.Providers": []}`
	siteConfig, err = s.SiteCreateIfUpToDate(ctx, &lastID, 1, config, false)
	require.NoError(t, err, "failed to create site config")

	lastID = siteConfig.ID

	//  Create another entry with the same config - ID: 4
	siteConfig, err = s.SiteCreateIfUpToDate(ctx, &lastID, 1, config, false)
	require.NoError(t, err, "failed to create site config")

	lastID = siteConfig.ID

	// Change config again one last time, so that we have a new entry in the DB - ID: 5
	config = `{"disableAutoGitUpdates": true, "auth.Providers": []}`
	_, err = s.SiteCreateIfUpToDate(ctx, &lastID, 1, config, false)
	require.NoError(t, err, "failed to create site config")

	// By this point we have 5 entries instead of 4.
	// 3 and 4 are identical.
	// The unique list of configs is:
	// 5, 3, 2, 1
}

func TestGetSiteConfigCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	s := db.Conf()
	createDummySiteConfigs(t, ctx, s)

	count, err := s.GetSiteConfigCount(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// We have 5 entries in the DB, but we skip redundant ones so this returns 4.
	if count != 4 {
		t.Fatalf("Expected 4 site config entries, but got %d", count)
	}
}

func TestListSiteConfigs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	s := db.Conf()
	createDummySiteConfigs(t, ctx, s)

	if _, err := s.ListSiteConfigs(ctx, &PaginationArgs{}); err != nil {
		t.Error("Expected non-nil error but got nil")
	}

	testCases := []struct {
		name        string
		listOptions *PaginationArgs
		expectedIDs []int32
	}{
		{
			name:        "nil pagination args",
			expectedIDs: []int32{1, 2, 3, 5},
		},
		{
			name: "first: 2 (subset of data)",
			listOptions: &PaginationArgs{
				First: pointers.Ptr(2),
			},
			expectedIDs: []int32{5, 3},
		},
		{
			name: "last: 2 (subset of data)",
			listOptions: &PaginationArgs{
				Last: pointers.Ptr(2),
			},
			expectedIDs: []int32{1, 2},
		},
		{
			name: "first: 5 (all of data)",
			listOptions: &PaginationArgs{
				First: pointers.Ptr(5),
			},
			expectedIDs: []int32{5, 3, 2, 1},
		},
		{
			name: "last: 5 (all of data)",
			listOptions: &PaginationArgs{
				Last: pointers.Ptr(5),
			},
			expectedIDs: []int32{1, 2, 3, 5},
		},
		{
			name: "first: 10 (more than data)",
			listOptions: &PaginationArgs{
				First: pointers.Ptr(10),
			},
			expectedIDs: []int32{5, 3, 2, 1},
		},
		{
			name: "last: 10 (more than data)",
			listOptions: &PaginationArgs{
				Last: pointers.Ptr(10),
			},
			expectedIDs: []int32{1, 2, 3, 5},
		},
		{
			name: "first: 2, after: 5",
			listOptions: &PaginationArgs{
				First: pointers.Ptr(2),
				After: pointers.Ptr("5"),
			},
			expectedIDs: []int32{3, 2},
		},
		{
			name: "first: 6, after: 5 (overflow)",
			listOptions: &PaginationArgs{
				First: pointers.Ptr(6),
				After: pointers.Ptr("5"),
			},
			expectedIDs: []int32{3, 2, 1},
		},
		{
			name: "last: 2, after: 5",
			listOptions: &PaginationArgs{
				Last:  pointers.Ptr(2),
				After: pointers.Ptr("5"),
			},
			expectedIDs: []int32{1, 2},
		},
		{
			name: "last: 6, after: 5 (overflow)",
			listOptions: &PaginationArgs{
				Last:  pointers.Ptr(6),
				After: pointers.Ptr("5"),
			},
			expectedIDs: []int32{1, 2, 3},
		},
		{
			name: "first: 2, before: 1",
			listOptions: &PaginationArgs{
				First:  pointers.Ptr(2),
				Before: pointers.Ptr("1"),
			},
			expectedIDs: []int32{5, 3},
		},
		{
			name: "first: 6, before: 1 (overflow)",
			listOptions: &PaginationArgs{
				First:  pointers.Ptr(6),
				Before: pointers.Ptr("1"),
			},
			expectedIDs: []int32{5, 3, 2},
		},
		{
			name: "last: 2, before: 2",
			listOptions: &PaginationArgs{
				Last:   pointers.Ptr(2),
				Before: pointers.Ptr("2"),
			},
			expectedIDs: []int32{3, 5},
		},
		{
			name: "last: 6, before: 2 (overflow)",
			listOptions: &PaginationArgs{
				Last:   pointers.Ptr(6),
				Before: pointers.Ptr("2"),
			},
			expectedIDs: []int32{3, 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			siteConfigs, err := s.ListSiteConfigs(ctx, tc.listOptions)
			if err != nil {
				t.Fatal(err)
			}

			if len(siteConfigs) != len(tc.expectedIDs) {
				t.Fatalf("Expected %d site config entries but got %d", len(tc.expectedIDs), len(siteConfigs))
			}

			for i, siteConfig := range siteConfigs {
				if tc.expectedIDs[i] != siteConfig.ID {
					t.Errorf("Expected ID %d, but got %d", tc.expectedIDs[i], siteConfig.ID)
				}
			}
		})
	}
}
