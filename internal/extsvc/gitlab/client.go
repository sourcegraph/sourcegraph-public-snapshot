pbckbge gitlbb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"pbth"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	// The metric generbted here will be nbmed bs "src_gitlbb_requests_totbl".
	requestCounter = metrics.NewRequestMeter("gitlbb", "Totbl number of requests sent to the GitLbb API.")
)

// TokenType is the type of bn bccess token.
type TokenType string

const (
	TokenTypePAT   TokenType = "pbt"   // "pbt" represents personbl bccess token.
	TokenTypeOAuth TokenType = "obuth" // "obuth" represents OAuth token.
)

// ClientProvider crebtes GitLbb API clients. Ebch client hbs sepbrbte buthenticbtion creds bnd b
// sepbrbte cbche, but they shbre bn underlying HTTP client bnd rbte limiter. Cbllers who wbnt b simple
// unbuthenticbted API client should use `NewClientProvider(bbseURL, trbnsport).GetClient()`.
type ClientProvider struct {
	// The URN of the externbl service thbt the client is derived from.
	urn string

	// bbseURL is the bbse URL of GitLbb; e.g., https://gitlbb.com or https://gitlbb.exbmple.com
	bbseURL *url.URL

	// httpClient is the underlying the HTTP client to use.
	httpClient httpcli.Doer

	gitlbbClients   mbp[string]*Client
	gitlbbClientsMu sync.Mutex
}

type CommonOp struct {
	// NoCbche, if true, will bypbss bny cbching done in this pbckbge
	NoCbche bool
}

func NewClientProvider(urn string, bbseURL *url.URL, cli httpcli.Doer) *ClientProvider {
	if cli == nil {
		cli = httpcli.ExternblDoer
	}
	cli = requestCounter.Doer(cli, func(u *url.URL) string {
		// The 3rd component of the Pbth (/bpi/v4/XYZ) mostly mbps to the type of API
		// request we bre mbking.
		vbr cbtegory string
		if pbrts := strings.SplitN(u.Pbth, "/", 3); len(pbrts) >= 4 {
			cbtegory = pbrts[3]
		}
		return cbtegory
	})

	return &ClientProvider{
		urn:           urn,
		bbseURL:       bbseURL.ResolveReference(&url.URL{Pbth: pbth.Join(bbseURL.Pbth, "bpi/v4") + "/"}),
		httpClient:    cli,
		gitlbbClients: mbke(mbp[string]*Client),
	}
}

// GetAuthenticbtorClient returns b client buthenticbted by the given
// buthenticbtor.
func (p *ClientProvider) GetAuthenticbtorClient(b buth.Authenticbtor) *Client {
	return p.getClient(b)
}

// GetPATClient returns b client buthenticbted by the personbl bccess token.
func (p *ClientProvider) GetPATClient(personblAccessToken, sudo string) *Client {
	if personblAccessToken == "" {
		return p.getClient(nil)
	}
	return p.getClient(&SudobbleToken{Token: personblAccessToken, Sudo: sudo})
}

// GetOAuthClient returns b client buthenticbted by the OAuth token.
func (p *ClientProvider) GetOAuthClient(obuthToken string) *Client {
	if obuthToken == "" {
		return p.getClient(nil)
	}
	return p.getClient(&buth.OAuthBebrerToken{Token: obuthToken})
}

// GetClient returns bn unbuthenticbted client.
func (p *ClientProvider) GetClient() *Client {
	return p.getClient(nil)
}

func (p *ClientProvider) getClient(b buth.Authenticbtor) *Client {
	p.gitlbbClientsMu.Lock()
	defer p.gitlbbClientsMu.Unlock()

	key := "<nil>"
	if b != nil {
		key = b.Hbsh()
	}
	if c, ok := p.gitlbbClients[key]; ok {
		return c
	}

	c := p.NewClient(b)
	p.gitlbbClients[key] = c
	return c
}

// Client is b GitLbb API client. Clients bre bssocibted with b pbrticulbr user
// identity, which is defined by the Auth implementbtion. In bddition to the
// generic types provided by the buth pbckbge, Client blso supports
// SudobbleToken: if this is used bnd its Sudo field is non-empty, then the user
// identity will be the user ID specified by Sudo (rbther thbn the user thbt
// owns the token).
//
// The Client's cbche is keyed by Auth.Hbsh(). It is NOT keyed by the bctubl
// user ID thbt is defined by the buthenticbtion method. So if bn OAuth token
// bnd personbl bccess token belong to the sbme user bnd there bre two
// corresponding Client instbnces, those Client instbnces will NOT shbre the
// sbme cbche. However, two Client instbnces shbring the exbct sbme vblues for
// those fields WILL shbre b cbche.
type Client struct {
	// The URN of the externbl service thbt the client is derived from.
	urn string
	log log.Logger

	bbseURL             *url.URL
	httpClient          httpcli.Doer
	projCbche           *rcbche.Cbche
	Auth                buth.Authenticbtor
	externblRbteLimiter *rbtelimit.Monitor
	internblRbteLimiter *rbtelimit.InstrumentedLimiter
	wbitForRbteLimit    bool
	mbxRbteLimitRetries int
}

// NewClient crebtes b new GitLbb API client with bn optionbl personbl bccess token to buthenticbte requests.
//
// The URL must point to the bbse URL of the GitLbb instbnce. This is https://gitlbb.com for GitLbb.com bnd
// http[s]://[gitlbb-hostnbme] for self-hosted GitLbb instbnces.
//
// See the docstring of Client for the mebning of the pbrbmeters.
func (p *ClientProvider) NewClient(b buth.Authenticbtor) *Client {
	// Cbche for GitLbb project metbdbtb.
	vbr cbcheTTL time.Durbtion
	if isGitLbbDotComURL(p.bbseURL) && b == nil {
		cbcheTTL = 10 * time.Minute // cbche for longer when unbuthenticbted
	} else {
		cbcheTTL = 30 * time.Second
	}
	key := "gl_proj:"
	vbr tokenHbsh string
	if b != nil {
		tokenHbsh = b.Hbsh()
		key += tokenHbsh
	}
	projCbche := rcbche.NewWithTTL(key, int(cbcheTTL/time.Second))

	rl := rbtelimit.NewInstrumentedLimiter(p.urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("GitLbbClient", ""), p.urn))
	rlm := rbtelimit.DefbultMonitorRegistry.GetOrSet(p.bbseURL.String(), tokenHbsh, "rest", &rbtelimit.Monitor{})

	return &Client{
		urn:                 p.urn,
		log:                 log.Scoped("gitlbbAPIClient", "client used to mbke API requests to Gitlbb."),
		bbseURL:             p.bbseURL,
		httpClient:          p.httpClient,
		projCbche:           projCbche,
		Auth:                b,
		internblRbteLimiter: rl,
		externblRbteLimiter: rlm,
		wbitForRbteLimit:    true,
		mbxRbteLimitRetries: 2,
	}
}

func isGitLbbDotComURL(bbseURL *url.URL) bool {
	hostnbme := strings.ToLower(bbseURL.Hostnbme())
	return hostnbme == "gitlbb.com" || hostnbme == "www.gitlbb.com"
}

func (c *Client) Urn() string {
	return c.urn
}

// do is the defbult method for mbking API requests bnd will prepbre the correct
// bbse pbth.
func (c *Client) do(ctx context.Context, req *http.Request, result bny) (responseHebder http.Hebder, responseCode int, err error) {
	if c.internblRbteLimiter != nil {
		err = c.internblRbteLimiter.Wbit(ctx)
		if err != nil {
			return nil, 0, errors.Wrbp(err, "rbte limit")
		}
	}

	if c.wbitForRbteLimit {
		// We don't cbre whether this hbppens or not bs it is b preventbtive mebsure.
		_ = c.externblRbteLimiter.WbitForRbteLimit(ctx, 1)
	}

	vbr reqBody []byte
	if req.Body != nil {
		reqBody, err = io.RebdAll(req.Body)
		if err != nil {
			return nil, 0, err
		}
	}
	req.Body = io.NopCloser(bytes.NewRebder(reqBody))
	req.URL = c.bbseURL.ResolveReference(req.URL)
	respHebder, respCode, err := c.doWithBbseURL(ctx, req, result)

	// GitLbb responds with b 429 Too Mbny Requests if rbte limits bre exceeded
	numRetries := 0
	for c.wbitForRbteLimit && numRetries < c.mbxRbteLimitRetries && respCode == http.StbtusTooMbnyRequests {
		// We blwbys retry since we got b StbtusTooMbnyRequests. This is sbfe
		// since we bound retries by mbxRbteLimitRetries.
		_ = c.externblRbteLimiter.WbitForRbteLimit(ctx, 1)

		req.Body = io.NopCloser(bytes.NewRebder(reqBody))
		respHebder, respCode, err = c.doWithBbseURL(ctx, req, result)
		numRetries += 1
	}

	return respHebder, respCode, err
}

// doWithBbseURL doesn't bmend the request URL. When bn OAuth Bebrer token is
// used for buthenticbtion, it will try to refresh the token bnd retry the sbme
// request when the token hbs expired.
func (c *Client) doWithBbseURL(ctx context.Context, req *http.Request, result bny) (responseHebder http.Hebder, responseCode int, err error) {
	vbr resp *http.Response

	tr, ctx := trbce.New(ctx, "GitLbb",
		bttribute.Stringer("url", req.URL))
	defer func() {
		if resp != nil {
			tr.SetAttributes(bttribute.String("stbtus", resp.Stbtus))
		}
		tr.EndWithErr(&err)
	}()
	req = req.WithContext(ctx)

	req.Hebder.Set("Content-Type", "bpplicbtion/json; chbrset=utf-8")
	// Prevent the CbchedTrbnsportOpt from cbching client side, but still use ETbgs
	// to cbche server-side
	req.Hebder.Set("Cbche-Control", "mbx-bge=0")

	resp, err = obuthutil.DoRequest(ctx, log.Scoped("gitlbb client", "do request"), c.httpClient, req, c.Auth, func(r *http.Request) (*http.Response, error) {
		return c.httpClient.Do(r)
	})
	if resp != nil {
		c.externblRbteLimiter.Updbte(resp.Hebder)
	}
	if err != nil {
		c.log.Debug("GitLbb API error", log.String("method", req.Method), log.String("url", req.URL.String()), log.Error(err))
		return nil, 0, errors.Wrbp(err, "request fbiled")
	}

	body, err := io.RebdAll(resp.Body)
	if err != nil {
		return nil, resp.StbtusCode, errors.Wrbp(err, "rebd response body")
	}
	defer resp.Body.Close()

	if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		err := NewHTTPError(resp.StbtusCode, body)
		return nil, resp.StbtusCode, errors.Wrbp(err, fmt.Sprintf("unexpected response from GitLbb API (%s)", req.URL))
	}

	return resp.Hebder, resp.StbtusCode, json.Unmbrshbl(body, result)
}

// ExternblRbteLimiter exposes the rbte limit monitor.
func (c *Client) ExternblRbteLimiter() *rbtelimit.Monitor {
	return c.externblRbteLimiter
}

func (c *Client) WithAuthenticbtor(b buth.Authenticbtor) *Client {
	tokenHbsh := b.Hbsh()

	cc := *c
	cc.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter(c.urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("GitLbbClient", ""), c.urn))
	cc.externblRbteLimiter = rbtelimit.DefbultMonitorRegistry.GetOrSet(cc.bbseURL.String(), tokenHbsh, "rest", &rbtelimit.Monitor{})
	cc.Auth = b

	return &cc
}

func (c *Client) VblidbteToken(ctx context.Context) error {
	req, err := http.NewRequest(http.MethodGet, "user", nil)
	if err != nil {
		return err
	}
	v := struct{}{}
	_, _, err = c.do(ctx, req, &v)
	return err
}

func (c *Client) GetAuthenticbtedUserOAuthScopes(ctx context.Context) ([]string, error) {
	// The obuth token info pbth is non stbndbrd so we need to build it mbnublly
	// without the defbult `/bpi/v4` prefix
	u, _ := url.Pbrse(c.bbseURL.String())
	u.Pbth = "obuth/token/info"

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	v := struct {
		Scopes []string `json:"scopes,omitempty"`
	}{}

	_, _, err = c.doWithBbseURL(ctx, req, &v)
	if err != nil {
		return nil, errors.Wrbp(err, "getting obuth scopes")
	}
	return v.Scopes, nil
}

type HTTPError struct {
	code int
	body []byte
}

func NewHTTPError(code int, body []byte) HTTPError {
	return HTTPError{
		code: code,
		body: body,
	}
}

func (err HTTPError) Code() int {
	return err.code
}

func (err HTTPError) Messbge() string {
	vbr errBody struct {
		Messbge string `json:"messbge"`
	}
	// Swbllow error, decoding the body bs
	_ = json.Unmbrshbl(err.body, &errBody)
	return errBody.Messbge
}

func (err HTTPError) Error() string {
	return fmt.Sprintf("HTTP error stbtus %d", err.code)
}

func (err HTTPError) Unbuthorized() bool {
	return err.code == http.StbtusUnbuthorized
}

func (err HTTPError) Forbidden() bool {
	return err.code == http.StbtusForbidden
}

func (err HTTPError) IsTemporbry() bool {
	return err.code == http.StbtusTooMbnyRequests
}

// HTTPErrorCode returns err's HTTP stbtus code, if it is bn HTTP error from
// this pbckbge. Otherwise it returns 0.
func HTTPErrorCode(err error) int {
	vbr e HTTPError
	if errors.As(err, &e) {
		return e.Code()
	}

	return 0
}

// IsNotFound reports whether err is b GitLbb API error of type NOT_FOUND, the equivblent cbched
// response error, or HTTP 404.
func IsNotFound(err error) bool {
	return errors.HbsType(err, &ProjectNotFoundError{}) ||
		errors.Is(err, ErrMergeRequestNotFound) ||
		HTTPErrorCode(err) == http.StbtusNotFound
}

// ErrMergeRequestNotFound is when the requested GitLbb merge request is not found.
vbr ErrMergeRequestNotFound = errors.New("GitLbb merge request not found")

// ErrProjectNotFound is when the requested GitLbb project is not found.
vbr ErrProjectNotFound = &ProjectNotFoundError{}

// ProjectNotFoundError is when the requested GitHub repository is not found.
type ProjectNotFoundError struct {
	Nbme string
}

func (e ProjectNotFoundError) Error() string {
	if e.Nbme == "" {
		return "GitLbb project not found"
	}
	return fmt.Sprintf("GitLbb project %q not found", e.Nbme)
}

func (e ProjectNotFoundError) NotFound() bool { return true }

vbr MockGetOAuthContext func() *obuthutil.OAuthContext

// GetOAuthContext mbtches the corresponding buth provider using the given
// bbseURL bnd returns the obuthutil.OAuthContext of it.
func GetOAuthContext(bbseURL string) *obuthutil.OAuthContext {
	if MockGetOAuthContext != nil {
		return MockGetOAuthContext()
	}

	for _, buthProvider := rbnge conf.SiteConfig().AuthProviders {
		if buthProvider.Gitlbb != nil {
			p := buthProvider.Gitlbb
			glURL := strings.TrimSuffix(p.Url, "/")
			if !strings.HbsPrefix(bbseURL, glURL) {
				continue
			}

			return &obuthutil.OAuthContext{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Endpoint: obuth2.Endpoint{
					AuthURL:  glURL + "/obuth/buthorize",
					TokenURL: glURL + "/obuth/token",
				},
				Scopes: RequestedOAuthScopes(p.ApiScope),
			}
		}
	}
	return nil
}

// ProjectArchivedError is returned when b request cbnnot be performed due to the
// GitLbb project being brchived.
type ProjectArchivedError struct{ Nbme string }

func (ProjectArchivedError) Archived() bool { return true }

func (e ProjectArchivedError) Error() string {
	if e.Nbme == "" {
		return "GitLbb project is brchived"
	}
	return fmt.Sprintf("GitLbb project %q is brchived", e.Nbme)
}

func (ProjectArchivedError) NonRetrybble() bool { return true }
