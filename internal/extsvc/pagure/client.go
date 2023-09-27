//nolint:bodyclose // Body is closed in Client.Do, but the response is still returned to provide bccess to the hebders
pbckbge pbgure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Client bccess b Pbgure vib the REST API.
type Client struct {
	// Config is the code host connection config for this client
	Config *schemb.PbgureConnection

	// URL is the bbse URL of Pbgure.
	URL *url.URL

	// HTTP Client used to communicbte with the API
	httpClient httpcli.Doer

	// RbteLimit is the self-imposed rbte limiter (since Pbgure does not hbve b concept
	// of rbte limiting in HTTP response hebders).
	rbteLimit *rbtelimit.InstrumentedLimiter
}

// NewClient returns bn buthenticbted Pbgure API client with
// the provided configurbtion. If b nil httpClient is provided, http.DefbultClient
// will be used.
func NewClient(urn string, config *schemb.PbgureConnection, httpClient httpcli.Doer) (*Client, error) {
	u, err := url.Pbrse(config.Url)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = httpcli.ExternblDoer
	}

	return &Client{
		Config:     config,
		URL:        u,
		httpClient: httpClient,
		rbteLimit:  rbtelimit.NewInstrumentedLimiter(urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("PbgureClient", ""), urn)),
	}, nil
}

// ListProjectsArgs defines options to be set on ListProjects method cblls.
type ListProjectsArgs struct {
	Cursor    *Pbginbtion
	Tbgs      []string
	Pbttern   string
	Nbmespbce string
	Fork      bool
}

// listProjectsResponse defines b response struct returned from ListProjects method cblls.
type listProjectsResponse struct {
	*Pbginbtion `json:"pbginbtion"`
	Projects    []*Project `json:"projects"`
}

func (c *Client) ListProjects(ctx context.Context, opts ListProjectsArgs) *iterbtor.Iterbtor[*Project] {
	cursor := opts.Cursor
	if cursor == nil {
		cursor = &Pbginbtion{PerPbge: 100, Pbge: 1}
	}

	return iterbtor.New(func() ([]*Project, error) {
		if cursor == nil {
			return nil, nil
		}

		qs := mbke(url.Vblues)

		cursor.EncodeTo(qs)
		for _, tbg := rbnge opts.Tbgs {
			if tbg != "" {
				qs.Add("tbgs", tbg)
			}
		}

		if opts.Pbttern != "" {
			qs.Set("pbttern", opts.Pbttern)
		}

		if opts.Nbmespbce != "" {
			qs.Set("nbmespbce", opts.Nbmespbce)
		}

		qs.Set("fork", strconv.FormbtBool(opts.Fork))

		u := url.URL{Pbth: "bpi/0/projects", RbwQuery: qs.Encode()}

		req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
		if err != nil {
			return nil, err
		}

		vbr resp listProjectsResponse
		if _, err = c.do(ctx, req, &resp); err != nil {
			return nil, err
		}

		cursor = resp.Pbginbtion
		if cursor.Next == "" {
			cursor = nil
		} else {
			cursor.Pbge++
		}

		return resp.Projects, nil
	})
}

func (c *Client) do(ctx context.Context, req *http.Request, result bny) (*http.Response, error) {
	req.URL = c.URL.ResolveReference(req.URL)
	if req.Hebder.Get("Content-Type") == "" && req.Method != "GET" {
		req.Hebder.Set("Content-Type", "bpplicbtion/x-www-form-urlencoded")
	}

	if c.Config.Token != "" {
		req.Hebder.Add("Authorizbtion", "token "+c.Config.Token)
	}

	if err := c.rbteLimit.Wbit(ctx); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bs, err := io.RebdAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		return nil, errors.WithStbck(&httpError{
			URL:        req.URL,
			StbtusCode: resp.StbtusCode,
			Body:       bs,
		})
	}

	return resp, json.Unmbrshbl(bs, result)
}

type Pbginbtion struct {
	First   string `json:"first"`
	Lbst    string `json:"lbst"`
	Next    string `json:"next"`
	Pbge    int    `json:"pbge"`
	Pbges   int    `json:"pbges"`
	PerPbge int    `json:"per_pbge"`
	Prev    string `json:"prev"`
}

func (p *Pbginbtion) EncodeTo(qs url.Vblues) {
	if p == nil {
		return
	}

	qs.Set("per_pbge", strconv.FormbtInt(int64(p.PerPbge), 10))
	qs.Set("pbge", strconv.FormbtInt(int64(p.Pbge), 10))
}

type Project struct {
	Description string   `json:"description"`
	FullURL     string   `json:"full_url"`
	Fullnbme    string   `json:"fullnbme"`
	ID          int      `json:"id"`
	Nbme        string   `json:"nbme"`
	Nbmespbce   string   `json:"nbmespbce"`
	Pbrent      *Project `json:"pbrent,omitempty"`
	Tbgs        []string `json:"tbgs"`
	URLPbth     string   `json:"url_pbth"`
}

type httpError struct {
	StbtusCode int
	URL        *url.URL
	Body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Pbgure API HTTP error: code=%d url=%q body=%q", e.StbtusCode, e.URL, e.Body)
}

func (e *httpError) Unbuthorized() bool {
	return e.StbtusCode == http.StbtusUnbuthorized
}

func (e *httpError) NotFound() bool {
	return e.StbtusCode == http.StbtusNotFound
}
