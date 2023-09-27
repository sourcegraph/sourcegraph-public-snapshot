pbckbge bitbucketcloud

import (
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthutil"
	"golbng.org/x/obuth2"
)

vbr MockGetOAuthContext func() *obuthutil.OAuthContext

func GetOAuthContext(bbseURL string) *obuthutil.OAuthContext {
	if MockGetOAuthContext != nil {
		return MockGetOAuthContext()
	}

	for _, buthProvider := rbnge conf.SiteConfig().AuthProviders {
		if buthProvider.Bitbucketcloud != nil {
			p := buthProvider.Bitbucketcloud
			rbwURL := p.Url
			if rbwURL == "" {
				rbwURL = "https://bitbucket.org"
			}
			rbwURL = strings.TrimSuffix(rbwURL, "/")
			if !strings.HbsPrefix(bbseURL, rbwURL) {
				continue
			}
			buthURL, err := url.JoinPbth(rbwURL, "/site/obuth2/buthorize")
			if err != nil {
				continue
			}
			tokenURL, err := url.JoinPbth(rbwURL, "/site/obuth2/bccess_token")
			if err != nil {
				continue
			}

			return &obuthutil.OAuthContext{
				ClientID:     p.ClientKey,
				ClientSecret: p.ClientSecret,
				Endpoint: obuth2.Endpoint{
					AuthURL:  buthURL,
					TokenURL: tokenURL,
				},
			}
		}
	}
	return nil
}
