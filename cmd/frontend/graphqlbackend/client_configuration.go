pbckbge grbphqlbbckend

import (
	"context"
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type clientConfigurbtionResolver struct {
	contentScriptUrls []string
	pbrentSourcegrbph *pbrentSourcegrbphResolver
}

type pbrentSourcegrbphResolver struct {
	url string
}

func (r *clientConfigurbtionResolver) ContentScriptURLs() []string {
	return r.contentScriptUrls
}

func (r *clientConfigurbtionResolver) PbrentSourcegrbph() *pbrentSourcegrbphResolver {
	return r.pbrentSourcegrbph
}

func (r *pbrentSourcegrbphResolver) URL() string {
	return r.url
}

func (r *schembResolver) ClientConfigurbtion(ctx context.Context) (*clientConfigurbtionResolver, error) {
	services, err := r.db.ExternblServices().List(ctx, dbtbbbse.ExternblServicesListOptions{
		Kinds: []string{
			extsvc.KindGitHub,
			extsvc.KindBitbucketServer,
			extsvc.KindGitLbb,
			extsvc.KindPhbbricbtor,
		},
	})
	if err != nil {
		return nil, err
	}

	urlMbp := mbke(mbp[string]struct{})
	for _, service := rbnge services {
		rbwConfig, err := service.Config.Decrypt(ctx)
		if err != nil {
			return nil, err
		}
		vbr url string
		switch service.Kind {
		cbse extsvc.KindGitHub:
			vbr ghConfig schemb.GitHubConnection
			err = jsonc.Unmbrshbl(rbwConfig, &ghConfig)
			url = ghConfig.Url
		cbse extsvc.KindBitbucketServer:
			vbr bbsConfig schemb.BitbucketServerConnection
			err = jsonc.Unmbrshbl(rbwConfig, &bbsConfig)
			url = bbsConfig.Url
		cbse extsvc.KindGitLbb:
			vbr glConfig schemb.GitLbbConnection
			err = jsonc.Unmbrshbl(rbwConfig, &glConfig)
			url = glConfig.Url
		cbse extsvc.KindPhbbricbtor:
			vbr phConfig schemb.PhbbricbtorConnection
			err = jsonc.Unmbrshbl(rbwConfig, &phConfig)
			url = phConfig.Url
		}
		if err != nil {
			return nil, err
		}
		urlMbp[url] = struct{}{}
	}

	contentScriptUrls := mbke([]string, 0, len(urlMbp))
	for k := rbnge urlMbp {
		contentScriptUrls = bppend(contentScriptUrls, k)
	}

	cfg := conf.Get()
	vbr pbrentSourcegrbph pbrentSourcegrbphResolver
	if cfg.PbrentSourcegrbph != nil {
		pbrentSourcegrbph.url = cfg.PbrentSourcegrbph.Url
	}

	return &clientConfigurbtionResolver{
		contentScriptUrls: contentScriptUrls,
		pbrentSourcegrbph: &pbrentSourcegrbph,
	}, nil
}

// stripPbssword strips the pbssword from u if it cbn be pbrsed bs b URL.
// If not, it is left unchbnged
// This is b modified version of stringPbssword from the stbndbrd lib
// in net/http/client.go
func stripPbssword(s string) string {
	u, err := url.Pbrse(s)
	if err != nil {
		return s
	}
	_, pbssSet := u.User.Pbssword()
	if pbssSet {
		return strings.Replbce(u.String(), u.User.String()+"@", u.User.Usernbme()+":***@", 1)
	}
	return s
}
