package repos

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
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
			"secretAccessKey": "secret-secretAccessKey"
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

func formatJSON(t testing.TB, s string) string {
	formatted, err := jsonc.Format(s, true, 2)
	if err != nil {
		t.Fatal(err)
	}

	return formatted
}
