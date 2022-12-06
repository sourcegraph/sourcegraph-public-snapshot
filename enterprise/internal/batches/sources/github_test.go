package sources

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGithubSource_CreateChangeset(t *testing.T) {
	// Repository used: sourcegraph/automation-testing
	//
	// The requests here cannot be easily rerun with `-update` since you can only
	// open a pull request once. To update, push a new branch to
	// automation-testing, and put the branch names into the `success` case
	// below.
	//
	// You can update just this test with `-update GithubSource_CreateChangeset`.
	repo := &types.Repo{
		Metadata: &github.Repository{
			ID:            "MDEwOlJlcG9zaXRvcnkyMjExNDc1MTM=",
			NameWithOwner: "sourcegraph/automation-testing",
		},
	}

	testCases := []struct {
		name   string
		cs     *Changeset
		err    string
		exists bool
	}{
		{
			name: "success",
			cs: &Changeset{
				Title:      "This is a test PR",
				Body:       "This is the description of the test PR",
				HeadRef:    "refs/heads/test-pr-10",
				BaseRef:    "refs/heads/master",
				RemoteRepo: repo,
				TargetRepo: repo,
				Changeset:  &btypes.Changeset{},
			},
		},
		{
			name: "already exists",
			cs: &Changeset{
				Title:      "This is a test PR",
				Body:       "This is the description of the test PR",
				HeadRef:    "refs/heads/always-open-pr",
				BaseRef:    "refs/heads/master",
				RemoteRepo: repo,
				TargetRepo: repo,
				Changeset:  &btypes.Changeset{},
			},
			// If PR already exists we'll just return it, no error
			err:    "",
			exists: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		tc.name = "GithubSource_CreateChangeset_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			// The GithubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_TOKEN"),
				})),
			}

			ctx := context.Background()
			githubSrc, err := NewGithubSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			exists, err := githubSrc.CreateChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			if have, want := exists, tc.exists; have != want {
				t.Errorf("exists:\nhave: %t\nwant: %t", have, want)
			}

			pr, ok := tc.cs.Changeset.Metadata.(*github.PullRequest)
			if !ok {
				t.Fatal("Metadata does not contain PR")
			}

			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), pr)
		})
	}
}

func TestGithubSource_CreateChangeset_CreationLimit(t *testing.T) {
	cli := new(mockDoer)
	// Version lookup
	versionMatchedBy := func(req *http.Request) bool {
		return req.Method == http.MethodGet && req.URL.Path == "/"
	}
	cli.On("Do", mock.MatchedBy(versionMatchedBy)).
		Once().
		Return(
			&http.Response{
				StatusCode: http.StatusOK,
				Header: map[string][]string{
					"X-GitHub-Enterprise-Version": {"99"},
				},
				Body: io.NopCloser(bytes.NewReader([]byte{})),
			},
			nil,
		)
	// Create Changeset mutation
	createChangesetMatchedBy := func(req *http.Request) bool {
		return req.Method == http.MethodPost && req.URL.Path == "/graphql"
	}
	cli.On("Do", mock.MatchedBy(createChangesetMatchedBy)).
		Once().
		Return(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"errors": [{"message": "error in GraphQL response: was submitted too quickly"}]}`))),
			},
			nil,
		)

	apiURL, err := url.Parse("https://fake.api.github.com")
	require.NoError(t, err)
	client := github.NewV4Client("extsvc:github:0", apiURL, nil, cli)
	source := &GithubSource{
		client: client,
	}

	repo := &types.Repo{
		Metadata: &github.Repository{
			ID:            "bLAhBLAh",
			NameWithOwner: "some-org/some-repo",
		},
	}
	cs := &Changeset{
		Title:      "This is a test PR",
		Body:       "This is the description of the test PR",
		HeadRef:    "refs/heads/always-open-pr",
		BaseRef:    "refs/heads/master",
		RemoteRepo: repo,
		TargetRepo: repo,
		Changeset:  &btypes.Changeset{},
	}

	exists, err := source.CreateChangeset(context.Background(), cs)
	assert.False(t, exists)
	assert.Error(t, err)
	assert.Equal(
		t,
		"reached GitHub's internal creation limit: see https://docs.sourcegraph.com/admin/config/batch_changes#avoiding-hitting-rate-limits: error in GraphQL response: error in GraphQL response: was submitted too quickly",
		err.Error(),
	)
}

type mockDoer struct {
	mock.Mock
}

func (d *mockDoer) Do(req *http.Request) (*http.Response, error) {
	args := d.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestGithubSource_CloseChangeset(t *testing.T) {
	// Repository used: sourcegraph/automation-testing
	//
	// This test can be run with `-update` provided:
	//
	// 1. https://github.com/sourcegraph/automation-testing/pull/468 is open.
	//
	// You can update just this test with `-update GithubSource_CloseChangeset`.
	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs: &Changeset{
				Changeset: &btypes.Changeset{
					Metadata: &github.PullRequest{
						ID: "PR_kwDODS5xec4waMkR",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GithubSource_CloseChangeset_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			// The GithubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_TOKEN"),
				})),
			}

			ctx := context.Background()
			githubSrc, err := NewGithubSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			err = githubSrc.CloseChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*github.PullRequest)
			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), pr)
		})
	}
}

func TestGithubSource_ReopenChangeset(t *testing.T) {
	// Repository used: sourcegraph/automation-testing
	//
	// This test can be run with `-update` provided:
	//
	// 1. https://github.com/sourcegraph/automation-testing/pull/353 is closed,
	//    but _not_ merged.
	//
	// You can update just this test with `-update GithubSource_ReopenChangeset`.
	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs: &Changeset{
				Changeset: &btypes.Changeset{
					Metadata: &github.PullRequest{
						// https://github.com/sourcegraph/automation-testing/pull/353
						ID: "MDExOlB1bGxSZXF1ZXN0NDg4MDI2OTk5",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GithubSource_ReopenChangeset_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			// The GithubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_TOKEN"),
				})),
			}

			ctx := context.Background()
			githubSrc, err := NewGithubSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			err = githubSrc.ReopenChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*github.PullRequest)
			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), pr)
		})
	}
}

func TestGithubSource_CreateComment(t *testing.T) {
	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs: &Changeset{
				Changeset: &btypes.Changeset{
					Metadata: &github.PullRequest{
						ID: "MDExOlB1bGxSZXF1ZXN0MzQ5NTIzMzE0",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GithubSource_CreateComment_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_TOKEN"),
				})),
			}

			ctx := context.Background()
			githubSrc, err := NewGithubSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			err = githubSrc.CreateComment(ctx, tc.cs, "test-comment")
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}
		})
	}
}

func TestGithubSource_UpdateChangeset(t *testing.T) {
	// Repository used: sourcegraph/automation-testing
	//
	// This test can be run with `-update` provided:
	//
	// 1. https://github.com/sourcegraph/automation-testing/pull/358 is open.
	//
	// You can update just this test with `-update GithubSource_UpdateChangeset`.
	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs: &Changeset{
				Title:   "This is a new title",
				Body:    "This is a new body",
				BaseRef: "refs/heads/master",
				Changeset: &btypes.Changeset{
					Metadata: &github.PullRequest{
						ID: "MDExOlB1bGxSZXF1ZXN0NTA0NDU4Njg1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GithubSource_UpdateChangeset_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			// The GithubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_TOKEN"),
				})),
			}

			ctx := context.Background()
			githubSrc, err := NewGithubSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			err = githubSrc.UpdateChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*github.PullRequest)
			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), pr)
		})
	}
}

func TestGithubSource_LoadChangeset(t *testing.T) {
	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "found",
			cs: &Changeset{
				RemoteRepo: &types.Repo{Metadata: &github.Repository{NameWithOwner: "sourcegraph/sourcegraph"}},
				TargetRepo: &types.Repo{Metadata: &github.Repository{NameWithOwner: "sourcegraph/sourcegraph"}},
				Changeset:  &btypes.Changeset{ExternalID: "5550"},
			},
		},
		{
			name: "not-found",
			cs: &Changeset{
				RemoteRepo: &types.Repo{Metadata: &github.Repository{NameWithOwner: "sourcegraph/sourcegraph"}},
				TargetRepo: &types.Repo{Metadata: &github.Repository{NameWithOwner: "sourcegraph/sourcegraph"}},
				Changeset:  &btypes.Changeset{ExternalID: "100000"},
			},
			err: "Changeset with external ID 100000 not found",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GithubSource_LoadChangeset_" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			// The GithubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_TOKEN"),
				})),
			}

			ctx := context.Background()
			githubSrc, err := NewGithubSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			err = githubSrc.LoadChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			meta := tc.cs.Changeset.Metadata.(*github.PullRequest)
			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), meta)
		})
	}
}

func TestGithubSource_WithAuthenticator(t *testing.T) {
	svc := &types.ExternalService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
		})),
	}

	ctx := context.Background()
	githubSrc, err := NewGithubSource(ctx, svc, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("supported", func(t *testing.T) {
		src, err := githubSrc.WithAuthenticator(&auth.OAuthBearerToken{})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if gs, ok := src.(*GithubSource); !ok {
			t.Error("cannot coerce Source into GithubSource")
		} else if gs == nil {
			t.Error("unexpected nil Source")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for name, tc := range map[string]auth.Authenticator{
			"nil":         nil,
			"BasicAuth":   &auth.BasicAuth{},
			"OAuthClient": &auth.OAuthClient{},
		} {
			t.Run(name, func(t *testing.T) {
				src, err := githubSrc.WithAuthenticator(tc)
				if err == nil {
					t.Error("unexpected nil error")
				} else if !errors.HasType(err, UnsupportedAuthenticatorError{}) {
					t.Errorf("unexpected error of type %T: %v", err, err)
				}
				if src != nil {
					t.Errorf("expected non-nil Source: %v", src)
				}
			})
		}
	})
}

func TestGithubSource_GetUserFork(t *testing.T) {
	ctx := context.Background()

	t.Run("failures", func(t *testing.T) {
		for name, tc := range map[string]struct {
			targetRepo *types.Repo
			client     githubClientFork
		}{
			"nil metadata": {
				targetRepo: &types.Repo{
					Metadata: nil,
				},
				client: nil,
			},
			"invalid metadata": {
				targetRepo: &types.Repo{
					Metadata: []string{},
				},
				client: nil,
			},
			"invalid NameWithOwner": {
				targetRepo: &types.Repo{
					Metadata: &github.Repository{
						NameWithOwner: "foo",
					},
				},
				client: nil,
			},
			"client error": {
				targetRepo: &types.Repo{
					Metadata: &github.Repository{
						NameWithOwner: "foo/bar",
					},
				},
				client: &mockGithubClientFork{err: errors.New("hello!")},
			},
		} {
			t.Run(name, func(t *testing.T) {
				fork, err := githubGetUserFork(ctx, tc.targetRepo, tc.client, nil)
				assert.Nil(t, fork)
				assert.NotNil(t, err)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		org := "org"
		user := "user"
		urn := extsvc.URN(extsvc.KindGitHub, 1)

		for name, tc := range map[string]struct {
			targetRepo    *types.Repo
			forkRepo      *github.Repository
			namespace     *string
			wantNamespace string
			client        githubClientFork
		}{
			"no namespace": {
				targetRepo: &types.Repo{
					Metadata: &github.Repository{
						NameWithOwner: "foo/bar",
					},
					Sources: map[string]*types.SourceInfo{
						urn: {
							ID:       urn,
							CloneURL: "https://github.com/foo/bar",
						},
					},
				},
				forkRepo:      &github.Repository{NameWithOwner: user + "/bar"},
				namespace:     nil,
				wantNamespace: user,
				client:        &mockGithubClientFork{fork: &github.Repository{NameWithOwner: user + "/bar"}},
			},
			"with namespace": {
				targetRepo: &types.Repo{
					Metadata: &github.Repository{
						NameWithOwner: "foo/bar",
					},
					Sources: map[string]*types.SourceInfo{
						urn: {
							ID:       urn,
							CloneURL: "https://github.com/foo/bar",
						},
					},
				},
				forkRepo:      &github.Repository{NameWithOwner: org + "/bar"},
				namespace:     &org,
				wantNamespace: org,
				client: &mockGithubClientFork{
					fork:    &github.Repository{NameWithOwner: org + "/bar"},
					wantOrg: &org,
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				fork, err := githubGetUserFork(ctx, tc.targetRepo, tc.client, tc.namespace)
				assert.Nil(t, err)
				assert.NotNil(t, fork)
				assert.NotEqual(t, fork, tc.targetRepo)
				assert.Equal(t, tc.forkRepo, fork.Metadata)
				assert.Equal(t, fork.Sources[urn].CloneURL, "https://github.com/"+tc.wantNamespace+"/bar")
			})
		}
	})
}

type mockGithubClientFork struct {
	wantOrg *string
	fork    *github.Repository
	err     error
}

var _ githubClientFork = &mockGithubClientFork{}

func (mock *mockGithubClientFork) Fork(ctx context.Context, owner, repo string, org *string, forkName string) (*github.Repository, error) {
	if (mock.wantOrg == nil && org != nil) || (mock.wantOrg != nil && org == nil) || (mock.wantOrg != nil && org != nil && *mock.wantOrg != *org) {
		return nil, errors.Newf("unexpected organisation: have=%v want=%v", org, mock.wantOrg)
	}

	return mock.fork, mock.err
}
