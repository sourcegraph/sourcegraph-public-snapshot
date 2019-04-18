package repos

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
)

func TestExternalService_IncludeExclude(t *testing.T) {
	now := time.Now()

	type testCase struct {
		method string
		name   string
		svcs   ExternalServices
		repos  Repos
		assert ExternalServicesAssertion
	}

	github := ExternalService{
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

	gitlab := ExternalService{
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

	bitbucketServer := ExternalService{
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
			Name: "github.com/org/foo",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: "github",
				ServiceID:   "https://github.com/",
				ID:          "foo",
			},
		},
		{
			Name: "gitlab.com/org/foo",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: "gitlab",
				ServiceID:   "https://gitlab.com/",
				ID:          "1",
			},
		},
		{
			Name: "github.com/org/baz",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: "github",
				ServiceID:   "https://github.mycorp.com/",
			},
		},
		{
			Name: "gitlab.com/org/baz",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: "gitlab",
				ServiceID:   "https://gitlab.mycorp.com/",
			},
		},
		{
			Name: "bitbucketserver.mycorp.com/org/foo",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "1",
				ServiceType: "bitbucketServer",
				ServiceID:   "https://bitbucketserver.mycorp.com/",
			},
		},
		{
			Name: "bitbucketserver.mycorp.com/org/baz",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: "bitbucketServer",
				ServiceID:   "https://bitbucketserver.mycorp.com/",
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
	}

	var testCases []testCase
	{
		svcs := ExternalServices{
			github.With(func(e *ExternalService) {
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
			gitlab.With(func(e *ExternalService) {
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
			bitbucketServer.With(func(e *ExternalService) {
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
			&otherService,
		}

		testCases = append(testCases, testCase{
			method: "exclude",
			name:   "already excluded repos are ignored",
			svcs:   svcs,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(svcs...),
		})
	}
	{
		svcs := ExternalServices{
			github.With(func(e *ExternalService) {
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
			gitlab.With(func(e *ExternalService) {
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
			bitbucketServer.With(func(e *ExternalService) {
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
			method: "exclude",
			name:   "repos are excluded",
			svcs:   svcs,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(
				github.With(func(e *ExternalService) {
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
				gitlab.With(func(e *ExternalService) {
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
				bitbucketServer.With(func(e *ExternalService) {
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
	{
		svcs := ExternalServices{
			github.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
					{
						// Some comment
						"url": "https://github.com",
						"token": "secret",
						"repositoryQuery": ["none"],
						"repos": [
							"org/FOO",
							"org/baz"
						]
					}`)
			}),
			gitlab.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://gitlab.com",
					"token": "secret",
					"projectQuery": ["none"],
					"projects": [
						{"id": 1},
						{"name": "org/baz"}
					]
				}`)
			}),
			bitbucketServer.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://bitbucketserver.mycorp.com",
					"username": "admin",
					"token": "secret",
					"repositoryQuery": ["none"],
					"repos": [
						"org/FOO",
						"org/baz"
					]
				}`)
			}),
			otherService.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					"repos": [
						"https://git-host.mycorp.com/org/baz",
						"https://git-host.mycorp.com/org/foo"
					]
				}`)
			}),
		}

		testCases = append(testCases, testCase{
			method: "include",
			name:   "already included repos are ignored",
			svcs:   svcs,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(svcs...),
		})
	}

	{
		svcs := ExternalServices{
			github.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://github.com",
					"token": "secret",
					"repositoryQuery": ["none"],
					"repos": [
						"org/boo"
					]
				}`)
			}),
			gitlab.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://gitlab.com",
					"token": "secret",
					"projectQuery": ["none"],
					"projects": [
						{"name": "org/boo"},
					]
				}`)
			}),
			bitbucketServer.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://bitbucketserver.mycorp.com",
					"username": "admin",
					"token": "secret",
					"repositoryQuery": ["none"],
					"repos": [
						"org/boo"
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
		}

		testCases = append(testCases, testCase{
			method: "include",
			name:   "repos are included",
			svcs:   svcs,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(
				github.With(func(e *ExternalService) {
					e.Config = formatJSON(t, `
					{
						// Some comment
						"url": "https://github.com",
						"token": "secret",
						"repositoryQuery": ["none"],
						"repos": [
							"org/boo",
							"org/foo",
							"org/baz"
						]
					}`)
				}),
				gitlab.With(func(e *ExternalService) {
					e.Config = formatJSON(t, `
					{
						// Some comment
						"url": "https://gitlab.com",
						"token": "secret",
						"projectQuery": ["none"],
						"projects": [
							{"name": "org/boo"},
							{"id": 1, "name": "org/foo"},
							{"name": "org/baz"}
						]
					}`)
				}),
				bitbucketServer.With(func(e *ExternalService) {
					e.Config = formatJSON(t, `
					{
						// Some comment
						"url": "https://bitbucketserver.mycorp.com",
						"username": "admin",
						"token": "secret",
						"repositoryQuery": ["none"],
						"repos": [
							"org/boo",
							"org/foo",
							"org/baz"
						]
					}`)
				}),
				otherService.With(func(e *ExternalService) {
					e.Config = formatJSON(t, `
					{
						"url": "https://git-host.mycorp.com",
						"repos": [
							"org/baz",
							"org/boo",
							"org/foo"
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
				switch tc.method {
				case "include":
					err = svc.Include(repos...)
				case "exclude":
					err = svc.Exclude(repos...)
				}

				if err != nil {
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
