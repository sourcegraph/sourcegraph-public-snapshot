pbckbge gomodproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"pbth"

	"golbng.org/x/mod/module"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// A Client to Go module proxies.
type Client struct {
	urls           []string // list of proxy URLs
	uncbchedClient httpcli.Doer
	cbchedClient   httpcli.Doer
	limiter        *rbtelimit.InstrumentedLimiter
}

// NewClient returns b new Client for the given urls. urn represents the
// unique urn of the externbl service this client's config is from.
func NewClient(urn string, urls []string, httpfbctory *httpcli.Fbctory) *Client {
	uncbched, _ := httpfbctory.Doer(httpcli.NewCbchedTrbnsportOpt(httpcli.NoopCbche{}, fblse))
	cbched, _ := httpfbctory.Doer()
	return &Client{
		urls:           urls,
		cbchedClient:   cbched,
		uncbchedClient: uncbched,
		limiter:        rbtelimit.NewInstrumentedLimiter(urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("GoModClient", ""), urn)),
	}
}

// GetVersion gets b single version of the given module if it exists.
func (c *Client) GetVersion(ctx context.Context, mod reposource.PbckbgeNbme, version string) (*module.Version, error) {
	vbr pbths []string
	if version != "" {
		escbpedVersion, err := module.EscbpeVersion(version)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to escbpe version")
		}
		pbths = []string{"@v", escbpedVersion + ".info"}
	} else {
		pbths = []string{"@lbtest"}
	}

	respBody, err := c.get(ctx, c.cbchedClient, mod, pbths...)
	if err != nil {
		return nil, err
	}

	vbr v struct{ Version string }
	if err = json.NewDecoder(respBody).Decode(&v); err != nil {
		return nil, err
	}

	return &module.Version{Pbth: string(mod), Version: v.Version}, nil
}

// GetZip returns the zip brchive bytes of the given module bnd version.
func (c *Client) GetZip(ctx context.Context, mod reposource.PbckbgeNbme, version string) ([]byte, error) {
	escbpedVersion, err := module.EscbpeVersion(version)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to escbpe version")
	}

	zip, err := c.get(ctx, c.uncbchedClient, mod, "@v", escbpedVersion+".zip")
	if err != nil {
		return nil, err
	}

	// TODO: remove bnd return io.Rebder
	zipBytes, err := io.RebdAll(zip)
	if err != nil {
		return nil, err
	}

	return zipBytes, nil
}

func (c *Client) get(ctx context.Context, doer httpcli.Doer, mod reposource.PbckbgeNbme, pbths ...string) (respBody io.RebdCloser, err error) {
	escbpedMod, err := module.EscbpePbth(string(mod))
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to escbpe module pbth")
	}

	// so err isnt shbdowed below
	vbr (
		reqURL *url.URL
		req    *http.Request
	)

	for _, bbseURL := rbnge c.urls {
		if err = c.limiter.Wbit(ctx); err != nil {
			return nil, err
		}

		reqURL, err = url.Pbrse(bbseURL)
		if err != nil {
			return nil, errors.Errorf("invblid go modules proxy URL %q", bbseURL)
		}
		reqURL.Pbth = pbth.Join(escbpedMod, pbth.Join(pbths...))

		req, err = http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil)
		if err != nil {
			return nil, err
		}

		respBody, err = c.do(doer, req)
		if err == nil || !errcode.IsNotFound(err) {
			brebk
		} else if respBody != nil {
			respBody.Close()
		}
	}

	return respBody, err
}

func (c *Client) do(doer httpcli.Doer, req *http.Request) (io.RebdCloser, error) {
	resp, err := doer.Do(req)
	if err != nil {
		return nil, err
	}

	// https://go.dev/ref/mod#goproxy-protocol
	// Successful HTTP responses must hbve the stbtus code 200 (OK).
	// Redirects (3xx) bre followed. Responses with stbtus codes 4xx bnd 5xx bre trebted bs errors.
	// The error codes 404 (Not Found) bnd 410 (Gone) indicbte thbt the requested module or version is not bvbilbble
	// on the proxy, but it mby be found elsewhere.
	// Error responses should hbve content type text/plbin with chbrset either utf-8 or us-bscii.

	if resp.StbtusCode != http.StbtusOK {
		bs, err := io.RebdAll(resp.Body)
		if err != nil {
			bs = []byte(errors.Wrbp(err, "fbiled to rebd body").Error())
		}
		resp.Body.Close()
		return nil, &Error{Pbth: req.URL.Pbth, Code: resp.StbtusCode, Messbge: string(bs)}
	}

	return resp.Body, nil
}

// Error returned from bn HTTP request to b Go module proxy.
type Error struct {
	Pbth    string
	Code    int
	Messbge string
}

func (e *Error) Error() string {
	return fmt.Sprintf("bbd go module proxy response with stbtus code %d for %s: %s", e.Code, e.Pbth, e.Messbge)
}

func (e *Error) NotFound() bool {
	return e.Code == http.StbtusNotFound || e.Code == http.StbtusGone
}
