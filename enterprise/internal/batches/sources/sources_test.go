package sources

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	mockassert "github.com/derision-test/go-mockgen/testutil/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestExtractCloneURL(t *testing.T) {
	tcs := []struct {
		name      string
		want      string
		cloneURLs []string
	}{
		{
			name: "https",
			want: "https://github.com/sourcegraph/sourcegraph",
			cloneURLs: []string{
				`https://github.com/sourcegraph/sourcegraph`,
			},
		},
		{
			name: "ssh",
			want: "git@github.com:sourcegraph/sourcegraph.git",
			cloneURLs: []string{
				`git@github.com:sourcegraph/sourcegraph.git`,
			},
		},
		{
			name: "https and ssh, favoring https",
			want: "https://github.com/sourcegraph/sourcegraph",
			cloneURLs: []string{
				`https://github.com/sourcegraph/sourcegraph`,
				`git@github.com:sourcegraph/sourcegraph.git`,
			},
		},
		{
			name: "https and ssh, favoring https different order",
			want: "https://github.com/sourcegraph/sourcegraph",
			cloneURLs: []string{
				`git@github.com:sourcegraph/sourcegraph.git`,
				`https://github.com/sourcegraph/sourcegraph`,
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			sources := map[string]*types.SourceInfo{}
			for i, c := range tc.cloneURLs {
				sources[strconv.Itoa(i)] = &types.SourceInfo{
					ID:       strconv.Itoa(i),
					CloneURL: c,
				}
			}
			repo := &types.Repo{
				Name:    api.RepoName("github.com/sourcegraph/sourcegraph"),
				URI:     "github.com/sourcegraph/sourcegraph",
				Sources: sources,
				Metadata: &github.Repository{
					NameWithOwner: "sourcegraph/sourcegraph",
					URL:           "https://github.com/sourcegraph/sourcegraph",
				},
			}

			have, err := extractCloneURL(repo)
			if err != nil {
				t.Fatal(err)
			}
			if have.String() != tc.want {
				t.Fatalf("invalid cloneURL returned, want=%q have=%q", tc.want, have)
			}
		})
	}
}

func TestLoadExternalService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	globalES := types.ExternalService{
		ID:          1,
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub global",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "authorization": {}}`),
	}
	userOwnedES := types.ExternalService{
		ID:              2,
		Kind:            extsvc.KindGitHub,
		DisplayName:     "GitHub user owned",
		NamespaceUserID: 1234,
		Config:          extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "authorization": {}}`),
	}
	newerGlobalES := types.ExternalService{
		ID:          3,
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub global newer",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "authorization": {}}`),
	}

	repo := &types.Repo{
		Name:    api.RepoName("test-repo"),
		URI:     "test-repo",
		Private: true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "external-id-123",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*types.SourceInfo{
			globalES.URN(): {
				ID:       globalES.URN(),
				CloneURL: "https://github.com/sourcegraph/sourcegraph",
			},
			userOwnedES.URN(): {
				ID:       userOwnedES.URN(),
				CloneURL: "https://123@github.com/sourcegraph/sourcegraph",
			},
			newerGlobalES.URN(): {
				ID:       newerGlobalES.URN(),
				CloneURL: "https://123456@github.com/sourcegraph/sourcegraph",
			},
		},
	}

	ess := database.NewMockExternalServiceStore()
	ess.ListFunc.SetDefaultHook(func(ctx context.Context, options database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		sources := make([]*types.ExternalService, 0)
		if _, ok := repo.Sources[globalES.URN()]; ok {
			sources = append(sources, &globalES)
		}
		if _, ok := repo.Sources[userOwnedES.URN()]; ok {
			sources = append(sources, &userOwnedES)
		}
		if _, ok := repo.Sources[newerGlobalES.URN()]; ok {
			sources = append(sources, &newerGlobalES)
		}
		return sources, nil
	})

	// Expect the newest public external service with a token to be returned.
	svc, err := loadExternalService(ctx, ess, database.ExternalServicesListOptions{IDs: repo.ExternalServiceIDs()})
	if err != nil {
		t.Fatalf("invalid error, expected nil, got %v", err)
	}
	if have, want := svc.ID, newerGlobalES.ID; have != want {
		t.Fatalf("invalid external service returned, want=%d have=%d", want, have)
	}

	// Now delete the global external services and expect the user owned external service to be returned.
	delete(repo.Sources, newerGlobalES.URN())
	delete(repo.Sources, globalES.URN())
	svc, err = loadExternalService(ctx, ess, database.ExternalServicesListOptions{IDs: repo.ExternalServiceIDs()})
	if err != nil {
		t.Fatalf("invalid error, expected nil, got %v", err)
	}
	if have, want := svc.ID, userOwnedES.ID; have != want {
		t.Fatalf("invalid external service returned, want=%d have=%d", want, have)
	}
}

func TestGitserverPushConfig(t *testing.T) {
	oauthHTTPSAuthenticator := auth.OAuthBearerToken{Token: "bearer-test"}
	oauthSSHAuthenticator := auth.OAuthBearerTokenWithSSH{
		OAuthBearerToken: oauthHTTPSAuthenticator,
		PrivateKey:       "private-key",
		Passphrase:       "passphrase",
		PublicKey:        "public-key",
	}
	basicHTTPSAuthenticator := auth.BasicAuth{Username: "basic", Password: "pw"}
	basicSSHAuthenticator := auth.BasicAuthWithSSH{
		BasicAuth:  basicHTTPSAuthenticator,
		PrivateKey: "private-key",
		Passphrase: "passphrase",
		PublicKey:  "public-key",
	}
	tcs := []struct {
		name                string
		repoName            string
		externalServiceType string
		cloneURLs           []string
		authenticator       auth.Authenticator
		wantPushConfig      *protocol.PushConfig
		wantErr             error
	}{
		// Without authenticator:
		{
			name:                "GitHub HTTPS no token",
			repoName:            "github.com/sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitHub,
			cloneURLs:           []string{"https://github.com/sourcegraph/sourcegraph"},
			authenticator:       nil,
			wantErr:             ErrNoPushCredentials{},
		},
		// With authenticator:
		{
			name:                "GitHub HTTPS with authenticator",
			repoName:            "github.com/sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitHub,
			cloneURLs:           []string{"https://github.com/sourcegraph/sourcegraph"},
			authenticator:       &oauthHTTPSAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://bearer-test@github.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitHub SSH with authenticator",
			repoName:            "github.com/sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitHub,
			cloneURLs:           []string{"git@github.com:sourcegraph/sourcegraph.git"},
			authenticator:       &oauthSSHAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL:  "git@github.com:sourcegraph/sourcegraph.git",
				PrivateKey: "private-key",
				Passphrase: "passphrase",
			},
		},
		{
			name:                "GitLab HTTPS with authenticator",
			repoName:            "sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitLab,
			cloneURLs:           []string{"https://gitlab.com/sourcegraph/sourcegraph"},
			authenticator:       &oauthHTTPSAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://git:bearer-test@gitlab.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitLab SSH with authenticator",
			repoName:            "sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitLab,
			cloneURLs:           []string{"git@gitlab.com:sourcegraph/sourcegraph.git"},
			authenticator:       &oauthSSHAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL:  "git@gitlab.com:sourcegraph/sourcegraph.git",
				PrivateKey: "private-key",
				Passphrase: "passphrase",
			},
		},
		{
			name:                "Bitbucket server HTTPS with authenticator",
			repoName:            "bitbucket.sgdev.org/sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeBitbucketServer,
			cloneURLs:           []string{"https://bitbucket.sgdev.org/sourcegraph/sourcegraph"},
			authenticator:       &basicHTTPSAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://basic:pw@bitbucket.sgdev.org/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "Bitbucket server SSH with authenticator",
			repoName:            "bitbucket.sgdev.org/sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeBitbucketServer,
			authenticator:       &basicSSHAuthenticator,
			cloneURLs:           []string{"git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph.git"},
			wantPushConfig: &protocol.PushConfig{
				RemoteURL:  "git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph.git",
				PrivateKey: "private-key",
				Passphrase: "passphrase",
			},
		},
		// Errors
		{
			name:                "Bitbucket server SSH no keypair",
			externalServiceType: extsvc.TypeBitbucketServer,
			cloneURLs:           []string{"git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph.git"},
			authenticator:       &basicHTTPSAuthenticator,
			wantErr:             ErrNoSSHCredential,
		},
		{
			name:                "Invalid credential type",
			externalServiceType: extsvc.TypeBitbucketServer,
			cloneURLs:           []string{"https://bitbucket.sgdev.org/sourcegraph/sourcegraph"},
			authenticator:       &auth.OAuthClient{},
			wantErr:             ErrNoPushCredentials{CredentialsType: "*auth.OAuthClient"},
		},
	}
	for _, tt := range tcs {
		t.Run(tt.name, func(t *testing.T) {
			sources := map[string]*types.SourceInfo{}
			for i, c := range tt.cloneURLs {
				sources[strconv.Itoa(i)] = &types.SourceInfo{
					ID:       strconv.Itoa(i),
					CloneURL: c,
				}
			}
			repo := &types.Repo{
				Name:    api.RepoName(tt.repoName),
				URI:     tt.repoName,
				Sources: sources,
				ExternalRepo: api.ExternalRepoSpec{
					ServiceType: tt.externalServiceType,
				},
			}

			havePushConfig, haveErr := GitserverPushConfig(repo, tt.authenticator)
			if haveErr != tt.wantErr {
				t.Fatalf("invalid error returned, want=%v have=%v", tt.wantErr, haveErr)
			}
			if diff := cmp.Diff(havePushConfig, tt.wantPushConfig); diff != "" {
				t.Fatalf("invalid push config returned: %s", diff)
			}
		})
	}
}

func TestSourcer_ForChangeset(t *testing.T) {
	ctx := context.Background()

	es := &types.ExternalService{
		ID:          1,
		Kind:        extsvc.KindGitLab,
		DisplayName: "GitHub.com",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "token": "123", "authorization": {}}`),
	}
	repo := &types.Repo{
		Name:    api.RepoName("test-repo"),
		URI:     "test-repo",
		Private: true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "external-id-123",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*types.SourceInfo{
			es.URN(): {
				ID:       es.URN(),
				CloneURL: "https://123@github.com/sourcegraph/sourcegraph",
			},
		},
	}

	siteToken := &auth.OAuthBearerToken{Token: "site"}
	userToken := &auth.OAuthBearerToken{Token: "user"}

	t.Run("created changesets", func(t *testing.T) {
		bc := &btypes.BatchChange{ID: 1, LastApplierID: 3}
		ch := &btypes.Changeset{ID: 2, OwnedByBatchChangeID: 1}

		t.Run("with user credential", func(t *testing.T) {
			credStore := database.NewMockUserCredentialsStore()
			credStore.GetByScopeFunc.SetDefaultHook(func(ctx context.Context, opts database.UserCredentialScope) (*database.UserCredential, error) {
				assert.EqualValues(t, repo.ExternalRepo.ServiceID, opts.ExternalServiceID)
				assert.EqualValues(t, repo.ExternalRepo.ServiceType, opts.ExternalServiceType)
				assert.EqualValues(t, bc.LastApplierID, opts.UserID)
				cred := &database.UserCredential{Credential: database.NewEmptyCredential()}
				cred.SetAuthenticator(ctx, userToken)
				return cred, nil
			})

			tx := NewMockSourcerStore()
			rs := database.NewMockRepoStore()
			rs.GetFunc.SetDefaultReturn(repo, nil)
			tx.ReposFunc.SetDefaultReturn(rs)
			tx.GetBatchChangeFunc.SetDefaultHook(func(ctx context.Context, opts store.GetBatchChangeOpts) (*btypes.BatchChange, error) {
				assert.EqualValues(t, bc.ID, opts.ID)
				return bc, nil
			})
			tx.UserCredentialsFunc.SetDefaultReturn(credStore)

			css := NewMockChangesetSource()
			want := NewMockChangesetSource()
			css.WithAuthenticatorFunc.SetDefaultHook(func(a auth.Authenticator) (ChangesetSource, error) {
				assert.Equal(t, userToken, a)
				return want, nil
			})

			have, err := newMockSourcer(css).ForChangeset(ctx, tx, ch)
			assert.NoError(t, err)
			assert.Same(t, want, have)
		})

		t.Run("with site credential", func(t *testing.T) {
			credStore := database.NewMockUserCredentialsStore()
			credStore.GetByScopeFunc.SetDefaultHook(func(ctx context.Context, opts database.UserCredentialScope) (*database.UserCredential, error) {
				assert.EqualValues(t, repo.ExternalRepo.ServiceID, opts.ExternalServiceID)
				assert.EqualValues(t, repo.ExternalRepo.ServiceType, opts.ExternalServiceType)
				assert.EqualValues(t, bc.LastApplierID, opts.UserID)
				return nil, &errcode.Mock{IsNotFound: true}
			})

			tx := NewMockSourcerStore()
			rs := database.NewMockRepoStore()
			rs.GetFunc.SetDefaultReturn(repo, nil)
			tx.ReposFunc.SetDefaultReturn(rs)
			tx.GetBatchChangeFunc.SetDefaultHook(func(ctx context.Context, opts store.GetBatchChangeOpts) (*btypes.BatchChange, error) {
				assert.EqualValues(t, bc.ID, opts.ID)
				return bc, nil
			})
			tx.GetSiteCredentialFunc.SetDefaultHook(func(ctx context.Context, opts store.GetSiteCredentialOpts) (*btypes.SiteCredential, error) {
				assert.EqualValues(t, repo.ExternalRepo.ServiceID, opts.ExternalServiceID)
				assert.EqualValues(t, repo.ExternalRepo.ServiceType, opts.ExternalServiceType)
				cred := &btypes.SiteCredential{Credential: database.NewEmptyCredential()}
				cred.SetAuthenticator(ctx, siteToken)
				return cred, nil
			})
			tx.UserCredentialsFunc.SetDefaultReturn(credStore)

			css := NewMockChangesetSource()
			want := NewMockChangesetSource()
			css.WithAuthenticatorFunc.SetDefaultHook(func(a auth.Authenticator) (ChangesetSource, error) {
				assert.Equal(t, siteToken, a)
				return want, nil
			})

			have, err := newMockSourcer(css).ForChangeset(ctx, tx, ch)
			assert.NoError(t, err)
			assert.Same(t, want, have)
		})

		t.Run("without site credential", func(t *testing.T) {
			// When we remove the fallback to the external service
			// configuration, this test is expected to fail.
			credStore := database.NewMockUserCredentialsStore()
			credStore.GetByScopeFunc.SetDefaultHook(func(ctx context.Context, opts database.UserCredentialScope) (*database.UserCredential, error) {
				assert.EqualValues(t, repo.ExternalRepo.ServiceID, opts.ExternalServiceID)
				assert.EqualValues(t, repo.ExternalRepo.ServiceType, opts.ExternalServiceType)
				assert.EqualValues(t, bc.LastApplierID, opts.UserID)
				return nil, &errcode.Mock{IsNotFound: true}
			})

			tx := NewMockSourcerStore()
			rs := database.NewMockRepoStore()
			rs.GetFunc.SetDefaultReturn(repo, nil)
			tx.ReposFunc.SetDefaultReturn(rs)
			tx.GetBatchChangeFunc.SetDefaultHook(func(ctx context.Context, opts store.GetBatchChangeOpts) (*btypes.BatchChange, error) {
				assert.EqualValues(t, bc.ID, opts.ID)
				return bc, nil
			})
			tx.GetSiteCredentialFunc.SetDefaultHook(func(ctx context.Context, opts store.GetSiteCredentialOpts) (*btypes.SiteCredential, error) {
				assert.EqualValues(t, repo.ExternalRepo.ServiceID, opts.ExternalServiceID)
				assert.EqualValues(t, repo.ExternalRepo.ServiceType, opts.ExternalServiceType)
				return nil, store.ErrNoResults
			})
			tx.UserCredentialsFunc.SetDefaultReturn(credStore)

			css := NewMockChangesetSource()
			_, err := newMockSourcer(css).ForChangeset(ctx, tx, ch)
			assert.Error(t, err)
		})
	})

	t.Run("imported changesets", func(t *testing.T) {
		ch := &btypes.Changeset{ID: 2, OwnedByBatchChangeID: 0}

		t.Run("with site credential", func(t *testing.T) {
			tx := NewMockSourcerStore()
			rs := database.NewMockRepoStore()
			rs.GetFunc.SetDefaultReturn(repo, nil)
			tx.ReposFunc.SetDefaultReturn(rs)
			tx.GetSiteCredentialFunc.SetDefaultHook(func(ctx context.Context, opts store.GetSiteCredentialOpts) (*btypes.SiteCredential, error) {
				assert.EqualValues(t, repo.ExternalRepo.ServiceID, opts.ExternalServiceID)
				assert.EqualValues(t, repo.ExternalRepo.ServiceType, opts.ExternalServiceType)
				cred := &btypes.SiteCredential{Credential: database.NewEmptyCredential()}
				cred.SetAuthenticator(ctx, siteToken)
				return cred, nil
			})

			css := NewMockChangesetSource()
			want := NewMockChangesetSource()
			css.WithAuthenticatorFunc.SetDefaultHook(func(a auth.Authenticator) (ChangesetSource, error) {
				assert.Equal(t, siteToken, a)
				return want, nil
			})

			have, err := newMockSourcer(css).ForChangeset(ctx, tx, ch)
			assert.NoError(t, err)
			assert.Same(t, want, have)
		})

		t.Run("without site credential", func(t *testing.T) {
			// When we remove the fallback to the external service
			// configuration, this test is expected to fail.
			tx := NewMockSourcerStore()
			rs := database.NewMockRepoStore()
			rs.GetFunc.SetDefaultReturn(repo, nil)
			tx.ReposFunc.SetDefaultReturn(rs)
			tx.GetSiteCredentialFunc.SetDefaultHook(func(ctx context.Context, opts store.GetSiteCredentialOpts) (*btypes.SiteCredential, error) {
				assert.EqualValues(t, repo.ExternalRepo.ServiceID, opts.ExternalServiceID)
				assert.EqualValues(t, repo.ExternalRepo.ServiceType, opts.ExternalServiceType)
				return nil, store.ErrNoResults
			})

			css := NewMockChangesetSource()
			want := errors.New("validator was called")
			css.ValidateAuthenticatorFunc.SetDefaultReturn(want)

			_, err := newMockSourcer(css).ForChangeset(ctx, tx, ch)
			assert.Error(t, err)
		})
	})
}

func TestGetRemoteRepo(t *testing.T) {
	ctx := context.Background()
	targetRepo := &types.Repo{}

	t.Run("forks disabled", func(t *testing.T) {
		t.Run("unforked changeset", func(t *testing.T) {
			// Set up a changeset source that will panic if any methods are invoked.
			css := NewStrictMockChangesetSource()

			// This should succeed, since loadRemoteRepo() should early return with
			// forks disabled.
			remoteRepo, err := GetRemoteRepo(ctx, css, targetRepo, &btypes.Changeset{}, &btypes.ChangesetSpec{
				ForkNamespace: nil,
			})
			assert.Nil(t, err)
			assert.Same(t, targetRepo, remoteRepo)
		})

		t.Run("forked changeset", func(t *testing.T) {
			forkNamespace := "fork"
			want := &types.Repo{}
			css := NewMockForkableChangesetSource()
			css.GetNamespaceForkFunc.SetDefaultReturn(want, nil)

			// This should succeed, since loadRemoteRepo() should early return with
			// forks disabled.
			remoteRepo, err := GetRemoteRepo(ctx, css, targetRepo, &btypes.Changeset{}, &btypes.ChangesetSpec{
				ForkNamespace: &forkNamespace,
			})
			assert.Nil(t, err)
			assert.Same(t, want, remoteRepo)
			mockassert.CalledOnce(t, css.GetNamespaceForkFunc)
		})
	})

	t.Run("forks enabled", func(t *testing.T) {
		forkNamespace := "<user>"

		t.Run("unforkable changeset source", func(t *testing.T) {
			css := NewMockChangesetSource()

			repo, err := GetRemoteRepo(ctx, css, targetRepo, &btypes.Changeset{}, &btypes.ChangesetSpec{
				ForkNamespace: &forkNamespace,
			})
			assert.Nil(t, repo)
			assert.ErrorIs(t, err, ErrChangesetSourceCannotFork)
		})

		t.Run("forkable changeset source", func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				want := &types.Repo{}
				css := NewMockForkableChangesetSource()
				css.GetUserForkFunc.SetDefaultReturn(want, nil)

				have, err := GetRemoteRepo(ctx, css, targetRepo, &btypes.Changeset{}, &btypes.ChangesetSpec{
					ForkNamespace: &forkNamespace,
				})
				assert.Nil(t, err)
				assert.Same(t, want, have)
				mockassert.CalledOnce(t, css.GetUserForkFunc)
			})

			t.Run("error from the source", func(t *testing.T) {
				want := errors.New("source error")
				css := NewMockForkableChangesetSource()
				css.GetUserForkFunc.SetDefaultReturn(nil, want)

				repo, err := GetRemoteRepo(ctx, css, targetRepo, &btypes.Changeset{}, &btypes.ChangesetSpec{
					ForkNamespace: &forkNamespace,
				})
				assert.Nil(t, repo)
				assert.Same(t, want, err)
				mockassert.CalledOnce(t, css.GetUserForkFunc)
			})
		})
	})
}

func newMockSourcer(css ChangesetSource) Sourcer {
	return newSourcer(nil, func(ctx context.Context, tx SourcerStore, cf *httpcli.Factory, externalServiceIDs []int64) (ChangesetSource, error) {
		return css, nil
	})
}
