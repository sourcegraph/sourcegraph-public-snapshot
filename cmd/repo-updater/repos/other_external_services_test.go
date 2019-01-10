package repos

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"sync"
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
		results SyncResults
		err     error
	}{
		{
			name:   "new repos from external service",
			svcs:   []*api.ExternalService{svcs["github.com"]},
			before: []*api.Repo{},
			after:  []*api.Repo{repos["github.com/foo/bar"]},
			results: SyncResults{
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
			results: SyncResults{
				{
					Service: svcs["github.com"],
					Synced:  []*protocol.RepoInfo{repoInfo(repos["github.com/foo/bar"])},
				},
			},
		},
		{
			name:   "invalid external service configs return an error",
			svcs:   []*api.ExternalService{svcs["bad"]},
			before: []*api.Repo{},
			after:  []*api.Repo{},
			results: SyncResults{
				{
					Service: svcs["bad"],
					Errors: SyncErrors{
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

// FakeInternalAPI implements the InternalAPI interface with the given in memory data.
// It's safe for concurrent use.
type FakeInternalAPI struct {
	mu     sync.RWMutex
	svcs   map[string][]*api.ExternalService
	repos  map[api.RepoName]*api.Repo
	repoID api.RepoID
}

// NewFakeInternalAPI returns a new FakeInternalAPI initialised with the given data.
func NewFakeInternalAPI(svcs []*api.ExternalService, repos []*api.Repo) *FakeInternalAPI {
	fa := FakeInternalAPI{
		svcs:  map[string][]*api.ExternalService{},
		repos: map[api.RepoName]*api.Repo{},
	}

	for _, svc := range svcs {
		fa.svcs[svc.Kind] = append(fa.svcs[svc.Kind], svc)
	}

	for _, repo := range repos {
		fa.repos[repo.Name] = repo
	}

	return &fa
}

// ExternalServicesList lists external services of the given Kind. A non-existent kind
// will result in an error being returned.
func (a *FakeInternalAPI) ExternalServicesList(
	_ context.Context,
	req api.ExternalServicesListRequest,
) ([]*api.ExternalService, error) {

	a.mu.RLock()
	defer a.mu.RUnlock()

	svcs, ok := a.svcs[req.Kind]
	if !ok {
		return nil, fmt.Errorf("no external services of kind %q", req.Kind)
	}

	return svcs, nil
}

// ReposList returns the list of all repos in the API.
func (a *FakeInternalAPI) ReposList() []*api.Repo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	repos := make([]*api.Repo, 0, len(a.repos))
	for _, repo := range a.repos {
		repos = append(repos, repo)
	}

	return repos
}

// RepoCreateIfNotExists creates the given repo if it doesn't exist. Repos with
// with an empty name are invalid and will result in an error to be returned.
func (a *FakeInternalAPI) ReposCreateIfNotExists(
	ctx context.Context,
	req api.RepoCreateOrUpdateRequest,
) (*api.Repo, error) {

	if req.RepoName == "" {
		return nil, errors.New("invalid empty repo name")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	repo, ok := a.repos[req.RepoName]
	if !ok {
		a.repoID++
		repo = &api.Repo{
			ID:           a.repoID,
			Name:         req.RepoName,
			Enabled:      req.Enabled,
			ExternalRepo: req.ExternalRepo,
		}
		a.repos[req.RepoName] = repo
	}

	return repo, nil
}

// ReposUpdateMetdata updates the metadata of repo with the given name.
// Non-existent repos return an error.
func (a *FakeInternalAPI) ReposUpdateMetadata(
	ctx context.Context,
	repoName api.RepoName,
	description string,
	fork, archived bool,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ok := a.repos[repoName]
	if !ok {
		return fmt.Errorf("repo %q not found", repoName)
	}

	// Fake success, no updates needed because the returned types (api.Repo)
	// don't include these fields.

	return nil
}
