package bitbucketcloud

import (
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"golang.org/x/oauth2"
)

var MockGetOAuthContext func() *oauthutil.OAuthContext

func GetOAuthContext(baseURL string) *oauthutil.OAuthContext {
	if MockGetOAuthContext != nil {
		return MockGetOAuthContext()
	}

	for _, authProvider := range conf.SiteConfig().AuthProviders {
		if authProvider.Bitbucketcloud != nil {
			p := authProvider.Bitbucketcloud
			rawURL := p.Url
			if rawURL == "" {
				rawURL = "https://bitbucket.org"
			}
			rawURL = strings.TrimSuffix(rawURL, "/")
			if !strings.HasPrefix(baseURL, rawURL) {
				continue
			}
			authURL, err := url.JoinPath(rawURL, "/site/oauth2/authorize")
			if err != nil {
				continue
			}
			tokenURL, err := url.JoinPath(rawURL, "/site/oauth2/access_token")
			if err != nil {
				continue
			}

			return &oauthutil.OAuthContext{
				ClientID:     p.ClientKey,
				ClientSecret: p.ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  authURL,
					TokenURL: tokenURL,
				},
			}
		}
	}
	return nil
}
