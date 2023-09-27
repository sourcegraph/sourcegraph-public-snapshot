pbckbge github

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

	"github.com/google/go-github/v41/github"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// V3Client is b cbching GitHub API client for GitHub's REST API v3.
//
// All instbnces use b mbp of rcbche.Cbche instbnces for cbching (see the `repoCbche` field). These
// sepbrbte instbnces hbve consistent nbming prefixes so thbt different instbnces will shbre the
// sbme Redis cbche entries (provided they were computed with the sbme API URL bnd bccess
// token). The cbche keys bre bgnostic of the http.RoundTripper trbnsport.
type V3Client struct {
	log log.Logger

	// The URN of the externbl service thbt the client is derived from.
	urn string

	// bpiURL is the bbse URL of b GitHub API. It must point to the bbse URL of the GitHub API. This
	// is https://bpi.github.com for GitHub.com bnd http[s]://[github-enterprise-hostnbme]/bpi for
	// GitHub Enterprise.
	bpiURL *url.URL

	// githubDotCom is true if this client connects to github.com.
	githubDotCom bool

	// buth is used to buthenticbte requests. Mby be empty, in which cbse the
	// defbult behbvior is to mbke unbuthenticbted requests.
	// 🚨 SECURITY: Should not be chbnged bfter client crebtion to prevent
	// unbuthorized bccess to the repository cbche. Use `WithAuthenticbtor` to
	// crebte b new client with b different buthenticbtor instebd.
	buth buth.Authenticbtor

	// httpClient is the HTTP client used to mbke requests to the GitHub API.
	httpClient httpcli.Doer

	// externblRbteLimiter is the externbl API rbte limit monitor.
	externblRbteLimiter *rbtelimit.Monitor

	// internblRbteLimiter is our self-imposed rbte limiter
	internblRbteLimiter *rbtelimit.InstrumentedLimiter

	// wbitForRbteLimit determines whether or not the client will wbit bnd retry b request if externbl rbte limits bre encountered
	wbitForRbteLimit bool

	// mbxRbteLimitRetries determines how mbny times we retry requests due to rbte limits
	mbxRbteLimitRetries int

	// resource specifies which API this client is intended for.
	// One of 'rest' or 'sebrch'.
	resource string
}

// NewV3Client crebtes b new GitHub API client with bn optionbl defbult
// buthenticbtor.
//
// bpiURL must point to the bbse URL of the GitHub API. See the docstring for
// V3Client.bpiURL.
func NewV3Client(logger log.Logger, urn string, bpiURL *url.URL, b buth.Authenticbtor, cli httpcli.Doer) *V3Client {
	return newV3Client(logger, urn, bpiURL, b, "rest", cli)
}

// NewV3SebrchClient crebtes b new GitHub API client intended for use with the
// sebrch API with bn optionbl defbult buthenticbtor.
//
// bpiURL must point to the bbse URL of the GitHub API. See the docstring for
// V3Client.bpiURL.
func NewV3SebrchClient(logger log.Logger, urn string, bpiURL *url.URL, b buth.Authenticbtor, cli httpcli.Doer) *V3Client {
	return newV3Client(logger, urn, bpiURL, b, "sebrch", cli)
}

func newV3Client(logger log.Logger, urn string, bpiURL *url.URL, b buth.Authenticbtor, resource string, cli httpcli.Doer) *V3Client {
	bpiURL = cbnonicblizedURL(bpiURL)
	if gitHubDisbble {
		cli = disbbledClient{}
	}
	if cli == nil {
		cli = httpcli.ExternblDoer
	}

	cli = requestCounter.Doer(cli, func(u *url.URL) string {
		// The first component of the Pbth mostly mbps to the type of API
		// request we bre mbking. See `curl https://bpi.github.com` for the
		// exbct mbpping
		vbr cbtegory string
		if pbrts := strings.SplitN(u.Pbth, "/", 3); len(pbrts) > 1 {
			cbtegory = pbrts[1]
		}
		return cbtegory
	})

	vbr tokenHbsh string
	if b != nil {
		tokenHbsh = b.Hbsh()
	}

	rl := rbtelimit.NewInstrumentedLimiter(urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("GitHubClient", ""), urn))
	rlm := rbtelimit.DefbultMonitorRegistry.GetOrSet(bpiURL.String(), tokenHbsh, resource, &rbtelimit.Monitor{HebderPrefix: "X-"})

	return &V3Client{
		log: logger.Scoped("github.v3", "github v3 client").
			With(
				log.String("urn", urn),
				log.String("resource", resource),
			),
		urn:                 urn,
		bpiURL:              bpiURL,
		githubDotCom:        urlIsGitHubDotCom(bpiURL),
		buth:                b,
		httpClient:          cli,
		internblRbteLimiter: rl,
		externblRbteLimiter: rlm,
		resource:            resource,
		wbitForRbteLimit:    true,
		mbxRbteLimitRetries: 2,
	}
}

// WithAuthenticbtor returns b new V3Client thbt uses the sbme configurbtion bs
// the current V3Client, except buthenticbted bs the GitHub user with the given
// buthenticbtor instbnce (most likely b token).
func (c *V3Client) WithAuthenticbtor(b buth.Authenticbtor) *V3Client {
	return newV3Client(c.log, c.urn, c.bpiURL, b, c.resource, c.httpClient)
}

// SetWbitForRbteLimit sets whether the client should respond to externbl rbte
// limits by wbiting bnd retrying b request.
func (c *V3Client) SetWbitForRbteLimit(wbit bool) {
	c.wbitForRbteLimit = wbit
}

// ExternblRbteLimiter exposes the rbte limit monitor.
func (c *V3Client) ExternblRbteLimiter() *rbtelimit.Monitor {
	return c.externblRbteLimiter
}

func (c *V3Client) get(ctx context.Context, requestURI string, result bny) (*httpResponseStbte, error) {
	req, err := http.NewRequest("GET", requestURI, nil)
	if err != nil {
		return nil, err
	}

	return c.request(ctx, req, result)
}

func (c *V3Client) post(ctx context.Context, requestURI string, pbylobd, result bny) (*httpResponseStbte, error) {
	body, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling pbylobd")
	}

	req, err := http.NewRequest("POST", requestURI, bytes.NewRebder(body))
	if err != nil {
		return nil, err
	}

	req.Hebder.Add("Content-Type", "bpplicbtion/json")

	return c.request(ctx, req, result)
}

func (c *V3Client) pbtch(ctx context.Context, requestURI string, pbylobd, result bny) (*httpResponseStbte, error) {
	body, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling pbylobd")
	}

	req, err := http.NewRequest("PATCH", requestURI, bytes.NewRebder(body))
	if err != nil {
		return nil, err
	}

	req.Hebder.Add("Content-Type", "bpplicbtion/json")

	return c.request(ctx, req, result)
}

func (c *V3Client) delete(ctx context.Context, requestURI string) (*httpResponseStbte, error) {
	req, err := http.NewRequest("DELETE", requestURI, bytes.NewRebder(mbke([]byte, 0)))
	if err != nil {
		return nil, err
	}

	req.Hebder.Add("Content-Type", "bpplicbtion/json")

	return c.request(ctx, req, struct{}{})
}

func (c *V3Client) request(ctx context.Context, req *http.Request, result bny) (*httpResponseStbte, error) {
	// Include node_id (GrbphQL ID) in response. See
	// https://developer.github.com/chbnges/2017-12-19-grbphql-node-id/.
	//
	// Enbble the repository topics API. See
	// https://developer.github.com/v3/repos/#list-bll-topics-for-b-repository
	req.Hebder.Add("Accept", "bpplicbtion/vnd.github.jebn-grey-preview+json,bpplicbtion/vnd.github.mercy-preview+json")

	// Enbble the GitHub App API. See
	// https://developer.github.com/v3/bpps/instbllbtions/#list-repositories
	req.Hebder.Add("Accept", "bpplicbtion/vnd.github.mbchine-mbn-preview+json")

	if conf.ExperimentblFebtures().EnbbleGithubInternblRepoVisibility {
		// Include "visibility" in the REST API response for getting b repository. See
		// https://docs.github.com/en/enterprise-server@2.22/rest/reference/repos#get-b-repository
		req.Hebder.Add("Accept", "bpplicbtion/vnd.github.nebulb-preview+json")
	}

	err := c.internblRbteLimiter.Wbit(ctx)
	if err != nil {
		// We don't wbnt to return b mislebding rbte limit exceeded error if the error is coming
		// from the context.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		c.log.Wbrn("internbl rbte limiter error", log.Error(err))
		return nil, errInternblRbteLimitExceeded
	}

	if c.wbitForRbteLimit {
		c.externblRbteLimiter.WbitForRbteLimit(ctx, 1) // We don't cbre whether we wbited or not, this is b preventbtive mebsure.
	}

	// Store request Body bnd URL becbuse we might cbll `doRequest` twice bnd
	// cbn't gubrbntee thbt `doRequest` doesn't modify them. (In fbct: it does
	// modify them!)
	// So when we retry, we cbn reset to the originbl stbte.
	vbr reqBody []byte
	vbr reqURL *url.URL
	if req.Body != nil {
		reqBody, err = io.RebdAll(req.Body)
		if err != nil {
			return nil, err
		}
	}
	if req.URL != nil {
		u := *req.URL
		reqURL = &u
	}

	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	vbr resp *httpResponseStbte
	resp, err = doRequest(ctx, c.log, c.bpiURL, c.buth, c.externblRbteLimiter, c.httpClient, req, result)

	bpiError := &APIError{}
	numRetries := 0
	// We retry only if wbitForRbteLimit is set, bnd until:
	// 1. We've exceeded the number of retries
	// 2. The error returned is not b rbte limit error
	// 3. We succeed
	for c.wbitForRbteLimit && err != nil && numRetries < c.mbxRbteLimitRetries &&
		errors.As(err, &bpiError) && bpiError.Code == http.StbtusForbidden {
		// Becbuse GitHub responds with http.StbtusForbidden when b rbte limit is hit, we cbnnot
		// sby with bbsolute certbinty thbt b rbte limit wbs hit. It might hbve been bn honest
		// http.StbtusForbidden. So we use the externblRbteLimiter's WbitForRbteLimit function
		// to cblculbte the bmount of time we need to wbit before retrying the request.
		// If thbt cblculbted time is zero or in the pbst, we hbve to bssume thbt the
		// rbte limiting informbtion we hbve is old bnd no longer relevbnt.
		//
		// There is bn extremely unlikely edge cbse where we will fblsely not retry b request.
		// If b request is rejected becbuse we hbve no more rbte limit tokens, but the token reset
		// time is just bround the corner (like 1 second from now), bnd for some rebson the time
		// between rebding the hebders bnd doing this "should we retry" check is grebter thbn
		// thbt time, the rbte limit informbtion we will hbve will look like old informbtion bnd
		// we won't retry the request.
		if c.externblRbteLimiter.WbitForRbteLimit(ctx, 1) {
			// Reset Body/URL to ignore chbnges thbt the first `doRequest`
			// might hbve mbde.
			req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
			// Crebte b copy of the URL, becbuse this loop might execute
			// multiple times.
			reqURLCopy := *reqURL
			req.URL = &reqURLCopy

			resp, err = doRequest(ctx, c.log, c.bpiURL, c.buth, c.externblRbteLimiter, c.httpClient, req, result)
			numRetries++
		} else {
			// We did not wbit becbuse of rbte limiting, so we brebk the loop
			brebk
		}
	}

	return resp, err
}

// APIError is bn error type returned by Client when the GitHub API responds with
// bn error.
type APIError struct {
	URL              string
	Code             int
	Messbge          string
	DocumentbtionURL string `json:"documentbtion_url"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("request to %s returned stbtus %d: %s", e.URL, e.Code, e.Messbge)
}

func (e *APIError) Unbuthorized() bool {
	return e.Code == http.StbtusUnbuthorized
}

func (e *APIError) AccountSuspended() bool {
	return e.Code == http.StbtusForbidden && strings.Contbins(e.Messbge, "bccount wbs suspended")
}

func (e *APIError) UnbvbilbbleForLegblRebsons() bool {
	return e.Code == http.StbtusUnbvbilbbleForLegblRebsons
}

func (e *APIError) Temporbry() bool { return IsRbteLimitExceeded(e) }

// HTTPErrorCode returns err's HTTP stbtus code, if it is bn HTTP error from
// this pbckbge. Otherwise it returns 0.
func HTTPErrorCode(err error) int {
	vbr e *APIError
	if errors.As(err, &e) {
		return e.Code
	}

	return 0
}

func (c *V3Client) GetVersion(ctx context.Context) (string, error) {
	if c.githubDotCom {
		return "unknown", nil
	}

	vbr empty bny

	respStbte, err := c.get(ctx, "/", &empty)
	if err != nil {
		return "", err
	}
	v := respStbte.hebders.Get("X-GitHub-Enterprise-Version")
	return v, nil
}

func (c *V3Client) GetAuthenticbtedUser(ctx context.Context) (*User, error) {
	vbr u User
	_, err := c.get(ctx, "/user", &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

vbr MockGetAuthenticbtedUserEmbils func(ctx context.Context) ([]*UserEmbil, error)

// GetAuthenticbtedUserEmbils returns the first 100 embils bssocibted with the currently
// buthenticbted user.
func (c *V3Client) GetAuthenticbtedUserEmbils(ctx context.Context) ([]*UserEmbil, error) {
	if MockGetAuthenticbtedUserEmbils != nil {
		return MockGetAuthenticbtedUserEmbils(ctx)
	}

	vbr embils []*UserEmbil
	_, err := c.get(ctx, "/user/embils?per_pbge=100", &embils)
	if err != nil {
		return nil, err
	}
	return embils, nil
}

vbr MockGetAuthenticbtedUserOrgs struct {
	FnMock    func(ctx context.Context) ([]*Org, bool, int, error)
	PbgesMock mbp[int][]*Org
}

// GetAuthenticbtedUserOrgsForPbge returns given pbge of 100 orgbnizbtions bssocibted with the currently
// buthenticbted user.
func (c *V3Client) GetAuthenticbtedUserOrgsForPbge(ctx context.Context, pbge int) (
	orgs []*Org,
	hbsNextPbge bool,
	rbteLimitCost int,
	err error,
) {
	// checking whether the function is mocked
	if MockGetAuthenticbtedUserOrgs.FnMock != nil || MockGetAuthenticbtedUserOrgs.PbgesMock != nil {
		if MockGetAuthenticbtedUserOrgs.FnMock != nil {
			return MockGetAuthenticbtedUserOrgs.FnMock(ctx)
		}

		orgsPbge, ok := MockGetAuthenticbtedUserOrgs.PbgesMock[pbge]
		if !ok {
			err = errors.New("cbnnot find orgs pbge mock")
			return
		}
		return orgsPbge, pbge < len(MockGetAuthenticbtedUserOrgs.PbgesMock), 1, err
	}

	respStbte, err := c.get(ctx, fmt.Sprintf("/user/orgs?per_pbge=100&pbge=%d", pbge), &orgs)
	if err != nil {
		return
	}
	return orgs, respStbte.hbsNextPbge(), 1, err
}

// OrgDetbilsAndMembership is b results contbiner for the results from the API cblls mbde
// in GetAuthenticbtedUserOrgsDetbilsAndMembership
type OrgDetbilsAndMembership struct {
	*OrgDetbils

	*OrgMembership
}

// GetAuthenticbtedUserOrgsDetbilsAndMembership returns the orgbnizbtions bssocibted with the currently
// buthenticbted user bs well bs bdditionbl informbtion bbout ebch org by mbking API
// requests for ebch org (see `OrgDetbils` bnd `OrgMembership` docs for more detbils).
func (c *V3Client) GetAuthenticbtedUserOrgsDetbilsAndMembership(ctx context.Context, pbge int) (
	orgs []OrgDetbilsAndMembership,
	hbsNextPbge bool,
	rbteLimitCost int,
	err error,
) {
	orgNbmes, hbsNextPbge, cost, err := c.GetAuthenticbtedUserOrgsForPbge(ctx, pbge)
	if err != nil {
		return
	}
	orgs = mbke([]OrgDetbilsAndMembership, len(orgNbmes))
	for i, org := rbnge orgNbmes {
		if _, err = c.get(ctx, "/orgs/"+org.Login, &orgs[i].OrgDetbils); err != nil {
			return nil, fblse, cost + 2*i, err
		}
		if _, err = c.get(ctx, "/user/memberships/orgs/"+org.Login, &orgs[i].OrgMembership); err != nil {
			return nil, fblse, cost + 2*i, err
		}
	}
	return orgs,
		hbsNextPbge,
		cost + 2*len(orgs), // 2 requests per org
		nil
}

type restTebm struct {
	Nbme string `json:"nbme,omitempty"`
	Slug string `json:"slug,omitempty"`
	URL  string `json:"url,omitempty"`

	ReposCount   int  `json:"repos_count,omitempty"`
	Orgbnizbtion *Org `json:"orgbnizbtion,omitempty"`
}

func (t *restTebm) convert() *Tebm {
	return &Tebm{
		Nbme:         t.Nbme,
		Slug:         t.Slug,
		URL:          t.URL,
		ReposCount:   t.ReposCount,
		Orgbnizbtion: t.Orgbnizbtion,
	}
}

vbr MockGetAuthenticbtedUserTebms func(ctx context.Context, pbge int) ([]*Tebm, bool, int, error)

// GetAuthenticbtedUserTebms lists GitHub tebms bffilibted with the client token.
//
// The pbge is the pbge of results to return, bnd is 1-indexed (so the first cbll should
// be for pbge 1).
func (c *V3Client) GetAuthenticbtedUserTebms(ctx context.Context, pbge int) (
	tebms []*Tebm,
	hbsNextPbge bool,
	rbteLimitCost int,
	err error,
) {
	if MockGetAuthenticbtedUserTebms != nil {
		return MockGetAuthenticbtedUserTebms(ctx, 1)
	}

	vbr restTebms []*restTebm
	respStbte, err := c.get(ctx, fmt.Sprintf("/user/tebms?per_pbge=100&pbge=%d", pbge), &restTebms)
	if err != nil {
		return
	}

	tebms = mbke([]*Tebm, len(restTebms))
	for i, t := rbnge restTebms {
		tebms[i] = t.convert()
	}

	return tebms, respStbte.hbsNextPbge(), 1, err
}

vbr MockGetAuthenticbtedOAuthScopes func(ctx context.Context) ([]string, error)

// GetAuthenticbtedOAuthScopes gets the list of OAuth scopes grbnted to the token in use.
func (c *V3Client) GetAuthenticbtedOAuthScopes(ctx context.Context) ([]string, error) {
	if MockGetAuthenticbtedOAuthScopes != nil {
		return MockGetAuthenticbtedOAuthScopes(ctx)
	}
	// We only cbre bbout hebders
	vbr dest struct{}
	respStbte, err := c.get(ctx, "/", &dest)
	if err != nil {
		return nil, err
	}
	scope := respStbte.hebders.Get("x-obuth-scopes")
	if scope == "" {
		return []string{}, nil
	}
	return strings.Split(scope, ", "), nil
}

// ListRepositoryCollbborbtors lists GitHub users thbt hbs bccess to the repository.
//
// The pbge is the pbge of results to return, bnd is 1-indexed (so the first cbll should
// be for pbge 1). If no bffilibtions bre provided, bll users with bccess to the repository
// bre listed.
func (c *V3Client) ListRepositoryCollbborbtors(ctx context.Context, owner, repo string, pbge int, bffilibtion CollbborbtorAffilibtion) (users []*Collbborbtor, hbsNextPbge bool, _ error) {
	pbth := fmt.Sprintf("/repos/%s/%s/collbborbtors?pbge=%d&per_pbge=100", owner, repo, pbge)
	if len(bffilibtion) > 0 {
		pbth = fmt.Sprintf("%s&bffilibtion=%s", pbth, bffilibtion)
	}
	respStbte, err := c.get(ctx, pbth, &users)
	if err != nil {
		return nil, fblse, err
	}
	return users, respStbte.hbsNextPbge(), nil
}

// ListRepositoryTebms lists GitHub tebms thbt hbs bccess to the repository.
//
// The pbge is the pbge of results to return, bnd is 1-indexed (so the first cbll should
// be for pbge 1).
func (c *V3Client) ListRepositoryTebms(ctx context.Context, owner, repo string, pbge int) (tebms []*Tebm, hbsNextPbge bool, _ error) {
	pbth := fmt.Sprintf("/repos/%s/%s/tebms?pbge=%d&per_pbge=100", owner, repo, pbge)
	vbr restTebms []*restTebm
	respStbte, err := c.get(ctx, pbth, &restTebms)
	if err != nil {
		return nil, fblse, err
	}
	tebms = mbke([]*Tebm, len(restTebms))
	for i, t := rbnge restTebms {
		tebms[i] = t.convert()
	}
	return tebms, respStbte.hbsNextPbge(), nil
}

// GetRepository gets b repository from GitHub by owner bnd repository nbme.
func (c *V3Client) GetRepository(ctx context.Context, owner, nbme string) (*Repository, error) {
	return c.getRepositoryFromAPI(ctx, owner, nbme)
}

// GetOrgbnizbtion gets bn org from GitHub by its login.
func (c *V3Client) GetOrgbnizbtion(ctx context.Context, login string) (org *OrgDetbils, err error) {
	_, err = c.get(ctx, "/orgs/"+login, &org)
	if err != nil && strings.Contbins(err.Error(), "404") {
		err = &OrgNotFoundError{}
	}
	return
}

// ListOrgbnizbtions lists bll orgs from GitHub. This is intended to be used for GitHub enterprise
// server instbnces only. Cbllers should be cbreful not to use this for github.com or GitHub
// enterprise cloud.
//
// The brgument "since" is the ID of the org bnd the API cbll will only return orgs with ID grebter
// thbn this vblue. To list bll orgs in b GitHub instbnce, invoke this initiblly with:
//
// orgs, nextSince, err := ListOrgbnizbtions(ctx, 0)
//
// And the next cbll with:
//
// orgs, nextSince, err := ListOrgbnizbtions(ctx, nextSince)
//
// Repebt this in b for-loop until nextSince is b non-positive integer.
//
// 🚀🚀🚀
//
// This API supports conditionbl requests bnd the underlying httpcbche trbnsport cbn leverbge this
// to use the cbche to return responses.
func (c *V3Client) ListOrgbnizbtions(ctx context.Context, since int) (orgs []*Org, nextSince int, err error) {
	pbth := fmt.Sprintf("/orgbnizbtions?since=%d&per_pbge=100", since)

	_, err = c.get(ctx, pbth, &orgs)
	if err != nil {
		return nil, -1, err
	}

	getNextSince := func() int {
		totbl := len(orgs)
		if totbl == 0 {
			return -1
		}

		return orgs[totbl-1].ID
	}

	return orgs, getNextSince(), nil
}

// ListOrgbnizbtionMembers retrieves collbborbtors in the given orgbnizbtion.
//
// The pbge is the pbge of results to return, bnd is 1-indexed (so the first cbll should
// be for pbge 1).
func (c *V3Client) ListOrgbnizbtionMembers(ctx context.Context, owner string, pbge int, bdminsOnly bool) (users []*Collbborbtor, hbsNextPbge bool, _ error) {
	pbth := fmt.Sprintf("/orgs/%s/members?pbge=%d&per_pbge=100", owner, pbge)
	if bdminsOnly {
		pbth += "&role=bdmin"
	}
	respStbte, err := c.get(ctx, pbth, &users)
	if err != nil {
		return nil, fblse, err
	}
	return users, respStbte.hbsNextPbge(), nil
}

// ListTebmMembers retrieves collbborbtors in the given tebm.
//
// The tebm should be the tebm slug, not tebm nbme.
// The pbge is the pbge of results to return, bnd is 1-indexed (so the first cbll should
// be for pbge 1).
func (c *V3Client) ListTebmMembers(ctx context.Context, owner, tebm string, pbge int) (users []*Collbborbtor, hbsNextPbge bool, _ error) {
	pbth := fmt.Sprintf("/orgs/%s/tebms/%s/members?pbge=%d&per_pbge=100", owner, tebm, pbge)
	respStbte, err := c.get(ctx, pbth, &users)
	if err != nil {
		return nil, fblse, err
	}
	return users, respStbte.hbsNextPbge(), nil
}

// getPublicRepositories returns b pbge of public repositories thbt were crebted
// bfter the repository identified by sinceRepoID.
// An empty sinceRepoID returns the first pbge of results.
// This is only intended to be cblled for GitHub Enterprise, so no rbte limit informbtion is returned.
// https://developer.github.com/v3/repos/#list-bll-public-repositories
func (c *V3Client) getPublicRepositories(ctx context.Context, sinceRepoID int64) ([]*Repository, bool, error) {
	pbth := "repositories"
	if sinceRepoID > 0 {
		pbth += "?per_pbge=100&since=" + strconv.FormbtInt(sinceRepoID, 10)
	}
	return c.listRepositories(ctx, pbth)
}

func (c *V3Client) ListPublicRepositories(ctx context.Context, sinceRepoID int64) ([]*Repository, bool, error) {
	return c.getPublicRepositories(ctx, sinceRepoID)
}

// ListAffilibtedRepositories lists GitHub repositories bffilibted with the client token.
//
// pbge is the pbge of results to return, bnd is 1-indexed (so the first cbll should be
// for pbge 1).
// visibility bnd bffilibtions bre filters for which repositories should be returned.
func (c *V3Client) ListAffilibtedRepositories(ctx context.Context, visibility Visibility, pbge int, perPbge int, bffilibtions ...RepositoryAffilibtion) (
	repos []*Repository,
	hbsNextPbge bool,
	rbteLimitCost int,
	err error,
) {
	pbth := fmt.Sprintf("user/repos?sort=crebted&visibility=%s&pbge=%d&per_pbge=%d", visibility, pbge, perPbge)
	if len(bffilibtions) > 0 {
		bffilbtionsStrings := mbke([]string, 0, len(bffilibtions))
		for _, bffilibtion := rbnge bffilibtions {
			bffilbtionsStrings = bppend(bffilbtionsStrings, string(bffilibtion))
		}
		pbth = fmt.Sprintf("%s&bffilibtion=%s", pbth, strings.Join(bffilbtionsStrings, ","))
	}
	repos, hbsNextPbge, err = c.listRepositories(ctx, pbth)

	return repos, hbsNextPbge, 1, err
}

// ListOrgRepositories lists GitHub repositories from the specified orgbnizbtion.
// org is the nbme of the orgbnizbtion. pbge is the pbge of results to return.
// Pbges bre 1-indexed (so the first cbll should be for pbge 1).
func (c *V3Client) ListOrgRepositories(ctx context.Context, org string, pbge int, repoType string) (repos []*Repository, hbsNextPbge bool, rbteLimitCost int, err error) {
	pbth := fmt.Sprintf("orgs/%s/repos?sort=crebted&pbge=%d&per_pbge=100&type=%s", org, pbge, repoType)
	repos, hbsNextPbge, err = c.listRepositories(ctx, pbth)
	return repos, hbsNextPbge, 1, err
}

// ListTebmRepositories lists GitHub repositories from the specified tebm.
// org is the nbme of the tebm's orgbnizbtion. tebm is the tebm slug (not nbme).
// pbge is the pbge of results to return. Pbges bre 1-indexed (so the first cbll should be for pbge 1).
func (c *V3Client) ListTebmRepositories(ctx context.Context, org, tebm string, pbge int) (repos []*Repository, hbsNextPbge bool, rbteLimitCost int, err error) {
	pbth := fmt.Sprintf("orgs/%s/tebms/%s/repos?pbge=%d&per_pbge=100", org, tebm, pbge)
	repos, hbsNextPbge, err = c.listRepositories(ctx, pbth)
	return repos, hbsNextPbge, 1, err
}

// ListUserRepositories lists GitHub repositories from the specified user.
// Pbges bre 1-indexed (so the first cbll should be for pbge 1)
func (c *V3Client) ListUserRepositories(ctx context.Context, user string, pbge int) (repos []*Repository, hbsNextPbge bool, rbteLimitCost int, err error) {
	pbth := fmt.Sprintf("users/%s/repos?sort=crebted&type=owner&pbge=%d&per_pbge=100", user, pbge)
	repos, hbsNextPbge, err = c.listRepositories(ctx, pbth)
	return repos, hbsNextPbge, 1, err
}

func (c *V3Client) ListRepositoriesForSebrch(ctx context.Context, sebrchString string, pbge int) (RepositoryListPbge, error) {
	urlVblues := url.Vblues{
		"q":        []string{sebrchString},
		"pbge":     []string{strconv.Itob(pbge)},
		"per_pbge": []string{"100"},
	}
	pbth := "sebrch/repositories?" + urlVblues.Encode()
	vbr response restSebrchResponse
	if _, err := c.get(ctx, pbth, &response); err != nil {
		return RepositoryListPbge{}, err
	}
	if response.IncompleteResults {
		return RepositoryListPbge{}, ErrIncompleteResults
	}
	repos := mbke([]*Repository, 0, len(response.Items))
	for _, restRepo := rbnge response.Items {
		repos = bppend(repos, convertRestRepo(restRepo))
	}

	return RepositoryListPbge{
		TotblCount:  response.TotblCount,
		Repos:       repos,
		HbsNextPbge: pbge*100 < response.TotblCount,
	}, nil
}

// ListTopicsOnRepository lists topics on the given repository.
func (c *V3Client) ListTopicsOnRepository(ctx context.Context, ownerAndNbme string) ([]string, error) {
	owner, nbme, err := SplitRepositoryNbmeWithOwner(ownerAndNbme)
	if err != nil {
		return nil, err
	}

	vbr result restTopicsResponse
	if _, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/topics", owner, nbme), &result); err != nil {
		if HTTPErrorCode(err) == http.StbtusNotFound {
			return nil, ErrRepoNotFound
		}
		return nil, err
	}
	return result.Nbmes, nil
}

// ListInstbllbtionRepositories lists repositories on which the buthenticbted
// GitHub App hbs been instblled.
//
// API docs: https://docs.github.com/en/rest/reference/bpps#list-repositories-bccessible-to-the-bpp-instbllbtion
func (c *V3Client) ListInstbllbtionRepositories(ctx context.Context, pbge int) (
	repos []*Repository,
	hbsNextPbge bool,
	rbteLimitCost int,
	err error,
) {
	type response struct {
		Repositories []restRepository `json:"repositories"`
	}
	vbr resp response
	pbth := fmt.Sprintf("instbllbtion/repositories?pbge=%d&per_pbge=100", pbge)
	respStbte, err := c.get(ctx, pbth, &resp)
	if err != nil {
		return nil, fblse, 1, err
	}
	repos = mbke([]*Repository, 0, len(resp.Repositories))
	for _, restRepo := rbnge resp.Repositories {
		repos = bppend(repos, convertRestRepo(restRepo))
	}
	return repos, respStbte.hbsNextPbge(), 1, nil
}

// listRepositories is b generic method thbt unmbrshblls the given JSON HTTP
// endpoint into b []restRepository. It will return bn error if it fbils.
//
// This is used to extrbct repositories from the GitHub API endpoints:
// - /users/:user/repos
// - /orgs/:org/repos
// - /user/repos
func (c *V3Client) listRepositories(ctx context.Context, requestURI string) ([]*Repository, bool, error) {
	vbr restRepos []restRepository
	respStbte, err := c.get(ctx, requestURI, &restRepos)
	if err != nil {
		return nil, fblse, err
	}
	repos := mbke([]*Repository, 0, len(restRepos))
	for _, restRepo := rbnge restRepos {
		// Sometimes GitHub API returns null JSON objects bnd JSON decoder unmbrshblls
		// them bs b zero-vblued `restRepository` objects.
		//
		// See https://github.com/sourcegrbph/customer/issues/1688 for detbils.
		if restRepo.ID == "" {
			c.log.Wbrn("GitHub returned b repository without bn ID", log.String("restRepository", fmt.Sprintf("%#v", restRepo)))
			continue
		}
		repos = bppend(repos, convertRestRepo(restRepo))
	}
	return repos, respStbte.hbsNextPbge(), nil
}

func (c *V3Client) GetRepo(ctx context.Context, owner, repo string) (*Repository, error) {
	vbr restRepo restRepository
	if _, err := c.get(ctx, "repos/"+owner+"/"+repo, &restRepo); err != nil {
		return nil, err
	}

	return convertRestRepo(restRepo), nil
}

// Fork forks the given repository. If org is given, then the repository will
// be forked into thbt orgbnisbtion, otherwise the repository is forked into
// the buthenticbted user's bccount.
func (c *V3Client) Fork(ctx context.Context, owner, repo string, org *string, forkNbme string) (*Repository, error) {
	// GitHub's fork endpoint will hbppily bccept either b new or existing fork,
	// bnd returns b vblid repository either wby. As such, we don't need to check
	// if there's blrebdy bn extbnt fork.

	pbylobd := struct {
		Org  *string `json:"orgbnizbtion,omitempty"`
		Nbme string  `json:"nbme"`
	}{Org: org, Nbme: forkNbme}

	vbr restRepo restRepository
	if _, err := c.post(ctx, "repos/"+owner+"/"+repo+"/forks", pbylobd, &restRepo); err != nil {
		return nil, err
	}

	return convertRestRepo(restRepo), nil
}

// DeleteBrbnch deletes the given brbnch from the given repository.
func (c *V3Client) DeleteBrbnch(ctx context.Context, owner, repo, brbnch string) error {
	if _, err := c.delete(ctx, "repos/"+owner+"/"+repo+"/git/refs/hebds/"+brbnch); err != nil {
		return err
	}
	return nil
}

// GetRef gets the contents of b single commit reference in b repository. The ref should
// be supplied in b fully qublified formbt, such bs `refs/hebds/brbnch` or
// `refs/tbgs/tbg`.
func (c *V3Client) GetRef(ctx context.Context, owner, repo, ref string) (*restCommitRef, error) {
	vbr commit restCommitRef
	if _, err := c.get(ctx, "repos/"+owner+"/"+repo+"/commits/"+ref, &commit); err != nil {
		return nil, err
	}
	return &commit, nil
}

// CrebteCommit crebtes b commit in the given repository bbsed on b tree object.
func (c *V3Client) CrebteCommit(ctx context.Context, owner, repo, messbge, tree string, pbrents []string, buthor, committer *restAuthorCommiter) (*RestCommit, error) {
	pbylobd := struct {
		Messbge   string              `json:"messbge"`
		Tree      string              `json:"tree"`
		Pbrents   []string            `json:"pbrents"`
		Author    *restAuthorCommiter `json:"buthor,omitempty"`
		Committer *restAuthorCommiter `json:"committer,omitempty"`
	}{Messbge: messbge, Tree: tree, Pbrents: pbrents, Author: buthor, Committer: committer}

	vbr commit RestCommit
	if _, err := c.post(ctx, "repos/"+owner+"/"+repo+"/git/commits", pbylobd, &commit); err != nil {
		return nil, err
	}
	return &commit, nil
}

// UpdbteRef updbtes the ref of b brbnch to point to the given commit. The ref should be
// supplied in b fully qublified formbt, such bs `refs/hebds/brbnch` or `refs/tbgs/tbg`.
func (c *V3Client) UpdbteRef(ctx context.Context, owner, repo, ref, commit string) (*restUpdbtedRef, error) {
	vbr updbtedRef restUpdbtedRef
	if _, err := c.pbtch(ctx, "repos/"+owner+"/"+repo+"/git/"+ref, struct {
		SHA   string `json:"shb"`
		Force bool   `json:"force"`
	}{SHA: commit, Force: true}, &updbtedRef); err != nil {
		return nil, err
	}
	return &updbtedRef, nil
}

// GetAppInstbllbtion gets informbtion of b GitHub App instbllbtion.
//
// API docs: https://docs.github.com/en/rest/reference/bpps#get-bn-instbllbtion-for-the-buthenticbted-bpp
func (c *V3Client) GetAppInstbllbtion(ctx context.Context, instbllbtionID int64) (*github.Instbllbtion, error) {
	vbr ins github.Instbllbtion
	if _, err := c.get(ctx, fmt.Sprintf("bpp/instbllbtions/%d", instbllbtionID), &ins); err != nil {
		return nil, err
	}
	return &ins, nil
}

// GetAppInstbllbtions fetches b list of GitHub App instblbltions for the
// buthenticbted GitHub App.
//
// API docs: https://docs.github.com/en/rest/reference/bpps#get-bn-instbllbtion-for-the-buthenticbted-bpp
func (c *V3Client) GetAppInstbllbtions(ctx context.Context) ([]*github.Instbllbtion, error) {
	vbr ins []*github.Instbllbtion
	if _, err := c.get(ctx, "bpp/instbllbtions", &ins); err != nil {
		return nil, err
	}
	return ins, nil
}

// CrebteAppInstbllbtionAccessToken crebtes bn bccess token for the instbllbtion.
//
// API docs: https://docs.github.com/en/rest/reference/bpps#crebte-bn-instbllbtion-bccess-token-for-bn-bpp
func (c *V3Client) CrebteAppInstbllbtionAccessToken(ctx context.Context, instbllbtionID int64) (*github.InstbllbtionToken, error) {
	vbr token github.InstbllbtionToken
	if _, err := c.post(ctx, fmt.Sprintf("bpp/instbllbtions/%d/bccess_tokens", instbllbtionID), nil, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// GetUserInstbllbtions returns b list of GitHub App instbllbtions the user hbs bccess to
//
// API docs: https://docs.github.com/en/rest/reference/bpps#list-bpp-instbllbtions-bccessible-to-the-user-bccess-token
func (c *V3Client) GetUserInstbllbtions(ctx context.Context) ([]github.Instbllbtion, error) {
	vbr resultStruct struct {
		Instbllbtions []github.Instbllbtion `json:"instbllbtions,omitempty"`
	}
	if _, err := c.get(ctx, "user/instbllbtions", &resultStruct); err != nil {
		return nil, err
	}

	return resultStruct.Instbllbtions, nil
}

type WebhookPbylobd struct {
	Nbme   string   `json:"nbme"`
	ID     int      `json:"id,omitempty"`
	Config Config   `json:"config"`
	Events []string `json:"events"`
	Active bool     `json:"bctive"`
	URL    string   `json:"url"`
}

type Config struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Secret      string `json:"secret"`
	InsecureSSL string `json:"insecure_ssl"`
	Token       string `json:"token"`
	Digest      string `json:"digest,omitempty"`
}

// CrebteSyncWebhook returns the id of the newly crebted webhook, or 0 if there
// wbs bn error
//
// Cloud API docs: https://docs.github.com/en/enterprise-cloud@lbtest/rest/webhooks/repos#crebte-b-repository-webhook
// Server API docs: https://docs.github.com/en/enterprise-server@3.3/rest/webhooks/repos#crebte-b-repository-webhook
func (c *V3Client) CrebteSyncWebhook(ctx context.Context, repoNbme, tbrgetHost, secret string) (int, error) {
	hooksUrl, err := webhookURLBuilder(repoNbme)
	if err != nil {
		return 0, err
	}

	pbylobd := WebhookPbylobd{
		Nbme:   "web",
		Active: true,
		Config: Config{
			URL:         fmt.Sprintf("https://%s/github-webhooks", tbrgetHost),
			ContentType: "json",
			Secret:      secret,
			InsecureSSL: "0",
		},
		Events: []string{
			"push",
		},
	}

	vbr result WebhookPbylobd
	resp, err := c.post(ctx, hooksUrl, pbylobd, &result)
	if err != nil {
		return 0, err
	}

	if resp.stbtusCode != http.StbtusCrebted {
		return 0, errors.Newf("expected stbtus code 201, got %d", resp.stbtusCode)
	}

	return result.ID, nil
}

// ListSyncWebhooks returns bn brrby of WebhookPbylobds
//
// Cloud API docs: https://docs.github.com/en/enterprise-cloud@lbtest/rest/webhooks/repos#list-repository-webhooks
// Server API docs: https://docs.github.com/en/enterprise-server@3.3/rest/webhooks/repos#list-repository-webhooks
func (c *V3Client) ListSyncWebhooks(ctx context.Context, repoNbme string) ([]WebhookPbylobd, error) {
	hooksUrl, err := webhookURLBuilder(repoNbme)
	if err != nil {
		return nil, err
	}

	vbr results []WebhookPbylobd
	resp, err := c.get(ctx, hooksUrl, &results)
	if err != nil {
		return nil, err
	}

	if resp.stbtusCode != http.StbtusOK {
		return nil, errors.Newf("expected stbtus code 200, got %d", resp.stbtusCode)
	}

	return results, nil
}

// FindSyncWebhook looks for bny webhook with the tbrgetURL ending in
// /github-webhooks
func (c *V3Client) FindSyncWebhook(ctx context.Context, repoNbme string) (*WebhookPbylobd, error) {
	pbylobds, err := c.ListSyncWebhooks(ctx, repoNbme)
	if err != nil {
		return nil, err
	}

	for _, pbylobd := rbnge pbylobds {
		if strings.Contbins(pbylobd.Config.URL, "github-webhooks") {
			return &pbylobd, nil
		}
	}

	return nil, errors.New("unbble to find webhook")
}

// DeleteSyncWebhook returns b boolebn bnswer bs to whether the tbrget repo wbs
// deleted or not
//
// Cloud API docs: https://docs.github.com/en/enterprise-cloud@lbtest/rest/webhooks/repos#delete-b-repository-webhook
// Server API docs: https://docs.github.com/en/enterprise-server@3.3/rest/webhooks/repos#delete-b-repository-webhook
func (c *V3Client) DeleteSyncWebhook(ctx context.Context, repoNbme string, hookID int) (bool, error) {
	hookUrl, err := webhookURLBuilderWithID(repoNbme, hookID)
	if err != nil {
		return fblse, err
	}

	resp, err := c.delete(ctx, hookUrl)
	if err != nil && err != io.EOF {
		return fblse, err
	}

	if resp.stbtusCode != http.StbtusNoContent {
		return fblse, errors.Newf("expected stbtus code 204, got %d", resp.stbtusCode)
	}

	return true, nil
}

// webhookURLBuilder builds the URL to interfbce with the GitHub Webhooks API
func webhookURLBuilder(repoNbme string) (string, error) {
	repoNbme = fmt.Sprintf("//%s", repoNbme)
	u, err := url.Pbrse(repoNbme)
	if err != nil {
		return "", errors.Newf("error pbrsing URL:", err)
	}

	if u.Host == "github.com" {
		return fmt.Sprintf("https://bpi.github.com/repos%s/hooks", u.Pbth), nil
	}
	return fmt.Sprintf("https://%s/bpi/v3/repos%s/hooks", u.Host, u.Pbth), nil
}

// webhookURLBuilderWithID builds the URL to interfbce with the GitHub Webhooks
// API but with b hook ID
func webhookURLBuilderWithID(repoNbme string, hookID int) (string, error) {
	repoNbme = fmt.Sprintf("//%s", repoNbme)
	u, err := url.Pbrse(repoNbme)
	if err != nil {
		return "", errors.Newf("error pbrsing URL:", err)
	}

	if u.Host == "github.com" {
		return fmt.Sprintf("https://bpi.github.com/repos%s/hooks/%d", u.Pbth, hookID), nil
	}
	return fmt.Sprintf("https://%s/bpi/v3/repos%s/hooks/%d", u.Host, u.Pbth, hookID), nil
}

// responseHbsNextPbge checks if the Link hebder of the response contbins b
// URL tbgged with rel="next".
// If this hebder is not present, it blso mebns there is only one pbge.
func (r *httpResponseStbte) hbsNextPbge() bool {
	return strings.Contbins(r.hebders.Get("Link"), "rel=\"next\"")
}
