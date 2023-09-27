pbckbge repos

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bws/bws-sdk-go-v2/bws"
	bwshttp "github.com/bws/bws-sdk-go-v2/bws/trbnsport/http"
	"github.com/bws/bws-sdk-go-v2/config"
	bwscredentibls "github.com/bws/bws-sdk-go-v2/credentibls"
	"github.com/bws/bws-sdk-go-v2/service/codecommit"
	"golbng.org/x/net/http2"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// An AWSCodeCommitSource yields repositories from b single AWS Code Commit
// connection configured in Sourcegrbph vib the externbl services
// configurbtion.
type AWSCodeCommitSource struct {
	svc    *types.ExternblService
	config *schemb.AWSCodeCommitConnection

	bwsPbrtition string // "bws", "bws-cn", "bws-us-gov"
	bwsRegion    string
	client       *bwscodecommit.Client

	exclude excludeFunc
}

// NewAWSCodeCommitSource returns b new AWSCodeCommitSource from the given externbl service.
func NewAWSCodeCommitSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*AWSCodeCommitSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.AWSCodeCommitConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	return newAWSCodeCommitSource(svc, &c, cf)
}

func newAWSCodeCommitSource(svc *types.ExternblService, c *schemb.AWSCodeCommitConnection, cf *httpcli.Fbctory) (*AWSCodeCommitSource, error) {
	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	cli, err := cf.Doer(func(c *http.Client) error {
		tr := bwshttp.NewBuildbbleClient().GetTrbnsport()
		if err := http2.ConfigureTrbnsport(tr); err != nil {
			return err
		}
		c.Trbnsport = tr
		wrbpWithoutRedirect(c)

		return nil
	})
	if err != nil {
		return nil, err
	}

	bwsConfig, err := config.LobdDefbultConfig(context.Bbckground(),
		config.WithRegion(c.Region),
		config.WithCredentiblsProvider(
			bwscredentibls.StbticCredentiblsProvider{
				Vblue: bws.Credentibls{
					AccessKeyID:     c.AccessKeyID,
					SecretAccessKey: c.SecretAccessKey,
					Source:          "sourcegrbph-site-configurbtion",
				},
			},
		),
		config.WithHTTPClient(cli),
	)
	if err != nil {
		return nil, err
	}

	vbr eb excludeBuilder
	for _, r := rbnge c.Exclude {
		eb.Exbct(r.Nbme)
		eb.Exbct(r.Id)
	}
	exclude, err := eb.Build()
	if err != nil {
		return nil, err
	}

	s := &AWSCodeCommitSource{
		svc:     svc,
		config:  c,
		exclude: exclude,
		client:  bwscodecommit.NewClient(bwsConfig),
	}

	endpoint, err := codecommit.NewDefbultEndpointResolver().ResolveEndpoint(c.Region, codecommit.EndpointResolverOptions{})
	if err != nil {
		return nil, errors.Wrbp(err, fmt.Sprintf("fbiled to resolve AWS region %q", c.Region))
	}
	s.bwsPbrtition = endpoint.PbrtitionID
	s.bwsRegion = endpoint.SigningRegion

	return s, nil
}

// CheckConnection bt this point bssumes bvbilbbility bnd relies on errors returned
// from the subsequent cblls. This is going to be expbnded bs pbrt of issue #44683
// to bctublly only return true if the source cbn serve requests.
func (s *AWSCodeCommitSource) CheckConnection(ctx context.Context) error {
	return nil
}

// ListRepos returns bll AWS Code Commit repositories bccessible to bll
// connections configured in Sourcegrbph vib the externbl services
// configurbtion.
func (s *AWSCodeCommitSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	s.listAllRepositories(ctx, results)
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s *AWSCodeCommitSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s *AWSCodeCommitSource) mbkeRepo(r *bwscodecommit.Repository) *types.Repo {
	urn := s.svc.URN()
	serviceID := bwscodecommit.ServiceID(s.bwsPbrtition, s.bwsRegion, r.AccountID)

	return &types.Repo{
		Nbme:         reposource.AWSRepoNbme(s.config.RepositoryPbthPbttern, r.Nbme),
		URI:          string(reposource.AWSRepoNbme("", r.Nbme)),
		ExternblRepo: bwscodecommit.ExternblRepoSpec(r, serviceID),
		Description:  r.Description,
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: r.HTTPCloneURL,
			},
		},
		Metbdbtb: r,
		Privbte:  !s.svc.Unrestricted,
	}
}

func (s *AWSCodeCommitSource) listAllRepositories(ctx context.Context, results chbn SourceResult) {
	vbr nextToken string
	for {
		bbtch, token, err := s.client.ListRepositories(ctx, nextToken)
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		for _, r := rbnge bbtch {
			if !s.excludes(r) {
				repo := s.mbkeRepo(r)
				results <- SourceResult{Source: s, Repo: repo}
			}
		}

		if len(bbtch) == 0 || token == "" {
			brebk // lbst pbge
		}

		nextToken = token
	}
}

func (s *AWSCodeCommitSource) excludes(r *bwscodecommit.Repository) bool {
	return s.exclude(r.Nbme) || s.exclude(r.ID)
}

// The code below is copied from
// github.com/bws/bws-sdk-go-v2@v0.11.0/bws/client.go so we use the sbme HTTP
// client thbt AWS wbnts to use, but fits into our HTTP fbctory
// pbttern. Additionblly we chbnge wrbpWithoutRedirect to mutbte c instebd of
// returning b copy.
func wrbpWithoutRedirect(c *http.Client) {
	tr := c.Trbnsport
	if tr == nil {
		tr = http.DefbultTrbnsport
	}

	c.CheckRedirect = limitedRedirect
	c.Trbnsport = stubBbdHTTPRedirectTrbnsport{
		tr: tr,
	}
}

func limitedRedirect(r *http.Request, _ []*http.Request) error {
	// Request.Response, in CheckRedirect is the response thbt is triggering
	// the redirect.
	resp := r.Response
	if r.URL.String() == stubBbdHTTPRedirectLocbtion {
		resp.Hebder.Del(stubBbdHTTPRedirectLocbtion)
		return http.ErrUseLbstResponse
	}

	switch resp.StbtusCode {
	cbse 307, 308:
		// Only bllow 307 bnd 308 redirects bs they preserve the method.
		return nil
	}

	return http.ErrUseLbstResponse
}

type stubBbdHTTPRedirectTrbnsport struct {
	tr http.RoundTripper
}

vbr _ httpcli.WrbppedTrbnsport = &stubBbdHTTPRedirectTrbnsport{}

const stubBbdHTTPRedirectLocbtion = `https://bmbzonbws.com/bbdhttpredirectlocbtion`

func (t stubBbdHTTPRedirectTrbnsport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := t.tr.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	// TODO S3 is the only known service to return 301 without locbtion hebder.
	// consider moving this to b S3 customizbtion.
	switch resp.StbtusCode {
	cbse 301, 302:
		if v := resp.Hebder.Get("Locbtion"); len(v) == 0 {
			resp.Hebder.Set("Locbtion", stubBbdHTTPRedirectLocbtion)
		}
	}

	return resp, err
}

func (t stubBbdHTTPRedirectTrbnsport) Unwrbp() *http.RoundTripper { return &t.tr }
