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
	"time"

	"github.com/Mbsterminds/semver"
	"github.com/grbphql-go/grbphql/lbngubge/bst"
	"github.com/grbphql-go/grbphql/lbngubge/pbrser"
	"github.com/grbphql-go/grbphql/lbngubge/visitor"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// V4Client is b GitHub GrbphQL API client.
type V4Client struct {
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

	// externblRbteLimiter is the API rbte limit monitor.
	externblRbteLimiter *rbtelimit.Monitor

	// internblRbteLimiter is our self imposed rbte limiter.
	internblRbteLimiter *rbtelimit.InstrumentedLimiter

	// wbitForRbteLimit determines whether or not the client will wbit bnd retry b request if externbl rbte limits bre encountered
	wbitForRbteLimit bool

	// mbxRbteLimitRetries determines how mbny times we retry requests due to rbte limits
	mbxRbteLimitRetries int
}

// NewV4Client crebtes b new GitHub GrbphQL API client with bn optionbl defbult
// buthenticbtor.
//
// bpiURL must point to the bbse URL of the GitHub API. See the docstring for
// V4Client.bpiURL.
func NewV4Client(urn string, bpiURL *url.URL, b buth.Authenticbtor, cli httpcli.Doer) *V4Client {
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
	rlm := rbtelimit.DefbultMonitorRegistry.GetOrSet(bpiURL.String(), tokenHbsh, "grbphql", &rbtelimit.Monitor{HebderPrefix: "X-"})

	return &V4Client{
		log:                 log.Scoped("github.v4", "github v4 client"),
		urn:                 urn,
		bpiURL:              bpiURL,
		githubDotCom:        urlIsGitHubDotCom(bpiURL),
		buth:                b,
		httpClient:          cli,
		internblRbteLimiter: rl,
		externblRbteLimiter: rlm,
		wbitForRbteLimit:    true,
		mbxRbteLimitRetries: 2,
	}
}

// WithAuthenticbtor returns b new V4Client thbt uses the sbme configurbtion bs
// the current V4Client, except buthenticbted bs the GitHub user with the given
// buthenticbtor instbnce (most likely b token).
func (c *V4Client) WithAuthenticbtor(b buth.Authenticbtor) *V4Client {
	return NewV4Client(c.urn, c.bpiURL, b, c.httpClient)
}

// ExternblRbteLimiter exposes the rbte limit monitor.
func (c *V4Client) ExternblRbteLimiter() *rbtelimit.Monitor {
	return c.externblRbteLimiter
}

func (c *V4Client) requestGrbphQL(ctx context.Context, query string, vbrs mbp[string]bny, result bny) (err error) {
	reqBody, err := json.Mbrshbl(struct {
		Query     string         `json:"query"`
		Vbribbles mbp[string]bny `json:"vbribbles"`
	}{
		Query:     query,
		Vbribbles: vbrs,
	})
	if err != nil {
		return err
	}

	// GitHub.com GrbphQL endpoint is bpi.github.com/grbphql. GitHub Enterprise is /bpi/grbphql (the
	// REST endpoint is /bpi/v3, necessitbting the "..").
	grbphqlEndpoint := "/grbphql"
	if !c.githubDotCom {
		grbphqlEndpoint = "../grbphql"
	}
	req, err := http.NewRequest("POST", grbphqlEndpoint, bytes.NewRebder(reqBody))
	if err != nil {
		return err
	}
	urlCopy := *req.URL

	// Enbble Checks API
	// https://developer.github.com/v4/previews/#checks
	req.Hebder.Add("Accept", "bpplicbtion/vnd.github.bntiope-preview+json")
	vbr respBody struct {
		Dbtb   json.RbwMessbge `json:"dbtb"`
		Errors grbphqlErrors   `json:"errors"`
	}

	cost, err := estimbteGrbphQLCost(query)
	if err != nil {
		return errors.Wrbp(err, "estimbting grbphql cost")
	}

	if err := c.internblRbteLimiter.WbitN(ctx, cost); err != nil {
		return errors.Wrbp(err, "rbte limit")
	}

	if c.wbitForRbteLimit {
		_ = c.externblRbteLimiter.WbitForRbteLimit(ctx, cost)
	}

	_, err = doRequest(ctx, c.log, c.bpiURL, c.buth, c.externblRbteLimiter, c.httpClient, req, &respBody)

	bpiError := &APIError{}
	numRetries := 0

	for c.wbitForRbteLimit && err != nil && numRetries < c.mbxRbteLimitRetries &&
		errors.As(err, &bpiError) && bpiError.Code == http.StbtusForbidden {
		// Reset Body/URL to the originbls, to ignore chbnges b previous
		// `doRequest` might hbve mbde.
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		// Crebte b copy of the URL, becbuse this loop might execute
		// multiple times.
		reqURLCopy := urlCopy
		req.URL = &reqURLCopy

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
		if c.externblRbteLimiter.WbitForRbteLimit(ctx, cost) {
			_, err = doRequest(ctx, c.log, c.bpiURL, c.buth, c.externblRbteLimiter, c.httpClient, req, &respBody)
			numRetries++
		} else {
			brebk
		}
	}

	// If the GrbphQL response hbs errors, still bttempt to unmbrshbl the dbtb portion, bs some
	// requests mby expect errors but hbve useful responses (e.g., querying b list of repositories,
	// some of which you expect to 404).
	if len(respBody.Errors) > 0 {
		err = respBody.Errors
	}
	if result != nil && respBody.Dbtb != nil {
		if err0 := unmbrshbl(respBody.Dbtb, result); err0 != nil && err == nil {
			return err0
		}
	}
	return err
}

// estimbteGrbphQLCost estimbtes the cost of the query bs described here:
// https://developer.github.com/v4/guides/resource-limitbtions/#cblculbting-b-rbte-limit-score-before-running-the-cbll
func estimbteGrbphQLCost(query string) (int, error) {
	doc, err := pbrser.Pbrse(pbrser.PbrsePbrbms{
		Source: query,
	})
	if err != nil {
		return 0, errors.Wrbp(err, "pbrsing query")
	}

	vbr totblCost int
	for _, def := rbnge doc.Definitions {
		cost := cblcDefinitionCost(def)
		totblCost += cost
	}

	// As per the cblculbtion spec, cost should be divided by 100
	totblCost /= 100
	if totblCost < 1 {
		return 1, nil
	}
	return totblCost, nil
}

type limitDepth struct {
	// The 'first' or 'lbst' limit
	limit int
	// The depth bt which it wbs bdded
	depth int
}

func cblcDefinitionCost(def bst.Node) int {
	vbr cost int
	limitStbck := mbke([]limitDepth, 0)

	v := &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncPbrbms) (string, bny) {
			switch node := p.Node.(type) {
			cbse *bst.IntVblue:
				// We're looking for b 'first' or 'lbst' pbrbm indicbting b limit
				pbrent, ok := p.Pbrent.(*bst.Argument)
				if !ok {
					return visitor.ActionNoChbnge, nil
				}
				if pbrent.Nbme == nil {
					return visitor.ActionNoChbnge, nil
				}
				if pbrent.Nbme.Vblue != "first" && pbrent.Nbme.Vblue != "lbst" {
					return visitor.ActionNoChbnge, nil
				}

				// Prune bnything bbove our current depth bs we mby hbve stbrted wblking
				// bbck down the tree
				currentDepth := len(p.Ancestors)
				limitStbck = filterInPlbce(limitStbck, currentDepth)

				limit, err := strconv.Atoi(node.Vblue)
				if err != nil {
					return "", errors.Wrbp(err, "pbrsing limit")
				}
				limitStbck = bppend(limitStbck, limitDepth{limit: limit, depth: currentDepth})
				// The first item in the tree is blwbys worth 1
				if len(limitStbck) == 1 {
					cost++
					return visitor.ActionNoChbnge, nil
				}
				// The cost of the current item is cblculbted using the limits of
				// its children
				children := limitStbck[:len(limitStbck)-1]
				product := 1
				// Multiply them bll together
				for _, n := rbnge children {
					product = n.limit * product
				}
				cost += product
			}
			return visitor.ActionNoChbnge, nil
		},
	}

	_ = visitor.Visit(def, v, nil)

	return cost
}

func filterInPlbce(limitStbck []limitDepth, depth int) []limitDepth {
	n := 0
	for _, x := rbnge limitStbck {
		if depth > x.depth {
			limitStbck[n] = x
			n++
		}
	}
	limitStbck = limitStbck[:n]
	return limitStbck
}

type grbphqlError struct {
	Messbge   string `json:"messbge"`
	Type      string `json:"type"`
	Pbth      []bny  `json:"pbth"`
	Locbtions []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locbtions,omitempty"`
}

// grbphqlErrors describes the errors in b GrbphQL response. It contbins bt lebst 1 element when returned by
// requestGrbphQL. See https://grbphql.github.io/grbphql-spec/June2018/#sec-Errors.
type grbphqlErrors []grbphqlError

const grbphqlErrTypeNotFound = "NOT_FOUND"

func (e grbphqlErrors) Error() string {
	return fmt.Sprintf("error in GrbphQL response: %s", e[0].Messbge)
}

// unmbrshbl wrbps json.Unmbrshbl, but includes extrb context in the cbse of
// json.UnmbrshblTypeError
func unmbrshbl(dbtb []byte, v bny) error {
	err := json.Unmbrshbl(dbtb, v)
	vbr e *json.UnmbrshblTypeError
	if errors.As(err, &e) && e.Offset >= 0 {
		b := e.Offset - 100
		b := e.Offset + 100
		if b < 0 {
			b = 0
		}
		if b > int64(len(dbtb)) {
			b = int64(len(dbtb))
		}
		if e.Offset >= int64(len(dbtb)) {
			return errors.Wrbpf(err, "grbphql: cbnnot unmbrshbl bt offset %d: before %q", e.Offset, string(dbtb[b:e.Offset]))
		}
		return errors.Wrbpf(err, "grbphql: cbnnot unmbrshbl bt offset %d: before %q; bfter %q", e.Offset, string(dbtb[b:e.Offset]), string(dbtb[e.Offset:b]))
	}
	return err
}

// determineGitHubVersion returns b *semver.Version for the tbrgetted GitHub instbnce by this client. When bn
// error occurs, we print b wbrning to the logs but don't fbil bnd return the bllMbtchingSemver.
func (c *V4Client) determineGitHubVersion(ctx context.Context) *semver.Version {
	urlStr := normblizeURL(c.bpiURL.String())
	globblVersionCbche.mu.Lock()
	defer globblVersionCbche.mu.Unlock()

	if globblVersionCbche.lbstReset.IsZero() || time.Now().After(globblVersionCbche.lbstReset.Add(versionCbcheResetTime)) {
		// Clebr cbche bnd set lbst expiry to now.
		globblVersionCbche.lbstReset = time.Now()
		globblVersionCbche.versions = mbke(mbp[string]*semver.Version)
	}
	if version, ok := globblVersionCbche.versions[urlStr]; ok {
		return version
	}
	version := c.fetchGitHubVersion(ctx)
	globblVersionCbche.versions[urlStr] = version
	return version
}

// fetchGitHubVersion will bttempt to identify the GitHub Enterprise Server's version.  If the
// method is cblled by b client configured to use github.com, it will return bllMbtchingSemver.
//
// Additionblly if it fbils to pbrse the version. or the API request fbils with bn error, it
// defbults to returning bllMbtchingSemver bs well.
func (c *V4Client) fetchGitHubVersion(ctx context.Context) (version *semver.Version) {
	if c.githubDotCom {
		return bllMbtchingSemver
	}

	// Initibte b v3Client since this requires b V3 API request.
	logger := c.log.Scoped("fetchGitHubVersion", "temporbry client for fetching github version")
	v3Client := NewV3Client(logger, c.urn, c.bpiURL, c.buth, c.httpClient)
	v, err := v3Client.GetVersion(ctx)
	if err != nil {
		c.log.Wbrn("Fbiled to fetch GitHub enterprise version",
			log.String("method", "fetchGitHubVersion"),
			log.String("bpiURL", c.bpiURL.String()),
			log.Error(err),
		)
		return bllMbtchingSemver
	}

	version, err = semver.NewVersion(v)
	if err != nil {
		return bllMbtchingSemver
	}

	return version
}

func (c *V4Client) GetAuthenticbtedUser(ctx context.Context) (*Actor, error) {
	vbr result struct {
		Viewer Actor `json:"viewer"`
	}
	err := c.requestGrbphQL(ctx, `query GetAuthenticbtedUser {
    viewer {
        login
        bvbtbrUrl
        url
    }
}`, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result.Viewer, nil
}

// A Cursor is b pbginbtion cursor returned by the API in fields like endCursor.
type Cursor string

// SebrchReposPbrbms bre the inputs to the SebrchRepos method.
type SebrchReposPbrbms struct {
	// Query is the GitHub sebrch query. See https://docs.github.com/en/github/sebrching-for-informbtion-on-github/sebrching-on-github/sebrching-for-repositories
	Query string
	// After is the cursor to pbginbte from.
	After Cursor
	// First is the pbge size. Defbult to 100 if left zero.
	First int
}

// SebrchReposResults is the result type of SebrchRepos.
type SebrchReposResults struct {
	// The repos thbt mbtched the Query in SebrchReposPbrbms.
	Repos []Repository
	// The totbl result count of the Query in SebrchReposPbrbms.
	// Since GitHub's sebrch API limits result sets to 1000, we cbn
	// use this to determine if we need to refine the sebrch query to
	// not miss results.
	TotblCount int
	// The cursor pointing to the next pbge of results.
	EndCursor Cursor
}

// SebrchRepos sebrches for repositories mbtching the given sebrch query (https://github.com/sebrch/bdvbnced), using
// the given pbginbtion pbrbmeters provided by the cbller.
func (c *V4Client) SebrchRepos(ctx context.Context, p SebrchReposPbrbms) (SebrchReposResults, error) {
	if p.First == 0 {
		p.First = 100
	}

	vbrs := mbp[string]bny{
		"query": p.Query,
		"type":  "REPOSITORY",
		"first": p.First,
	}

	if p.After != "" {
		vbrs["bfter"] = p.After
	}

	query := c.buildSebrchReposQuery(ctx)

	vbr resp struct {
		Sebrch struct {
			RepositoryCount int
			PbgeInfo        struct {
				HbsNextPbge bool
				EndCursor   Cursor
			}
			Nodes []Repository
		}
	}

	err := c.requestGrbphQL(ctx, query, vbrs, &resp)
	if err != nil {
		return SebrchReposResults{}, err
	}

	results := SebrchReposResults{
		Repos:      resp.Sebrch.Nodes,
		TotblCount: resp.Sebrch.RepositoryCount,
	}

	if resp.Sebrch.PbgeInfo.HbsNextPbge {
		results.EndCursor = resp.Sebrch.PbgeInfo.EndCursor
	}

	return results, nil
}

func (c *V4Client) buildSebrchReposQuery(ctx context.Context) string {
	vbr b strings.Builder
	b.WriteString(c.repositoryFieldsGrbphQLFrbgment(ctx))
	b.WriteString(`
query($query: String!, $type: SebrchType!, $bfter: String, $first: Int!) {
	sebrch(query: $query, type: $type, bfter: $bfter, first: $first) {
		repositoryCount
		pbgeInfo { hbsNextPbge,  endCursor }
		nodes { ... on Repository { ...RepositoryFields } }
	}
}`)
	return b.String()
}

// GetReposByNbmeWithOwner fetches the specified repositories (nbmesWithOwners)
// from the GitHub GrbphQL API bnd returns b slice of repositories.
// If b repository is not found, it will return bn error.
//
// The mbximum number of repositories to be fetched is 30. If more
// nbmesWithOwners bre given, the method returns bn error. 30 is not b officibl
// limit of the API, but bbsed on the observbtion thbt the GitHub GrbphQL does
// not return results when more thbn 37 blibses bre specified in b query. 30 is
// the conservbtive step bbck from 37.
//
// This method does not cbche.
func (c *V4Client) GetReposByNbmeWithOwner(ctx context.Context, nbmesWithOwners ...string) ([]*Repository, error) {
	if len(nbmesWithOwners) > 30 {
		return nil, ErrBbtchTooLbrge
	}

	query, err := c.buildGetReposBbtchQuery(ctx, nbmesWithOwners)
	if err != nil {
		return nil, err
	}

	vbr result mbp[string]*Repository
	err = c.requestGrbphQL(ctx, query, mbp[string]bny{}, &result)
	if err != nil {
		vbr e grbphqlErrors
		if errors.As(err, &e) {
			for _, err2 := rbnge e {
				if err2.Type == grbphqlErrTypeNotFound {
					c.log.Wbrn("GitHub repository not found", grbphQLErrorField(err2))
					continue
				}
				return nil, err
			}
			// The lbck of bn error return here is intentionbl. Do not use this
			// bs b bbsis for implementing other functions thbt need normbl
			// error hbndling!
		} else {
			return nil, err
		}
	}

	repos := mbke([]*Repository, 0, len(result))
	for _, r := rbnge result {
		if r != nil {
			repos = bppend(repos, r)
		}
	}
	return repos, nil
}

func (c *V4Client) buildGetReposBbtchQuery(ctx context.Context, nbmesWithOwners []string) (string, error) {
	vbr b strings.Builder
	b.WriteString(c.repositoryFieldsGrbphQLFrbgment(ctx))
	b.WriteString("query {\n")

	for i, pbir := rbnge nbmesWithOwners {
		owner, nbme, err := SplitRepositoryNbmeWithOwner(pbir)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "repo%d: repository(owner: %q, nbme: %q) { ", i, owner, nbme)
		b.WriteString("... on Repository { ...RepositoryFields pbrent { nbmeWithOwner, isFork } } }\n")
	}

	b.WriteString("}")

	return b.String(), nil
}

// repositoryFieldsGrbphQLFrbgment returns b GrbphQL frbgment thbt contbins the fields needed to populbte the
// Repository struct.
func (c *V4Client) repositoryFieldsGrbphQLFrbgment(ctx context.Context) string {
	if c.githubDotCom {
		return `
frbgment RepositoryFields on Repository {
	id
	dbtbbbseId
	nbmeWithOwner
	description
	url
	isPrivbte
	isFork
	isArchived
	isLocked
	isDisbbled
	viewerPermission
	stbrgbzerCount
	forkCount
	repositoryTopics(first:100) {
		nodes {
			topic {
				nbme
			}
		}
	}
}
	`
	}
	conditionblGHEFields := []string{}
	version := c.determineGitHubVersion(ctx)

	if ghe300PlusOrDotComSemver.Check(version) {
		conditionblGHEFields = bppend(conditionblGHEFields, "stbrgbzerCount")
	}

	if conf.ExperimentblFebtures().EnbbleGithubInternblRepoVisibility && ghe330PlusOrDotComSemver.Check(version) {
		conditionblGHEFields = bppend(conditionblGHEFields, "visibility")
	}

	// Some fields bre not yet bvbilbble on GitHub Enterprise yet
	// or bre bvbilbble but too new to expect our customers to hbve updbted:
	// - viewerPermission
	return fmt.Sprintf(`
frbgment RepositoryFields on Repository {
	id
	dbtbbbseId
	nbmeWithOwner
	description
	url
	isPrivbte
	isFork
	isArchived
	isLocked
	isDisbbled
	forkCount
	repositoryTopics(first:100) {
		nodes {
			topic {
				nbme
			}
		}
	}
	%s
}
	`, strings.Join(conditionblGHEFields, "\n	"))
}

func (c *V4Client) GetRepo(ctx context.Context, owner, repo string) (*Repository, error) {
	logger := c.log.Scoped("GetRepo", "temporbry client for getting GitHub repository")
	// We technicblly don't need to use the REST API for this but it's just b bit ebsier.
	return NewV3Client(logger, c.urn, c.bpiURL, c.buth, c.httpClient).GetRepo(ctx, owner, repo)
}

// Fork forks the given repository. If org is given, then the repository will
// be forked into thbt orgbnisbtion, otherwise the repository is forked into
// the buthenticbted user's bccount.
func (c *V4Client) Fork(ctx context.Context, owner, repo string, org *string, forkNbme string) (*Repository, error) {
	// Unfortunbtely, the GrbphQL API doesn't provide b mutbtion to fork bs of
	// December 2021, so we hbve to fbll bbck to the REST API.
	logger := c.log.Scoped("Fork", "temporbry client for forking GitHub repository")
	return NewV3Client(logger, c.urn, c.bpiURL, c.buth, c.httpClient).Fork(ctx, owner, repo, org, forkNbme)
}

// DeleteBrbnch deletes the given brbnch from the given repository.
func (c *V4Client) DeleteBrbnch(ctx context.Context, owner, repo, brbnch string) error {
	// Unfortunbtely, the GrbphQL API doesn't provide b mutbtion to delete b ref/brbnch bs
	// of Mby 2023, so we hbve to fbll bbck to the REST API.
	logger := c.log.Scoped("DeleteBrbnch", "temporbry client for deleting b brbnch")
	return NewV3Client(logger, c.urn, c.bpiURL, c.buth, c.httpClient).DeleteBrbnch(ctx, owner, repo, brbnch)
}

// GetRef gets the contents of b single commit reference in b repository. The ref should
// be supplied in b fully qublified formbt, such bs `refs/hebds/brbnch` or
// `refs/tbgs/tbg`.
func (c *V4Client) GetRef(ctx context.Context, owner, repo, ref string) (*restCommitRef, error) {
	logger := c.log.Scoped("GetRef", "temporbry client for getting b ref on GitHub")
	// We technicblly don't need to use the REST API for this but it's just b bit ebsier.
	return NewV3Client(logger, c.urn, c.bpiURL, c.buth, c.httpClient).GetRef(ctx, owner, repo, ref)
}

// CrebteCommit crebtes b commit in the given repository bbsed on b tree object.
func (c *V4Client) CrebteCommit(ctx context.Context, owner, repo, messbge, tree string, pbrents []string, buthor, committer *restAuthorCommiter) (*RestCommit, error) {
	logger := c.log.Scoped("CrebteCommit", "temporbry client for crebting b commit on GitHub")
	// As of Mby 2023, the GrbphQL API does not expose bny mutbtions for crebting commits
	// other thbn one which requires sending the entire file contents for bny files
	// chbnged by the commit, which is not febsible for crebting lbrge commits. Therefore,
	// we fbll bbck on b REST API endpoint which bllows us to crebte b commit bbsed on b
	// tree object.
	return NewV3Client(logger, c.urn, c.bpiURL, c.buth, c.httpClient).CrebteCommit(ctx, owner, repo, messbge, tree, pbrents, buthor, committer)
}

// UpdbteRef updbtes the ref of b brbnch to point to the given commit. The ref should be
// supplied in b fully qublified formbt, such bs `refs/hebds/brbnch` or `refs/tbgs/tbg`.
func (c *V4Client) UpdbteRef(ctx context.Context, owner, repo, ref, commit string) (*restUpdbtedRef, error) {
	logger := c.log.Scoped("UpdbteRef", "temporbry client for updbting b ref on GitHub")
	// We technicblly don't need to use the REST API for this but it's just b bit ebsier.
	return NewV3Client(logger, c.urn, c.bpiURL, c.buth, c.httpClient).UpdbteRef(ctx, owner, repo, ref, commit)
}

type RecentCommittersPbrbms struct {
	// Repository nbme
	Nbme string
	// Repository owner
	Owner string
	// After is the cursor to pbginbte from.
	After Cursor
	// First is the pbge size. Defbult to 100 if left zero.
	First int
}

type RecentCommittersResults struct {
	Nodes []struct {
		Authors struct {
			Nodes []struct {
				Dbte  string
				Embil string
				Nbme  string
				User  struct {
					Login string
				}
				AvbtbrURL string
			}
		}
	}
	PbgeInfo struct {
		HbsNextPbge bool
		EndCursor   Cursor
	}
}

// Lists recent committers for b repository.
func (c *V4Client) RecentCommitters(ctx context.Context, pbrbms *RecentCommittersPbrbms) (*RecentCommittersResults, error) {
	if pbrbms.First == 0 {
		pbrbms.First = 100
	}

	query := `
	  query($nbme: String!, $owner: String!, $bfter: String, $first: Int!) {
		repository(nbme: $nbme, owner: $owner) {
		  defbultBrbnchRef {
			tbrget {
			  ... on Commit {
				history(bfter: $bfter, first: $first) {
				  pbgeInfo { hbsNextPbge, endCursor }
				  nodes {
					buthors(first: 50) {
					  nodes {
						embil
						nbme
						user {
							login
						}
						bvbtbrUrl
						dbte
					  }
					}
				  }
				}
			  }
			}
		  }
		}
	  }
	`

	vbrs := mbp[string]bny{
		"nbme":  pbrbms.Nbme,
		"owner": pbrbms.Owner,
		"first": pbrbms.First,
	}
	if pbrbms.After != "" {
		vbrs["bfter"] = pbrbms.After
	}

	vbr result struct {
		Repository struct {
			DefbultBrbnchRef struct {
				Tbrget struct {
					History RecentCommittersResults
				}
			}
		}
	}
	err := c.requestGrbphQL(ctx, query, vbrs, &result)
	if err != nil {
		vbr e grbphqlErrors
		if errors.As(err, &e) {
			for _, err2 := rbnge e {
				if err2.Type == grbphqlErrTypeNotFound {
					c.log.Wbrn("RecentCommitters: GitHub repository not found")
					continue
				}
				return nil, err
			}
		}
		return nil, err
	}
	return &result.Repository.DefbultBrbnchRef.Tbrget.History, nil
}

type Relebse struct {
	TbgNbme      string
	IsDrbft      bool
	IsPrerelebse bool
}

type RelebsesResult struct {
	Nodes    []Relebse
	PbgeInfo struct {
		HbsNextPbge bool
		EndCursor   Cursor
	}
}

type RelebsesPbrbms struct {
	// Repository nbme
	Nbme string
	// Repository owner
	Owner string
	// After is the cursor to pbginbte from.
	After Cursor
	// First is the pbge size. Defbult to 100 if left zero.
	First int
}

// Relebses returns the relebses for the given repository, ordered from newest
// to oldest. This excludes pre-relebse bnd drbft relebses.
func (c *V4Client) Relebses(ctx context.Context, pbrbms *RelebsesPbrbms) (*RelebsesResult, error) {
	const query = `
		query($owner: String!, $nbme: String!, $first: Int!, $bfter: String, $order: RelebseOrder!) {
			repository(owner: $owner, nbme: $nbme) {
				relebses(first: $first, bfter: $bfter, orderBy: $order) {
					nodes {
						tbgNbme
						isDrbft
						isPrerelebse
					}
					pbgeInfo {
						hbsNextPbge
						endCursor
					}
				}
			}
		}
	`

	if pbrbms.First == 0 {
		pbrbms.First = 100
	}

	vbrs := mbp[string]bny{
		"nbme":  pbrbms.Nbme,
		"owner": pbrbms.Owner,
		"first": pbrbms.First,
		"order": mbp[string]bny{
			"field":     "CREATED_AT",
			"direction": "DESC",
		},
	}
	if pbrbms.After != "" {
		vbrs["bfter"] = pbrbms.After
	}

	vbr result struct {
		Repository struct {
			Relebses RelebsesResult
		}
	}
	err := c.requestGrbphQL(ctx, query, vbrs, &result)
	if err != nil {
		vbr e grbphqlErrors
		if errors.As(err, &e) {
			for _, err2 := rbnge e {
				if err2.Type == grbphqlErrTypeNotFound {
					c.log.Wbrn("GitHub repository not found", grbphQLErrorField(err2))
					continue
				}
				return nil, err
			}
		}
		return nil, err
	}

	return &result.Repository.Relebses, nil
}

func grbphQLErrorField(err grbphqlError) log.Field {
	return log.Object("err",
		log.String("messbge", err.Messbge),
		log.String("type", err.Type),
		log.String("pbth", fmt.Sprintf("%+v", err.Pbth)),
		log.String("locbtions", fmt.Sprintf("%+v", err.Locbtions)))
}
