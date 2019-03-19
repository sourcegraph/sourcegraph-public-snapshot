package repos

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
)

func TestExternalService_IncludeExcludeGithubRepos(t *testing.T) {
	now := time.Now()
	github := ExternalService{
		Kind:        "GITHUB",
		DisplayName: "Github",
		Config: `{
			// Some comment
			"url": "https://github.com"
		}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	repos := Repos{
		{
			Name: "foo",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: "github",
				ServiceID:   "https://github.com/",
				ID:          "foo",
			},
		},
		{
			Name: "bar",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: "gitlab",
				ServiceID:   "https://gitlab.com/",
				ID:          "bar",
			},
		},
		{
			Name: "baz",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: "github",
				ServiceID:   "https://github.mycorp.com/",
			},
		},
	}

	type testCase struct {
		method string
		name   string
		svc    *ExternalService
		repos  Repos
		assert ExternalServicesAssertion
		err    string
	}

	var testCases []testCase
	{
		svc := github.With(func(e *ExternalService) {
			e.Config = formatJSON(t, `
			{
				// Some comment
				"url": "https://github.com",
				"exclude": [
					{"id": "foo"},
					{"name": "BAZ"}
				]
			}`)
		})

		testCases = append(testCases, testCase{
			method: "exclude",
			name:   "already excluded repos and non-github repos are ignored",
			svc:    svc,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(svc),
			err:    "<nil>",
		})
	}
	{
		svc := ExternalService{Kind: "GITLAB"}
		testCases = append(testCases, testCase{
			method: "exclude",
			name:   "non github external services return an error",
			svc:    &svc,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(&svc),
			err:    `config: unexpected external service kind "GITLAB"`,
		})
	}
	{
		svc := github.With(func(e *ExternalService) {
			e.Config = formatJSON(t, `
			{
				// Some comment
				"url": "https://github.com",
				"exclude": [
					{"name": "boo"}
				]
			}`)
		})

		testCases = append(testCases, testCase{
			method: "exclude",
			name:   "github repos are excluded",
			svc:    svc,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(svc.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://github.com",
					"exclude": [
						{"name": "boo"},
						{"id": "foo", "name": "foo"},
						{"name": "baz"}
					]
				}`)
			})),
			err: `<nil>`,
		})
	}
	{
		svc := github.With(func(e *ExternalService) {
			e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://github.com",
					"repos": [
						"FOO",
						"baz"
					]
				}`)
		})

		testCases = append(testCases, testCase{
			method: "include",
			name:   "already included repos and non-github repos are ignored",
			svc:    svc,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(svc),
			err:    "<nil>",
		})
	}
	{
		svc := ExternalService{Kind: "GITLAB"}
		testCases = append(testCases, testCase{
			method: "include",
			name:   "non github external services return an error",
			svc:    &svc,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(&svc),
			err:    `config: unexpected external service kind "GITLAB"`,
		})
	}
	{
		svc := github.With(func(e *ExternalService) {
			e.Config = formatJSON(t, `
				{
					// Some comment
					"url": "https://github.com",
					"repos": [
						"boo"
					]
				}`)
		})

		testCases = append(testCases, testCase{
			method: "include",
			name:   "github repos are included",
			svc:    svc,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(svc.With(func(e *ExternalService) {
				e.Config = formatJSON(t, `
					{
						// Some comment
						"url": "https://github.com",
						"repos": [
							"boo",
							"foo",
							"baz"
						]
					}`)
			})),
			err: `<nil>`,
		})
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repos := tc.svc.Clone(), tc.repos.Clone()

			var err error
			switch tc.method {
			case "include":
				err = svc.IncludeGithubRepos(repos...)
			case "exclude":
				err = svc.ExcludeGithubRepos(repos...)
			}

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, ExternalServices{svc})
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

func marshalJSON(t testing.TB, v interface{}) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return formatJSON(t, string(bs))
}
