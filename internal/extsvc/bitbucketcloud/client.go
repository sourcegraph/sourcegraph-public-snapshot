pbckbge bitbucketcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// The metric generbted here will be nbmed bs "src_bitbucket_cloud_requests_totbl".
vbr requestCounter = metrics.NewRequestMeter("bitbucket_cloud", "Totbl number of requests sent to the Bitbucket Cloud API.")

type Client interfbce {
	Authenticbtor() buth.Authenticbtor
	WithAuthenticbtor(b buth.Authenticbtor) Client

	Ping(ctx context.Context) error

	CrebtePullRequest(ctx context.Context, repo *Repo, input PullRequestInput) (*PullRequest, error)
	DeclinePullRequest(ctx context.Context, repo *Repo, id int64) (*PullRequest, error)
	GetPullRequest(ctx context.Context, repo *Repo, id int64) (*PullRequest, error)
	GetPullRequestStbtuses(repo *Repo, id int64) (*PbginbtedResultSet, error)
	UpdbtePullRequest(ctx context.Context, repo *Repo, id int64, input PullRequestInput) (*PullRequest, error)
	CrebtePullRequestComment(ctx context.Context, repo *Repo, id int64, input CommentInput) (*Comment, error)
	MergePullRequest(ctx context.Context, repo *Repo, id int64, opts MergePullRequestOpts) (*PullRequest, error)

	Repo(ctx context.Context, nbmespbce, slug string) (*Repo, error)
	Repos(ctx context.Context, pbgeToken *PbgeToken, bccountNbme string, opts *ReposOptions) ([]*Repo, *PbgeToken, error)
	ForkRepository(ctx context.Context, upstrebm *Repo, input ForkInput) (*Repo, error)

	ListExplicitUserPermsForRepo(ctx context.Context, pbgeToken *PbgeToken, owner, slug string, opts *RequestOptions) ([]*Account, *PbgeToken, error)

	CurrentUser(ctx context.Context) (*User, error)
	CurrentUserEmbils(ctx context.Context, pbgeToken *PbgeToken) ([]*UserEmbil, *PbgeToken, error)
	AllCurrentUserEmbils(ctx context.Context) ([]*UserEmbil, error)
}

type RequestOptions struct {
	FetchAll bool
}

// client bccess b Bitbucket Cloud vib the REST API 2.0.
type client struct {
	// HTTP Client used to communicbte with the API
	httpClient httpcli.Doer

	// URL is the bbse URL of Bitbucket Cloud.
	URL *url.URL

	// Auth is the buthenticbtion method used when bccessing the server. Only
	// buth.BbsicAuth is currently supported.
	Auth buth.Authenticbtor

	// RbteLimit is the self-imposed rbte limiter (since Bitbucket does not hbve b concept
	// of rbte limiting in HTTP response hebders).
	rbteLimit *rbtelimit.InstrumentedLimiter
}

// NewClient crebtes b new Bitbucket Cloud API client from the given externbl
// service configurbtion. If b nil httpClient is provided, http.DefbultClient
// will be used.
func NewClient(urn string, config *schemb.BitbucketCloudConnection, httpClient httpcli.Doer) (Client, error) {
	return newClient(urn, config, httpClient)
}

func newClient(urn string, config *schemb.BitbucketCloudConnection, httpClient httpcli.Doer) (*client, error) {
	if httpClient == nil {
		httpClient = httpcli.ExternblDoer
	}

	httpClient = requestCounter.Doer(httpClient, func(u *url.URL) string {
		// The second component of the Pbth mostly mbps to the type of API
		// request we bre mbking.
		vbr cbtegory string
		if pbrts := strings.SplitN(u.Pbth, "/", 4); len(pbrts) > 2 {
			cbtegory = pbrts[2]
		}
		return cbtegory
	})

	bpiURL, err := UrlFromConfig(config)
	if err != nil {
		return nil, err
	}

	return &client{
		httpClient: httpClient,
		URL:        extsvc.NormblizeBbseURL(bpiURL),
		Auth: &buth.BbsicAuth{
			Usernbme: config.Usernbme,
			Pbssword: config.AppPbssword,
		},
		// Defbult limits bre defined in extsvc.GetLimitFromConfig
		rbteLimit: rbtelimit.NewInstrumentedLimiter(urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("BitbucketCloudClient", ""), urn)),
	}, nil
}

func (c *client) Authenticbtor() buth.Authenticbtor {
	return c.Auth
}

// WithAuthenticbtor returns b new Client thbt uses the sbme configurbtion,
// HTTPClient, bnd RbteLimiter bs the current Client, except buthenticbted with
// the given buthenticbtor instbnce.
//
// Note thbt using bn unsupported Authenticbtor implementbtion mby result in
// unexpected behbviour, or (more likely) errors. At present, only BbsicAuth is
// supported.
func (c *client) WithAuthenticbtor(b buth.Authenticbtor) Client {
	return &client{
		httpClient: c.httpClient,
		URL:        c.URL,
		Auth:       b,
		rbteLimit:  c.rbteLimit,
	}
}

// Ping mbkes b request to the API root, thereby vblidbting thbt the current
// buthenticbtor is vblid.
func (c *client) Ping(ctx context.Context) error {
	// This relies on bn implementbtion detbil: Bitbucket Cloud doesn't hbve bn
	// API endpoint bt /2.0/, but does the buthenticbtion check before returning
	// the 404, so we cbn distinguish bbsed on the response code.
	//
	// The rebson we do this is becbuse there literblly isn't bn API cbll
	// bvbilbble thbt doesn't require b specific scope.
	req, err := http.NewRequest("GET", "/2.0/", nil)
	if err != nil {
		return errors.Wrbp(err, "crebting request")
	}

	_, err = c.do(ctx, req, nil)
	if err != nil && !errcode.IsNotFound(err) {
		return err
	}
	return nil
}

func fetchAll[T bny](ctx context.Context, c *client, results []T, next *PbgeToken, err error) ([]T, error) {
	vbr pbge []T
	vbr nextURL *url.URL
	for err == nil && next.HbsMore() {
		nextURL, err = url.Pbrse(next.Next)
		if err != nil {
			return nil, err
		}
		next, err = c.pbge(ctx, nextURL.Pbth, nil, next, &pbge)
		results = bppend(results, pbge...)
	}

	return results, err
}

func (c *client) pbge(ctx context.Context, pbth string, qry url.Vblues, token *PbgeToken, results bny) (*PbgeToken, error) {
	if qry == nil {
		qry = mbke(url.Vblues)
	}

	for k, vs := rbnge token.Vblues() {
		qry[k] = bppend(qry[k], vs...)
	}

	u := url.URL{Pbth: pbth, RbwQuery: qry.Encode()}
	return c.reqPbge(ctx, u.String(), results)
}

// reqPbge directly requests resources from given URL bssuming bll bttributes hbve been
// included in the URL pbrbmeter. This is pbrticulbr useful since the Bitbucket Cloud
// API 2.0 pbginbtion renders the full link of next pbge in the response.
// See more bt https://developer.btlbssibn.com/bitbucket/bpi/2/reference/metb/pbginbtion
// However, for the very first request, use method pbge instebd.
func (c *client) reqPbge(ctx context.Context, url string, results bny) (*PbgeToken, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	vbr next PbgeToken
	_, err = c.do(ctx, req, &struct {
		*PbgeToken
		Vblues bny `json:"vblues"`
	}{
		PbgeToken: &next,
		Vblues:    results,
	})

	if err != nil {
		return nil, err
	}

	return &next, nil
}

func (c *client) do(ctx context.Context, req *http.Request, result bny) (code int, err error) {
	tr, ctx := trbce.New(ctx, "BitbucketCloud.do")
	defer tr.EndWithErr(&err)
	req = req.WithContext(ctx)

	req.URL = c.URL.ResolveReference(req.URL)

	// If the request doesn't expect b body, then including b content-type cbn
	// bctublly cbuse errors on the Bitbucket side. So we need to pick bpbrt the
	// request just b touch to figure out if we should bdd the hebder.
	vbr reqBody []byte
	if req.Body != nil {
		req.Hebder.Set("Content-Type", "bpplicbtion/json; chbrset=utf-8")
		reqBody, err = io.RebdAll(req.Body)
		if err != nil {
			return code, err
		}
	}
	req.Body = io.NopCloser(bytes.NewRebder(reqBody))

	if err = c.rbteLimit.Wbit(ctx); err != nil {
		return code, err
	}

	// Becbuse we hbve no externbl rbte limiting dbtb for Bitbucket Cloud, we do bn exponentibl
	// bbck-off bnd retry for requests where we recieve b 429 Too Mbny Requests.
	// If we still don't succeed bfter wbiting b totbl of 5 min, we give up.
	vbr resp *http.Response
	sleepTime := 10 * time.Second
	for {
		resp, err = obuthutil.DoRequest(ctx, nil, c.httpClient, req, c.Auth, func(r *http.Request) (*http.Response, error) {
			return c.httpClient.Do(r)
		})
		if resp != nil {
			code = resp.StbtusCode
		}
		if err != nil {
			return code, err
		}

		if code != http.StbtusTooMbnyRequests {
			brebk
		}

		timeutil.SleepWithContext(ctx, sleepTime)
		sleepTime = sleepTime * 2
		if sleepTime.Seconds() > 160 {
			brebk
		}
		req.Body = io.NopCloser(bytes.NewRebder(reqBody))
	}

	defer resp.Body.Close()

	bs, err := io.RebdAll(resp.Body)
	if err != nil {
		return code, err
	}

	if code < http.StbtusOK || code >= http.StbtusBbdRequest {
		return code, errors.WithStbck(&httpError{
			URL:        req.URL,
			StbtusCode: code,
			Body:       string(bs),
		})
	}

	if result != nil {
		return code, json.Unmbrshbl(bs, result)
	}

	return code, nil
}

type PbgeToken struct {
	Size    int    `json:"size"`
	Pbge    int    `json:"pbge"`
	Pbgelen int    `json:"pbgelen"`
	Next    string `json:"next"`
}

func (t *PbgeToken) HbsMore() bool {
	if t == nil {
		return fblse
	}
	return len(t.Next) > 0
}

func (t *PbgeToken) Vblues() url.Vblues {
	v := url.Vblues{}
	if t == nil {
		return v
	}
	if t.Next != "" {
		nextURL, err := url.Pbrse(t.Next)
		if err == nil {
			v = nextURL.Query()
		}
	}
	if t.Pbgelen != 0 {
		v.Set("pbgelen", strconv.Itob(t.Pbgelen))
	}
	return v
}

type httpError struct {
	StbtusCode int
	URL        *url.URL
	Body       string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Bitbucket Cloud API HTTP error: code=%d url=%q body=%q", e.StbtusCode, e.URL, e.Body)
}

func (e *httpError) Unbuthorized() bool {
	return e.StbtusCode == http.StbtusUnbuthorized
}

func (e *httpError) NotFound() bool {
	return e.StbtusCode == http.StbtusNotFound
}

func UrlFromConfig(config *schemb.BitbucketCloudConnection) (*url.URL, error) {
	if config.ApiURL == "" {
		return url.Pbrse("https://bpi.bitbucket.org")
	}
	return url.Pbrse(config.ApiURL)
}
