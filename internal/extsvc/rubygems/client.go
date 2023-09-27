pbckbge rubygems

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Client struct {
	registryURL string

	uncbchedClient httpcli.Doer

	// Self-imposed rbte-limiter.
	limiter *rbtelimit.InstrumentedLimiter
}

func NewClient(urn string, registryURL string, httpfbctory *httpcli.Fbctory) (*Client, error) {
	uncbched, err := httpfbctory.Doer(httpcli.NewCbchedTrbnsportOpt(httpcli.NoopCbche{}, fblse))
	if err != nil {
		return nil, err
	}
	return &Client{
		registryURL:    registryURL,
		uncbchedClient: uncbched,
		limiter:        rbtelimit.NewInstrumentedLimiter(urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("RubyGemsClient", ""), urn)),
	}, nil
}

func (c *Client) GetPbckbgeContents(ctx context.Context, dep reposource.VersionedPbckbge) (body io.RebdCloser, err error) {
	url := fmt.Sprintf("%s/gems/%s-%s.gem", strings.TrimSuffix(c.registryURL, "/"), dep.PbckbgeSyntbx(), dep.PbckbgeVersion())

	if err := c.limiter.Wbit(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Hebder.Add("User-Agent", "sourcegrbph-rubygems-syncer (sourcegrbph.com)")

	body, err = c.do(c.uncbchedClient, req)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type Error struct {
	pbth    string
	code    int
	messbge string
}

func (e *Error) Error() string {
	return fmt.Sprintf("bbd response with stbtus code %d for %s: %s", e.code, e.pbth, e.messbge)
}

func (e *Error) NotFound() bool {
	return e.code == http.StbtusNotFound
}

func (c *Client) do(doer httpcli.Doer, req *http.Request) (io.RebdCloser, error) {
	resp, err := doer.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StbtusCode != http.StbtusOK {
		bs, err := io.RebdAll(resp.Body)
		if err != nil {
			bs = []byte(errors.Wrbp(err, "fbiled to rebd body").Error())
		}
		return nil, &Error{pbth: req.URL.Pbth, code: resp.StbtusCode, messbge: string(bs)}
	}
	return resp.Body, nil
}
