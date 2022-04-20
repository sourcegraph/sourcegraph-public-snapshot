package sources

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

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
	client := NewStrictMockBitbucketCloudClient()
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

	s := &BitbucketCloudSource{client: client}
	pushConfig, err := s.GitserverPushConfig(ctx, store, repo)
	assert.Nil(t, err)
	assert.NotNil(t, pushConfig)
	assert.Equal(t, "https://user:pass@bitbucket.org/clone/link", pushConfig.RemoteURL)
}

func TestBitbucketCloudSource_WithAuthenticator(t *testing.T) {
	t.Run("unsupported types", func(t *testing.T) {
		client := NewStrictMockBitbucketCloudClient()
		s := &BitbucketCloudSource{client: client}

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

				client := NewStrictMockBitbucketCloudClient()
				client.WithAuthenticatorFunc.SetDefaultHook(func(a auth.Authenticator) bitbucketcloud.Client {
					assert.Same(t, au, a)
					return newClient
				})
				s := &BitbucketCloudSource{client: client}

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
			client := NewStrictMockBitbucketCloudClient()
			client.PingFunc.SetDefaultReturn(want)
			s := &BitbucketCloudSource{client: client}

			assert.Equal(t, want, s.ValidateAuthenticator(ctx))
		})
	}
}
