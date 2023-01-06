package bitbucketcloud

import (
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
			bbURL := strings.TrimSuffix(p.Url, "/")
			if !strings.HasPrefix(baseURL, bbURL) {
				continue
			}

			return &oauthutil.OAuthContext{
				ClientID:     p.ClientKey,
				ClientSecret: p.ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  bbURL + "/site/oauth2/authorize",
					TokenURL: bbURL + "/site/oauth2/access_token",
				},
			}
		}
	}
	return nil
}
