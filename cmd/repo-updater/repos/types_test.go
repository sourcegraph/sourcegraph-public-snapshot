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
			"token": "secret"
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
			"token": "secret"
		}`,
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
					"exclude": [
						{"id": 1},
						{"name": "org/baz"}
					]
				}`)
			}),
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
					"exclude": [
						{"name": "org/boo"},
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
						"exclude": [
							{"name": "org/boo"},
							{"id": 1, "name": "org/foo"},
							{"name": "org/baz"}
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
					"projects": [
						{"id": 1},
						{"name": "org/baz"}
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
					"projects": [
						{"name": "org/boo"},
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
						"projects": [
							{"name": "org/boo"},
							{"id": 1, "name": "org/foo"},
							{"name": "org/baz"}
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
