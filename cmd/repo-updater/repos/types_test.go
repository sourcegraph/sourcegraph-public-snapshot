package repos

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"golang.org/x/time/rate"
)

func TestExternalService_Exclude(t *testing.T) {
	now := time.Now()

	type testCase struct {
		name   string
		svcs   ExternalServices
		repos  Repos
		assert ExternalServicesAssertion
	}

	githubService := ExternalService{
		Kind:        "GITHUB",
		DisplayName: "Github",
		Config: `{
			// Some comment
			"url": "https://github.com",
			"token": "secret",
			"repositoryQuery": ["none"]
		}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	gitlabService := ExternalService{
		Kind:        "GITLAB",
		DisplayName: "GitLab",
		Config: `{
			// Some comment
			"url": "https://gitlab.com",
			"token": "secret",
			"projectQuery": ["none"]
		}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	bitbucketServerService := ExternalService{
		Kind:        "BITBUCKETSERVER",
		DisplayName: "Bitbucket Server",
		Config: `{
			// Some comment
			"url": "https://bitbucketserver.mycorp.com",
			"username: "admin",
			"token": "secret",
			"repositoryQuery": ["none"]
		}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	awsCodeCommitService := ExternalService{
		ID:          9,
		Kind:        "AWSCODECOMMIT",
		DisplayName: "AWS CodeCommit",
		Config: `{
			"region": "us-west-1",
			"accessKeyID": "secret-accessKeyID",
			"secretAccessKey": "secret-secretAccessKey",
			"gitCredentials": {"username": "user", "password": "pw"},
		}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	gitoliteService := ExternalService{
		Kind:        "GITOLITE",
		DisplayName: "Gitolite",
		Config: `{
			// Some comment
			"host": "git@gitolite.mycorp.com",
			"prefix": "gitolite.mycorp.com/"
		}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	otherService := ExternalService{
		Kind:        "OTHER",
		DisplayName: "Other code hosts",
		Config: formatJSON(t, `{
			"url": "https://git-host.mycorp.com",
			"repos": []
		}`),
		CreatedAt: now,
		UpdatedAt: now,
	}

	repos := Repos{
		{
			Metadata: &github.Repository{
				ID:            "foo",
				NameWithOwner: "org/foo",
			},
		},
		{
			Metadata: &gitlab.Project{
				ProjectCommon: gitlab.ProjectCommon{
					ID:                1,
					PathWithNamespace: "org/foo",
				},
			},
		},
		{
			Metadata: &github.Repository{
				NameWithOwner: "org/baz",
			},
		},
		{
			Metadata: &gitlab.Project{
				ProjectCommon: gitlab.ProjectCommon{
					PathWithNamespace: "org/baz",
				},
			},
		},
		{
			Metadata: &bitbucketserver.Repo{
				ID:   1,
				Slug: "foo",
				Project: &bitbucketserver.Project{
					Key: "org",
				},
			},
		},
		{
			Metadata: &bitbucketserver.Repo{
				Slug: "baz",
				Project: &bitbucketserver.Project{
					Key: "org",
				},
			},
		},
		{
			Metadata: &awscodecommit.Repository{
				ID:   "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
				Name: "foo",
			},
		},
		{
			Metadata: &awscodecommit.Repository{
				ID:   "b4455554-4444-5555-b7d2-888c9EXAMPLE",
				Name: "baz",
			},
		},
		{
			Name: "git-host.mycorp.com/org/foo",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "1",
				ServiceType: "other",
				ServiceID:   "https://git-host.mycorp.com/",
			},
		},
		{
			Name: "git-host.mycorp.com/org/baz",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: "other",
				ServiceID:   "https://git-host.mycorp.com/",
			},
		},
		{
			Metadata: &gitolite.Repo{Name: "foo"},
		},
	}

	var testCases []testCase
	{
		svcs := ExternalServices{
			githubService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://github.com",
					"token": "secret",
					"repositoryQuery": ["none"],
					"exclude": [
						{"id": "foo"},
						{"name": "org/BAZ"}
					]
				}`)
			}),
			gitlabService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://gitlab.com",
					"token": "secret",
					"projectQuery": ["none"],
					"exclude": [
						{"id": 1},
						{"name": "org/baz"}
					]
				}`)
			}),
			bitbucketServerService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://bitbucketserver.mycorp.com",
					"username": "admin",
					"token": "secret",
					"repositoryQuery": ["none"],
					"exclude": [
						{"id": 1},
						{"name": "org/baz"}
					]
				}`)
			}),
			awsCodeCommitService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"region": "us-west-1",
					"accessKeyID": "secret-accessKeyID",
					"secretAccessKey": "secret-secretAccessKey",
					"gitCredentials": {"username": "user", "password": "pw"},
					"exclude": [
						{"id": "f001337a-3450-46fd-b7d2-650c0EXAMPLE"},
						{"name": "baz"}
					]
				}`)
			}),
			gitoliteService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"host": "git@gitolite.mycorp.com",
					"prefix": "gitolite.mycorp.com/",
					"exclude": [
						{"name": "foo"}
					]
				}`)
			}),
			&otherService,
		}

		testCases = append(testCases, testCase{
			name:   "already excluded repos are ignored",
			svcs:   svcs,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(svcs...),
		})
	}
	{
		svcs := ExternalServices{
			githubService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://github.com",
					"token": "secret",
					"repositoryQuery": ["none"],
					"exclude": [
						{"name": "org/boo"},
					]
				}`)
			}),
			gitlabService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://gitlab.com",
					"token": "secret",
					"projectQuery": ["none"],
					"exclude": [
						{"name": "org/boo"},
					]
				}`)
			}),
			bitbucketServerService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://gitlab.com",
					"username": "admin",
					"token": "secret",
					"repositoryQuery": ["none"],
					"exclude": [
						{"name": "org/boo"},
					]
				}`)
			}),
			awsCodeCommitService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"region": "us-west-1",
					"accessKeyID": "secret-accessKeyID",
					"secretAccessKey": "secret-secretAccessKey",
					"gitCredentials": {"username": "user", "password": "pw"},
					"exclude": [
						{"name": "boo"}
					]
				}`)
			}),
			gitoliteService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"host": "git@gitolite.mycorp.com",
					"prefix": "gitolite.mycorp.com/",
					"exclude": [
						{"name": "boo"}
					]
				}`)
			}),
			otherService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					"url": "https://git-host.mycorp.com",
					"repos": [
						"org/foo",
						"org/boo",
						"org/baz"
					]
				}`)
			}),
		}

		testCases = append(testCases, testCase{
			name:  "repos are excluded",
			svcs:  svcs,
			repos: repos,
			assert: Assert.ExternalServicesEqual(
				githubService.With(func(e *ExternalService) {
					e.Config = formatJSON(t, `
					{
						// Some comment
						"url": "https://github.com",
						"token": "secret",
						"repositoryQuery": ["none"],
						"exclude": [
							{"name": "org/boo"},
							{"id": "foo", "name": "org/foo"},
							{"name": "org/baz"}
						]
					}`)
				}),
				gitlabService.With(func(e *ExternalService) {
					e.Config = formatJSON(t, `
					{
						// Some comment
						"url": "https://gitlab.com",
						"token": "secret",
						"projectQuery": ["none"],
						"exclude": [
							{"name": "org/boo"},
							{"id": 1, "name": "org/foo"},
							{"name": "org/baz"}
						]
					}`)
				}),
				bitbucketServerService.With(func(e *ExternalService) {
					e.Config = formatJSON(t, `
					{
						// Some comment
						"url": "https://gitlab.com",
						"username": "admin",
						"token": "secret",
						"repositoryQuery": ["none"],
						"exclude": [
							{"name": "org/boo"},
							{"id": 1, "name": "org/foo"},
							{"name": "org/baz"}
						]
					}`)
				}),
				awsCodeCommitService.With(func(e *ExternalService) {
					e.Config = formatJSON(t, `
					{
						// Some comment
						"region": "us-west-1",
						"accessKeyID": "secret-accessKeyID",
						"secretAccessKey": "secret-secretAccessKey",
						"gitCredentials": {"username": "user", "password": "pw"},
						"exclude": [
							{"name": "boo"},
							{"id": "f001337a-3450-46fd-b7d2-650c0EXAMPLE", "name": "foo"},
							{"id": "b4455554-4444-5555-b7d2-888c9EXAMPLE", "name": "baz"}
						]
					}`)
				}),
				gitoliteService.With(func(e *ExternalService) {
					e.Config = formatJSON(t, `
					{
						// Some comment
						"host": "git@gitolite.mycorp.com",
						"prefix": "gitolite.mycorp.com/",
						"exclude": [
							{"name": "boo"},
							{"name": "foo"}
						]
					}`)
				}),
				otherService.With(func(e *ExternalService) {
					e.Config = formatJSON(t, `
					{
						"url": "https://git-host.mycorp.com",
						"repos": [
							"org/boo"
						]
					}`)
				}),
			),
		})
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svcs, repos := tc.svcs.Clone(), tc.repos.Clone()

			var err error
			for _, svc := range svcs {
				if err = svc.Exclude(repos...); err != nil {
					t.Fatal(err)
				}
			}

			if tc.assert != nil {
				tc.assert(t, svcs)
			}
		})
	}
}

// Our uses of pick happen from iterating through a map. So we can't guarantee
// that we test both pick(a, b) and pick(b, a) without writing this specific
// test.
func TestPick(t *testing.T) {
	eid := func(id string) api.ExternalRepoSpec {
		return api.ExternalRepoSpec{
			ID:          id,
			ServiceType: "fake",
			ServiceID:   "https://fake.com",
		}
	}
	a := &Repo{Name: "bar", ExternalRepo: eid("1")}
	b := &Repo{Name: "bar", ExternalRepo: eid("2")}

	for _, args := range [][2]*Repo{{a, b}, {b, a}} {
		keep, discard := pick(args[0], args[1])
		if keep != a || discard != b {
			t.Errorf("unexpected pick(%v, %v)", args[0], args[1])
		}
	}
}

func formatJSON(t testing.TB, s string) string {
	formatted, err := jsonc.Format(s, nil)
	if err != nil {
		t.Fatal(err)
	}

	return formatted
}

func TestRateLimiterRegistry(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	baseURL := "http://gitlab.com/"
	makeConfig := func(u string, perHour int) string {
		return fmt.Sprintf(
			`
{
  "url": "%s",
  "rateLimit": {
    "enabled": true,
    "requestsPerHour": %d,
  }
}
`, u, perHour)
	}

	// Two services for the same code host
	mockLister := &MockExternalServicesLister{
		listExternalServices: func(ctx context.Context, args StoreListExternalServicesArgs) ([]*ExternalService, error) {
			return []*ExternalService{
				{
					ID:          1,
					Kind:        "GitLab",
					DisplayName: "GitLab",
					Config:      makeConfig(baseURL, 3600),
					CreatedAt:   now,
					UpdatedAt:   now,
					DeletedAt:   time.Time{},
				},
				{
					ID:          2,
					Kind:        "GitLab",
					DisplayName: "GitLab",
					Config:      makeConfig(baseURL, 7200),
					CreatedAt:   now,
					UpdatedAt:   now,
					DeletedAt:   time.Time{},
				},
			}, nil
		},
	}

	r := &RateLimiterRegistry{
		serviceLister: mockLister,
		rateLimiters:  make(map[string]*rate.Limiter),
	}

	l := r.GetRateLimiter(baseURL)
	if l == nil {
		t.Fatalf("Expected a limiter")
	}
	expectedLimit := rate.Inf
	if l.Limit() != expectedLimit {
		t.Fatalf("Expected limit %f, got %f", expectedLimit, l.Limit())
	}

	err := r.SyncRateLimiters(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// We should have default limit
	l = r.GetRateLimiter(baseURL)
	if l == nil {
		t.Fatalf("Expected a limiter")
	}
	expectedLimit = rate.Limit(1)
	if l.Limit() != expectedLimit {
		t.Fatalf("Expected limit %f, got %f", expectedLimit, l.Limit())
	}

	err = r.SyncRateLimiters(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// We should have new limit
	l = r.GetRateLimiter(baseURL)
	if l == nil {
		t.Fatalf("Expected a limiter")
	}
	expectedLimit = rate.Limit(1)
	if l.Limit() != expectedLimit {
		t.Fatalf("Expected limit %f, got %f", expectedLimit, l.Limit())
	}
}

type MockExternalServicesLister struct {
	listExternalServices func(context.Context, StoreListExternalServicesArgs) ([]*ExternalService, error)
}

func (m MockExternalServicesLister) ListExternalServices(ctx context.Context, args StoreListExternalServicesArgs) ([]*ExternalService, error) {
	return m.listExternalServices(ctx, args)
}
