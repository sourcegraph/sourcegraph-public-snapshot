package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

func TestSetDefaultQueryCount(t *testing.T) {
	for in, want := range map[string]string{
		"":                     hardCodedCount,
		"count:10":             "count:10",
		"r:foo":                "r:foo" + hardCodedCount,
		"r:foo count:10":       "r:foo count:10",
		"r:foo count:10 f:bar": "r:foo count:10 f:bar",
		"r:foo count:":         "r:foo count:" + hardCodedCount,
		"r:foo count:xyz":      "r:foo count:xyz" + hardCodedCount,
	} {
		t.Run(in, func(t *testing.T) {
			have := setDefaultQueryCount(in)
			if have != want {
				t.Errorf("unexpected query: have %q; want %q", have, want)
			}
		})
	}
}

func TestResolveRepositorySearch(t *testing.T) {
	client, done := mockGraphQLClient(testResolveRepositorySearchResult)
	defer done()

	svc := &Service{client: client}

	repos, err := svc.resolveRepositorySearch(context.Background(), "sourcegraph reconciler or src-cli")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// We want multiple FileMatches in the same repository to result in one repository
	// and Repository and FileMatches should be folded into the same repository.
	if have, want := len(repos), 3; have != want {
		t.Fatalf("wrong number of repos. want=%d, have=%d", want, have)
	}
}

const testResolveRepositorySearchResult = `{
  "data": {
    "search": {
      "results": {
        "results": [
          {
            "__typename": "FileMatch",
            "file": { "path": "website/src/components/PricingTable.tsx" },
            "repository": {
              "id": "UmVwb3NpdG9yeTo0MDc=",
              "name": "github.com/sd9/about",
              "url": "/github.com/sd9/about",
              "externalRepository": { "serviceType": "github" },
              "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "1576f235209fbbfc918129db35f3d108347b74cb" } }
            }
          },
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeToxOTM=",
            "name": "github.com/sd9/src-cli",
            "url": "/github.com/sd9/src-cli",
            "externalRepository": { "serviceType": "github" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "21dd58b08d64620942401b5543f5b0d33498bacb" } }
          },
          {
            "__typename": "FileMatch",
            "file": { "path": "cmd/src/api.go" },
            "repository": {
              "id": "UmVwb3NpdG9yeToxOTM=",
              "name": "github.com/sd9/src-cli",
              "url": "/github.com/sd9/src-cli",
              "externalRepository": { "serviceType": "github" },
              "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "21dd58b08d64620942401b5543f5b0d33498bacb" } }
            }
          },
          {
            "__typename": "FileMatch",
            "file": { "path": "client/web/src/components/externalServices/externalServices.tsx" },
            "repository": {
              "id": "UmVwb3NpdG9yeTo2",
              "name": "github.com/sourcegraph/sourcegraph",
              "url": "/github.com/sourcegraph/sourcegraph",
              "externalRepository": { "serviceType": "github" },
              "defaultBranch": { "name": "refs/heads/main", "target": { "oid": "e612c34f2e27005928ff3dfdd8e5ead5a0cb1526" } }
            }
          },
          {
            "__typename": "FileMatch",
            "file": { "path": "cmd/frontend/internal/httpapi/src_cli.go" },
            "repository": {
              "id": "UmVwb3NpdG9yeTo2",
              "name": "github.com/sourcegraph/sourcegraph",
              "url": "/github.com/sourcegraph/sourcegraph",
              "externalRepository": { "serviceType": "github" },
              "defaultBranch": { "name": "refs/heads/main", "target": { "oid": "e612c34f2e27005928ff3dfdd8e5ead5a0cb1526" } }
            }
          }
        ]
      }
    }
  }
}
`

func mockGraphQLClient(responses ...string) (client api.Client, done func()) {
	mux := http.NewServeMux()

	var count int
	mux.HandleFunc("/.api/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_, _ = w.Write([]byte(responses[count]))

		if count < len(responses)-1 {
			count += 1
		}
	})

	ts := httptest.NewServer(mux)

	var clientBuffer bytes.Buffer
	client = api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

	return client, ts.Close
}

func TestResolveRepositories_Unsupported(t *testing.T) {
	client, done := mockGraphQLClient(testResolveRepositoriesUnsupported)
	defer done()

	spec := &batches.BatchSpec{
		On: []batches.OnQueryOrRepository{
			{RepositoriesMatchingQuery: "testquery"},
		},
	}

	t.Run("allowUnsupported:true", func(t *testing.T) {
		svc := &Service{client: client, allowUnsupported: true, allowIgnored: true}

		repos, err := svc.ResolveRepositories(context.Background(), spec)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(repos) != 4 {
			t.Fatalf("wrong number of repos. want=%d, have=%d", 4, len(repos))
		}
	})

	t.Run("allowUnsupported:false", func(t *testing.T) {
		svc := &Service{client: client, allowUnsupported: false, allowIgnored: true}

		repos, err := svc.ResolveRepositories(context.Background(), spec)
		repoSet, ok := err.(batches.UnsupportedRepoSet)
		if !ok {
			t.Fatalf("err is not UnsupportedRepoSet")
		}
		if len(repoSet) != 1 {
			t.Fatalf("wrong number of repos. want=%d, have=%d", 1, len(repoSet))
		}
		if len(repos) != 3 {
			t.Fatalf("wrong number of repos. want=%d, have=%d", 3, len(repos))
		}
	})
}

const testResolveRepositoriesUnsupported = `{
  "data": {
    "search": {
      "results": {
        "results": [
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeToxMw==",
            "name": "bitbucket.sgdev.org/SOUR/automation-testing",
            "url": "/bitbucket.sgdev.org/SOUR/automation-testing",
            "externalRepository": { "serviceType": "bitbucketserver" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "b978d56de5578a935ca0bf07b56528acc99d5a61" } }
          },
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeTo0",
            "name": "github.com/sourcegraph/automation-testing",
            "url": "/github.com/sourcegraph/automation-testing",
            "externalRepository": { "serviceType": "github" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "6ac8a32ecaf6c4dc5ce050b9af51bce3db8efd63" } }
          },
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeTo2MQ==",
            "name": "gitlab.sgdev.org/sourcegraph/automation-testing",
            "url": "/gitlab.sgdev.org/sourcegraph/automation-testing",
            "externalRepository": { "serviceType": "gitlab" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "3b79a5d479d2af9cfe91e0aad4e9dddca7278150" } }
          },
		  {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeToxNDg=",
            "name": "git-codecommit.us-est-1.amazonaws.com/stripe-go",
            "url": "/git-codecommit.us-est-1.amazonaws.com/stripe-go",
            "externalRepository": { "serviceType": "awscodecommit" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "3b79a5d479d2af9cfe91e0aad4e9dddca7278150" } }
          }
        ]
      }
    }
  }
}
`

func TestResolveRepositories_Ignored(t *testing.T) {
	spec := &batches.BatchSpec{
		On: []batches.OnQueryOrRepository{
			{RepositoriesMatchingQuery: "testquery"},
		},
	}

	t.Run("allowIgnored:true", func(t *testing.T) {
		client, done := mockGraphQLClient(testResolveRepositories, testBatchIgnoreInRepos)
		defer done()

		svc := &Service{client: client, allowIgnored: true}

		repos, err := svc.ResolveRepositories(context.Background(), spec)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(repos) != 3 {
			t.Fatalf("wrong number of repos. want=%d, have=%d", 3, len(repos))
		}
	})

	t.Run("allowIgnored:false", func(t *testing.T) {
		client, done := mockGraphQLClient(testResolveRepositories, testBatchIgnoreInRepos)
		defer done()

		svc := &Service{client: client, allowIgnored: false}

		repos, err := svc.ResolveRepositories(context.Background(), spec)
		ignored, ok := err.(batches.IgnoredRepoSet)
		if !ok {
			t.Fatalf("err is not IgnoredRepoSet: %s", err)
		}
		if len(ignored) != 1 {
			t.Fatalf("wrong number of ignored repos. want=%d, have=%d", 1, len(ignored))
		}
		if len(repos) != 2 {
			t.Fatalf("wrong number of repos. want=%d, have=%d", 2, len(repos))
		}
	})
}

const testResolveRepositories = `{
  "data": {
    "search": {
      "results": {
        "results": [
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeToxMw==",
            "name": "bitbucket.sgdev.org/SOUR/automation-testing",
            "url": "/bitbucket.sgdev.org/SOUR/automation-testing",
            "externalRepository": { "serviceType": "bitbucketserver" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "b978d56de5578a935ca0bf07b56528acc99d5a61" } }
          },
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeTo0",
            "name": "github.com/sourcegraph/automation-testing",
            "url": "/github.com/sourcegraph/automation-testing",
            "externalRepository": { "serviceType": "github" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "6ac8a32ecaf6c4dc5ce050b9af51bce3db8efd63" } }
          },
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeTo2MQ==",
            "name": "gitlab.sgdev.org/sourcegraph/automation-testing",
            "url": "/gitlab.sgdev.org/sourcegraph/automation-testing",
            "externalRepository": { "serviceType": "gitlab" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "3b79a5d479d2af9cfe91e0aad4e9dddca7278150" } }
          }
        ]
      }
    }
  }
}
`

const testBatchIgnoreInRepos = `{
    "data": {
        "repo_0": {
            "results": {
                "results": [
                    {
                        "__typename": "FileMatch",
                        "file": {
                            "path": ".batchignore"
                        }
                    }
                ]
            }
        },
        "repo_1": { "results": { "results": [] } },
        "repo_2": { "results": { "results": [] } }
    }
}
`

func TestService_FindDirectoriesInRepos(t *testing.T) {
	client, done := mockGraphQLClient(testFindDirectoriesInRepos)
	defer done()

	fileName := "package.json"
	repos := []*graphql.Repository{
		{ID: "repo-id-0", Name: "github.com/sourcegraph/automation-testing"},
		{ID: "repo-id-1", Name: "github.com/sourcegraph/sourcegraph"},
	}

	svc := &Service{client: client}

	results, err := svc.FindDirectoriesInRepos(context.Background(), fileName, repos...)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(results) != len(repos) {
		t.Fatalf("wrong number of repos. want=%d, have=%d", 4, len(repos))
	}

	want := map[*graphql.Repository][]string{
		repos[0]: {"examples/project3", "project1", "project2"},
		repos[1]: {"docs/client1", ".", "docs/client2/examples"},
	}

	if !cmp.Equal(want, results, cmpopts.SortSlices(sortStrings)) {
		t.Errorf("wrong tasks built (-want +got):\n%s", cmp.Diff(want, results))
	}
}

func sortStrings(a, b string) bool { return a < b }

const testFindDirectoriesInRepos = `{
    "data": {
        "repo_0": {
            "results": {
                "results": [
                    {
                        "__typename": "FileMatch",
                        "file": {
                            "path": "examples/project3/package.json"
                        }
                    },
                    {
                        "__typename": "FileMatch",
                        "file": {
                            "path": "project1/package.json"
                        }
                    },
                    {
                        "__typename": "FileMatch",
                        "file": {
                            "path": "project2/package.json"
                        }
                    }
                ]
            }
        },
        "repo_1": {
            "results": {
                "results": [
                    {
                        "__typename": "FileMatch",
                        "file": {
                            "path": "docs/client1/package.json"
                        }
                    },
                    {
                        "__typename": "FileMatch",
                        "file": {
                            "path": "package.json"
                        }
                    },
                    {
                        "__typename": "FileMatch",
                        "file": {
                            "path": "docs/client2/examples/package.json"
                        }
                    }
                ]
            }
        }
    }
}
`

func TestService_ValidateChangesetSpecs(t *testing.T) {
	repo1 := &graphql.Repository{ID: "repo-graphql-id-1", Name: "github.com/sourcegraph/src-cli"}
	repo2 := &graphql.Repository{ID: "repo-graphql-id-2", Name: "github.com/sourcegraph/sourcegraph"}

	tests := map[string]struct {
		repos []*graphql.Repository
		specs []*batches.ChangesetSpec

		wantErrInclude string
	}{
		"no errors": {
			repos: []*graphql.Repository{repo1, repo2},
			specs: []*batches.ChangesetSpec{
				{CreatedChangeset: &batches.CreatedChangeset{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-1"},
				},
				{CreatedChangeset: &batches.CreatedChangeset{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-2"},
				},
				{CreatedChangeset: &batches.CreatedChangeset{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-1"},
				},
				{CreatedChangeset: &batches.CreatedChangeset{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-2"},
				},
			},
		},

		"imported changeset": {
			repos: []*graphql.Repository{repo1},
			specs: []*batches.ChangesetSpec{
				{ExternalChangeset: &batches.ExternalChangeset{
					ExternalID: "123",
				}},
			},
			// This should not fail validation ever.
		},

		"duplicate branches": {
			repos: []*graphql.Repository{repo1, repo2},
			specs: []*batches.ChangesetSpec{
				{CreatedChangeset: &batches.CreatedChangeset{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-1"},
				},
				{CreatedChangeset: &batches.CreatedChangeset{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-2"},
				},
				{CreatedChangeset: &batches.CreatedChangeset{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-1"},
				},
				{CreatedChangeset: &batches.CreatedChangeset{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-1"},
				},
			},
			wantErrInclude: `github.com/sourcegraph/sourcegraph: 2 changeset specs have the branch "branch-1"`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := &Service{}
			haveErr := svc.ValidateChangesetSpecs(tt.repos, tt.specs)
			if tt.wantErrInclude != "" {
				if haveErr == nil {
					t.Fatalf("expected %q to be included in error, but got none", tt.wantErrInclude)
				} else {
					if !strings.Contains(haveErr.Error(), tt.wantErrInclude) {
						t.Fatalf("expected %q to be included in error, but was not. error=%q", tt.wantErrInclude, haveErr.Error())
					}
				}
			} else {
				if haveErr != nil {
					t.Fatalf("unexpected error: %s", haveErr)
				}
			}
		})
	}
}
