package sources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	bbcs "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources/bitbucketcloud"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestNewBitbucketCloudSource(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		for name, input := range map[string]string{
			"invalid JSON":   "invalid JSON",
			"invalid schema": `{"appPassword": ["not a string"]}`,
			"bad URN":        `{"apiURL": "http://[::1]:namedport"}`,
		} {
			t.Run(name, func(t *testing.T) {
				ctx := context.Background()
				s, err := NewBitbucketCloudSource(ctx, &types.ExternalService{
					Config: extsvc.NewUnencryptedConfig(input),
				}, nil)
				assert.Nil(t, s)
				assert.NotNil(t, err)
			})
		}
	})

	t.Run("valid", func(t *testing.T) {
		ctx := context.Background()
		s, err := NewBitbucketCloudSource(ctx, &types.ExternalService{Config: extsvc.NewEmptyConfig()}, nil)
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
	s, client := mockBitbucketCloudSource()
	client.AuthenticatorFunc.SetDefaultReturn(&au)

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

	pushConfig, err := s.GitserverPushConfig(repo)
	assert.Nil(t, err)
	assert.NotNil(t, pushConfig)
	assert.Equal(t, "https://user:pass@bitbucket.org/clone/link", pushConfig.RemoteURL)
}

func TestBitbucketCloudSource_WithAuthenticator(t *testing.T) {
	t.Run("unsupported types", func(t *testing.T) {
		s, _ := mockBitbucketCloudSource()

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

				s, client := mockBitbucketCloudSource()
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
			s, client := mockBitbucketCloudSource()
			client.PingFunc.SetDefaultReturn(want)

			assert.Equal(t, want, s.ValidateAuthenticator(ctx))
		})
	}
}

func TestBitbucketCloudSource_LoadChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid external ID", func(t *testing.T) {
		s, _ := mockBitbucketCloudSource()

		cs, _, _ := mockBitbucketCloudChangeset()
		cs.ExternalID = "not a number"

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
	})

	t.Run("error getting pull request", func(t *testing.T) {
		cs, repo, _ := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
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
		cs, repo, _ := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
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

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, repo, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
		want := mockAnnotatePullRequestError(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, repo.Metadata, r)
			assert.EqualValues(t, 42, i)
			return pr, nil
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, repo, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
		mockAnnotatePullRequestSuccess(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, repo.Metadata, r)
			assert.EqualValues(t, 42, i)
			return pr, nil
		})

		err := s.LoadChangeset(ctx, cs)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestBitbucketCloudSource_CreateChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error creating pull request", func(t *testing.T) {
		cs, repo, _ := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()

		want := errors.New("error")
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, repo.Metadata, r)
			assert.Equal(t, cs.Title, pri.Title)
			return nil, want
		})

		exists, err := s.CreateChangeset(ctx, cs)
		assert.False(t, exists)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, repo, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
		want := mockAnnotatePullRequestError(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, repo.Metadata, r)
			assert.Equal(t, cs.Title, pri.Title)
			return pr, nil
		})

		exists, err := s.CreateChangeset(ctx, cs)
		assert.False(t, exists)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, repo, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
		mockAnnotatePullRequestSuccess(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, repo.Metadata, r)
			assert.Equal(t, cs.Title, pri.Title)
			assert.Nil(t, pri.SourceRepo)
			return pr, nil
		})

		exists, err := s.CreateChangeset(ctx, cs)
		assert.True(t, exists)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})

	t.Run("success with fork", func(t *testing.T) {
		cs, repo, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
		mockAnnotatePullRequestSuccess(client)

		fork := &bitbucketcloud.Repo{
			UUID:     "fork-uuid",
			FullName: "fork/repo",
			Slug:     "repo",
		}
		cs.RemoteRepo = &types.Repo{
			Metadata: fork,
		}

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, repo.Metadata, r)
			assert.Equal(t, cs.Title, pri.Title)
			assert.Equal(t, fork, pri.SourceRepo)
			return pr, nil
		})

		exists, err := s.CreateChangeset(ctx, cs)
		assert.True(t, exists)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestBitbucketCloudSource_CloseChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error declining pull request", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		want := errors.New("error")
		client.DeclinePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			return nil, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.CloseChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
		want := mockAnnotatePullRequestError(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.DeclinePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			return pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.CloseChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
		mockAnnotatePullRequestSuccess(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.DeclinePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			return pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.CloseChangeset(ctx, cs)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestBitbucketCloudSource_UpdateChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error updating pull request", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		want := errors.New("error")
		client.UpdatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			assert.Equal(t, cs.Title, pri.Title)
			return nil, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
		want := mockAnnotatePullRequestError(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.UpdatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			assert.Equal(t, cs.Title, pri.Title)
			return pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
		mockAnnotatePullRequestSuccess(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.UpdatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			assert.Equal(t, cs.Title, pri.Title)
			return pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.UpdateChangeset(ctx, cs)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestBitbucketCloudSource_CreateComment(t *testing.T) {
	ctx := context.Background()

	t.Run("error creating comment", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		want := errors.New("error")
		client.CreatePullRequestCommentFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, ci bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			assert.Equal(t, "comment", ci.Content)
			return nil, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.CreateComment(ctx, cs, "comment")
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.CreatePullRequestCommentFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, ci bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			assert.Equal(t, "comment", ci.Content)
			return &bitbucketcloud.Comment{}, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.CreateComment(ctx, cs, "comment")
		assert.Nil(t, err)
	})
}

func TestBitbucketCloudSource_MergeChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error merging pull request", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		want := errors.New("error")
		client.MergePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, mpro bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			assert.Nil(t, mpro.MergeStrategy)
			return nil, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.MergeChangeset(ctx, cs, false)
		assert.NotNil(t, err)
		target := ChangesetNotMergeableError{}
		assert.ErrorAs(t, err, &target)
		assert.Equal(t, want.Error(), target.ErrorMsg)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		want := &notFoundError{}
		client.MergePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, mpro bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			assert.Nil(t, mpro.MergeStrategy)
			return nil, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.MergeChangeset(ctx, cs, false)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChangeset()
		s, client := mockBitbucketCloudSource()
		want := mockAnnotatePullRequestError(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.MergePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, mpro bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			assert.Nil(t, mpro.MergeStrategy)
			return pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.MergeChangeset(ctx, cs, false)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		squash := bitbucketcloud.MergeStrategySquash
		for name, tc := range map[string]struct {
			squash bool
			want   *bitbucketcloud.MergeStrategy
		}{
			"no squash": {false, nil},
			"squash":    {true, &squash},
		} {
			t.Run(name, func(t *testing.T) {
				cs, _, bbRepo := mockBitbucketCloudChangeset()
				s, client := mockBitbucketCloudSource()
				mockAnnotatePullRequestSuccess(client)

				pr := mockBitbucketCloudPullRequest(bbRepo)
				client.MergePullRequestFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, mpro bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
					assert.Same(t, bbRepo, r)
					assert.EqualValues(t, 420, i)
					assert.Equal(t, tc.want, mpro.MergeStrategy)
					return pr, nil
				})

				annotateChangesetWithPullRequest(cs, pr)
				err := s.MergeChangeset(ctx, cs, tc.squash)
				assert.Nil(t, err)
				assertChangesetMatchesPullRequest(t, cs, pr)
			})
		}
	})
}

func TestBitbucketCloudSource_Fork(t *testing.T) {
	// We'll test both GetNamespaceFork and GetUserFork in here, since they're
	// closely related anyway.

	ctx := context.Background()

	upstream := &bitbucketcloud.Repo{
		UUID:     "repo-uuid",
		FullName: "upstream/repo",
		Slug:     "repo",
	}
	upstreamRepo := &types.Repo{Metadata: upstream}

	fork := &bitbucketcloud.Repo{
		UUID:     "fork-uuid",
		FullName: "fork/repo",
		Slug:     "repo",
	}

	t.Run("GetNamespaceFork", func(t *testing.T) {
		t.Run("error checking for repo", func(t *testing.T) {
			s, client := mockBitbucketCloudSource()

			want := errors.New("error")
			client.RepoFunc.SetDefaultHook(func(ctx context.Context, namespace, slug string) (*bitbucketcloud.Repo, error) {
				assert.Equal(t, "fork", namespace)
				assert.Equal(t, "repo", slug)
				return nil, want
			})

			repo, err := s.GetNamespaceFork(ctx, upstreamRepo, "fork")
			assert.Nil(t, repo)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, want)
		})

		t.Run("forked repo already exists", func(t *testing.T) {
			s, client := mockBitbucketCloudSource()

			client.RepoFunc.SetDefaultHook(func(ctx context.Context, namespace, slug string) (*bitbucketcloud.Repo, error) {
				assert.Equal(t, "fork", namespace)
				assert.Equal(t, "repo", slug)
				return fork, nil
			})

			repo, err := s.GetNamespaceFork(ctx, upstreamRepo, "fork")
			assert.Nil(t, err)
			assert.NotNil(t, repo)
			assert.Same(t, fork, repo.Metadata)
		})

		t.Run("fork error", func(t *testing.T) {
			s, client := mockBitbucketCloudSource()

			client.RepoFunc.SetDefaultHook(func(ctx context.Context, namespace, slug string) (*bitbucketcloud.Repo, error) {
				assert.Equal(t, "fork", namespace)
				assert.Equal(t, "repo", slug)
				return nil, &notFoundError{}
			})

			want := errors.New("error")
			client.ForkRepositoryFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, fi bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
				assert.Same(t, upstream, r)
				assert.EqualValues(t, "fork", fi.Workspace)
				return nil, want
			})

			repo, err := s.GetNamespaceFork(ctx, upstreamRepo, "fork")
			assert.Nil(t, repo)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, want)
		})

		t.Run("success", func(t *testing.T) {
			s, client := mockBitbucketCloudSource()

			client.RepoFunc.SetDefaultHook(func(ctx context.Context, namespace, slug string) (*bitbucketcloud.Repo, error) {
				assert.Equal(t, "fork", namespace)
				assert.Equal(t, "repo", slug)
				return nil, &notFoundError{}
			})

			client.ForkRepositoryFunc.SetDefaultHook(func(ctx context.Context, r *bitbucketcloud.Repo, fi bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
				assert.Same(t, upstream, r)
				assert.EqualValues(t, "fork", fi.Workspace)
				return fork, nil
			})

			repo, err := s.GetNamespaceFork(ctx, upstreamRepo, "fork")
			assert.Nil(t, err)
			assert.NotNil(t, repo)
			assert.Same(t, fork, repo.Metadata)
		})
	})

	t.Run("GetUserFork", func(t *testing.T) {
		t.Run("error getting current user", func(t *testing.T) {
			s, client := mockBitbucketCloudSource()

			want := errors.New("error")
			client.CurrentUserFunc.SetDefaultReturn(nil, want)

			repo, err := s.GetUserFork(ctx, upstreamRepo)
			assert.Nil(t, repo)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, want)
		})

		t.Run("success", func(t *testing.T) {
			s, client := mockBitbucketCloudSource()

			user := &bitbucketcloud.User{
				Account: bitbucketcloud.Account{
					Username: "user",
				},
			}
			client.CurrentUserFunc.SetDefaultReturn(user, nil)

			client.RepoFunc.SetDefaultHook(func(ctx context.Context, namespace, slug string) (*bitbucketcloud.Repo, error) {
				assert.Equal(t, "user", namespace)
				assert.Equal(t, "repo", slug)
				return fork, nil
			})

			repo, err := s.GetUserFork(ctx, upstreamRepo)
			assert.Nil(t, err)
			assert.NotNil(t, repo)
			assert.Same(t, fork, repo.Metadata)
		})
	})
}

func TestBitbucketCloudSource_annotatePullRequest(t *testing.T) {
	// The case where GetPullRequestStatuses errors and where it returns an
	// empty result set are thoroughly covered in other tests, so we'll just
	// handle the other branches of annotatePullRequest.

	ctx := context.Background()

	t.Run("error getting all statuses", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()
		_, _, bbRepo := mockBitbucketCloudChangeset()
		pr := mockBitbucketCloudPullRequest(bbRepo)

		want := errors.New("error")
		client.GetPullRequestStatusesFunc.SetDefaultHook(func(r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PaginatedResultSet, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			return bitbucketcloud.NewPaginatedResultSet(mockBitbucketCloudURL(), func(ctx context.Context, r *http.Request) (*bitbucketcloud.PageToken, []any, error) {
				return nil, nil, want
			}), nil
		})

		apr, err := s.annotatePullRequest(ctx, bbRepo, pr)
		assert.Nil(t, apr)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()
		_, _, bbRepo := mockBitbucketCloudChangeset()
		pr := mockBitbucketCloudPullRequest(bbRepo)

		want := []*bitbucketcloud.PullRequestStatus{
			{UUID: "1"},
			{UUID: "2"},
		}
		client.GetPullRequestStatusesFunc.SetDefaultHook(func(r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PaginatedResultSet, error) {
			assert.Same(t, bbRepo, r)
			assert.EqualValues(t, 420, i)
			first := true
			return bitbucketcloud.NewPaginatedResultSet(mockBitbucketCloudURL(), func(ctx context.Context, r *http.Request) (*bitbucketcloud.PageToken, []any, error) {
				if first {
					out := []any{}
					for _, status := range want {
						out = append(out, status)
					}

					first = false
					return &bitbucketcloud.PageToken{}, out, nil
				} else {
					return &bitbucketcloud.PageToken{}, nil, nil
				}
			}), nil
		})

		apr, err := s.annotatePullRequest(ctx, bbRepo, pr)
		assert.Nil(t, err)
		assert.NotNil(t, apr)
		assert.Same(t, pr, apr.PullRequest)
		assert.Equal(t, want, apr.Statuses)
	})
}

func TestBitbucketCloudSource_setChangesetMetadata(t *testing.T) {
	// The only interesting case we didn't cover in any other test is what
	// happens if Changeset.SetMetadata returns an error, so let's set that up.

	ctx := context.Background()
	s, client := mockBitbucketCloudSource()
	mockAnnotatePullRequestSuccess(client)

	cs, _, bbRepo := mockBitbucketCloudChangeset()
	pr := mockBitbucketCloudPullRequest(bbRepo)
	pr.Source.Repo.FullName = "no-slash"
	pr.Source.Repo.UUID = "a-different-uuid"

	err := s.setChangesetMetadata(ctx, bbRepo, pr, cs)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "setting changeset metadata")
}

func assertChangesetMatchesPullRequest(t *testing.T, cs *Changeset, pr *bitbucketcloud.PullRequest) {
	t.Helper()

	// We're not thoroughly testing setChangesetMetadata() et al in this
	// assertion, but we do want to ensure that the PR was used to populate
	// fields on the Changeset.
	assert.EqualValues(t, strconv.FormatInt(pr.ID, 10), cs.ExternalID)
	assert.Equal(t, "refs/heads/"+pr.Source.Branch.Name, cs.ExternalBranch)

	if pr.Source.Repo.UUID != pr.Destination.Repo.UUID {
		ns, err := pr.Source.Repo.Namespace()
		assert.Nil(t, err)
		assert.Equal(t, ns, cs.ExternalForkNamespace)
	} else {
		assert.Empty(t, cs.ExternalForkNamespace)
	}
}

// mockBitbucketCloudChangeset creates a plausible non-forked changeset, repo,
// and Bitbucket Cloud specific repo.
func mockBitbucketCloudChangeset() (*Changeset, *types.Repo, *bitbucketcloud.Repo) {
	bbRepo := &bitbucketcloud.Repo{FullName: "org/repo", UUID: "repo-uuid"}
	repo := &types.Repo{Metadata: bbRepo}
	cs := &Changeset{
		Changeset: &btypes.Changeset{
			ExternalID: "42",
		},
		RemoteRepo: repo,
		TargetRepo: repo,
	}

	return cs, repo, bbRepo
}

// mockBitbucketCloudPullRequest returns a plausible pull request that would be
// returned from Bitbucket Cloud for a non-forked changeset.
func mockBitbucketCloudPullRequest(repo *bitbucketcloud.Repo) *bitbucketcloud.PullRequest {
	return &bitbucketcloud.PullRequest{
		ID: 420,
		Source: bitbucketcloud.PullRequestEndpoint{
			Branch: bitbucketcloud.PullRequestBranch{Name: "branch"},
			Repo:   *repo,
		},
		Destination: bitbucketcloud.PullRequestEndpoint{
			Branch: bitbucketcloud.PullRequestBranch{Name: "main"},
			Repo:   *repo,
		},
	}
}

func annotateChangesetWithPullRequest(cs *Changeset, pr *bitbucketcloud.PullRequest) {
	cs.Metadata = &bbcs.AnnotatedPullRequest{
		PullRequest: pr,
		Statuses:    []*bitbucketcloud.PullRequestStatus{},
	}
}

func mockBitbucketCloudSource() (*BitbucketCloudSource, *MockBitbucketCloudClient) {
	client := NewStrictMockBitbucketCloudClient()
	s := &BitbucketCloudSource{client: client}

	return s, client
}

// mockAnnotatePullRequestError configures the mock client to return an error
// when GetPullRequestStatuses is invoked by annotatePullRequest.
func mockAnnotatePullRequestError(client *MockBitbucketCloudClient) error {
	err := errors.New("error")
	client.GetPullRequestStatusesFunc.SetDefaultReturn(nil, err)

	return err
}

// mockAnnotatePullRequestSuccess configures the mock client to be able to
// return a valid, empty set of statuses.
func mockAnnotatePullRequestSuccess(client *MockBitbucketCloudClient) {
	client.GetPullRequestStatusesFunc.SetDefaultReturn(mockEmptyResultSet(), nil)
}

func mockEmptyResultSet() *bitbucketcloud.PaginatedResultSet {
	return bitbucketcloud.NewPaginatedResultSet(mockBitbucketCloudURL(), func(ctx context.Context, r *http.Request) (*bitbucketcloud.PageToken, []any, error) {
		return &bitbucketcloud.PageToken{}, nil, nil
	})
}

func mockBitbucketCloudURL() *url.URL {
	u, err := url.Parse("https://bitbucket.org/")
	if err != nil {
		panic(err)
	}

	return u
}

type notFoundError struct{}

var _ error = &notFoundError{}

func (notFoundError) Error() string {
	return "not found"
}

func (notFoundError) NotFound() bool {
	return true
}
