package repos_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/dnaeon/go-vcr/recorder"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGithubSource_ListRepos(t *testing.T) {
	config := func(cfg *schema.GitHubConnection) string {
		t.Helper()
		bs, err := json.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		return string(bs)
	}

	for _, tc := range []struct {
		name  string
		ctx   context.Context
		svc   api.ExternalService
		repos []*repos.Repo
		err   string
	}{
		{
			name: "blacklisted repos are never returned",
			svc: api.ExternalService{
				Kind: "GITHUB",
				Config: config(&schema.GitHubConnection{
					Url: "https://github.com",
					RepositoryQuery: []string{
						"user:tsenart in:name patrol",
					},
					Repos: []string{
						"sourcegraph/sourcegraph",
						"tsenart/vegeta",
					},
					Blacklist: []*schema.Blacklist{
						{Name: "tsenart/vegeta"},
						// {Id: "tsenart/vegeta"}, patrol's id
					},
				}),
			},
			repos: []*repos.Repo{
				{
					Name: "github.com/sourcegraph/sourcegraph",
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r, err := recorder.New("fixtures/github-source")
			if err != nil {
				t.Fatal(err)
			}
			defer r.Stop()

			s, err := repos.NewGithubSource(&tc.svc)
			if err != nil {
				t.Error(err)
				return // Let defers run
			}

			ctx := tc.ctx
			if ctx == nil {
				ctx = context.Background()
			}

			repos, err := s.ListRepos(ctx)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if have, want := repos, tc.repos; !reflect.DeepEqual(have, want) {
				t.Errorf("repos: %s", cmp.Diff(have, want))
			}
		})
	}
}
