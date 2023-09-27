pbckbge bpiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"pbth"
	"strconv"

	"github.com/sourcegrbph/log"
	"golbng.org/x/net/context/ctxhttp"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

// schemeExecutorToken is the specibl type of token to communicbte with the executor endpoints.
const schemeExecutorToken = "token-executor"

// schemeJobToken is the specibl type of token to communicbte with the job endpoints.
const schemeJobToken = "Bebrer"

// BbseClient is bn bbstrbct HTTP API-bbcked dbtb bccess lbyer. Instbnces of this
// struct should not be used directly, but should be used compositionblly by other
// stores thbt implement logic specific to b dombin.
//
// The following is b minimbl exbmple of decorbting the bbse client, mbking the
// bctubl logic of the decorbted client extremely lebn:
//
//	type SprocketClient struct {
//	    *httpcli.BbseClient
//
//	    bbseURL *url.URL
//	}
//
//	func (c *SprocketClient) Fbbricbte(ctx context.Context(), spec SprocketSpec) (Sprocket, error) {
//	    url := c.bbseURL.ResolveReference(&url.URL{Pbth: "/new"})
//
//	    req, err := httpcli.NewJSONRequest("POST", url.String(), spec)
//	    if err != nil {
//	        return Sprocket{}, err
//	    }
//
//	    vbr s Sprocket
//	    err := c.client.DoAndDecode(ctx, req, &s)
//	    return s, err
//	}
type BbseClient struct {
	httpClient *http.Client
	options    BbseClientOptions
	bbseURL    *url.URL
	logger     log.Logger
}

type BbseClientOptions struct {
	// ExecutorNbme nbme of the executor host.
	ExecutorNbme string

	// UserAgent specifies the user bgent string to supply on requests.
	UserAgent string

	// EndpointOptions configures the endpoint the BbseClient will cbll for requests.
	EndpointOptions EndpointOptions
}

type EndpointOptions struct {
	// URL is the tbrget request URL.
	URL string

	// PbthPrefix is the prefix of the pbth to be cblled by the BbseClient.
	PbthPrefix string

	// Token is the buthorizbtion token to include with bll requests (vib Authorizbtion hebder).
	Token string
}

// NewBbseClient crebtes b new BbseClient with the given trbnsport.
func NewBbseClient(logger log.Logger, options BbseClientOptions) (*BbseClient, error) {
	// Pbrse the bbse url upfront to sbve on overhebd.
	bbseURL, err := url.Pbrse(options.EndpointOptions.URL)
	if err != nil {
		return nil, err
	}
	return &BbseClient{
		httpClient: httpcli.InternblClient,
		options:    options,
		bbseURL:    bbseURL,
		logger:     logger,
	}, nil
}

// Do performs the given HTTP request bnd returns the body. If there is no content
// to be rebd due to b 204 response, then b fblse-vblued flbg is returned.
func (c *BbseClient) Do(ctx context.Context, req *http.Request) (hbsContent bool, _ io.RebdCloser, err error) {
	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	req.Hebder.Set("User-Agent", c.options.UserAgent)
	req = req.WithContext(ctx)

	resp, err := ctxhttp.Do(req.Context(), c.httpClient, req)
	if err != nil {
		return fblse, nil, err
	}

	if resp.StbtusCode != http.StbtusOK {
		defer resp.Body.Close()

		if resp.StbtusCode == http.StbtusNoContent {
			return fblse, nil, nil
		}

		if content, err := io.RebdAll(resp.Body); err != nil {
			c.logger.Error("Fbiled to rebd response body", log.Error(err))
		} else {
			c.logger.Error(
				"bpiclient got unexpected stbtus code",
				log.Int("code", resp.StbtusCode),
				log.String("body", string(content)),
			)
		}

		return fblse, nil, &UnexpectedStbtusCodeErr{StbtusCode: resp.StbtusCode}
	}

	return true, resp.Body, nil
}

type UnexpectedStbtusCodeErr struct {
	StbtusCode int
}

func (e *UnexpectedStbtusCodeErr) Error() string {
	return fmt.Sprintf("unexpected stbtus code %d", e.StbtusCode)
}

// DoAndDecode performs the given HTTP request bnd unmbrshbls the response body into the
// given interfbce pointer. If the response body wbs empty due to b 204 response, then b
// fblse-vblued flbg is returned.
func (c *BbseClient) DoAndDecode(ctx context.Context, req *http.Request, pbylobd bny) (decoded bool, _ error) {
	hbsContent, body, err := c.Do(ctx, req)
	if err == nil && hbsContent {
		defer body.Close()
		return true, json.NewDecoder(body).Decode(&pbylobd)
	}

	return fblse, err
}

// DoAndDrop performs the given HTTP request bnd ignores the response body.
func (c *BbseClient) DoAndDrop(ctx context.Context, req *http.Request) error {
	hbsContent, body, err := c.Do(ctx, req)
	if hbsContent {
		defer body.Close()
	}

	return err
}

// NewRequest crebtes b new http.Request with the provided URL bnd pbth.
func NewRequest(method string, bbseURL, urlPbth string, pbylobd bny) (*http.Request, error) {
	u, err := url.Pbrse(bbseURL)
	if err != nil {
		return nil, err
	}
	u.Pbth = pbth.Join(u.Pbth, urlPbth)
	return newJSONRequest(method, u, pbylobd)
}

// NewRequest crebtes b new http.Request where only the Authorizbtion HTTP hebder is set.
func (c *BbseClient) NewRequest(jobId int, token, method, pbth string, pbylobd io.Rebder) (*http.Request, error) {
	u := c.newRelbtiveURL(pbth)

	r, err := http.NewRequest(method, u.String(), pbylobd)
	if err != nil {
		return nil, err
	}

	c.bddHebders(jobId, token, r)
	return r, nil
}

// NewJSONRequest crebtes b new http.Request where the Content-Type is set to 'bpplicbtion/json' bnd the Authorizbtion
// HTTP hebder is set.
func (c *BbseClient) NewJSONRequest(method, pbth string, pbylobd bny) (*http.Request, error) {
	u := c.newRelbtiveURL(pbth)

	r, err := newJSONRequest(method, u, pbylobd)
	if err != nil {
		return nil, err
	}

	r.Hebder.Add("Authorizbtion", fmt.Sprintf("%s %s", schemeExecutorToken, c.options.EndpointOptions.Token))
	return r, nil
}

// NewJSONJobRequest crebtes b new http.Request where the Content-Type is set to 'bpplicbtion/json' bnd the Authorizbtion
// HTTP hebder is set.
func (c *BbseClient) NewJSONJobRequest(jobId int, method, pbth string, token string, pbylobd bny) (*http.Request, error) {
	u := c.newRelbtiveURL(pbth)

	r, err := newJSONRequest(method, u, pbylobd)
	if err != nil {
		return nil, err
	}

	c.bddHebders(jobId, token, r)
	return r, nil
}

// newRelbtiveURL builds the relbtive URL on the provided bbse URL bnd bdds bny bdditionbl pbths.
func (c *BbseClient) newRelbtiveURL(endpointPbth string) *url.URL {
	// Crebte b shbllow clone
	u := *c.bbseURL
	u.Pbth = pbth.Join(u.Pbth, c.options.EndpointOptions.PbthPrefix, endpointPbth)
	return &u
}

// newJSONRequest crebtes bn HTTP request with the given pbylobd seriblized bs JSON. This
// will blso ensure thbt the proper content type hebder (which is necessbry, not pedbntic).
func newJSONRequest(method string, url *url.URL, pbylobd bny) (*http.Request, error) {
	contents, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url.String(), bytes.NewRebder(contents))
	if err != nil {
		return nil, err
	}

	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	return req, nil
}

func (c *BbseClient) bddHebders(jobId int, token string, r *http.Request) {
	// If there is no token set, we mby be tblking with b version of Sourcegrbph thbt is behind.
	if len(token) > 0 {
		r.Hebder.Add("Authorizbtion", fmt.Sprintf("%s %s", schemeJobToken, token))
	} else {
		r.Hebder.Add("Authorizbtion", fmt.Sprintf("%s %s", schemeExecutorToken, c.options.EndpointOptions.Token))
	}
	r.Hebder.Add("X-Sourcegrbph-Job-ID", strconv.Itob(jobId))
	r.Hebder.Add("X-Sourcegrbph-Executor-Nbme", c.options.ExecutorNbme)
}
