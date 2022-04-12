package graphqlbackend

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"

	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestEstimateQueryCost(t *testing.T) {
	for _, tc := range []struct {
		name      string
		query     string
		variables map[string]interface{}
		want      QueryCost
	}{
		{
			name: "Multiple top level queries",
			query: `query {
  thing
}
query{
  thing
}
`,
			want: QueryCost{
				FieldCount: 2,
				MaxDepth:   1,
			},
		},
		{
			name: "Simple query, no variables",
			query: `
query SiteProductVersion {
                site {
                    productVersion
                    buildVersion
                    hasCodeIntelligence
                }
            }
`,
			want: QueryCost{
				FieldCount: 4,
				MaxDepth:   2,
			},
		},
		{
			name: "nodes field should not be counted",
			query: `
query{
  externalServices(first: 10){
    nodes{
      displayName
      webhookURL
    }
  }
  somethingElse
}
`,
			want: QueryCost{
				FieldCount: 22,
				MaxDepth:   3,
			},
		},
		{
			name: "Query with variables",
			query: `
query Extensions($first: Int!, $prioritizeExtensionIDs: [String!]!) {
                    extensionRegistry {
                        extensions(first: $first, prioritizeExtensionIDs: $prioritizeExtensionIDs) {
                            nodes {
                                id
                                extensionID
                                url
                                manifest {
                                    raw
                                }
                                viewerCanAdminister
                            }
                        }
                    }
                }
`,
			variables: map[string]interface{}{
				"first": 10,
			},
			want: QueryCost{
				FieldCount: 62,
				MaxDepth:   5,
			},
		},
		{
			name: "Query with default variables",
			query: `
query fetchExternalServices($first: Int = 10){
  externalServices(first: $first){
    nodes{
      displayName
      webhookURL
    }
  }
}
`,
			variables: map[string]interface{}{
				"first": 5,
			},
			want: QueryCost{
				FieldCount: 11,
				MaxDepth:   3,
			},
		},
		{
			name: "Query with default variables, non supplied",
			query: `
query fetchExternalServices($first: Int = 10){
  externalServices(first: $first){
    nodes{
      displayName
      webhookURL
    }
  }
}
`,
			variables: map[string]interface{}{},
			want: QueryCost{
				FieldCount: 21,
				MaxDepth:   3,
			},
		},
		{
			name: "Query with fragments",
			query: `
query StatusMessages {
	 statusMessages {
		 ...StatusMessageFields
	 }
 }
 fragment StatusMessageFields on StatusMessage {
	 __typename
	 ... on CloningProgress {
		 message
	 }
	 ... on SyncError {
		 message
	 }
	 ... on ExternalServiceSyncError {
		 message
		 externalService {
			 id
			 displayName
		 }
	 }
 }
`,
			want: QueryCost{
				FieldCount: 5,
				MaxDepth:   2,
			},
		},
		{
			name: "Simple inline fragments",
			query: `
query{
    __typename
	... on Foo {
         one
         two
     }
     ... on Bar {
         one
     }
}
`,
			want: QueryCost{
				FieldCount: 2,
				MaxDepth:   2,
			},
		},
		{
			name: "Search query",
			query: `
query Search($query: String!, $version: SearchVersion!, $patternType: SearchPatternType!) {
  search(
    query: $query
    version: $version
    patternType: $patternType
  ) {
    results {
      __typename
      limitHit
      matchCount
      approximateResultCount
      missing {
        name
      }
      cloning {
        name
      }
      repositoriesCount
      timedout {
        name
      }
      indexUnavailable
      dynamicFilters {
        value
        label
        count
        limitHit
        kind
      }
      results {
        __typename
        ... on Repository {
          id
          name
          label {
            html
          }
          url
          icon
          detail {
            html
          }
          matches {
            url
            body {
              text
              html
            }
            highlights {
              line
              character
              length
            }
          }
        }
        ... on FileMatch {
          file {
            path
            url
            commit {
              oid
            }
          }
          repository {
            name
            url
          }
          revSpec {
            __typename
            ... on GitRef {
              displayName
              url
            }
            ... on GitRevSpecExpr {
              expr
              object {
                commit {
                  url
                }
              }
            }
            ... on GitObject {
              abbreviatedOID
              commit {
                url
              }
            }
          }
          limitHit
          symbols {
            name
            containerName
            url
            kind
          }
          lineMatches {
            preview
            lineNumber
            offsetAndLengths
          }
        }
        ... on CommitSearchResult {
          label {
            html
          }
          url
          icon
          detail {
            html
          }
          matches {
            url
            body {
              text
              html
            }
            highlights {
              line
              character
              length
            }
          }
        }
      }
      alert {
        title
        description
        proposedQueries {
          description
          query
        }
      }
      elapsedMilliseconds
    }
  }
}
`,
			want: QueryCost{
				FieldCount: 50,
				MaxDepth:   9,
			},
		},
		{
			name: "Allow null variables",
			// NOTE: $first is nullable
			query: `
query RepositoryComparisonDiff($repo: String!, $base: String, $head: String, $first: Int) {
  repository(name: $repo) {
    comparison(base: $base, head: $head) {
      fileDiffs(first: $first) {
        nodes {
          ...FileDiffFields
        }
        totalCount
      }
    }
  }
}

fragment FileDiffFields on FileDiff {
  oldPath
  newPath
  internalID
}
`,
			want: QueryCost{
				FieldCount: 7,
				MaxDepth:   5,
			},
			variables: map[string]interface{}{
				"base": "a46cf4a8b6dc42ea7b7b716e53c49dd3508a8678",
				"head": "0fd3fb1f4e41ae1f95970beeec1c1f7b2d8a7d06",
				"repo": "github.com/presslabs/mysql-operator",
			},
		},
		{
			name: "Nested named fragments",
			query: `
query{
    __typename
	...FooFields
}
fragment FooFields on Foo {
	...BarFields
}
fragment BarFields on Bar {
	one
}
`,
			want: QueryCost{
				FieldCount: 1,
				MaxDepth:   1,
			},
		},
		{
			name: "More nested fragments",
			query: `
{
  node {
    ...FileFragment
  }
}

fragment FileFragment on File {
  ... on Usable {
    ...UsableFields
  }
}

fragment UsableFields on Usable {
  isUsable
}
`,
			want: QueryCost{
				FieldCount: 3,
				MaxDepth:   2,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.want
			want.Version = costEstimateVersion
			have, err := EstimateQueryCost(tc.query, tc.variables)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(want, *have); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestRatelimitFromConfig(t *testing.T) {
	testCases := []struct {
		name   string
		config *schema.ApiRatelimit

		uid  string
		isIP bool
		cost int

		enabled bool

		wantLimited bool
		wantResult  throttled.RateLimitResult
	}{
		{
			name:   "Nil config",
			config: nil,

			enabled: false,
		},
		{
			name: "Disabled config",
			config: &schema.ApiRatelimit{
				Enabled: false,
			},
			enabled: false,
		},
		{
			name: "No overrides",
			config: &schema.ApiRatelimit{
				Enabled:   true,
				Overrides: nil,
				PerIP:     5000,
				PerUser:   5000,
			},
			enabled: true,

			uid:  "test",
			isIP: false,
			cost: 1,

			wantLimited: false,
			wantResult: throttled.RateLimitResult{
				Limit:      1001,
				Remaining:  1000,
				ResetAfter: 720 * time.Millisecond,
				RetryAfter: -1,
			},
		},
		{
			name: "With value override",
			config: &schema.ApiRatelimit{
				Enabled: true,
				PerIP:   5000,
				PerUser: 5000,
				Overrides: []*schema.Overrides{
					{
						Key:   "test",
						Limit: 2500,
					},
				},
			},
			enabled: true,

			uid:  "test",
			isIP: false,
			cost: 1,

			wantLimited: false,
			wantResult: throttled.RateLimitResult{
				Limit:      501,
				Remaining:  500,
				ResetAfter: 720 * time.Millisecond * 2,
				RetryAfter: -1,
			},
		},
		{
			name: "With blocked override",
			config: &schema.ApiRatelimit{
				Enabled: true,
				PerIP:   5000,
				PerUser: 5000,
				Overrides: []*schema.Overrides{
					{
						Key:   "test",
						Limit: "blocked",
					},
				},
			},
			enabled: true,

			uid:  "test",
			isIP: false,
			cost: 1,

			wantLimited: true,
			wantResult: throttled.RateLimitResult{
				Limit:      0,
				Remaining:  0,
				ResetAfter: 0,
				RetryAfter: 0,
			},
		},
		{
			name: "With unlimited override",
			config: &schema.ApiRatelimit{
				Enabled: true,
				PerIP:   5000,
				PerUser: 5000,
				Overrides: []*schema.Overrides{
					{
						Key:   "test",
						Limit: "unlimited",
					},
				},
			},
			enabled: true,

			uid:  "test",
			isIP: false,
			cost: 1,

			wantLimited: false,
			wantResult: throttled.RateLimitResult{
				Limit:      100000,
				Remaining:  100000,
				ResetAfter: 0,
				RetryAfter: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store, err := memstore.New(1024)
			if err != nil {
				t.Fatal(err)
			}
			rlw := &RateLimitWatcher{
				store: store,
			}

			rlw.updateFromConfig(tc.config)
			rl, enabled := rlw.Get()

			if tc.enabled != enabled {
				t.Fatalf("Want %v, got %v", tc.enabled, enabled)
			}
			if !tc.enabled {
				return
			}
			limited, result, err := rl.RateLimit(tc.uid, tc.cost, LimiterArgs{IsIP: tc.isIP})
			if err != nil {
				t.Fatal(err)
			}
			if limited != tc.wantLimited {
				t.Fatalf("Limited, want %v got %v", tc.wantLimited, limited)
			}
			if diff := cmp.Diff(tc.wantResult, result); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestBasicLimiterEnabled(t *testing.T) {
	tests := []struct {
		limit       int
		wantEnabled bool
	}{
		{
			limit:       1,
			wantEnabled: true,
		},
		{
			limit:       100,
			wantEnabled: true,
		},
		{
			limit:       0,
			wantEnabled: false,
		},
		{
			limit:       -1,
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("limit:%d", tt.limit), func(t *testing.T) {
			store, err := memstore.New(1)
			if err != nil {
				t.Fatal(err)
			}
			bl := BasicLimitWatcher{store: store}
			bl.updateFromConfig(tt.limit)

			_, enabled := bl.Get()

			if got := enabled; got != tt.wantEnabled {
				t.Fatalf("got %t, want %t", got, tt.wantEnabled)
			}
		})
	}
}

func TestBasicLimiter(t *testing.T) {
	store, err := memstore.New(1)
	if err != nil {
		t.Fatal(err)
	}
	bl := BasicLimitWatcher{store: store}
	bl.updateFromConfig(1)

	limiter, enabled := bl.Get()
	if !enabled {
		t.Fatalf("got %t, want true", enabled)
	}

	// These arguments correspond to call we want to limit.
	limiterArgs := LimiterArgs{
		Anonymous:     true,
		RequestName:   "unknown",
		RequestSource: trace.SourceOther,
	}

	// 1st call should not be limited.
	limited, _, err := limiter.RateLimit("", 1, limiterArgs)
	if err != nil {
		t.Fatal(err)
	}
	if limited {
		t.Fatalf("got %t, want false", limited)
	}

	// 2nd call should be limited.
	limited, _, err = limiter.RateLimit("", 1, limiterArgs)
	if err != nil {
		t.Fatal(err)
	}
	if !limited {
		t.Fatalf("got %t, want true", limited)
	}
}
