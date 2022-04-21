package sources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestNewBitbucketCloudSource(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		for name, input := range map[string]string{
			"invalid JSON":    "invalid JSON",
			"unparsable JSON": `{"appPassword": ["not a string"]}`,
			"bad URN":         `{"apiURL": "http://[::1]:namedport"}`,
		} {
			t.Run(name, func(t *testing.T) {
				s, err := NewBitbucketCloudSource(&types.ExternalService{
					Config: input,
				}, nil)
				assert.Nil(t, s)
				assert.NotNil(t, err)
			})
		}
	})

	t.Run("valid", func(t *testing.T) {
		s, err := NewBitbucketCloudSource(&types.ExternalService{}, nil)
		assert.NotNil(t, s)
		assert.Nil(t, err)
	})
}

func TestBitbucketCloudSource_GitserverPushConfig(t *testing.T) {
	// This isn't a full blown test of all the possibilities of
	// gitserverPushConfig(), but we do need to validate that the authenticator
	// on the client affects the eventual URL in the correct way, and that
	// requires a bunch of boilerplate to make it look like we have a valid
	// external service and repo.
	//
	// So, cue the boilerplate:
	au := auth.BasicAuthWithSSH{
		BasicAuth: auth.BasicAuth{Username: "user", Password: "pass"},
	}
	s, client := mockBitbucketCloudSource(t)
	client.AuthenticatorFunc.SetDefaultReturn(&au)

	ctx := context.Background()

	svc := types.ExternalService{
		Kind:   extsvc.KindBitbucketCloud,
		Config: `{}`,
	}
	store := database.NewStrictMockExternalServiceStore()
	store.ListFunc.SetDefaultReturn([]*types.ExternalService{&svc}, nil)

	repo := &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: extsvc.TypeBitbucketCloud,
		},
		Metadata: &bitbucketcloud.Repo{
			Links: bitbucketcloud.RepoLinks{
				Clone: bitbucketcloud.CloneLinks{
					bitbucketcloud.Link{
						Name: "https",
						Href: "https://bitbucket.org/clone/link",
					},
				},
			},
		},
		Sources: map[string]*types.SourceInfo{
			"1": {
				ID:       "extsvc:bitbucketcloud:1",
				CloneURL: "https://bitbucket.org/clone/link",
			},
		},
	}

	pushConfig, err := s.GitserverPushConfig(ctx, store, repo)
	assert.Nil(t, err)
	assert.NotNil(t, pushConfig)
	assert.Equal(t, "https://user:pass@bitbucket.org/clone/link", pushConfig.RemoteURL)
}

func TestBitbucketCloudSource_WithAuthenticator(t *testing.T) {
	t.Run("unsupported types", func(t *testing.T) {
		s, _ := mockBitbucketCloudSource(t)

		for _, au := range []auth.Authenticator{
			&auth.OAuthBearerToken{},
			&auth.OAuthBearerTokenWithSSH{},
			&auth.OAuthClient{},
		} {
			t.Run(fmt.Sprintf("%T", au), func(t *testing.T) {
				newSource, err := s.WithAuthenticator(au)
				assert.Nil(t, newSource)
				assert.NotNil(t, err)
				assert.ErrorAs(t, err, &UnsupportedAuthenticatorError{})
			})
		}
	})

	t.Run("supported types", func(t *testing.T) {
		for _, au := range []auth.Authenticator{
			&auth.BasicAuth{},
			&auth.BasicAuthWithSSH{},
		} {
			t.Run(fmt.Sprintf("%T", au), func(t *testing.T) {
				newClient := NewStrictMockBitbucketCloudClient()

				s, client := mockBitbucketCloudSource(t)
				client.WithAuthenticatorFunc.SetDefaultHook(func(a auth.Authenticator) bitbucketcloud.Client {
					assert.Same(t, au, a)
					return newClient
				})

				newSource, err := s.WithAuthenticator(au)
				assert.Nil(t, err)
				assert.Same(t, newClient, newSource.(*BitbucketCloudSource).client)
			})
		}
	})
}

func TestBitbucketCloudSource_ValidateAuthenticator(t *testing.T) {
	ctx := context.Background()

	for name, want := range map[string]error{
		"nil":   nil,
		"error": errors.New("error"),
	} {
		t.Run(name, func(t *testing.T) {
			s, client := mockBitbucketCloudSource(t)
			client.PingFunc.SetDefaultReturn(want)

			assert.Equal(t, want, s.ValidateAuthenticator(ctx))
		})
	}
}

func TestBitbucketCloudSource_LoadChangeset(t *testing.T) {
	ctx := context.Background()

	mockChangeset := func() (*Changeset, *types.Repo, *bitbucketcloud.Repo) {
		bbRepo := &bitbucketcloud.Repo{FullName: "org/repo", UUID: "repo-uuid"}
		repo := &types.Repo{Metadata: bbRepo}
		cs := &Changeset{
			Changeset: &btypes.Changeset{
				ExternalID: "42",
			},
			TargetRepo: repo,
		}

		return cs, repo, bbRepo
	}

	t.Run("invalid external ID", func(t *testing.T) {
		s, _ := mockBitbucketCloudSource(t)

		cs, _, _ := mockChangeset()
		cs.ExternalID = "not a number"

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
	})

	t.Run("error getting pull request", func(t *testing.T) {
		cs, repo, _ := mockChangeset()
		s, client := mockBitbucketCloudSource(t)
		want := errors.New("error")
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, repo.Metadata, r)
			assert.EqualValues(t, 42, i)
			return nil, want
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, repo, _ := mockChangeset()
		s, client := mockBitbucketCloudSource(t)
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, repo.Metadata, r)
			assert.EqualValues(t, 42, i)
			return nil, &notFoundError{}
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		target := ChangesetNotFoundError{}
		assert.ErrorAs(t, err, &target)
		assert.Same(t, target.Changeset, cs)
	})

	t.Run("success", func(t *testing.T) {
		cs, repo, bbRepo := mockChangeset()
		s, client := mockBitbucketCloudSource(t)
		mockAnnotatePullRequest(t, client)
		pr := &bitbucketcloud.PullRequest{
			ID: 420,
			Source: bitbucketcloud.PullRequestEndpoint{
				Branch: bitbucketcloud.PullRequestBranch{Name: "branch"},
				Repo:   *bbRepo,
			},
			Destination: bitbucketcloud.PullRequestEndpoint{
				Branch: bitbucketcloud.PullRequestBranch{Name: "main"},
				Repo:   *bbRepo,
			},
		}
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, repo.Metadata, r)
			assert.EqualValues(t, 42, i)
			return pr, nil
		})

		err := s.LoadChangeset(ctx, cs)
		assert.Nil(t, err)

		// We're not thoroughly testing setChangesetMetadata() et al in this
		// test, but we do want to ensure that the PR was used to populate
		// fields on the Changeset.
		assert.Equal(t, "420", cs.ExternalID)
		assert.Equal(t, "refs/heads/branch", cs.ExternalBranch)
		assert.Empty(t, cs.ExternalForkNamespace)
	})
}

// TODO: annotatePullRequest and setChangesetMetadata need explicit unit tests.

func mockBitbucketCloudSource(t *testing.T) (*BitbucketCloudSource, *MockBitbucketCloudClient) {
	t.Helper()

	client := NewStrictMockBitbucketCloudClient()
	s := &BitbucketCloudSource{client: client}

	return s, client
}

// mockAnnotatePullRequest configures the mock client to be able to return a
// valid, empty set of statuses.
func mockAnnotatePullRequest(t *testing.T, client *MockBitbucketCloudClient) {
	t.Helper()
	client.GetPullRequestStatusesFunc.SetDefaultReturn(mockEmptyResultSet(t), nil)
}

func mockEmptyResultSet(t *testing.T) *bitbucketcloud.PaginatedResultSet {
	t.Helper()

	u, err := url.Parse("https://bitbucket.org/")
	if err != nil {
		t.Fatal(err)
	}

	return bitbucketcloud.NewPaginatedResultSet(u, func(ctx context.Context, r *http.Request) (*bitbucketcloud.PageToken, []interface{}, error) {
		return &bitbucketcloud.PageToken{}, nil, nil
	})
}

type notFoundError struct{}

var _ error = &notFoundError{}

func (notFoundError) Error() string {
	return "not found"
}

func (notFoundError) NotFound() bool {
	return true
}
