package repos

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestOtherRepoName(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		out  api.RepoName
	}{
		{"user and password elided", "https://user:pass@foo.bar/baz", "foo.bar/baz"},
		{"scheme elided", "https://user@foo.bar/baz", "foo.bar/baz"},
		{"raw query elided", "https://foo.bar/baz?secret_token=12345", "foo.bar/baz"},
		{"fragment elided", "https://foo.bar/baz#fragment", "foo.bar/baz"},
		{": replaced with -", "git://foo.bar/baz:bam", "foo.bar/baz-bam"},
		{"@ replaced with -", "ssh://foo.bar/baz@bam", "foo.bar/baz-bam"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cloneURL, err := url.Parse(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := otherRepoName(cloneURL), tc.out; have != want {
				t.Errorf("otherRepoName(%q):\nhave: %q\nwant: %q", tc.in, have, want)
			}
		})
	}
}

func TestOtherReposSyncer_syncAll(t *testing.T) {
	repoInfo := func(r *api.Repo) *protocol.RepoInfo {
		return &protocol.RepoInfo{
			Name:         r.Name,
			VCS:          protocol.VCSInfo{URL: "https://" + string(r.Name)},
			ExternalRepo: r.ExternalRepo,
		}
	}

	svcs := map[string]*api.ExternalService{
		"github.com": {
			ID:     0,
			Kind:   "OTHER",
			Config: `{"repos": ["https://github.com/foo/bar"]}`,
		},
		"bad": {
			ID:     1,
			Kind:   "OTHER",
			Config: `{"repos": [""]}`,
		},
		"invalid-json": {
			ID:     2,
			Kind:   "OTHER",
			Config: `{`,
		},
	}

	repos := map[string]*api.Repo{
		"bad": { // bad repo
			ExternalRepo: &api.ExternalRepoSpec{ServiceType: "other"},
		},
		"github.com/foo/bar": {
			ID:      1,
			Name:    "github.com/foo/bar",
			Enabled: true,
			ExternalRepo: &api.ExternalRepoSpec{
				ID:          string("github.com/foo/bar"),
				ServiceType: "other",
				ServiceID:   "https://github.com",
			},
		},
	}

	for _, tc := range []struct {
		name    string
		svcs    []*api.ExternalService
		before  []*api.Repo
		after   []*api.Repo
		results OtherSyncResults
		err     error
	}{
		{
			name:   "new repos from external service",
			svcs:   []*api.ExternalService{svcs["github.com"]},
			before: []*api.Repo{},
			after:  []*api.Repo{repos["github.com/foo/bar"]},
			results: OtherSyncResults{
				{
					Service: svcs["github.com"],
					Synced:  []*protocol.RepoInfo{repoInfo(repos["github.com/foo/bar"])},
				},
			},
		},
		{
			name:   "existing repos from external service",
			svcs:   []*api.ExternalService{svcs["github.com"]},
			before: []*api.Repo{repos["github.com/foo/bar"]},
			after:  []*api.Repo{repos["github.com/foo/bar"]},
			results: OtherSyncResults{
				{
					Service: svcs["github.com"],
					Synced:  []*protocol.RepoInfo{repoInfo(repos["github.com/foo/bar"])},
				},
			},
		},
		{
			name:   "external service listing error",
			svcs:   []*api.ExternalService{},
			before: []*api.Repo{},
			after:  []*api.Repo{},
			err:    errors.New(`no external services of kind "OTHER"`),
		},
		{
			name:   "invalid JSON in exernal service config",
			svcs:   []*api.ExternalService{svcs["invalid-json"]},
			before: []*api.Repo{},
			after:  []*api.Repo{},
			results: OtherSyncResults{
				{
					Service: svcs["invalid-json"],
					Errors: OtherSyncErrors{
						{
							Service: svcs["invalid-json"],
							Err:     "config error: failed to parse JSON: [CloseBraceExpected]",
						},
					},
				},
			},
		},
		{
			name:   "invalid external service configs return an error",
			svcs:   []*api.ExternalService{svcs["bad"]},
			before: []*api.Repo{},
			after:  []*api.Repo{},
			results: OtherSyncResults{
				{
					Service: svcs["bad"],
					Errors: OtherSyncErrors{
						{
							Service: svcs["bad"],
							Repo: &protocol.RepoInfo{
								Name:         repos["bad"].Name,
								VCS:          protocol.VCSInfo{},
								ExternalRepo: repos["bad"].ExternalRepo,
							},
							Err: "invalid empty repo name",
						},
					},
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fa := NewFakeInternalAPI(tc.svcs, tc.before)
			results, err := NewOtherReposSyncer(fa, nil).syncAll(ctx)
			after := fa.ReposList()

			for _, exp := range []struct {
				name       string
				have, want interface{}
			}{
				{name: "repos", have: after, want: tc.after},
				{name: "results", have: results, want: tc.results},
				{name: "error", have: fmt.Sprint(err), want: fmt.Sprint(tc.err)},
			} {
				if !reflect.DeepEqual(exp.have, exp.want) {
					t.Errorf("unexpected %q:\n%s", exp.name, pretty.Compare(exp.have, exp.want))
				}
			}
		})
	}
}
