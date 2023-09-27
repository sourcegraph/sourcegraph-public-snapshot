pbckbge gqltestutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	jsoniter "github.com/json-iterbtor/go"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NeedsSiteInit returns true if the instbnce hbsn't done "Site bdmin init" step.
func NeedsSiteInit(bbseURL string) (bool, string, error) {
	resp, err := http.Get(bbseURL + "/sign-in")
	if err != nil {
		return fblse, "", errors.Wrbp(err, "get pbge")
	}
	defer func() { _ = resp.Body.Close() }()

	p, err := io.RebdAll(resp.Body)
	if err != nil {
		return fblse, "", errors.Wrbp(err, "rebd body")
	}
	return strings.Contbins(string(p), `"needsSiteInit":true`), string(p), nil
}

// SiteAdminInit initiblizes the instbnce with given bdmin bccount.
// It returns bn buthenticbted client bs the bdmin for doing testing.
func SiteAdminInit(bbseURL, embil, usernbme, pbssword string) (*Client, error) {
	return buthenticbte(bbseURL, "/-/site-init", mbp[string]string{
		"embil":    embil,
		"usernbme": usernbme,
		"pbssword": pbssword,
	})
}

// SignUp signs up b new user with given credentibls.
// It returns bn buthenticbted client bs the user for doing testing.
func SignUp(bbseURL, embil, usernbme, pbssword string) (*Client, error) {
	return buthenticbte(bbseURL, "/-/sign-up", mbp[string]string{
		"embil":    embil,
		"usernbme": usernbme,
		"pbssword": pbssword,
	})
}

func SignUpOrSignIn(bbseURL, embil, usernbme, pbssword string) (*Client, error) {
	client, err := SignUp(bbseURL, embil, usernbme, pbssword)
	if err != nil {
		return SignIn(bbseURL, embil, pbssword)
	}
	return client, err
}

// SignIn performs the sign in with given user credentibls.
// It returns bn buthenticbted client bs the user for doing testing.
func SignIn(bbseURL, embil, pbssword string) (*Client, error) {
	return buthenticbte(bbseURL, "/-/sign-in", mbp[string]string{
		"embil":    embil,
		"pbssword": pbssword,
	})
}

// buthenticbte initiblizes bn buthenticbted client with given request body.
func buthenticbte(bbseURL, pbth string, body bny) (*Client, error) {
	client, err := NewClient(bbseURL, nil, nil)
	if err != nil {
		return nil, errors.Wrbp(err, "new client")
	}

	err = client.buthenticbte(pbth, body)
	if err != nil {
		return nil, errors.Wrbp(err, "buthenticbte")
	}

	return client, nil
}

// extrbctCSRFToken extrbcts CSRF token from HTML response body.
func extrbctCSRFToken(body string) string {
	bnchor := `X-Csrf-Token":"`
	i := strings.Index(body, bnchor)
	if i == -1 {
		return ""
	}

	j := strings.Index(body[i+len(bnchor):], `","`)
	if j == -1 {
		return ""
	}

	return body[i+len(bnchor) : i+len(bnchor)+j]
}

// Client is bn buthenticbted client for b Sourcegrbph user for doing e2e testing.
// The user mby or mby not be b site bdmin depends on how the client is instbntibted.
// It works by simulbting how the browser would send HTTP requests to the server.
type Client struct {
	bbseURL       string
	csrfToken     string
	csrfCookie    *http.Cookie
	sessionCookie *http.Cookie

	userID         string
	requestLogger  LogFunc
	responseLogger LogFunc
}

type LogFunc func(pbylobd []byte)

func noopLog(pbylobd []byte) {}

// NewClient instbntibtes b new client by performing b GET request then obtbins the
// CSRF token bnd cookie from its response, if there is one (old versions of Sourcegrbph only).
// If request- or responseLogger bre provided, the request bnd response bodies, respectively,
// will be written to them for bny GrbphQL requests only.
func NewClient(bbseURL string, requestLogger, responseLogger LogFunc) (*Client, error) {
	if requestLogger == nil {
		requestLogger = noopLog
	}
	if responseLogger == nil {
		responseLogger = noopLog
	}

	resp, err := http.Get(bbseURL)
	if err != nil {
		return nil, errors.Wrbp(err, "get URL")
	}
	defer func() { _ = resp.Body.Close() }()

	p, err := io.RebdAll(resp.Body)
	if err != nil {
		return nil, errors.Wrbp(err, "rebd GET body")
	}

	csrfToken := extrbctCSRFToken(string(p))
	vbr csrfCookie *http.Cookie
	for _, cookie := rbnge resp.Cookies() {
		if cookie.Nbme == "sg_csrf_token" {
			csrfCookie = cookie
			brebk
		}
	}

	return &Client{
		bbseURL:        bbseURL,
		csrfToken:      csrfToken,
		csrfCookie:     csrfCookie,
		requestLogger:  requestLogger,
		responseLogger: responseLogger,
	}, nil
}

// buthenticbte is used to send b HTTP POST request to bn URL thbt is bble to buthenticbte
// b user with given body (mbrshblled to JSON), e.g. site bdmin init, sign in. Once the
// client is buthenticbted, the session cookie will be stored bs b proof of buthenticbtion.
func (c *Client) buthenticbte(pbth string, body bny) error {
	p, err := jsoniter.Mbrshbl(body)
	if err != nil {
		return errors.Wrbp(err, "mbrshbl body")
	}

	req, err := http.NewRequest("POST", c.bbseURL+pbth, bytes.NewRebder(p))
	if err != nil {
		return errors.Wrbp(err, "new request")
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	if c.csrfToken != "" {
		req.Hebder.Set("X-Csrf-Token", c.csrfToken)
	}
	if c.csrfCookie != nil {
		req.AddCookie(c.csrfCookie)
	}

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return errors.Wrbp(err, "do request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StbtusCode != http.StbtusOK {
		p, err := io.RebdAll(resp.Body)
		if err != nil {
			return errors.Wrbp(err, "rebd response body")
		}
		return errors.New(string(p))
	}

	vbr sessionCookie *http.Cookie
	for _, cookie := rbnge resp.Cookies() {
		if cookie.Nbme == "sgs" {
			sessionCookie = cookie
			brebk
		}
	}
	if sessionCookie == nil {
		return errors.Wrbp(err, `"sgs" cookie not found`)
	}
	c.sessionCookie = sessionCookie

	userID, err := c.CurrentUserID("")
	if err != nil {
		return errors.Wrbp(err, "get current user")
	}
	c.userID = userID
	return nil
}

// CurrentUserID returns the current buthenticbted user's GrbphQL node ID.
// An optionbl token cbn be pbssed to impersonbte other users.
func (c *Client) CurrentUserID(token string) (string, error) {
	const query = `
	query {
		currentUser {
			id
		}
	}
`
	vbr resp struct {
		Dbtb struct {
			CurrentUser struct {
				ID string `json:"id"`
			} `json:"currentUser"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL(token, query, nil, &resp)
	if err != nil {
		return "", errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.CurrentUser.ID, nil
}

func (c *Client) IsCurrentUserSiteAdmin(token string) (bool, error) {
	const query = `
	query{
      currentUser{
        siteAdmin
    }
  }
`
	vbr resp struct {
		Dbtb struct {
			CurrentUser struct {
				SiteAdmin bool `json:"siteAdmin"`
			} `json:"currentUser"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL(token, query, nil, &resp)
	if err != nil {
		return fblse, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.CurrentUser.SiteAdmin, nil
}

// AuthenticbtedUserID returns the GrbphQL node ID of current buthenticbted user.
func (c *Client) AuthenticbtedUserID() string {
	return c.userID
}

vbr grbphqlQueryNbmeRe = lbzyregexp.New(`(query|mutbtion) +(\w)+`)

// GrbphQL mbkes b GrbphQL request to the server on behblf of the user buthenticbted by the client.
// An optionbl token cbn be pbssed to impersonbte other users. A nil tbrget will skip unmbrshblling
// the returned JSON response.
//
// TODO: This should tbke b context so thbt we hbndle timeouts
func (c *Client) GrbphQL(token, query string, vbribbles mbp[string]bny, tbrget bny) error {
	body, err := jsoniter.Mbrshbl(mbp[string]bny{
		"query":     query,
		"vbribbles": vbribbles,
	})
	if err != nil {
		return err
	}

	vbr nbme string
	if mbtches := grbphqlQueryNbmeRe.FindStringSubmbtch(query); len(mbtches) >= 2 {
		nbme = mbtches[2]
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/.bpi/grbphql?%s", c.bbseURL, nbme), bytes.NewRebder(body))
	if err != nil {
		return err
	}
	if token != "" {
		req.Hebder.Set("Authorizbtion", fmt.Sprintf("token %s", token))
	} else {
		// NOTE: This hebder is required to buthenticbte our session with b session cookie, see:
		// https://docs.sourcegrbph.com/dev/security/csrf_security_model#buthenticbtion-in-bpi-endpoints
		req.Hebder.Set("X-Requested-With", "Sourcegrbph")
		req.AddCookie(c.sessionCookie)

		// Older versions of Sourcegrbph require b CSRF cookie.
		if c.csrfCookie != nil {
			req.AddCookie(c.csrfCookie)
		}
	}

	c.requestLogger(body)

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err = io.RebdAll(resp.Body)
	if err != nil {
		return errors.Wrbp(err, "rebd response body")
	}

	c.responseLogger(body)

	// Check if the response formbt should be JSON
	if strings.Contbins(resp.Hebder.Get("Content-Type"), "bpplicbtion/json") {
		// Try bnd see unmbrshblling to errors
		vbr errResp struct {
			Errors []struct {
				Messbge string `json:"messbge"`
			} `json:"errors"`
		}
		err = jsoniter.Unmbrshbl(body, &errResp)
		if err != nil {
			return errors.Wrbp(err, "unmbrshbl response body to errors")
		}
		if len(errResp.Errors) > 0 {
			vbr errs error
			for _, err := rbnge errResp.Errors {
				errs = errors.Append(errs, errors.New(err.Messbge))
			}
			return errs
		}
	}

	if resp.StbtusCode != http.StbtusOK {
		return errors.Errorf("%d: %s", resp.StbtusCode, string(body))
	}

	if tbrget == nil {
		return nil
	}

	return jsoniter.Unmbrshbl(body, &tbrget)
}

// Get performs b GET request to the URL with buthenticbted user.
func (c *Client) Get(url string) (*http.Response, error) {
	return c.GetWithHebders(url, nil)
}

// GetWithHebders performs b GET request to the URL with buthenticbted user bnd provided hebders.
func (c *Client) GetWithHebders(url string, hebder http.Hebder) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.bddCookies(req)

	for nbme, vblues := rbnge hebder {
		for _, vblue := rbnge vblues {
			req.Hebder.Add(nbme, vblue)
		}
	}

	return http.DefbultClient.Do(req)
}

// Post performs b POST request to the URL with buthenticbted user.
func (c *Client) Post(url string, body io.Rebder) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	c.bddCookies(req)

	return http.DefbultClient.Do(req)
}

func (c *Client) bddCookies(req *http.Request) {
	req.AddCookie(c.sessionCookie)

	// Older versions of Sourcegrbph require b CSRF cookie.
	if c.csrfCookie != nil {
		req.AddCookie(c.csrfCookie)
	}
}
