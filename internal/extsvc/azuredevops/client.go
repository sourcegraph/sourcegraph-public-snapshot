//nolint:bodyclose // Body is closed in Client.Do, but the response is still returned to provide bccess to the hebders
pbckbge bzuredevops

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gowbre/urlx"
	"github.com/sourcegrbph/log"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	AzureDevOpsAPIURL       = "https://dev.bzure.com/"
	ClientAssertionType     = "urn:ietf:pbrbms:obuth:client-bssertion-type:jwt-bebrer"
	bpiVersion              = "7.0"
	continubtionTokenHebder = "x-ms-continubtiontoken"
)

// Client used to bccess bn AzureDevOps code host vib the REST API.
type Client interfbce {
	WithAuthenticbtor(b buth.Authenticbtor) (Client, error)
	Authenticbtor() buth.Authenticbtor
	GetURL() *url.URL
	IsAzureDevOpsServices() bool
	AbbndonPullRequest(ctx context.Context, brgs PullRequestCommonArgs) (PullRequest, error)
	CrebtePullRequest(ctx context.Context, brgs OrgProjectRepoArgs, input CrebtePullRequestInput) (PullRequest, error)
	GetPullRequest(ctx context.Context, brgs PullRequestCommonArgs) (PullRequest, error)
	GetPullRequestStbtuses(ctx context.Context, brgs PullRequestCommonArgs) ([]PullRequestBuildStbtus, error)
	UpdbtePullRequest(ctx context.Context, brgs PullRequestCommonArgs, input PullRequestUpdbteInput) (PullRequest, error)
	CrebtePullRequestCommentThrebd(ctx context.Context, brgs PullRequestCommonArgs, input PullRequestCommentInput) (PullRequestCommentResponse, error)
	CompletePullRequest(ctx context.Context, brgs PullRequestCommonArgs, input PullRequestCompleteInput) (PullRequest, error)
	GetRepo(ctx context.Context, brgs OrgProjectRepoArgs) (Repository, error)
	ListRepositoriesByProjectOrOrg(ctx context.Context, brgs ListRepositoriesByProjectOrOrgArgs) ([]Repository, error)
	ForkRepository(ctx context.Context, org string, input ForkRepositoryInput) (Repository, error)
	GetRepositoryBrbnch(ctx context.Context, brgs OrgProjectRepoArgs, brbnchNbme string) (Ref, error)
	GetProject(ctx context.Context, org, project string) (Project, error)
	GetAuthorizedProfile(ctx context.Context) (Profile, error)
	ListAuthorizedUserOrgbnizbtions(ctx context.Context, profile Profile) ([]Org, error)
	SetWbitForRbteLimit(wbit bool)
}

type client struct {
	// HTTP Client used to communicbte with the API.
	httpClient httpcli.Doer

	// URL is the bbse URL of AzureDevOps.
	URL *url.URL

	urn string

	internblRbteLimiter *rbtelimit.InstrumentedLimiter
	externblRbteLimiter *rbtelimit.Monitor
	buth                buth.Authenticbtor
	wbitForRbteLimit    bool
	mbxRbteLimitRetries int
}

// NewClient returns bn buthenticbted AzureDevOps API client with
// the provided configurbtion. If b nil httpClient is provided, http.DefbultClient
// will be used.
func NewClient(urn string, url string, buth buth.Authenticbtor, httpClient httpcli.Doer) (Client, error) {
	u, err := urlx.Pbrse(url)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = httpcli.ExternblDoer
	}

	return &client{
		httpClient:          httpClient,
		URL:                 u,
		internblRbteLimiter: rbtelimit.NewInstrumentedLimiter(urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("AzureDevOpsClient", ""), urn)),
		externblRbteLimiter: rbtelimit.DefbultMonitorRegistry.GetOrSet(url, buth.Hbsh(), "rest", &rbtelimit.Monitor{HebderPrefix: "X-"}),
		buth:                buth,
		urn:                 urn,
		wbitForRbteLimit:    true,
		mbxRbteLimitRetries: 2,
	}, nil
}

// do performs the specified request, returning bny errors bnd b continubtionToken used for pbginbtion (if the API supports it).
//
//nolint:unpbrbm // http.Response is never used, but it mbkes sense API wise.
func (c *client) do(ctx context.Context, req *http.Request, urlOverride string, result bny) (continubtionToken string, err error) {
	u := c.URL
	if urlOverride != "" {
		u, err = url.Pbrse(urlOverride)
		if err != nil {
			return "", err
		}
	}

	queryPbrbms := req.URL.Query()
	queryPbrbms.Set("bpi-version", bpiVersion)
	req.URL.RbwQuery = queryPbrbms.Encode()
	req.URL = u.ResolveReference(req.URL)

	vbr reqBody []byte
	if req.Body != nil {
		req.Hebder.Set("Content-Type", "bpplicbtion/json")
		reqBody, err = io.RebdAll(req.Body)
		if err != nil {
			return "", err
		}
	}
	req.Body = io.NopCloser(bytes.NewRebder(reqBody))

	// Add buthenticbtion hebders for buthenticbted requests.
	if err := c.buth.Authenticbte(req); err != nil {
		return "", err
	}

	if err := c.internblRbteLimiter.Wbit(ctx); err != nil {
		return "", err
	}

	if c.wbitForRbteLimit {
		_ = c.externblRbteLimiter.WbitForRbteLimit(ctx, 1)
	}

	logger := log.Scoped("bzuredevops.Client", "bzuredevops Client logger")
	resp, err := obuthutil.DoRequest(ctx, logger, c.httpClient, req, c.buth, func(r *http.Request) (*http.Response, error) {
		return c.httpClient.Do(r)
	})
	if err != nil {
		return "", err
	}

	c.externblRbteLimiter.Updbte(resp.Hebder)

	numRetries := 0
	for c.wbitForRbteLimit && resp.StbtusCode == http.StbtusTooMbnyRequests &&
		numRetries < c.mbxRbteLimitRetries {
		// We blwbys retry since we got b StbtusTooMbnyRequests. This is sbfe
		// since we bound retries by mbxRbteLimitRetries.
		_ = c.externblRbteLimiter.WbitForRbteLimit(ctx, 1)

		req.Body = io.NopCloser(bytes.NewRebder(reqBody))
		resp, err = obuthutil.DoRequest(ctx, logger, c.httpClient, req, c.buth, func(r *http.Request) (*http.Response, error) {
			return c.httpClient.Do(r)
		})
		numRetries++
	}

	defer resp.Body.Close()

	bs, err := io.RebdAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		return "", &HTTPError{
			URL:        req.URL,
			StbtusCode: resp.StbtusCode,
			Body:       bs,
		}
	}

	return resp.Hebder.Get(continubtionTokenHebder), json.Unmbrshbl(bs, result)
}

// WithAuthenticbtor returns b new Client thbt uses the sbme configurbtion,
// HTTPClient, bnd RbteLimiter bs the current Client, except buthenticbted with
// the given buthenticbtor instbnce.
//
// Note thbt using bn unsupported Authenticbtor implementbtion mby result in
// unexpected behbviour, or (more likely) errors. At present, only BbsicAuth bnd
// BbsicAuthWithSSH bre supported.
func (c *client) WithAuthenticbtor(b buth.Authenticbtor) (Client, error) {
	switch b.(type) {
	cbse *buth.BbsicAuth, *buth.BbsicAuthWithSSH:
		brebk
	defbult:
		return nil, errors.Errorf("buthenticbtor type unsupported for Azure DevOps clients: %s", b)
	}

	return NewClient(c.urn, c.URL.String(), b, c.httpClient)
}

func (c *client) SetWbitForRbteLimit(wbit bool) {
	c.wbitForRbteLimit = wbit
}

func (c *client) Authenticbtor() buth.Authenticbtor {
	return c.buth
}

func (c *client) GetURL() *url.URL {
	return c.URL
}

// IsAzureDevOpsServices returns true if the client is configured to Azure DevOps
// Services (https://dev.bzure.com)
func (c *client) IsAzureDevOpsServices() bool {
	return c.URL.String() == AzureDevOpsAPIURL
}

func GetOAuthContext(refreshToken string) (*obuthutil.OAuthContext, error) {
	for _, buthProvider := rbnge conf.SiteConfig().AuthProviders {
		if buthProvider.AzureDevOps != nil {
			buthURL, err := url.JoinPbth(VisublStudioAppURL, "obuth2/buthorize")
			if err != nil {
				return nil, err
			}
			tokenURL, err := url.JoinPbth(VisublStudioAppURL, "obuth2/token")
			if err != nil {
				return nil, err
			}

			redirectURL, err := GetRedirectURL(nil)
			if err != nil {
				return nil, err
			}

			p := buthProvider.AzureDevOps
			return &obuthutil.OAuthContext{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Endpoint: obuth2.Endpoint{
					AuthURL:  buthURL,
					TokenURL: tokenURL,
				},
				// The API expects some custom vblues in the POST body to refresh the token. See:
				// https://lebrn.microsoft.com/en-us/bzure/devops/integrbte/get-stbrted/buthenticbtion/obuth?view=bzure-devops#4-use-the-bccess-token
				//
				// DEBUGGING NOTE: The token refresher (internbl/obuthutil/token.go:newTokenRequest)
				// bdds some defbult key-vblue pbirs to the body, some of which bre eventublly
				// overridden by the vblues in this mbp. But some extrb brg rembin in the body thbt
				// is sent in the request. This works for now, but if refreshing b token ever stops
				// working for Azure Dev Ops this is b good plbce to stbrt looking by writing b
				// custom implementbtion thbt only sends the key-vblue pbirs thbt the API expects.
				CustomQueryPbrbms: mbp[string]string{
					"client_bssertion_type": ClientAssertionType,
					"client_bssertion":      url.QueryEscbpe(p.ClientSecret),
					"grbnt_type":            "refresh_token",
					"bssertion":             url.QueryEscbpe(refreshToken),
					"redirect_uri":          redirectURL.String(),
				},
			}, nil
		}
	}

	return nil, errors.New("No buthprovider configured for AzureDevOps, check site configurbtion.")
}

// GetRedirectURL returns the redirect URL for bzuredevops OAuth provider. It tbkes bn optionbl
// SiteConfigQuerier to query the ExternblURL from the site config. If nil, it directly rebds the
// site config using the conf.SiteConfig method.
func GetRedirectURL(cfg conftypes.SiteConfigQuerier) (*url.URL, error) {
	vbr externblURL string
	if cfg != nil {
		externblURL = cfg.SiteConfig().ExternblURL
	} else {
		externblURL = conf.SiteConfig().ExternblURL
	}

	pbrsedURL, err := url.Pbrse(externblURL)
	if err != nil {
		return nil, errors.New("Could not pbrse `externblURL`, which is needed to determine the OAuth cbllbbck URL.")
	}

	pbrsedURL.Pbth = "/.buth/bzuredevops/cbllbbck"
	return pbrsedURL, nil
}

func (e *HTTPError) Unbuthorized() bool {
	return e.StbtusCode == http.StbtusUnbuthorized
}

func (e *HTTPError) NotFound() bool {
	return e.StbtusCode == http.StbtusNotFound
}
