package sources

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestExtractCloneURL(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name      string
		want      string
		cloneURLs []string
	}{
		{
			name:      "https",
			want:      "https://secrettoken@github.com/sourcegraph/sourcegraph",
			cloneURLs: []string{"https://secrettoken@github.com/sourcegraph/sourcegraph"},
		},
		{
			name:      "https user password",
			want:      "https://git:secrettoken@github.com/sourcegraph/sourcegraph",
			cloneURLs: []string{"https://git:secrettoken@github.com/sourcegraph/sourcegraph"},
		},
		{
			name:      "ssh no protocol specified",
			want:      "ssh://git@github.com/sourcegraph/sourcegraph.git",
			cloneURLs: []string{"git@github.com:sourcegraph/sourcegraph.git"},
		},
		{
			name:      "ssh protocol specified",
			want:      "ssh://git@github.com/sourcegraph/sourcegraph.git",
			cloneURLs: []string{"ssh://git@github.com/sourcegraph/sourcegraph.git"},
		},
		{
			name: "https and ssh, favoring https",
			want: "https://secrettoken@github.com/sourcegraph/sourcegraph",
			cloneURLs: []string{
				"https://secrettoken@github.com/sourcegraph/sourcegraph",
				"git@github.com:sourcegraph/sourcegraph.git",
				"ssh://git@github.com/sourcegraph/sourcegraph.git",
			},
		},
		{
			name: "https and ssh, favoring https different order",
			want: "https://secrettoken@github.com/sourcegraph/sourcegraph",
			cloneURLs: []string{
				"git@github.com:sourcegraph/sourcegraph.git",
				"ssh://git@github.com/sourcegraph/sourcegraph.git",
				"https://secrettoken@github.com/sourcegraph/sourcegraph",
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			repo := &types.Repo{
				Sources: map[string]*types.SourceInfo{},
			}
			for _, cloneURL := range tc.cloneURLs {
				repo.Sources[cloneURL] = &types.SourceInfo{
					CloneURL: cloneURL,
				}
			}
			have, err := extractCloneURL(repo)
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

	database.Mocks.ExternalServices.List = func(opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
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
	}
	t.Cleanup(func() {
		database.Mocks.ExternalServices.List = nil
	})

	// Expect the newest public external service with a token to be returned.
	svc, err := loadExternalService(ctx, &database.ExternalServiceStore{}, database.ExternalServicesListOptions{IDs: repo.ExternalServiceIDs()})
	if err != nil {
		t.Fatalf("invalid error, expected nil, got %v", err)
	}
	if have, want := svc.ID, withTokenNewer.ID; have != want {
		t.Fatalf("invalid external service returned, want=%d have=%d", want, have)
	}

	// Now delete the global external services and expect the user owned external service to be returned.
	delete(repo.Sources, withTokenNewer.URN())
	delete(repo.Sources, withToken.URN())
	svc, err = loadExternalService(ctx, &database.ExternalServiceStore{}, database.ExternalServicesListOptions{IDs: repo.ExternalServiceIDs()})
	if err != nil {
		t.Fatalf("invalid error, expected nil, got %v", err)
	}
	if have, want := svc.ID, userOwnedWithToken.ID; have != want {
		t.Fatalf("invalid external service returned, want=%d have=%d", want, have)
	}
}

func TestBatchesSource_GitserverPushConfig(t *testing.T) {
	t.Parallel()

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
		externalServiceType string
		cloneURL            string
		authenticator       auth.Authenticator
		wantPushConfig      *protocol.PushConfig
		wantErr             error
	}{
		// Without authenticator:
		{
			name:                "GitHub HTTPS no token",
			externalServiceType: extsvc.TypeGitHub,
			cloneURL:            "https://github.com/sourcegraph/sourcegraph",
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://github.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitHub HTTPS token",
			externalServiceType: extsvc.TypeGitHub,
			cloneURL:            "https://token@github.com/sourcegraph/sourcegraph",
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://token@github.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitHub SSH",
			externalServiceType: extsvc.TypeGitHub,
			cloneURL:            "git@github.com:sourcegraph/sourcegraph.git",
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "ssh://git@github.com/sourcegraph/sourcegraph.git",
			},
		},
		{
			name:                "GitLab HTTPS no token",
			externalServiceType: extsvc.TypeGitLab,
			cloneURL:            "https://gitlab.com/sourcegraph/sourcegraph",
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://gitlab.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitLab HTTPS token",
			externalServiceType: extsvc.TypeGitLab,
			cloneURL:            "https://git:token@gitlab.com/sourcegraph/sourcegraph",
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://git:token@gitlab.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitLab SSH",
			externalServiceType: extsvc.TypeGitLab,
			cloneURL:            "git@gitlab.com:sourcegraph/sourcegraph.git",
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "ssh://git@gitlab.com/sourcegraph/sourcegraph.git",
			},
		},
		{
			name:                "Bitbucket server HTTPS no token",
			externalServiceType: extsvc.TypeBitbucketServer,
			cloneURL:            "https://bitbucket.sgdev.org/sourcegraph/sourcegraph",
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://bitbucket.sgdev.org/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "Bitbucket server HTTPS token",
			externalServiceType: extsvc.TypeBitbucketServer,
			cloneURL:            "https://token@bitbucket.sgdev.org/sourcegraph/sourcegraph",
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://token@bitbucket.sgdev.org/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "Bitbucket server SSH",
			externalServiceType: extsvc.TypeBitbucketServer,
			cloneURL:            "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph",
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph",
			},
		},
		// With authenticator:
		{
			name:                "GitHub HTTPS no token with authenticator",
			externalServiceType: extsvc.TypeGitHub,
			cloneURL:            "https://github.com/sourcegraph/sourcegraph",
			authenticator:       &oauthHTTPSAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://bearer-test@github.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitHub HTTPS token with authenticator",
			externalServiceType: extsvc.TypeGitHub,
			cloneURL:            "https://token@github.com/sourcegraph/sourcegraph",
			authenticator:       &oauthHTTPSAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://bearer-test@github.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitHub SSH with authenticator",
			externalServiceType: extsvc.TypeGitHub,
			cloneURL:            "git@github.com:sourcegraph/sourcegraph.git",
			authenticator:       &oauthSSHAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL:  "ssh://git@github.com/sourcegraph/sourcegraph.git",
				PrivateKey: "private-key",
				Passphrase: "passphrase",
			},
		},
		{
			name:                "GitLab HTTPS no token with authenticator",
			externalServiceType: extsvc.TypeGitLab,
			cloneURL:            "https://gitlab.com/sourcegraph/sourcegraph",
			authenticator:       &oauthHTTPSAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://git:bearer-test@gitlab.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitLab HTTPS token with authenticator",
			externalServiceType: extsvc.TypeGitLab,
			cloneURL:            "https://git:token@gitlab.com/sourcegraph/sourcegraph",
			authenticator:       &oauthHTTPSAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://git:bearer-test@gitlab.com/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "GitLab SSH with authenticator",
			externalServiceType: extsvc.TypeGitLab,
			cloneURL:            "git@gitlab.com:sourcegraph/sourcegraph.git",
			authenticator:       &oauthSSHAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL:  "ssh://git@gitlab.com/sourcegraph/sourcegraph.git",
				PrivateKey: "private-key",
				Passphrase: "passphrase",
			},
		},
		{
			name:                "Bitbucket server HTTPS no token with authenticator",
			externalServiceType: extsvc.TypeBitbucketServer,
			cloneURL:            "https://bitbucket.sgdev.org/sourcegraph/sourcegraph",
			authenticator:       &basicHTTPSAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://basic:pw@bitbucket.sgdev.org/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "Bitbucket server HTTPS token with authenticator",
			externalServiceType: extsvc.TypeBitbucketServer,
			cloneURL:            "https://token@bitbucket.sgdev.org/sourcegraph/sourcegraph",
			authenticator:       &basicHTTPSAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL: "https://basic:pw@bitbucket.sgdev.org/sourcegraph/sourcegraph",
			},
		},
		{
			name:                "Bitbucket server SSH with authenticator",
			externalServiceType: extsvc.TypeBitbucketServer,
			cloneURL:            "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph",
			authenticator:       &basicSSHAuthenticator,
			wantPushConfig: &protocol.PushConfig{
				RemoteURL:  "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph",
				PrivateKey: "private-key",
				Passphrase: "passphrase",
			},
		},
		// Errors
		{
			name:                "Bitbucket server SSH no keypair",
			externalServiceType: extsvc.TypeBitbucketServer,
			cloneURL:            "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/sourcegraph",
			authenticator:       &basicHTTPSAuthenticator,
			wantErr:             ErrNoSSHCredential,
		},
		{
			name:                "Invalid credential type",
			externalServiceType: extsvc.TypeGitHub,
			cloneURL:            "https://github.com/sourcegraph/sourcegraph",
			authenticator:       &auth.OAuthClient{},
			wantErr:             ErrNoPushCredentials{CredentialsType: "*auth.OAuthClient"},
		},
	}
	for _, tt := range tcs {
		t.Run(tt.name, func(t *testing.T) {
			repo := &types.Repo{
				ExternalRepo: api.ExternalRepoSpec{
					ServiceType: tt.externalServiceType,
				},
				Sources: map[string]*types.SourceInfo{tt.cloneURL: {CloneURL: tt.cloneURL}},
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
