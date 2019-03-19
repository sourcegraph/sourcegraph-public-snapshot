package repos

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExternalService_IncludeExcludeGithubRepos(t *testing.T) {
	now := time.Now()
	github := ExternalService{
		Kind:        "GITHUB",
		DisplayName: "Github",
		CreatedAt:   now,
		UpdatedAt:   now,
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
			e.Config = marshalJSON(t, &schema.GitHubConnection{
				Url: "https://github.com",
				Exclude: []*schema.Exclude{
					{Id: repos[0].ExternalRepo.ID},
					{Name: strings.ToUpper(repos[2].Name)}, // Test case insensitvity
				},
			})
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
			e.Config = marshalJSON(t, &schema.GitHubConnection{
				Url: "https://github.com",
				Exclude: []*schema.Exclude{
					{Name: "boo"},
				},
			})
		})

		testCases = append(testCases, testCase{
			method: "exclude",
			name:   "github repos are excluded",
			svc:    svc,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(svc.With(func(e *ExternalService) {
				e.Config = marshalJSON(t, &schema.GitHubConnection{
					Url: "https://github.com",
					Exclude: []*schema.Exclude{
						{Name: "boo"},
						{Name: repos[0].Name, Id: repos[0].ExternalRepo.ID},
						{Name: repos[2].Name},
					},
				})
			})),
			err: `<nil>`,
		})
	}
	{
		svc := github.With(func(e *ExternalService) {
			e.Config = marshalJSON(t, &schema.GitHubConnection{
				Url: "https://github.com",
				Repos: []string{
					strings.ToUpper(repos[0].Name), // Test case insensitvity
					repos[2].Name,
				},
			})
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
			e.Config = marshalJSON(t, &schema.GitHubConnection{
				Url:   "https://github.com",
				Repos: []string{"boo"},
			})
		})

		testCases = append(testCases, testCase{
			method: "include",
			name:   "github repos are included",
			svc:    svc,
			repos:  repos,
			assert: Assert.ExternalServicesEqual(svc.With(func(e *ExternalService) {
				e.Config = marshalJSON(t, &schema.GitHubConnection{
					Url: "https://github.com",
					Repos: []string{
						"boo",
						repos[0].Name,
						repos[2].Name,
					},
				})
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

func marshalJSON(t testing.TB, v interface{}) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	formatted, err := jsonc.Format(string(bs), true, 2)
	if err != nil {
		t.Fatal(err)
	}

	return formatted
}
