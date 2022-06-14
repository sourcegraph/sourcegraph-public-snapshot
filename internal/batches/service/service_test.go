package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/docker"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/mock"
)

func TestSetDefaultQueryCount(t *testing.T) {
	for in, want := range map[string]string{
		"":                     hardCodedCount,
		"count:10":             "count:10",
		"count:all":            "count:all",
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

	spec := &batcheslib.BatchSpec{
		On: []batcheslib.OnQueryOrRepository{
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
	spec := &batcheslib.BatchSpec{
		On: []batcheslib.OnQueryOrRepository{
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

func TestResolveRepositories_RepoWithoutBranch(t *testing.T) {
	spec := &batcheslib.BatchSpec{
		On: []batcheslib.OnQueryOrRepository{
			{RepositoriesMatchingQuery: "testquery"},
		},
	}

	client, done := mockGraphQLClient(testResolveRepositoriesNoBranch, testBatchIgnoreInReposNoBranch)
	defer done()

	svc := &Service{client: client, allowIgnored: false}

	repos, err := svc.ResolveRepositories(context.Background(), spec)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(repos) != 1 {
		t.Fatalf("wrong number of repos. want=%d, have=%d", 2, len(repos))
	}
}

const testResolveRepositoriesNoBranch = `{
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
            "defaultBranch": null
          },
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeTo0",
            "name": "github.com/sourcegraph/automation-testing",
            "url": "/github.com/sourcegraph/automation-testing",
            "externalRepository": { "serviceType": "github" },
            "defaultBranch": null
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

const testBatchIgnoreInReposNoBranch = `{
    "data": {
        "repo_0": { "results": { "results": [] } }
    }
}
`

func TestService_FindDirectoriesInRepos(t *testing.T) {
	client, done := mockGraphQLClient(testFindDirectoriesInRepos)
	defer done()

	fileName := "package.json"
	repos := []*graphql.Repository{
		{
			ID:     "repo-id-0",
			Name:   "github.com/sourcegraph/automation-testing",
			Branch: graphql.Branch{Name: "dev"},
			DefaultBranch: &graphql.Branch{
				Name: "main",
				Target: graphql.Target{
					OID: "d34db33f",
				},
			},
		},
		{
			ID:     "repo-id-1",
			Name:   "github.com/sourcegraph/sourcegraph",
			Branch: graphql.Branch{Name: "dev"},
			DefaultBranch: &graphql.Branch{
				Name: "main",
				Target: graphql.Target{
					OID: "d34db33f",
				},
			},
		},
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
		repos[1]: {"docs/client1", "", "docs/client2/examples"},
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
		specs []*batcheslib.ChangesetSpec

		wantErrInclude string
	}{
		"no errors": {
			repos: []*graphql.Repository{repo1, repo2},
			specs: []*batcheslib.ChangesetSpec{
				{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-1",
				},
				{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-2",
				},
				{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-1",
				},
				{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-2",
				},
			},
		},

		"imported changeset": {
			repos: []*graphql.Repository{repo1},
			specs: []*batcheslib.ChangesetSpec{
				{
					ExternalID: "123",
				},
			},
			// This should not fail validation ever.
		},

		"duplicate branches": {
			repos: []*graphql.Repository{repo1, repo2},
			specs: []*batcheslib.ChangesetSpec{
				{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-1",
				},
				{
					HeadRepository: repo1.ID, HeadRef: "refs/heads/branch-2",
				},
				{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-1",
				},
				{
					HeadRepository: repo2.ID, HeadRef: "refs/heads/branch-1",
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
				} else if !strings.Contains(haveErr.Error(), tt.wantErrInclude) {
					t.Fatalf("expected %q to be included in error, but was not. error=%q", tt.wantErrInclude, haveErr.Error())
				}
			} else {
				if haveErr != nil {
					t.Fatalf("unexpected error: %s", haveErr)
				}
			}
		})
	}
}

func TestEnsureDockerImages(t *testing.T) {
	ctx := context.Background()
	parallelCases := []int{0, 1, 2, 4, 8}

	newServiceWithImages := func(images map[string]docker.Image) *Service {
		return &Service{
			imageCache: &mock.ImageCache{Images: images},
		}
	}

	t.Run("success", func(t *testing.T) {
		t.Run("single image", func(t *testing.T) {
			// A zeroed mock.Image should be usable for testing purposes.
			image := &mock.Image{}
			images := map[string]docker.Image{
				"image": image,
			}

			for name, steps := range map[string][]batcheslib.Step{
				"single step":    {{Container: "image"}},
				"multiple steps": {{Container: "image"}, {Container: "image"}},
			} {
				t.Run(name, func(t *testing.T) {
					for _, parallelism := range parallelCases {
						t.Run(fmt.Sprintf("%d worker(s)", parallelism), func(t *testing.T) {
							svc := newServiceWithImages(images)
							progress := &mock.Progress{}

							have, err := svc.EnsureDockerImages(ctx, steps, parallelism, progress.Callback())
							assert.Nil(t, err)
							assert.Equal(t, images, have)
							assert.Equal(t, []mock.ProgressCall{
								{Done: 0, Total: 1},
								{Done: 1, Total: 1},
							}, progress.Calls)
						})
					}
				})
			}
		})

		t.Run("multiple images", func(t *testing.T) {
			var (
				imageA = &mock.Image{}
				imageB = &mock.Image{}
				imageC = &mock.Image{}
				images = map[string]docker.Image{
					"a": imageA,
					"b": imageB,
					"c": imageC,
				}
			)

			for _, parallelism := range parallelCases {
				t.Run(fmt.Sprintf("%d worker(s)", parallelism), func(t *testing.T) {
					svc := newServiceWithImages(images)
					progress := &mock.Progress{}

					have, err := svc.EnsureDockerImages(ctx, []batcheslib.Step{
						{Container: "a"},
						{Container: "a"},
						{Container: "a"},
						{Container: "b"},
						{Container: "c"},
					}, parallelism, progress.Callback())
					assert.Nil(t, err)
					assert.Equal(t, images, have)
					assert.Equal(t, []mock.ProgressCall{
						{Done: 0, Total: 3},
						{Done: 1, Total: 3},
						{Done: 2, Total: 3},
						{Done: 3, Total: 3},
					}, progress.Calls)
				})
			}
		})
	})

	t.Run("errors", func(t *testing.T) {
		// The only really interesting case is where an image fails — we want to
		// ensure that the error is propagated, and that we don't end up
		// deadlocking while the context cancellation propagates. Let's set up a
		// good number of images (and steps) so we can give this a good test.
		wantErr := errors.New("expected error")
		images := map[string]docker.Image{}
		steps := []batcheslib.Step{}

		total := 100
		for i := 0; i < total; i++ {
			name := strconv.Itoa(i)
			if i%25 == 0 {
				images[name] = &mock.Image{EnsureErr: wantErr}
			} else {
				images[name] = &mock.Image{}
			}
			for j := 0; j < (i%10)+1; j++ {
				steps = append(steps, batcheslib.Step{Container: name})
			}
		}

		// Just verify we did that right!
		assert.Len(t, images, total)
		assert.True(t, len(steps) > total)

		for _, parallelism := range parallelCases {
			t.Run(fmt.Sprintf("%d worker(s)", parallelism), func(t *testing.T) {
				svc := newServiceWithImages(images)
				progress := &mock.Progress{}

				have, err := svc.EnsureDockerImages(ctx, steps, parallelism, progress.Callback())
				assert.ErrorIs(t, err, wantErr)
				assert.Nil(t, have)

				// Because there's no particular order the images will be fetched in,
				// the number of progress calls we get is non-deterministic, other than
				// that we should always get the first one.
				assert.Equal(t, mock.ProgressCall{Done: 0, Total: total}, progress.Calls[0])
			})
		}

	})
}

func TestService_ParseBatchSpec(t *testing.T) {
	svc := &Service{}

	tempDir := t.TempDir()
	tempOutsideDir := t.TempDir()
	// create temp files/dirs that can be used by the tests
	_, err := os.Create(filepath.Join(tempDir, "sample.sh"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(tempDir, "another.sh"))
	require.NoError(t, err)

	tests := []struct {
		name         string
		batchSpecDir string
		rawSpec      string
		isRemote     bool
		expectedSpec *batcheslib.BatchSpec
		expectedErr  error
	}{
		{
			name: "simple spec",
			rawSpec: `
name: test-spec
description: A test spec
`,
			expectedSpec: &batcheslib.BatchSpec{Name: "test-spec", Description: "A test spec"},
		},
		{
			name: "unknown field",
			rawSpec: `
name: test-spec
description: A test spec
some-new-field: Foo bar
`,
			expectedErr: errors.New("parsing batch spec: Additional property some-new-field is not allowed"),
		},
		{
			name:         "mount absolute file",
			batchSpecDir: tempDir,
			rawSpec: fmt.Sprintf(`
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: %s
        mountpoint: /tmp/sample.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`, filepath.Join(tempDir, "sample.sh")),
			expectedSpec: &batcheslib.BatchSpec{
				Name:        "test-spec",
				Description: "A test spec",
				Steps: []batcheslib.Step{
					{
						Run:       "/tmp/sample.sh",
						Container: "alpine:3",
						Mount: []batcheslib.Mount{
							{
								Path:       filepath.Join(tempDir, "sample.sh"),
								Mountpoint: "/tmp/sample.sh",
							},
						},
					},
				},
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title:  "Test Mount",
					Body:   "Test a mounted path",
					Branch: "test",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "Test",
					},
				},
			},
		},
		{
			name:         "mount relative file",
			batchSpecDir: tempDir,
			rawSpec: `
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: ./sample.sh
        mountpoint: /tmp/sample.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`,
			expectedSpec: &batcheslib.BatchSpec{
				Name:        "test-spec",
				Description: "A test spec",
				Steps: []batcheslib.Step{
					{
						Run:       "/tmp/sample.sh",
						Container: "alpine:3",
						Mount: []batcheslib.Mount{
							{
								Path:       filepath.Join(tempDir, "sample.sh"),
								Mountpoint: "/tmp/sample.sh",
							},
						},
					},
				},
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title:  "Test Mount",
					Body:   "Test a mounted path",
					Branch: "test",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "Test",
					},
				},
			},
		},
		{
			name:         "mount absolute directory",
			batchSpecDir: tempDir,
			rawSpec: fmt.Sprintf(`
name: test-spec
description: A test spec
steps:
  - run: /tmp/some-file.sh
    container: alpine:3
    mount:
      - path: %s
        mountpoint: /tmp
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`, tempDir),
			expectedSpec: &batcheslib.BatchSpec{
				Name:        "test-spec",
				Description: "A test spec",
				Steps: []batcheslib.Step{
					{
						Run:       "/tmp/some-file.sh",
						Container: "alpine:3",
						Mount: []batcheslib.Mount{
							{
								Path:       tempDir + string(filepath.Separator),
								Mountpoint: "/tmp",
							},
						},
					},
				},
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title:  "Test Mount",
					Body:   "Test a mounted path",
					Branch: "test",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "Test",
					},
				},
			},
		},
		{
			name:         "mount relative directory",
			batchSpecDir: tempDir,
			rawSpec: `
name: test-spec
description: A test spec
steps:
  - run: /tmp/some-file.sh
    container: alpine:3
    mount:
      - path: ./
        mountpoint: /tmp
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`,
			expectedSpec: &batcheslib.BatchSpec{
				Name:        "test-spec",
				Description: "A test spec",
				Steps: []batcheslib.Step{
					{
						Run:       "/tmp/some-file.sh",
						Container: "alpine:3",
						Mount: []batcheslib.Mount{
							{
								Path:       tempDir + string(filepath.Separator),
								Mountpoint: "/tmp",
							},
						},
					},
				},
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title:  "Test Mount",
					Body:   "Test a mounted path",
					Branch: "test",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "Test",
					},
				},
			},
		},
		{
			name:         "mount multiple files",
			batchSpecDir: tempDir,
			rawSpec: `
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh && /tmp/another.sh
    container: alpine:3
    mount:
      - path: ./sample.sh
        mountpoint: /tmp/sample.sh
      - path: ./another.sh
        mountpoint: /tmp/another.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`,
			expectedSpec: &batcheslib.BatchSpec{
				Name:        "test-spec",
				Description: "A test spec",
				Steps: []batcheslib.Step{
					{
						Run:       "/tmp/sample.sh && /tmp/another.sh",
						Container: "alpine:3",
						Mount: []batcheslib.Mount{
							{
								Path:       filepath.Join(tempDir, "sample.sh"),
								Mountpoint: "/tmp/sample.sh",
							},
							{
								Path:       filepath.Join(tempDir, "another.sh"),
								Mountpoint: "/tmp/another.sh",
							},
						},
					},
				},
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title:  "Test Mount",
					Body:   "Test a mounted path",
					Branch: "test",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "Test",
					},
				},
			},
		},
		{
			name:         "mount path does not exist",
			batchSpecDir: tempDir,
			rawSpec: fmt.Sprintf(`
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: %s
        mountpoint: /tmp
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`, filepath.Join(tempDir, "does", "not", "exist", "sample.sh")),
			expectedErr: errors.Newf("handling mount: step 1 mount path %s does not exist", filepath.Join(tempDir, "does", "not", "exist", "sample.sh")),
		},
		{
			name:         "mount path not subdirectory of spec",
			batchSpecDir: tempDir,
			rawSpec: fmt.Sprintf(`
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: %s
        mountpoint: /tmp
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`, tempOutsideDir),
			expectedErr: errors.New("handling mount: step 1 mount path is not in the same directory or subdirectory as the batch spec"),
		},
		{
			name:         "mount remote processing",
			batchSpecDir: tempDir,
			rawSpec: `
name: test-spec
description: A test spec
steps:
  - run: /tmp/foo/bar/sample.sh
    container: alpine:3
    mount:
      - path: /valid/sample.sh
        mountpoint: /tmp/foo/bar/sample.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`,
			isRemote:    true,
			expectedErr: errors.New("handling mount: mounts are not support for server-side processing"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec, err := svc.ParseBatchSpec(test.batchSpecDir, []byte(test.rawSpec), test.isRemote)
			if test.expectedErr != nil {
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedSpec, spec)
			}
		})
	}
}
