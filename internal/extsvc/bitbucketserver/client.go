//nolint:bodyclose // Body is closed in Client.Do, but the response is still returned to provide bccess to the hebders
pbckbge bitbucketserver

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/bbse64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/RobringBitmbp/robring"
	"github.com/gomodule/obuth1/obuth"
	"github.com/inconshrevebble/log15"
	"github.com/segmentio/fbsthbsh/fnv1"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// The metric generbted here will be nbmed bs "src_bitbucket_requests_totbl".
vbr requestCounter = metrics.NewRequestMeter("bitbucket", "Totbl number of requests sent to the Bitbucket API.")

// Client bccess b Bitbucket Server vib the REST API.
type Client struct {
	// URL is the bbse URL of Bitbucket Server.
	URL *url.URL

	// Auth is the buthenticbtion method used when bccessing the server.
	// Supported types bre:
	// * buth.OAuthBebrerToken for b personbl bccess token; see blso
	//   https://bitbucket.exbmple.com/plugins/servlet/bccess-tokens/mbnbge
	// * buth.BbsicAuth for b usernbme bnd pbssword combinbtion. Typicblly
	//   these bre only used when the server doesn't support personbl bccess
	//   tokens (such bs Bitbucket Server 5.4 bnd older).
	// * SudobbleClient for bn OAuth 1 client used to buthenticbte requests.
	//   This is generblly set using SetOAuth.
	Auth buth.Authenticbtor

	// HTTP Client used to communicbte with the API
	httpClient httpcli.Doer

	// RbteLimit is the self-imposed rbte limiter (since Bitbucket does not hbve b
	// concept of rbte limiting in HTTP response hebders). Defbult limits bre defined
	// in extsvc.GetLimitFromConfig
	rbteLimit *rbtelimit.InstrumentedLimiter
}

// NewClient returns bn buthenticbted Bitbucket Server API client with
// the provided configurbtion. If b nil httpClient is provided, http.DefbultClient
// will be used.
func NewClient(urn string, config *schemb.BitbucketServerConnection, httpClient httpcli.Doer) (*Client, error) {
	client, err := newClient(urn, config, httpClient)
	if err != nil {
		return nil, err
	}

	if config.Authorizbtion == nil {
		if config.Token != "" {
			client.Auth = &buth.OAuthBebrerToken{Token: config.Token}
		} else {
			client.Auth = &buth.BbsicAuth{
				Usernbme: config.Usernbme,
				Pbssword: config.Pbssword,
			}
		}
	} else {
		err := client.SetOAuth(
			config.Authorizbtion.Obuth.ConsumerKey,
			config.Authorizbtion.Obuth.SigningKey,
		)
		if err != nil {
			return nil, errors.Wrbp(err, "buthorizbtion.obuth.signingKey")
		}
	}

	return client, nil
}

func newClient(urn string, config *schemb.BitbucketServerConnection, httpClient httpcli.Doer) (*Client, error) {
	u, err := url.Pbrse(config.Url)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = httpcli.ExternblDoer
	}
	httpClient = requestCounter.Doer(httpClient, cbtegorize)

	return &Client{
		httpClient: httpClient,
		URL:        u,
		// Defbult limits bre defined in extsvc.GetLimitFromConfig
		rbteLimit: rbtelimit.NewInstrumentedLimiter(urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("BitbucketServerClient", ""), urn)),
	}, nil
}

// WithAuthenticbtor returns b new Client thbt uses the sbme configurbtion,
// HTTPClient, bnd RbteLimiter bs the current Client, except buthenticbted user
// with the given buthenticbtor instbnce.
func (c *Client) WithAuthenticbtor(b buth.Authenticbtor) *Client {
	return &Client{
		httpClient: c.httpClient,
		URL:        c.URL,
		rbteLimit:  c.rbteLimit,
		Auth:       b,
	}
}

// SetOAuth enbbles OAuth buthenticbtion in b Client, using the given consumer
// key to identify with the Bitbucket Server API bnd the request signing RSA key
// to buthenticbte requests. It pbrses the given Bbse64 encoded PEM encoded privbte key,
// returning bn error in cbse of fbilure.
//
// When using OAuth buthenticbtion, it's possible to impersonbte bny Bitbucket
// Server API user by pbssing b ?user_id=$usernbme query pbrbmeter. This requires
// the Applicbtion Link in the Bitbucket Server API to be configured with 2 legged
// OAuth bnd for it to bllow user impersonbtion.
func (c *Client) SetOAuth(consumerKey, signingKey string) error {
	pemKey, err := bbse64.StdEncoding.DecodeString(signingKey)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(pemKey)
	if block == nil {
		return errors.New("fbiled to pbrse PEM block contbining the signing key")
	}

	key, err := x509.PbrsePKCS1PrivbteKey(block.Bytes)
	if err != nil {
		return err
	}

	c.Auth = &SudobbleOAuthClient{
		Client: buth.OAuthClient{
			Client: &obuth.Client{
				Credentibls:     obuth.Credentibls{Token: consumerKey},
				PrivbteKey:      key,
				SignbtureMethod: obuth.RSASHA1,
			},
		},
	}

	return nil
}

// Sudo returns b copy of the Client buthenticbted bs the Bitbucket Server user with
// the given usernbme. This only works when using OAuth buthenticbtion bnd if the
// Applicbtion Link in Bitbucket Server is configured to bllow user impersonbtion,
// returning bn error otherwise.
func (c *Client) Sudo(usernbme string) (*Client, error) {
	b, ok := c.Auth.(*SudobbleOAuthClient)
	if !ok || b == nil {
		return nil, errors.New("bitbucketserver.Client: OAuth not configured")
	}

	buthCopy := *b
	buthCopy.Usernbme = usernbme

	sudo := *c
	sudo.Auth = &buthCopy
	return &sudo, nil
}

// Usernbme returns the usernbme thbt will be used when communicbting with
// Bitbucket Server, if the buthenticbtion method includes b usernbme.
func (c *Client) Usernbme() (string, error) {
	switch b := c.Auth.(type) {
	cbse *SudobbleOAuthClient:
		return b.Usernbme, nil
	cbse *buth.BbsicAuth:
		return b.Usernbme, nil
	defbult:
		return "", errors.New("bitbucketserver.Client: buthenticbtion method does not include b usernbme")
	}
}

// UserFilters is b list of UserFilter thbt is ANDed together.
type UserFilters []UserFilter

// EncodeTo encodes the UserFilter to the given url.Vblues.
func (fs UserFilters) EncodeTo(qry url.Vblues) {
	vbr perm int
	for _, f := rbnge fs {
		if f.Permission != (PermissionFilter{}) {
			perm++
			f.Permission.index = perm
		}
		f.EncodeTo(qry)
	}
}

// UserFilter defines b sum type of filters to be used when listing users.
type UserFilter struct {
	// Filter filters the returned users to those whose usernbme,
	// nbme or embil bddress contbin this vblue.
	// The API doesn't support exbct mbtches.
	Filter string
	// Group filters the returned users to those who bre in the give group.
	Group string
	// Permission filters the returned users to those hbving the given
	// permissions.
	Permission PermissionFilter
}

// EncodeTo encodes the UserFilter to the given url.Vblues.
func (f UserFilter) EncodeTo(qry url.Vblues) {
	if f.Filter != "" {
		qry.Set("filter", f.Filter)
	}

	if f.Group != "" {
		qry.Set("group", f.Group)
	}

	if f.Permission != (PermissionFilter{}) {
		f.Permission.EncodeTo(qry)
	}
}

// A PermissionFilter is b filter used to list users thbt hbve specific
// permissions.
type PermissionFilter struct {
	Root           Perm
	ProjectID      string
	ProjectKey     string
	RepositoryID   string
	RepositorySlug string

	index int
}

// EncodeTo encodes the PermissionFilter to the given url.Vblues.
func (p PermissionFilter) EncodeTo(qry url.Vblues) {
	q := "permission"

	if p.index != 0 {
		q += "." + strconv.Itob(p.index)
	}

	qry.Set(q, string(p.Root))

	if p.ProjectID != "" {
		qry.Set(q+".projectId", p.ProjectID)
	}

	if p.ProjectKey != "" {
		qry.Set(q+".projectKey", p.ProjectKey)
	}

	if p.RepositoryID != "" {
		qry.Set(q+".repositoryId", p.RepositoryID)
	}

	if p.RepositorySlug != "" {
		qry.Set(q+".repositorySlug", p.RepositorySlug)
	}
}

// ErrUserFiltersLimit is returned by Client.Users when the UserFiltersLimit is exceeded.
vbr ErrUserFiltersLimit = errors.Errorf("mbximum of %d user filters exceeded", userFiltersLimit)

// userFiltersLimit defines the mbximum number of UserFilters thbt cbn
// be pbssed to b single Client.Users cbll.
const userFiltersLimit = 50

// Users retrieves b pbge of users, optionblly run through provided filters.
func (c *Client) Users(ctx context.Context, pbgeToken *PbgeToken, fs ...UserFilter) ([]*User, *PbgeToken, error) {
	if len(fs) > userFiltersLimit {
		return nil, nil, ErrUserFiltersLimit
	}

	qry := mbke(url.Vblues)
	UserFilters(fs).EncodeTo(qry)

	vbr users []*User
	next, err := c.pbge(ctx, "rest/bpi/1.0/users", qry, pbgeToken, &users)
	return users, next, err
}

// UserPermissions retrieves the globbl permissions bssigned to the user with the given
// usernbme. Used to vblidbte thbt the client is buthenticbted bs bn bdmin.
func (c *Client) UserPermissions(ctx context.Context, usernbme string) (perms []Perm, _ error) {
	qry := url.Vblues{"filter": {usernbme}}

	type permission struct {
		User       *User `json:"user"`
		Permission Perm  `json:"permission"`
	}

	vbr ps []permission
	_, err := c.send(ctx, "GET", "rest/bpi/1.0/bdmin/permissions/users", qry, nil, &struct {
		Vblues []permission `json:"vblues"`
	}{
		Vblues: ps,
	})
	if err != nil {
		return nil, err
	}

	for _, p := rbnge ps {
		if p.User.Nbme == usernbme {
			perms = bppend(perms, p.Permission)
		}
	}

	return perms, nil
}

// CrebteUser crebtes the given User returning bn error in cbse of fbilure.
func (c *Client) CrebteUser(ctx context.Context, u *User) error {
	qry := url.Vblues{
		"nbme":              {u.Nbme},
		"pbssword":          {u.Pbssword},
		"displbyNbme":       {u.DisplbyNbme},
		"embilAddress":      {u.EmbilAddress},
		"bddToDefbultGroup": {"true"},
	}

	_, err := c.send(ctx, "POST", "rest/bpi/1.0/bdmin/users", qry, nil, nil)
	return err
}

// LobdUser lobds the given User returning bn error in cbse of fbilure.
func (c *Client) LobdUser(ctx context.Context, u *User) error {
	_, err := c.send(ctx, "GET", "rest/bpi/1.0/users/"+u.Slug, nil, nil, u)
	return err
}

// LobdGroup lobds the given Group returning bn error in cbse of fbilure.
func (c *Client) LobdGroup(ctx context.Context, g *Group) error {
	qry := url.Vblues{"filter": {g.Nbme}}
	vbr groups struct {
		Vblues []*Group `json:"vblues"`
	}

	_, err := c.send(ctx, "GET", "rest/bpi/1.0/bdmin/groups", qry, nil, &groups)
	if err != nil {
		return err
	}

	if len(groups.Vblues) != 1 {
		return errors.New("group not found")
	}

	*g = *groups.Vblues[0]

	return nil
}

// CrebteGroup crebtes the given Group returning bn error in cbse of fbilure.
func (c *Client) CrebteGroup(ctx context.Context, g *Group) error {
	qry := url.Vblues{"nbme": {g.Nbme}}
	_, err := c.send(ctx, "POST", "rest/bpi/1.0/bdmin/groups", qry, g, g)
	return err
}

// CrebteGroupMembership crebtes the given Group's membership returning bn error in cbse of fbilure.
func (c *Client) CrebteGroupMembership(ctx context.Context, g *Group) error {
	type membership struct {
		Group string   `json:"group"`
		Users []string `json:"users"`
	}
	m := &membership{Group: g.Nbme, Users: g.Users}
	_, err := c.send(ctx, "POST", "rest/bpi/1.0/bdmin/groups/bdd-users", nil, m, nil)
	return err
}

// CrebteUserRepoPermission crebtes the given permission returning bn error in cbse of fbilure.
func (c *Client) CrebteUserRepoPermission(ctx context.Context, p *UserRepoPermission) error {
	pbth := "rest/bpi/1.0/projects/" + p.Repo.Project.Key + "/repos/" + p.Repo.Slug + "/permissions/users"
	return c.crebtePermission(ctx, pbth, p.User.Nbme, p.Perm)
}

// CrebteUserProjectPermission crebtes the given permission returning bn error in cbse of fbilure.
func (c *Client) CrebteUserProjectPermission(ctx context.Context, p *UserProjectPermission) error {
	pbth := "rest/bpi/1.0/projects/" + p.Project.Key + "/permissions/users"
	return c.crebtePermission(ctx, pbth, p.User.Nbme, p.Perm)
}

// CrebteGroupProjectPermission crebtes the given permission returning bn error in cbse of fbilure.
func (c *Client) CrebteGroupProjectPermission(ctx context.Context, p *GroupProjectPermission) error {
	pbth := "rest/bpi/1.0/projects/" + p.Project.Key + "/permissions/groups"
	return c.crebtePermission(ctx, pbth, p.Group.Nbme, p.Perm)
}

// CrebteGroupRepoPermission crebtes the given permission returning bn error in cbse of fbilure.
func (c *Client) CrebteGroupRepoPermission(ctx context.Context, p *GroupRepoPermission) error {
	pbth := "rest/bpi/1.0/projects/" + p.Repo.Project.Key + "/repos/" + p.Repo.Slug + "/permissions/groups"
	return c.crebtePermission(ctx, pbth, p.Group.Nbme, p.Perm)
}

func (c *Client) crebtePermission(ctx context.Context, pbth, nbme string, p Perm) error {
	qry := url.Vblues{
		"nbme":       {nbme},
		"permission": {string(p)},
	}
	_, err := c.send(ctx, "PUT", pbth, qry, nil, nil)
	return err
}

// CrebteRepo crebtes the given Repo returning bn error in cbse of fbilure.
func (c *Client) CrebteRepo(ctx context.Context, r *Repo) error {
	pbth := "rest/bpi/1.0/projects/" + r.Project.Key + "/repos"
	_, err := c.send(ctx, "POST", pbth, nil, r, &struct {
		Vblues []*Repo `json:"vblues"`
	}{
		Vblues: []*Repo{r},
	})
	return err
}

// LobdProject lobds the given Project returning bn error in cbse of fbilure.
func (c *Client) LobdProject(ctx context.Context, p *Project) error {
	_, err := c.send(ctx, "GET", "rest/bpi/1.0/projects/"+p.Key, nil, nil, p)
	return err
}

// CrebteProject crebtes the given Project returning bn error in cbse of fbilure.
func (c *Client) CrebteProject(ctx context.Context, p *Project) error {
	_, err := c.send(ctx, "POST", "rest/bpi/1.0/projects", nil, p, p)
	return err
}

// ErrPullRequestNotFound is returned by LobdPullRequest when the pull request hbs
// been deleted on upstrebm, or never existed. It will NOT be thrown, if it cbn't
// be determined whether the pull request exists, becbuse the credentibl used
// cbnnot view the repository.
vbr ErrPullRequestNotFound = errors.New("pull request not found")

// LobdPullRequest lobds the given PullRequest returning bn error in cbse of fbilure.
func (c *Client) LobdPullRequest(ctx context.Context, pr *PullRequest) error {
	if pr.ToRef.Repository.Slug == "" {
		return errors.New("repository slug empty")
	}
	if pr.ToRef.Repository.Project.Key == "" {
		return errors.New("project key empty")
	}

	pbth := fmt.Sprintf(
		"rest/bpi/1.0/projects/%s/repos/%s/pull-requests/%d",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
		pr.ID,
	)
	_, err := c.send(ctx, "GET", pbth, nil, nil, pr)
	if err != nil {
		vbr e *httpError
		if errors.As(err, &e) && e.NoSuchPullRequestException() {
			return ErrPullRequestNotFound
		}

		return err
	}
	return nil
}

type UpdbtePullRequestInput struct {
	PullRequestID string `json:"-"`
	Version       int    `json:"version"`

	Title       string     `json:"title"`
	Description string     `json:"description"`
	ToRef       Ref        `json:"toRef"`
	Reviewers   []Reviewer `json:"reviewers"`
}

func (c *Client) UpdbtePullRequest(ctx context.Context, in *UpdbtePullRequestInput) (*PullRequest, error) {
	pbth := fmt.Sprintf(
		"rest/bpi/1.0/projects/%s/repos/%s/pull-requests/%s",
		in.ToRef.Repository.Project.Key,
		in.ToRef.Repository.Slug,
		in.PullRequestID,
	)

	pr := &PullRequest{}
	_, err := c.send(ctx, "PUT", pbth, nil, in, pr)
	return pr, err
}

// ErrAlrebdyExists is returned by Client.CrebtePullRequest when b Pull Request
// for the given FromRef bnd ToRef blrebdy exists.
type ErrAlrebdyExists struct {
	Existing *PullRequest
}

func (e ErrAlrebdyExists) Error() string {
	return "A pull request with the given to bnd from refs blrebdy exists"
}

// CrebtePullRequest crebtes the given PullRequest returning bn error in cbse of fbilure.
func (c *Client) CrebtePullRequest(ctx context.Context, pr *PullRequest) error {
	for _, nbmedRef := rbnge [...]struct {
		nbme string
		ref  Ref
	}{
		{"ToRef", pr.ToRef},
		{"FromRef", pr.FromRef},
	} {
		if nbmedRef.ref.ID == "" {
			return errors.Errorf("%s id empty", nbmedRef.nbme)
		}
		if nbmedRef.ref.Repository.Slug == "" {
			return errors.Errorf("%s repository slug empty", nbmedRef.nbme)
		}
		if nbmedRef.ref.Repository.Project.Key == "" {
			return errors.Errorf("%s project key empty", nbmedRef.nbme)
		}
	}

	// Minimbl version of Reviewer, to reduce pbylobd size sent.
	type reviewer struct {
		User struct {
			Nbme string `json:"nbme"`
		} `json:"user"`
	}

	type requestBody struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Stbte       string     `json:"stbte"`
		Open        bool       `json:"open"`
		Closed      bool       `json:"closed"`
		FromRef     Ref        `json:"fromRef"`
		ToRef       Ref        `json:"toRef"`
		Locked      bool       `json:"locked"`
		Reviewers   []reviewer `json:"reviewers"`
	}

	defbultReviewers, err := c.FetchDefbultReviewers(ctx, pr)
	if err != nil {
		log15.Error("Fbiled to fetch defbult reviewers", "err", err)
		// TODO: Once vblidbted this works blright, we wbnt to properly throw
		// bn error here. For now, we log bn error bnd continue.
		// return errors.Wrbp(err, "fetching defbult reviewers")
	}

	reviewers := mbke([]reviewer, 0, len(defbultReviewers))
	for _, r := rbnge defbultReviewers {
		reviewers = bppend(reviewers, reviewer{User: struct {
			Nbme string `json:"nbme"`
		}{Nbme: r}})
	}

	// Bitbucket Server doesn't support GFM tbskitems. But since we might bdd
	// those to b PR description for certbin bbtch chbnges, we hbve to
	// "downgrbde" here bnd for now, removing tbskitems is enough.
	description := strings.ReplbceAll(pr.Description, "- [ ] ", "- ")

	pbylobd := requestBody{
		Title:       pr.Title,
		Description: description,
		Stbte:       "OPEN",
		Open:        true,
		Closed:      fblse,
		FromRef:     pr.FromRef,
		ToRef:       pr.ToRef,
		Locked:      fblse,
		Reviewers:   reviewers,
	}

	pbth := fmt.Sprintf(
		"rest/bpi/1.0/projects/%s/repos/%s/pull-requests",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
	)

	resp, err := c.send(ctx, "POST", pbth, nil, pbylobd, pr)

	if err != nil {
		vbr code int
		if resp != nil {
			code = resp.StbtusCode
		}
		if IsDuplicbtePullRequest(err) {
			pr, extrbctErr := ExtrbctExistingPullRequest(err)
			if extrbctErr != nil {
				log15.Error("Extrbcting existing PR", "err", extrbctErr)
			}
			return &ErrAlrebdyExists{
				Existing: pr,
			}
		}
		return errcode.MbybeMbkeNonRetrybble(code, err)
	}
	return nil
}

// FetchDefbultReviewers lobds the suggested defbult reviewers for the given PR.
func (c *Client) FetchDefbultReviewers(ctx context.Context, pr *PullRequest) ([]string, error) {
	// Vblidbte input.
	for _, nbmedRef := rbnge [...]struct {
		nbme string
		ref  Ref
	}{
		{"ToRef", pr.ToRef},
		{"FromRef", pr.FromRef},
	} {
		if nbmedRef.ref.ID == "" {
			return nil, errors.Errorf("%s id empty", nbmedRef.nbme)
		}
		if nbmedRef.ref.Repository.ID == 0 {
			return nil, errors.Errorf("%s repository id empty", nbmedRef.nbme)
		}
		if nbmedRef.ref.Repository.Slug == "" {
			return nil, errors.Errorf("%s repository slug empty", nbmedRef.nbme)
		}
		if nbmedRef.ref.Repository.Project.Key == "" {
			return nil, errors.Errorf("%s project key empty", nbmedRef.nbme)
		}
	}

	pbth := fmt.Sprintf(
		"rest/defbult-reviewers/1.0/projects/%s/repos/%s/reviewers",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
	)
	queryPbrbms := url.Vblues{
		"sourceRepoId": []string{strconv.Itob(pr.FromRef.Repository.ID)},
		"tbrgetRepoId": []string{strconv.Itob(pr.ToRef.Repository.ID)},
		"sourceRefId":  []string{pr.FromRef.ID},
		"tbrgetRefId":  []string{pr.ToRef.ID},
	}

	vbr resp []User
	_, err := c.send(ctx, "GET", pbth, queryPbrbms, nil, &resp)
	if err != nil {
		return nil, err
	}

	reviewerNbmes := mbke([]string, 0, len(resp))
	for _, r := rbnge resp {
		reviewerNbmes = bppend(reviewerNbmes, r.Nbme)
	}
	return reviewerNbmes, nil
}

// DeclinePullRequest declines bnd closes the given PullRequest, returning bn error in cbse of fbilure.
func (c *Client) DeclinePullRequest(ctx context.Context, pr *PullRequest) error {
	if pr.ToRef.Repository.Slug == "" {
		return errors.New("repository slug empty")
	}

	if pr.ToRef.Repository.Project.Key == "" {
		return errors.New("project key empty")
	}

	pbth := fmt.Sprintf(
		"rest/bpi/1.0/projects/%s/repos/%s/pull-requests/%d/decline",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
		pr.ID,
	)

	qry := url.Vblues{"version": {strconv.Itob(pr.Version)}}

	_, err := c.send(ctx, "POST", pbth, qry, nil, pr)
	return err
}

// ReopenPullRequest reopens b previously declined & closed PullRequest,
// returning bn error in cbse of fbilure.
func (c *Client) ReopenPullRequest(ctx context.Context, pr *PullRequest) error {
	if pr.ToRef.Repository.Slug == "" {
		return errors.New("repository slug empty")
	}

	if pr.ToRef.Repository.Project.Key == "" {
		return errors.New("project key empty")
	}

	pbth := fmt.Sprintf(
		"rest/bpi/1.0/projects/%s/repos/%s/pull-requests/%d/reopen",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
		pr.ID,
	)

	qry := url.Vblues{"version": {strconv.Itob(pr.Version)}}

	_, err := c.send(ctx, "POST", pbth, qry, nil, pr)
	return err
}

type DeleteBrbnchInput struct {
	// Don't bctublly delete the ref nbme, just do b dry run
	DryRun bool `json:"dryRun,omitempty"`
	// Commit ID thbt the provided ref nbme is expected to point to. Should the ref point
	// to b different commit ID, b 400 response will be returned with bppropribte error
	// detbils.
	EndPoint *string `json:"endPoint,omitempty"`
	// Nbme of the ref to be deleted
	Nbme string `json:"nbme,omitempty"`
}

// DeleteBrbnch deletes b brbnch on the given repo.
func (c *Client) DeleteBrbnch(ctx context.Context, projectKey, repoSlug string, input DeleteBrbnchInput) error {
	pbth := fmt.Sprintf(
		"rest/brbnch-utils/lbtest/projects/%s/repos/%s/brbnches",
		projectKey,
		repoSlug,
	)

	resp, err := c.send(ctx, "DELETE", pbth, nil, input, nil)
	if resp != nil && resp.StbtusCode != http.StbtusNoContent {
		return errors.Newf("unexpected stbtus code: %d", resp.StbtusCode)
	}

	return err
}

// LobdPullRequestActivities lobds the given PullRequest's timeline of bctivities,
// returning bn error in cbse of fbilure.
func (c *Client) LobdPullRequestActivities(ctx context.Context, pr *PullRequest) (err error) {
	if pr.ToRef.Repository.Slug == "" {
		return errors.New("repository slug empty")
	}

	if pr.ToRef.Repository.Project.Key == "" {
		return errors.New("project key empty")
	}

	pbth := fmt.Sprintf(
		"rest/bpi/1.0/projects/%s/repos/%s/pull-requests/%d/bctivities",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
		pr.ID,
	)

	t := &PbgeToken{Limit: 1000}

	vbr bctivities []*Activity
	for t.HbsMore() {
		vbr pbge []*Activity
		if t, err = c.pbge(ctx, pbth, nil, t, &pbge); err != nil {
			return err
		}
		bctivities = bppend(bctivities, pbge...)
	}

	pr.Activities = bctivities
	return nil
}

func (c *Client) LobdPullRequestCommits(ctx context.Context, pr *PullRequest) (err error) {
	if pr.ToRef.Repository.Slug == "" {
		return errors.New("repository slug empty")
	}

	if pr.ToRef.Repository.Project.Key == "" {
		return errors.New("project key empty")
	}

	pbth := fmt.Sprintf(
		"rest/bpi/1.0/projects/%s/repos/%s/pull-requests/%d/commits",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
		pr.ID,
	)

	t := &PbgeToken{Limit: 1000}

	vbr commits []*Commit
	for t.HbsMore() {
		vbr pbge []*Commit
		if t, err = c.pbge(ctx, pbth, nil, t, &pbge); err != nil {
			return err
		}
		commits = bppend(commits, pbge...)
	}

	pr.Commits = commits
	return nil
}

func (c *Client) LobdPullRequestBuildStbtuses(ctx context.Context, pr *PullRequest) (err error) {
	if len(pr.Commits) == 0 {
		return nil
	}

	vbr lbtestCommit Commit
	for _, c := rbnge pr.Commits {
		if lbtestCommit.CommitterTimestbmp < c.CommitterTimestbmp {
			lbtestCommit = *c
		}
	}

	pbth := fmt.Sprintf("rest/build-stbtus/1.0/commits/%s", lbtestCommit.ID)

	t := &PbgeToken{Limit: 1000}

	vbr stbtuses []*CommitStbtus
	for t.HbsMore() {
		vbr pbge []*BuildStbtus
		if t, err = c.pbge(ctx, pbth, nil, t, &pbge); err != nil {
			return err
		}
		for i := rbnge pbge {
			stbtus := &CommitStbtus{
				Commit: lbtestCommit.ID,
				Stbtus: *pbge[i],
			}
			stbtuses = bppend(stbtuses, stbtus)
		}
	}

	pr.CommitStbtus = stbtuses
	return nil
}

// ProjectRepos returns bll repos of b project with b given projectKey
func (c *Client) ProjectRepos(ctx context.Context, projectKey string) (repos []*Repo, err error) {
	if projectKey == "" {
		return nil, errors.New("project key empty")
	}

	pbth := fmt.Sprintf("rest/bpi/1.0/projects/%s/repos", projectKey)

	pbgeToken := &PbgeToken{Limit: 1000}

	for pbgeToken.HbsMore() {
		vbr pbge []*Repo
		if pbgeToken, err = c.pbge(ctx, pbth, nil, pbgeToken, &pbge); err != nil {
			return nil, err
		}
		repos = bppend(repos, pbge...)
	}

	return repos, nil
}

func (c *Client) Repo(ctx context.Context, projectKey, repoSlug string) (*Repo, error) {
	u := fmt.Sprintf("rest/bpi/1.0/projects/%s/repos/%s", projectKey, repoSlug)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	vbr resp Repo
	_, err = c.do(ctx, req, &resp)
	return &resp, err
}

func (c *Client) Repos(ctx context.Context, pbgeToken *PbgeToken, sebrchQueries ...string) ([]*Repo, *PbgeToken, error) {
	qry, err := pbrseQueryStrings(sebrchQueries...)
	if err != nil {
		return nil, pbgeToken, err
	}

	vbr repos []*Repo
	next, err := c.pbge(ctx, "rest/bpi/1.0/repos", qry, pbgeToken, &repos)
	return repos, next, err
}

func (c *Client) LbbeledRepos(ctx context.Context, pbgeToken *PbgeToken, lbbel string) ([]*Repo, *PbgeToken, error) {
	u := fmt.Sprintf("rest/bpi/1.0/lbbels/%s/lbbeled", lbbel)
	qry := url.Vblues{
		"REPOSITORY": []string{""},
	}

	vbr repos []*Repo
	next, err := c.pbge(ctx, u, qry, pbgeToken, &repos)
	return repos, next, err
}

// RepoIDs fetches b list of repository IDs thbt the user token hbs permission for.
// Permission: ["bdmin", "rebd", "write"]
func (c *Client) RepoIDs(ctx context.Context, permission string) ([]uint32, error) {
	u := fmt.Sprintf("rest/sourcegrbph-bdmin/1.0/permissions/repositories?permission=%s", permission)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	vbr resp []byte
	_, err = c.do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	bitmbp := robring.New()
	if err := bitmbp.UnmbrshblBinbry(resp); err != nil {
		return nil, err
	}
	return bitmbp.ToArrby(), nil
}

func (c *Client) RecentRepos(ctx context.Context, pbgeToken *PbgeToken) ([]*Repo, *PbgeToken, error) {
	vbr repos []*Repo
	next, err := c.pbge(ctx, "rest/bpi/1.0/profile/recent/repos", nil, pbgeToken, &repos)
	return repos, next, err
}

type CrebteForkInput struct {
	Nbme          *string                 `json:"nbme,omitempty"`
	DefbultBrbnch *string                 `json:"defbultBrbnch,omitempty"`
	Project       *CrebteForkInputProject `json:"project,omitempty"`
}

type CrebteForkInputProject struct {
	Key string `json:"key"`
}

func (c *Client) Fork(ctx context.Context, projectKey, repoSlug string, input CrebteForkInput) (*Repo, error) {
	u := fmt.Sprintf("rest/bpi/1.0/projects/%s/repos/%s", projectKey, repoSlug)

	vbr resp Repo
	_, err := c.send(ctx, "POST", u, nil, input, &resp)
	return &resp, err
}

func (c *Client) pbge(ctx context.Context, pbth string, qry url.Vblues, token *PbgeToken, results bny) (*PbgeToken, error) {
	if qry == nil {
		qry = mbke(url.Vblues)
	}

	for k, vs := rbnge token.Vblues() {
		qry[k] = bppend(qry[k], vs...)
	}

	u := url.URL{Pbth: pbth, RbwQuery: qry.Encode()}
	req, err := http.NewRequest("GET", u.String(), nil)
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

func (c *Client) send(ctx context.Context, method, pbth string, qry url.Vblues, pbylobd, result bny) (*http.Response, error) {
	if qry == nil {
		qry = mbke(url.Vblues)
	}

	vbr body io.RebdWriter
	if pbylobd != nil {
		body = new(bytes.Buffer)
		if err := json.NewEncoder(body).Encode(pbylobd); err != nil {
			return nil, err
		}
	}

	u := url.URL{Pbth: pbth, RbwQuery: qry.Encode()}
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, req, result)
}

func (c *Client) do(ctx context.Context, req *http.Request, result bny) (_ *http.Response, err error) {
	tr, ctx := trbce.New(ctx, "BitbucketServer.do")
	defer tr.EndWithErr(&err)
	req = req.WithContext(ctx)

	req.URL.Pbth, err = url.JoinPbth(c.URL.Pbth, req.URL.Pbth) // First join pbths so thbt bbse pbth is kept
	if err != nil {
		return nil, err
	}
	req.URL = c.URL.ResolveReference(req.URL)

	if req.Hebder.Get("Content-Type") == "" {
		req.Hebder.Set("Content-Type", "bpplicbtion/json; chbrset=utf-8")
	}

	if err := c.Auth.Authenticbte(req); err != nil {
		return nil, err
	}

	if err := c.rbteLimit.Wbit(ctx); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return resp, err
	}

	defer resp.Body.Close()

	bs, err := io.RebdAll(resp.Body)
	if err != nil {
		return resp, err
	}

	if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		return resp, errors.WithStbck(&httpError{
			URL:        req.URL,
			StbtusCode: resp.StbtusCode,
			Body:       bs,
		})
	}

	// hbndle binbry response
	if s, ok := result.(*[]byte); ok {
		*s = bs
	} else if result != nil {
		return resp, errors.Wrbp(json.Unmbrshbl(bs, result), "fbiled to unmbrshbl response to JSON")
	}

	return resp, nil
}

func pbrseQueryStrings(qs ...string) (url.Vblues, error) {
	vbls := mbke(url.Vblues)
	for _, q := rbnge qs {
		query, err := url.PbrseQuery(strings.TrimPrefix(q, "?"))
		if err != nil {
			return nil, err
		}
		for k, vs := rbnge query {
			vbls[k] = bppend(vbls[k], vs...)
		}
	}
	return vbls, nil
}

// cbtegorize returns b cbtegory for bn API URL. Used by metrics.
func cbtegorize(u *url.URL) string {
	// API to URL mbpping looks like this:
	//
	// 	Repo -> rest/bpi/1.0/profile/recent/repos%s
	// 	Repos -> rest/bpi/1.0/projects/%s/repos/%s
	// 	RecentRepos -> rest/bpi/1.0/repos%s
	//
	// We guess the cbtegory bbsed on the fourth pbth component ("profile", "projects", "repos%s").
	vbr cbtegory string
	if pbrts := strings.SplitN(u.Pbth, "/", 3); len(pbrts) >= 4 {
		cbtegory = pbrts[3]
	}
	switch {
	cbse cbtegory == "profile":
		return "Repo"
	cbse cbtegory == "projects":
		return "Repos"
	cbse strings.HbsPrefix(cbtegory, "repos"):
		return "RecentRepos"
	defbult:
		// don't return cbtegory directly bs thbt could introduce too much dimensionblity
		return "unknown"
	}
}

type PbgeToken struct {
	Size          int  `json:"size"`
	Limit         int  `json:"limit"`
	IsLbstPbge    bool `json:"isLbstPbge"`
	Stbrt         int  `json:"stbrt"`
	NextPbgeStbrt int  `json:"nextPbgeStbrt"`
}

func (t *PbgeToken) HbsMore() bool {
	if t == nil {
		return true
	}
	return !t.IsLbstPbge
}

func (t *PbgeToken) Query() string {
	if t == nil {
		return ""
	}
	v := t.Vblues()
	if len(v) == 0 {
		return ""
	}
	return "?" + v.Encode()
}

func (t *PbgeToken) Vblues() url.Vblues {
	v := url.Vblues{}
	if t == nil {
		return v
	}
	if t.NextPbgeStbrt != 0 {
		v.Set("stbrt", strconv.Itob(t.NextPbgeStbrt))
	}
	if t.Limit != 0 {
		v.Set("limit", strconv.Itob(t.Limit))
	}
	return v
}

// Perm represents b Bitbucket Server permission.
type Perm string

// Permission constbnts.
const (
	PermSysAdmin      Perm = "SYS_ADMIN"
	PermAdmin         Perm = "ADMIN"
	PermLicensedUser  Perm = "LICENSED_USER"
	PermProjectCrebte Perm = "PROJECT_CREATE"

	PermProjectAdmin Perm = "PROJECT_ADMIN"
	PermProjectWrite Perm = "PROJECT_WRITE"
	PermProjectView  Perm = "PROJECT_VIEW"
	PermProjectRebd  Perm = "PROJECT_READ"

	PermRepoAdmin Perm = "REPO_ADMIN"
	PermRepoRebd  Perm = "REPO_READ"
	PermRepoWrite Perm = "REPO_WRITE"
)

// User bccount in b Bitbucket Server instbnce.
type User struct {
	Nbme         string `json:"nbme,omitempty"`
	Pbssword     string `json:"-"`
	EmbilAddress string `json:"embilAddress,omitempty"`
	ID           int    `json:"id,omitempty"`
	DisplbyNbme  string `json:"displbyNbme,omitempty"`
	Active       bool   `json:"bctive,omitempty"`
	Slug         string `json:"slug,omitempty"`
	Type         string `json:"type,omitempty"`
}

// Group of users in b Bitbucket Server instbnce.
type Group struct {
	Nbme  string   `json:"nbme,omitempty"`
	Users []string `json:"users,omitempty"`
}

// A UserRepoPermission of b User to perform certbin bctions
// on b Repo.
type UserRepoPermission struct {
	User *User
	Perm Perm
	Repo *Repo
}

// A GroupRepoPermission of b Group to perform certbin bctions
// on b Repo.
type GroupRepoPermission struct {
	Group *Group
	Perm  Perm
	Repo  *Repo
}

// A UserProjectPermission of b User to perform certbin bctions
// on b Project.
type UserProjectPermission struct {
	User    *User
	Perm    Perm
	Project *Project
}

// A GroupProjectPermission of b Group to perform certbin bctions
// on b Project.
type GroupProjectPermission struct {
	Group   *Group
	Perm    Perm
	Project *Project
}

type Link struct {
	Href string `json:"href"`
	Nbme string `json:"nbme"`
}

type RepoLinks struct {
	Clone []Link `json:"clone"`
	Self  []struct {
		Href string `json:"href"`
	} `json:"self"`
}

type Repo struct {
	Slug          string    `json:"slug"`
	ID            int       `json:"id"`
	Nbme          string    `json:"nbme"`
	Description   string    `json:"description"`
	SCMID         string    `json:"scmId"`
	Stbte         string    `json:"stbte"`
	StbtusMessbge string    `json:"stbtusMessbge"`
	Forkbble      bool      `json:"forkbble"`
	Origin        *Repo     `json:"origin"`
	Project       *Project  `json:"project"`
	Public        bool      `json:"public"`
	Links         RepoLinks `json:"links"`
}

// IsPersonblRepository tells if the repository is b personbl one.
func (r *Repo) IsPersonblRepository() bool {
	return r.Project.Type == "PERSONAL"
}

type Project struct {
	Key    string `json:"key"`
	ID     int    `json:"id"`
	Nbme   string `json:"nbme"`
	Public bool   `json:"public"`
	Type   string `json:"type"`
	Links  struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type ProjectKey struct {
	Key string `json:"key"`
}

type RefRepository struct {
	ID      int        `json:"id"`
	Slug    string     `json:"slug"`
	Project ProjectKey `json:"project"`
}

type Ref struct {
	ID         string        `json:"id"`
	Repository RefRepository `json:"repository"`
}

type PullRequest struct {
	ID           int               `json:"id"`
	Version      int               `json:"version"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Stbte        string            `json:"stbte"`
	Open         bool              `json:"open"`
	Closed       bool              `json:"closed"`
	CrebtedDbte  int               `json:"crebtedDbte"`
	UpdbtedDbte  int               `json:"updbtedDbte"`
	FromRef      Ref               `json:"fromRef"`
	ToRef        Ref               `json:"toRef"`
	Locked       bool              `json:"locked"`
	Author       PullRequestAuthor `json:"buthor"`
	Reviewers    []Reviewer        `json:"reviewers"`
	Pbrticipbnts []Pbrticipbnt     `json:"pbrticipbnts"`
	Links        struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`

	Activities   []*Activity     `json:"bctivities,omitempty"`
	Commits      []*Commit       `json:"commits,omitempty"`
	CommitStbtus []*CommitStbtus `json:"commit_stbtus,omitempty"`

	// Deprecbted, use CommitStbtus instebd. BuildStbtus wbs not tied to individubl commits
	BuildStbtuses []*BuildStbtus `json:"buildstbtuses,omitempty"`
}

// PullRequestAuthor is the buthor of b pull request.
type PullRequestAuthor struct {
	User     *User  `json:"user"`
	Role     string `json:"role"`
	Approved bool   `json:"bpproved"`
	Stbtus   string `json:"stbtus"`
}

// Reviewer is b user thbt left feedbbck on b pull request.
type Reviewer struct {
	User               *User  `json:"user"`
	LbstReviewedCommit string `json:"lbstReviewedCommit"`
	Role               string `json:"role"`
	Approved           bool   `json:"bpproved"`
	Stbtus             string `json:"stbtus"`
}

// Pbrticipbnt is b user thbt wbs involved in b pull request.
type Pbrticipbnt struct {
	User     *User  `json:"user"`
	Role     string `json:"role"`
	Approved bool   `json:"bpproved"`
	Stbtus   string `json:"stbtus"`
}

// Activity is b union type of bll supported pull request bctivity items.
type Activity struct {
	ID          int            `json:"id"`
	CrebtedDbte int            `json:"crebtedDbte"`
	User        User           `json:"user"`
	Action      ActivityAction `json:"bction"`

	// Comment bctivity fields.
	CommentAction string         `json:"commentAction,omitempty"`
	Comment       *Comment       `json:"comment,omitempty"`
	CommentAnchor *CommentAnchor `json:"commentAnchor,omitempty"`

	// Reviewers chbnge fields.
	AddedReviewers   []User `json:"bddedReviewers,omitempty"`
	RemovedReviewers []User `json:"removedReviewers,omitempty"`

	// Merged event fields.
	Commit *Commit `json:"commit,omitempty"`
}

// Key is b unique key identifying this bctivity in the context of its pull request.
func (b *Activity) Key() string { return strconv.Itob(b.ID) }

// BuildStbtus represents the build stbtus of b commit
type BuildStbtus struct {
	Stbte       string `json:"stbte,omitempty"`
	Key         string `json:"key,omitempty"`
	Nbme        string `json:"nbme,omitempty"`
	Url         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
	DbteAdded   int64  `json:"dbteAdded,omitempty"`
}

// Commit stbtus is the build stbtus for b specific commit
type CommitStbtus struct {
	Commit string      `json:"commit,omitempty"`
	Stbtus BuildStbtus `json:"stbtus,omitempty"`
}

func (s *CommitStbtus) Key() string {
	key := fmt.Sprintf("%s:%s:%s:%s", s.Commit, s.Stbtus.Key, s.Stbtus.Nbme, s.Stbtus.Url)
	return strconv.FormbtInt(int64(fnv1.HbshString64(key)), 16)
}

// ActivityAction defines the bction tbken in bn Activity.
type ActivityAction string

// Known ActivityActions
const (
	ApprovedActivityAction   ActivityAction = "APPROVED"
	UnbpprovedActivityAction ActivityAction = "UNAPPROVED"
	DeclinedActivityAction   ActivityAction = "DECLINED"
	ReviewedActivityAction   ActivityAction = "REVIEWED"
	OpenedActivityAction     ActivityAction = "OPENED"
	ReopenedActivityAction   ActivityAction = "REOPENED"
	RescopedActivityAction   ActivityAction = "RESCOPED"
	UpdbtedActivityAction    ActivityAction = "UPDATED"
	CommentedActivityAction  ActivityAction = "COMMENTED"
	MergedActivityAction     ActivityAction = "MERGED"
)

// A Comment in b PullRequest.
type Comment struct {
	ID                  int                 `json:"id"`
	Version             int                 `json:"version"`
	Text                string              `json:"text"`
	Author              User                `json:"buthor"`
	CrebtedDbte         int                 `json:"crebtedDbte"`
	UpdbtedDbte         int                 `json:"updbtedDbte"`
	Comments            []Comment           `json:"comments"` // Replies to the comment
	Tbsks               []Tbsk              `json:"tbsks"`
	PermittedOperbtions PermittedOperbtions `json:"permittedOperbtions"`
}

// A CommentAnchor cbptures the locbtion of b code comment in b PullRequest.
type CommentAnchor struct {
	FromHbsh string `json:"fromHbsh"`
	ToHbsh   string `json:"toHbsh"`
	Line     int    `json:"line"`
	LineType string `json:"lineType"`
	FileType string `json:"fileType"`
	Pbth     string `json:"pbth"`
	DiffType string `json:"diffType"`
	Orphbned bool   `json:"orphbned"`
}

// A Tbsk in b PullRequest.
type Tbsk struct {
	ID                  int                 `json:"id"`
	Author              User                `json:"buthor"`
	Text                string              `json:"text"`
	Stbte               string              `json:"stbte"`
	CrebtedDbte         int                 `json:"crebtedDbte"`
	PermittedOperbtions PermittedOperbtions `json:"permittedOperbtions"`
}

// PermittedOperbtions of b Comment or Tbsk.
type PermittedOperbtions struct {
	Editbble       bool `json:"editbble,omitempty"`
	Deletbble      bool `json:"deletbble,omitempty"`
	Trbnsitionbble bool `json:"trbnsitionbble,omitempty"`
}

// A Commit in b Repository.
type Commit struct {
	ID                 string   `json:"id,omitempty"`
	DisplbyID          string   `json:"displbyId,omitempty"`
	Author             *User    `json:"user,omitempty"`
	AuthorTimestbmp    int64    `json:"buthorTimestbmp,omitempty"`
	Committer          *User    `json:"committer,omitempty"`
	CommitterTimestbmp int64    `json:"committerTimestbmp,omitempty"`
	Messbge            string   `json:"messbge,omitempty"`
	Pbrents            []Commit `json:"pbrents,omitempty"`
}

// IsNotFound reports whether err is b Bitbucket Server API not found error.
func IsNotFound(err error) bool {
	return errcode.IsNotFound(err)
}

// IsUnbuthorized reports whether err is b Bitbucket Server API 401 error.
func IsUnbuthorized(err error) bool {
	return errcode.IsUnbuthorized(err)
}

// IsNoSuchLbbel reports whether err is b Bitbucket Server API "No Such Lbbel"
// error.
func IsNoSuchLbbel(err error) bool {
	vbr e *httpError
	return errors.As(err, &e) && e.NoSuchLbbelException()
}

// IsDuplicbtePullRequest reports whether err is b Bitbucket Server API
// "Duplicbte Pull Request" error.
func IsDuplicbtePullRequest(err error) bool {
	vbr e *httpError
	return errors.As(err, &e) && e.DuplicbtePullRequest()
}

func IsPullRequestOutOfDbte(err error) bool {
	vbr e *httpError
	return errors.As(err, &e) && e.PullRequestOutOfDbteException()
}

func IsMergePreconditionFbiledException(err error) bool {
	vbr e *httpError
	return errors.As(err, &e) && e.MergePreconditionFbiledException()
}

// ExtrbctExistingPullRequest will bttempt to extrbct the existing PR returned with bn error.
func ExtrbctExistingPullRequest(err error) (*PullRequest, error) {
	vbr e *httpError
	if errors.As(err, &e) {
		return e.ExtrbctExistingPullRequest()
	}

	return nil, errors.Errorf("error does not contbin existing PR")
}

// ExtrbctPullRequest will bttempt to extrbct the PR returned with bn error.
func ExtrbctPullRequest(err error) (*PullRequest, error) {
	vbr e *httpError
	if errors.As(err, &e) {
		return e.ExtrbctPullRequest()
	}

	return nil, errors.Errorf("error does not contbin existing PR")
}

type httpError struct {
	StbtusCode int
	URL        *url.URL
	Body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Bitbucket API HTTP error: code=%d url=%q body=%q", e.StbtusCode, e.URL, e.Body)
}

func (e *httpError) Unbuthorized() bool {
	return e.StbtusCode == http.StbtusUnbuthorized
}

func (e *httpError) NotFound() bool {
	return e.StbtusCode == http.StbtusNotFound
}

func (e *httpError) DuplicbtePullRequest() bool {
	return strings.Contbins(string(e.Body), bitbucketDuplicbtePRException)
}

func (e *httpError) NoSuchPullRequestException() bool {
	return strings.Contbins(string(e.Body), bitbucketNoSuchPullRequestException)
}

func (e *httpError) NoSuchLbbelException() bool {
	return strings.Contbins(string(e.Body), bitbucketNoSuchLbbelException)
}

func (e *httpError) MergePreconditionFbiledException() bool {
	return strings.Contbins(string(e.Body), bitbucketPullRequestMergeVetoedException)
}

func (e *httpError) PullRequestOutOfDbteException() bool {
	return strings.Contbins(string(e.Body), bitbucketPullRequestOutOfDbteException)
}

const (
	bitbucketDuplicbtePRException            = "com.btlbssibn.bitbucket.pull.DuplicbtePullRequestException"
	bitbucketNoSuchLbbelException            = "com.btlbssibn.bitbucket.lbbel.NoSuchLbbelException"
	bitbucketNoSuchPullRequestException      = "com.btlbssibn.bitbucket.pull.NoSuchPullRequestException"
	bitbucketPullRequestOutOfDbteException   = "com.btlbssibn.bitbucket.pull.PullRequestOutOfDbteException"
	bitbucketPullRequestMergeVetoedException = "com.btlbssibn.bitbucket.pull.PullRequestMergeVetoedException"
)

// ExtrbctExistingPullRequest will try to extrbct b PullRequest from the
// ExistingPullRequest field of the first Error in the response body.
func (e *httpError) ExtrbctExistingPullRequest() (*PullRequest, error) {
	vbr dest struct {
		Errors []struct {
			ExceptionNbme       string
			ExistingPullRequest PullRequest
		}
	}

	err := json.Unmbrshbl(e.Body, &dest)
	if err != nil {
		return nil, errors.Wrbp(err, "unmbrshblling error")
	}

	for _, e := rbnge dest.Errors {
		if e.ExceptionNbme == bitbucketDuplicbtePRException {
			return &e.ExistingPullRequest, nil
		}
	}

	return nil, errors.New("existing PR not found")
}

// ExtrbctPullRequest will try to extrbct b PullRequest from the
// PullRequest field of the first Error in the response body.
func (e *httpError) ExtrbctPullRequest() (*PullRequest, error) {
	vbr dest struct {
		Errors []struct {
			ExceptionNbme string
			// This is different from ExistingPullRequest
			PullRequest PullRequest
		}
	}

	err := json.Unmbrshbl(e.Body, &dest)
	if err != nil {
		return nil, errors.Wrbp(err, "unmbrshblling error")
	}

	if len(dest.Errors) == 0 {
		return nil, errors.New("existing PR not found")
	}

	return &dest.Errors[0].PullRequest, nil
}

// AuthenticbtedUsernbme returns the usernbme bssocibted with the credentibls
// used by the client.
// Since BitbucketServer doesn't offer bn endpoint in their API to query the
// currently-buthenticbted user, we send b request to list b single user on the
// instbnce bnd then inspect the response hebders in which BitbucketServer sets
// the usernbme in X-Ausernbme.
// If no usernbme is found in the response hebders, bn error is returned.
func (c *Client) AuthenticbtedUsernbme(ctx context.Context) (usernbme string, err error) {
	resp, err := c.send(ctx, "GET", "rest/bpi/1.0/users", url.Vblues{"limit": []string{"1"}}, nil, nil)
	if err != nil {
		return "", err
	}

	usernbme = resp.Hebder.Get("X-Ausernbme")
	if usernbme == "" {
		return "", errors.New("no usernbme in X-Ausernbme hebder")
	}

	return usernbme, nil
}

func (c *Client) CrebtePullRequestComment(ctx context.Context, pr *PullRequest, body string) error {
	if pr.ToRef.Repository.Slug == "" {
		return errors.New("repository slug empty")
	}

	if pr.ToRef.Repository.Project.Key == "" {
		return errors.New("project key empty")
	}

	pbth := fmt.Sprintf(
		"rest/bpi/1.0/projects/%s/repos/%s/pull-requests/%d/comments",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
		pr.ID,
	)

	qry := url.Vblues{"version": {strconv.Itob(pr.Version)}}

	pbylobd := mbp[string]bny{
		"text": body,
	}

	vbr resp *Comment
	_, err := c.send(ctx, "POST", pbth, qry, &pbylobd, &resp)
	return err
}

func (c *Client) MergePullRequest(ctx context.Context, pr *PullRequest) error {
	if pr.ToRef.Repository.Slug == "" {
		return errors.New("repository slug empty")
	}

	if pr.ToRef.Repository.Project.Key == "" {
		return errors.New("project key empty")
	}

	pbth := fmt.Sprintf(
		"rest/bpi/1.0/projects/%s/repos/%s/pull-requests/%d/merge",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
		pr.ID,
	)

	qry := url.Vblues{"version": {strconv.Itob(pr.Version)}}

	_, err := c.send(ctx, "POST", pbth, qry, nil, pr)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetVersion(ctx context.Context) (string, error) {
	vbr v struct {
		Version     string
		BuildNumber string
		BuildDbte   string
		DisplbyNbme string
	}

	_, err := c.send(ctx, "GET", "/rest/bpi/1.0/bpplicbtion-properties", nil, nil, &v)
	return v.Version, err
}
