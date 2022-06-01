package sources

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestExtractCloneURL(t *testing.T) {
	tcs := []struct {
		name            string
		want            string
		configs         []string
		overrideRepoURL string
	}{
		{
			name: "https",
			want: "https://secrettoken@github.com/sourcegraph/sourcegraph",
			configs: []string{
				`{"url": "https://github.com", "token": "secrettoken", "authorization": {}}`,
			},
		},
		{
			name: "ssh",
			want: "git@github.com:sourcegraph/sourcegraph.git",
			configs: []string{
				`{"url": "https://github.com", "gitURLType": "ssh", "authorization": {}}`,
			},
		},
		{
			name: "https and ssh, favoring https",
			want: "https://secrettoken@github.com/sourcegraph/sourcegraph",
			configs: []string{
				`{"url": "https://github.com", "token": "secrettoken", "authorization": {}}`,
				`{"url": "https://github.com", "gitURLType": "ssh", "authorization": {}}`,
			},
		},
		{
			name: "https and ssh, favoring https different order",
			want: "https://secrettoken@github.com/sourcegraph/sourcegraph",
			configs: []string{
				`{"url": "https://github.com", "gitURLType": "ssh", "authorization": {}}`,
				`{"url": "https://github.com", "token": "secrettoken", "authorization": {}}`,
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			repo := &types.Repo{
				Name:    api.RepoName("github.com/sourcegraph/sourcegraph"),
				URI:     "github.com/sourcegraph/sourcegraph",
				Sources: make(map[string]*types.SourceInfo),
				Metadata: &github.Repository{
					NameWithOwner: "sourcegraph/sourcegraph",
					URL:           "https://github.com/sourcegraph/sourcegraph",
				},
			}
			if tc.overrideRepoURL != "" {
				repo.Metadata.(*github.Repository).URL = tc.overrideRepoURL
			}

			for idx := range tc.configs {
				repo.Sources[fmt.Sprintf("%d", idx)] = &types.SourceInfo{
					ID: fmt.Sprintf("::%d", idx), // see SourceInfo.ExternalServiceID
				}
			}

			ess := database.NewMockExternalServiceStore()
			ess.ListFunc.SetDefaultHook(func(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				services := make([]*types.ExternalService, 0, len(opt.IDs))
				for _, id := range opt.IDs {
					services = append(services, &types.ExternalService{
						ID:     id,
						Kind:   extsvc.KindGitHub,
						Config: tc.configs[int(id)],
					})
				}
				return services, nil
			})

			have, err := extractCloneURL(context.Background(), ess, repo)
			if err != nil {
				t.Fatal(err)
			}
			if have != tc.want {
				t.Fatalf("invalid cloneURL returned, want=%q have=%q", tc.want, have)
			}
		})
	}
}

func TestLoadExternalService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	noToken := types.ExternalService{
		ID:          1,
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub no token",
		Config:      `{"url": "https://github.com", "authorization": {}}`,
	}
	userOwnedWithToken := types.ExternalService{
		ID:              2,
		Kind:            extsvc.KindGitHub,
		DisplayName:     "GitHub user owned",
		NamespaceUserID: 1234,
		Config:          `{"url": "https://github.com", "token": "123", "authorization": {}}`,
	}
	withToken := types.ExternalService{
		ID:          3,
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub token",
		Config:      `{"url": "https://github.com", "token": "123", "authorization": {}}`,
	}
	withTokenNewer := types.ExternalService{
		ID:          4,
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub newer token",
		Config:      `{"url": "https://github.com", "token": "123456", "authorization": {}}`,
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
			noToken.URN(): {
				ID:       noToken.URN(),
				CloneURL: "https://github.com/sourcegraph/sourcegraph",
			},
			userOwnedWithToken.URN(): {
				ID:       userOwnedWithToken.URN(),
				CloneURL: "https://123@github.com/sourcegraph/sourcegraph",
			},
			withToken.URN(): {
				ID:       withToken.URN(),
				CloneURL: "https://123@github.com/sourcegraph/sourcegraph",
			},
			withTokenNewer.URN(): {
				ID:       withTokenNewer.URN(),
				CloneURL: "https://123456@github.com/sourcegraph/sourcegraph",
			},
		},
	}

	ess := database.NewMockExternalServiceStore()
	ess.ListFunc.SetDefaultHook(func(ctx context.Context, options database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		sources := make([]*types.ExternalService, 0)
		if _, ok := repo.Sources[noToken.URN()]; ok {
			sources = append(sources, &noToken)
		}
		if _, ok := repo.Sources[userOwnedWithToken.URN()]; ok {
			sources = append(sources, &userOwnedWithToken)
		}
		if _, ok := repo.Sources[withToken.URN()]; ok {
			sources = append(sources, &withToken)
		}
		if _, ok := repo.Sources[withTokenNewer.URN()]; ok {
			sources = append(sources, &withTokenNewer)
		}
		return sources, nil
	})

	// Expect the newest public external service with a token to be returned.
	svc, err := loadExternalService(ctx, ess, database.ExternalServicesListOptions{IDs: repo.ExternalServiceIDs()})
	if err != nil {
		t.Fatalf("invalid error, expected nil, got %v", err)
	}
	if have, want := svc.ID, withTokenNewer.ID; have != want {
		t.Fatalf("invalid external service returned, want=%d have=%d", want, have)
	}

	// Now delete the global external services and expect the user owned external service to be returned.
	delete(repo.Sources, withTokenNewer.URN())
	delete(repo.Sources, withToken.URN())
	svc, err = loadExternalService(ctx, ess, database.ExternalServicesListOptions{IDs: repo.ExternalServiceIDs()})
	if err != nil {
		t.Fatalf("invalid error, expected nil, got %v", err)
	}
	if have, want := svc.ID, userOwnedWithToken.ID; have != want {
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
		config              string
		repoMetadata        any
		authenticator       auth.Authenticator
		wantPushConfig      *protocol.PushConfig
		wantErr             error
	}{
		// Without authenticator:
		{
			name:                "GitHub HTTPS no token",
			repoName:            "github.com/sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitHub,
			config:              `{"url": "https://github.com", "authorization": {}}`,
			repoMetadata: &github.Repository{
				NameWithOwner: "sourcegraph/sourcegraph",
				URL:           "https://github.com/sourcegraph/sourcegraph",
			},
			wantErr: ErrNoPushCredentials{},
		},
		// With authenticator:
		{
			name:                "GitHub HTTPS no token with authenticator",
			repoName:            "github.com/sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitHub,
			config:              `{"url": "https://github.com", "authorization": {}}`,
			authenticator:       &oauthHTTPSAuthenticator,
			repoMetadata: &github.Repository{
				NameWithOwner: "sourcegraph/sourcegraph",
				URL:           "https://github.com/sourcegraph/sourcegraph",
			},
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://bearer-test@github.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitHub HTTPS token with authenticator",
			repoName:            "github.com/sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitHub,
			config:              `{"url": "https://github.com", "token": "token", "authorization": {}}`,
			authenticator:       &oauthHTTPSAuthenticator,
			repoMetadata: &github.Repository{
				NameWithOwner: "sourcegraph/sourcegraph",
				URL:           "https://github.com/sourcegraph/sourcegraph",
			},
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://bearer-test@github.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitHub SSH with authenticator",
			repoName:            "github.com/sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitHub,
			config:              `{"url": "https://github.com", "gitURLType": "ssh", "authorization": {}}`,
			authenticator:       &oauthSSHAuthenticator,
			repoMetadata: &github.Repository{
				NameWithOwner: "sourcegraph/sourcegraph",
				URL:           "https://github.com/sourcegraph/sourcegraph",
			},
			wantPushConfig: &protocol.PushConfig{
				RemoteURL:  "git@github.com:sourcegraph/sourcegraph.git",
				PrivateKey: "private-key",
				Passphrase: "passphrase",
			},
		},
		{
			name:                "GitLab HTTPS no token with authenticator",
			repoName:            "sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitLab,
			config:              `{}`,
			authenticator:       &oauthHTTPSAuthenticator,
			repoMetadata: &gitlab.Project{
				ProjectCommon: gitlab.ProjectCommon{
					ID:                1,
					PathWithNamespace: "sourcegraph/sourcegraph",
					HTTPURLToRepo:     "https://gitlab.com/sourcegraph/sourcegraph",
				}},
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://git:bearer-test@gitlab.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitLab HTTPS token with authenticator",
			repoName:            "sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitLab,
			config:              `{}`,
			authenticator:       &oauthHTTPSAuthenticator,
			repoMetadata: &gitlab.Project{
				ProjectCommon: gitlab.ProjectCommon{
					ID:                1,
					PathWithNamespace: "sourcegraph/sourcegraph",
					HTTPURLToRepo:     "https://git:token@gitlab.com/sourcegraph/sourcegraph",
				}},
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://git:bearer-test@gitlab.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitLab SSH with authenticator",
			repoName:            "sourcegraph/sourcegraph",
			externalServiceType: extsvc.TypeGitLab,
			config:              `{"gitURLType": "ssh"}`,
			authenticator:       &oauthSSHAuthenticator,
			repoMetadata: &gitlab.Project{
				ProjectCommon: gitlab.ProjectCommon{
					ID:                1,
					PathWithNamespace: "sourcegraph/sourcegraph",
					SSHURLToRepo:      "git@gitlab.com:sourcegraph/sourcegraph.git",
				}},
			wantPushConfig: &protocol.PushConfig{
				RemoteURL:  "git@gitlab.com:sourcegraph/sourcegraph.git",
				PrivateKey: "private-key",
				Passphrase: "passphrase",
			},
		},
		{
			name:                "Bitbucket server HTTPS no token with authenticator",
			externalServiceType: extsvc.TypeBitbucketServer,
			config:              `{}`,
			authenticator:       &basicHTTPSAuthenticator,
			repoMetadata: &bitbucketserver.Repo{
				ID:   1,
				Slug: "sourcegraph/sourcegraph",
				Project: &bitbucketserver.Project{
					Key: "sourcegraph/sourcegraph",
				},
				Links: struct {
					Clone []struct {
						Href string `json:"href"`
						Name string `json:"name"`
					} `json:"clone"`
					Self []struct {
						Href string `json:"href"`
					} `json:"self"`
				}{
					Clone: []struct {
						Href string "json:\"href\""
						Name string "json:\"name\""
					}{
						{Name: "http", Href: "https://bitbucket.sgdev.org/sourcegraph/sourcegraph"},
					},
				},
			},
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://basic:pw@bitbucket.sgdev.org/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "Bitbucket server HTTPS token with authenticator",
			externalServiceType: extsvc.TypeBitbucketServer,
			config:              `{}`,
			authenticator:       &basicHTTPSAuthenticator,
			repoMetadata: &bitbucketserver.Repo{
				ID:   1,
				Slug: "sourcegraph/sourcegraph",
				Project: &bitbucketserver.Project{
					Key: "sourcegraph/sourcegraph",
				},
				Links: struct {
					Clone []struct {
						Href string `json:"href"`
						Name string `json:"name"`
					} `json:"clone"`
					Self []struct {
						Href string `json:"href"`
					} `json:"self"`
				}{
					Clone: []struct {
						Href string "json:\"href\""
						Name string "json:\"name\""
					}{
						{Name: "http", Href: "https://token@bitbucket.sgdev.org/sourcegraph/sourcegraph"},
					},
				},
			},
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://basic:pw@bitbucket.sgdev.org/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "Bitbucket server SSH with authenticator",
			externalServiceType: extsvc.TypeBitbucketServer,
			config:              `{"gitURLType": "ssh"}`,
			authenticator:       &basicSSHAuthenticator,
			repoMetadata: &bitbucketserver.Repo{
				ID:   1,
				Slug: "sourcegraph/sourcegraph",
				Project: &bitbucketserver.Project{
					Key: "sourcegraph/sourcegraph",
				},
				Links: struct {
					Clone []struct {
						Href string `json:"href"`
						Name string `json:"name"`
					} `json:"clone"`
					Self []struct {
						Href string `json:"href"`
					} `json:"self"`
				}{
					Clone: []struct {
						Href string "json:\"href\""
						Name string "json:\"name\""
					}{
						{Name: "ssh", Href: "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph.git"},
					},
				},
			},
			wantPushConfig: &protocol.PushConfig{
				RemoteURL:  "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph.git",
				PrivateKey: "private-key",
				Passphrase: "passphrase",
			},
		},
		// Errors
		{
			name:                "Bitbucket server SSH no keypair",
			externalServiceType: extsvc.TypeBitbucketServer,
			config:              `{"gitURLType": "ssh"}`,
			authenticator:       &basicHTTPSAuthenticator,
			repoMetadata: &bitbucketserver.Repo{
				ID:   1,
				Slug: "sourcegraph/sourcegraph",
				Project: &bitbucketserver.Project{
					Key: "sourcegraph/sourcegraph",
				},
				Links: struct {
					Clone []struct {
						Href string `json:"href"`
						Name string `json:"name"`
					} `json:"clone"`
					Self []struct {
						Href string `json:"href"`
					} `json:"self"`
				}{
					Clone: []struct {
						Href string "json:\"href\""
						Name string "json:\"name\""
					}{
						{Name: "ssh", Href: "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph.git"},
					},
				},
			},
			wantErr: ErrNoSSHCredential,
		},
		{
			name:                "Invalid credential type",
			externalServiceType: extsvc.TypeBitbucketServer,
			config:              `{}`,
			authenticator:       &auth.OAuthClient{},
			repoMetadata: &bitbucketserver.Repo{
				ID:   1,
				Slug: "sourcegraph/sourcegraph",
				Project: &bitbucketserver.Project{
					Key: "sourcegraph/sourcegraph",
				},
				Links: struct {
					Clone []struct {
						Href string `json:"href"`
						Name string `json:"name"`
					} `json:"clone"`
					Self []struct {
						Href string `json:"href"`
					} `json:"self"`
				}{
					Clone: []struct {
						Href string "json:\"href\""
						Name string "json:\"name\""
					}{
						{Name: "ssh", Href: "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph.git"},
					},
				},
			},
			wantErr: ErrNoPushCredentials{CredentialsType: "*auth.OAuthClient"},
		},
	}
	for _, tt := range tcs {
		t.Run(tt.name, func(t *testing.T) {
			repo := &types.Repo{
				ExternalRepo: api.ExternalRepoSpec{
					ServiceType: tt.externalServiceType,
				},
				Name:     api.RepoName(tt.repoName),
				URI:      tt.repoName,
				Sources:  make(map[string]*types.SourceInfo),
				Metadata: tt.repoMetadata,
			}

			repo.Sources["1"] = &types.SourceInfo{
				ID: "::1", // see SourceInfo.ExternalServiceID
			}

			ess := database.NewMockExternalServiceStore()
			ess.ListFunc.SetDefaultHook(func(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				services := make([]*types.ExternalService, 0, len(opt.IDs))
				for _, id := range opt.IDs {
					services = append(services, &types.ExternalService{
						ID:     id,
						Kind:   extsvc.TypeToKind(tt.externalServiceType),
						Config: tt.config,
					})
				}

				return services, nil
			})

			havePushConfig, haveErr := gitserverPushConfig(context.Background(), ess, repo, tt.authenticator)
			if haveErr != tt.wantErr {
				t.Fatalf("invalid error returned, want=%v have=%v", tt.wantErr, haveErr)
			}
			if diff := cmp.Diff(havePushConfig, tt.wantPushConfig); diff != "" {
				t.Fatalf("invalid push config returned: %s", diff)
			}
		})
	}
}
