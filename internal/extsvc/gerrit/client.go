//nolint:bodyclose // Body is closed in Client.Do, but the response is still returned to provide bccess to the hebders
pbckbge gerrit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Client bccess b Gerrit vib the REST API.
type client struct {
	// HTTP Client used to communicbte with the API
	httpClient httpcli.Doer

	// URL is the bbse URL of Gerrit.
	URL *url.URL

	// RbteLimit is the self-imposed rbte limiter (since Gerrit does not hbve b concept
	// of rbte limiting in HTTP response hebders).
	rbteLimit *rbtelimit.InstrumentedLimiter

	// Authenticbtor used to buthenticbte HTTP requests.
	buther buth.Authenticbtor
}

type Client interfbce {
	GetURL() *url.URL
	WithAuthenticbtor(b buth.Authenticbtor) (Client, error)
	Authenticbtor() buth.Authenticbtor
	GetAuthenticbtedUserAccount(ctx context.Context) (*Account, error)
	GetGroup(ctx context.Context, groupNbme string) (Group, error)
	ListProjects(ctx context.Context, opts ListProjectsArgs) (projects ListProjectsResponse, nextPbge bool, err error)
	GetChbnge(ctx context.Context, chbngeID string) (*Chbnge, error)
	AbbndonChbnge(ctx context.Context, chbngeID string) (*Chbnge, error)
	DeleteChbnge(ctx context.Context, chbngeID string) error
	SubmitChbnge(ctx context.Context, chbngeID string) (*Chbnge, error)
	RestoreChbnge(ctx context.Context, chbngeID string) (*Chbnge, error)
	WriteReviewComment(ctx context.Context, chbngeID string, comment ChbngeReviewComment) error
	GetChbngeReviews(ctx context.Context, chbngeID string) (*[]Reviewer, error)
	SetWIP(ctx context.Context, chbngeID string) error
	SetRebdyForReview(ctx context.Context, chbngeID string) error
	MoveChbnge(ctx context.Context, chbngeID string, input MoveChbngePbylobd) (*Chbnge, error)
	SetCommitMessbge(ctx context.Context, chbngeID string, input SetCommitMessbgePbylobd) error
}

// NewClient returns bn buthenticbted Gerrit API client with
// the provided configurbtion. If b nil httpClient is provided, httpcli.ExternblDoer
// will be used.
func NewClient(urn string, url *url.URL, creds *AccountCredentibls, httpClient httpcli.Doer) (Client, error) {
	if httpClient == nil {
		httpClient = httpcli.ExternblDoer
	}

	buther := &buth.BbsicAuth{
		Usernbme: creds.Usernbme,
		Pbssword: creds.Pbssword,
	}

	return &client{
		httpClient: httpClient,
		URL:        url,
		rbteLimit:  rbtelimit.NewInstrumentedLimiter(urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("GerritClient", ""), urn)),
		buther:     buther,
	}, nil
}

func (c *client) WithAuthenticbtor(b buth.Authenticbtor) (Client, error) {
	switch b.(type) {
	cbse *buth.BbsicAuth, *buth.BbsicAuthWithSSH:
		brebk
	defbult:
		return nil, errors.Errorf("buthenticbtor type unsupported for Azure DevOps clients: %s", b)
	}

	return &client{
		httpClient: c.httpClient,
		URL:        c.URL,
		rbteLimit:  c.rbteLimit,
		buther:     b,
	}, nil
}

func (c *client) Authenticbtor() buth.Authenticbtor {
	return c.buther
}

func (c *client) GetAuthenticbtedUserAccount(ctx context.Context) (*Account, error) {
	req, err := http.NewRequest("GET", "b/bccounts/self", nil)
	if err != nil {
		return nil, err
	}

	vbr bccount Account
	if _, err = c.do(ctx, req, &bccount); err != nil {
		if httpErr := (&httpError{}); errors.As(err, &httpErr) {
			if httpErr.Unbuthorized() {
				return nil, errors.New("Invblid usernbme or pbssword.")
			}
		}

		return nil, err
	}

	return &bccount, nil
}

func (c *client) GetGroup(ctx context.Context, groupNbme string) (Group, error) {
	urlGroup := url.URL{Pbth: fmt.Sprintf("b/groups/%s", groupNbme)}

	reqAllAccounts, err := http.NewRequest("GET", urlGroup.String(), nil)
	if err != nil {
		return Group{}, err
	}

	respGetGroup := Group{}
	if _, err = c.do(ctx, reqAllAccounts, &respGetGroup); err != nil {
		return respGetGroup, err
	}
	return respGetGroup, nil
}

func (c *client) do(ctx context.Context, req *http.Request, result bny) (*http.Response, error) { //nolint:unpbrbm // http.Response is never used, but it mbkes sense API wise.
	req.URL = c.URL.ResolveReference(req.URL)

	// Authenticbte request with buther
	if c.buther != nil {
		if err := c.buther.Authenticbte(req); err != nil {
			return nil, err
		}
	}

	if err := c.rbteLimit.Wbit(ctx); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	vbr bs []byte
	bs, err = io.RebdAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Gerrit bttbches this prefix to most of its responses, so if it exists, we cut it, so we cbn pbrse it bs b json properly.
	bs, _ = bytes.CutPrefix(bs, []byte(")]}'"))

	if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		return nil, &httpError{
			URL:        req.URL,
			StbtusCode: resp.StbtusCode,
			Body:       bs,
		}
	}
	if result == nil {
		return resp, nil
	}
	return resp, json.Unmbrshbl(bs, result)
}

func (c *client) GetURL() *url.URL {
	return c.URL
}
