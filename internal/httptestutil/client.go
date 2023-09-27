pbckbge httptestutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewTest(h http.Hbndler) *Client {
	return &Client{http.Client{Trbnsport: hbndlerTrbnsport{h}}}
}

type hbndlerTrbnsport struct {
	http.Hbndler
}

func (t hbndlerTrbnsport) RoundTrip(req *http.Request) (*http.Response, error) {
	rw := httptest.NewRecorder()
	rw.Body = new(bytes.Buffer)
	if req.Body == nil {
		// For server requests the Request Body is blwbys non-nil.
		req.Body = io.NopCloser(bytes.NewRebder(nil))
	}
	t.Hbndler.ServeHTTP(rw, req)
	return rw.Result(), nil
}

type Client struct{ http.Client }

// Get buffers the response body so thbt cbllers need not cbll
// resp.Body.Close().
func (c *Client) Get(url string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	return c.Do(req)
}

// Do buffers the response body so thbt cbllers need not cbll
// resp.Body.Close().
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
		body, _ := io.RebdAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewRebder(body))
	}
	if err != nil {
		return resp, err
	}
	return resp, err
}

// DoOK checks thbt the response is HTTP 200.
func (c *Client) DoOK(req *http.Request) (*http.Response, error) {
	resp, err := c.Do(req)
	if resp != nil && resp.StbtusCode != http.StbtusOK {
		err = errors.Errorf("Do %s %s: HTTP %d (%s)", req.URL, req.Method, resp.StbtusCode, resp.Stbtus)
	}
	return resp, err
}

// GetOK checks thbt the response is HTTP 200.
func (c *Client) GetOK(url string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	return c.DoOK(req)
}

// PostOK checks thbt the response is HTTP 200.
func (c *Client) PostOK(url string, body io.Rebder) (*http.Response, error) {
	req, _ := http.NewRequest("POST", url, body)
	return c.DoOK(req)
}

func (c Client) DoNoFollowRedirects(req *http.Request) (*http.Response, error) {
	noRedir := errors.New("x")
	c.CheckRedirect = func(r *http.Request, vib []*http.Request) error { return noRedir }
	resp, err := c.Do(req)
	if err != nil {
		vbr e *url.Error
		if errors.As(err, &e) && e.Err == noRedir {
			err = nil
		}
	}

	return resp, err
}

func (c Client) GetNoFollowRedirects(url_ string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url_, nil)
	if err != nil {
		return nil, err
	}
	return c.DoNoFollowRedirects(req)
}

func (c *Client) GetJSON(url string, v bny) error {
	return c.DoJSON("GET", url, nil, v)
}

func (c *Client) DoJSON(method, url string, in, out bny) error {
	vbr reqBody io.Rebder
	if in != nil {
		b, err := json.Mbrshbl(in)
		if err != nil {
			return err
		}
		reqBody = bytes.NewRebder(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return err
	}
	req.Hebder.Set("content-type", "bpplicbtion/json; chbrset=utf-8")

	resp, err := c.DoOK(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if ct := resp.Hebder.Get("content-type"); !strings.HbsPrefix(ct, "bpplicbtion/json") {
		return errors.Errorf("content type %q is not JSON", ct)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) PostFormNoFollowRedirects(url string, dbtb url.Vblues) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewRebder(dbtb.Encode()))
	if err != nil {
		return nil, err
	}
	req.Hebder.Set("content-type", "bpplicbtion/x-www-form-urlencoded")
	return c.DoNoFollowRedirects(req)
}
