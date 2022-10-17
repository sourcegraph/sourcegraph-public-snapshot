package gqltestutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NeedsSiteInit returns true if the instance hasn't done "Site admin init" step.
func NeedsSiteInit(baseURL string) (bool, string, error) {
	resp, err := http.Get(baseURL + "/sign-in")
	if err != nil {
		return false, "", errors.Wrap(err, "get page")
	}
	defer func() { _ = resp.Body.Close() }()

	p, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", errors.Wrap(err, "read body")
	}
	return strings.Contains(string(p), `"needsSiteInit":true`), string(p), nil
}

// SiteAdminInit initializes the instance with given admin account.
// It returns an authenticated client as the admin for doing testing.
func SiteAdminInit(baseURL, email, username, password string) (*Client, error) {
	return authenticate(baseURL, "/-/site-init", map[string]string{
		"email":    email,
		"username": username,
		"password": password,
	})
}

// SignUp signs up a new user with given credentials.
// It returns an authenticated client as the user for doing testing.
func SignUp(baseURL, email, username, password string) (*Client, error) {
	return authenticate(baseURL, "/-/sign-up", map[string]string{
		"email":    email,
		"username": username,
		"password": password,
	})
}

func SignUpOrSignIn(baseURL, email, username, password string) (*Client, error) {
	client, err := SignUp(baseURL, email, username, password)
	if err != nil {
		return SignIn(baseURL, email, password)
	}
	return client, err
}

// SignIn performs the sign in with given user credentials.
// It returns an authenticated client as the user for doing testing.
func SignIn(baseURL, email, password string) (*Client, error) {
	return authenticate(baseURL, "/-/sign-in", map[string]string{
		"email":    email,
		"password": password,
	})
}

// authenticate initializes an authenticated client with given request body.
func authenticate(baseURL, path string, body any) (*Client, error) {
	client, err := NewClient(baseURL, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "new client")
	}

	err = client.authenticate(path, body)
	if err != nil {
		return nil, errors.Wrap(err, "authenticate")
	}

	return client, nil
}

// extractCSRFToken extracts CSRF token from HTML response body.
func extractCSRFToken(body string) string {
	anchor := `X-Csrf-Token":"`
	i := strings.Index(body, anchor)
	if i == -1 {
		return ""
	}

	j := strings.Index(body[i+len(anchor):], `","`)
	if j == -1 {
		return ""
	}

	return body[i+len(anchor) : i+len(anchor)+j]
}

// Client is an authenticated client for a Sourcegraph user for doing e2e testing.
// The user may or may not be a site admin depends on how the client is instantiated.
// It works by simulating how the browser would send HTTP requests to the server.
type Client struct {
	baseURL       string
	csrfToken     string
	csrfCookie    *http.Cookie
	sessionCookie *http.Cookie

	userID         string
	requestLogger  LogFunc
	responseLogger LogFunc
}

type LogFunc func(payload []byte)

func noopLog(payload []byte) {}

// NewClient instantiates a new client by performing a GET request then obtains the
// CSRF token and cookie from its response, if there is one (old versions of Sourcegraph only).
// If request- or responseLogger are provided, the request and response bodies, respectively,
// will be written to them for any GraphQL requests only.
func NewClient(baseURL string, requestLogger, responseLogger LogFunc) (*Client, error) {
	if requestLogger == nil {
		requestLogger = noopLog
	}
	if responseLogger == nil {
		responseLogger = noopLog
	}

	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "get URL")
	}
	defer func() { _ = resp.Body.Close() }()

	p, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read GET body")
	}

	csrfToken := extractCSRFToken(string(p))
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "sg_csrf_token" {
			csrfCookie = cookie
			break
		}
	}

	return &Client{
		baseURL:        baseURL,
		csrfToken:      csrfToken,
		csrfCookie:     csrfCookie,
		requestLogger:  requestLogger,
		responseLogger: responseLogger,
	}, nil
}

// authenticate is used to send a HTTP POST request to an URL that is able to authenticate
// a user with given body (marshalled to JSON), e.g. site admin init, sign in. Once the
// client is authenticated, the session cookie will be stored as a proof of authentication.
func (c *Client) authenticate(path string, body any) error {
	p, err := jsoniter.Marshal(body)
	if err != nil {
		return errors.Wrap(err, "marshal body")
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(p))
	if err != nil {
		return errors.Wrap(err, "new request")
	}
	req.Header.Set("Content-Type", "application/json")
	if c.csrfToken != "" {
		req.Header.Set("X-Csrf-Token", c.csrfToken)
	}
	if c.csrfCookie != nil {
		req.AddCookie(c.csrfCookie)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "do request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		p, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
		return errors.New(string(p))
	}

	var sessionCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "sgs" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		return errors.Wrap(err, `"sgs" cookie not found`)
	}
	c.sessionCookie = sessionCookie

	userID, err := c.CurrentUserID("")
	if err != nil {
		return errors.Wrap(err, "get current user")
	}
	c.userID = userID
	return nil
}

// CurrentUserID returns the current authenticated user's GraphQL node ID.
// An optional token can be passed to impersonate other users.
func (c *Client) CurrentUserID(token string) (string, error) {
	const query = `
	query {
		currentUser {
			id
		}
	}
`
	var resp struct {
		Data struct {
			CurrentUser struct {
				ID string `json:"id"`
			} `json:"currentUser"`
		} `json:"data"`
	}
	err := c.GraphQL(token, query, nil, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.CurrentUser.ID, nil
}

func (c *Client) IsCurrentUserSiteAdmin(token string) (bool, error) {
	const query = `
	query{
      currentUser{
        siteAdmin
    }
  }
`
	var resp struct {
		Data struct {
			CurrentUser struct {
				SiteAdmin bool `json:"siteAdmin"`
			} `json:"currentUser"`
		} `json:"data"`
	}
	err := c.GraphQL(token, query, nil, &resp)
	if err != nil {
		return false, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.CurrentUser.SiteAdmin, nil
}

// AuthenticatedUserID returns the GraphQL node ID of current authenticated user.
func (c *Client) AuthenticatedUserID() string {
	return c.userID
}

var graphqlQueryNameRe = lazyregexp.New(`(query|mutation) +(\w)+`)

// GraphQL makes a GraphQL request to the server on behalf of the user authenticated by the client.
// An optional token can be passed to impersonate other users. A nil target will skip unmarshalling
// the returned JSON response.
func (c *Client) GraphQL(token, query string, variables map[string]any, target any) error {
	body, err := jsoniter.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return err
	}

	var name string
	if matches := graphqlQueryNameRe.FindStringSubmatch(query); len(matches) >= 2 {
		name = matches[2]
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/.api/graphql?%s", c.baseURL, name), bytes.NewReader(body))
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	} else {
		// NOTE: This header is required to authenticate our session with a session cookie, see:
		// https://docs.sourcegraph.com/dev/security/csrf_security_model#authentication-in-api-endpoints
		req.Header.Set("X-Requested-With", "Sourcegraph")
		req.AddCookie(c.sessionCookie)

		// Older versions of Sourcegraph require a CSRF cookie.
		if c.csrfCookie != nil {
			req.AddCookie(c.csrfCookie)
		}
	}

	c.requestLogger(body)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read response body")
	}

	c.responseLogger(body)

	// Check if the response format should be JSON
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		// Try and see unmarshalling to errors
		var errResp struct {
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		err = jsoniter.Unmarshal(body, &errResp)
		if err != nil {
			return errors.Wrap(err, "unmarshal response body to errors")
		}
		if len(errResp.Errors) > 0 {
			var errs error
			for _, err := range errResp.Errors {
				errs = errors.Append(errs, errors.New(err.Message))
			}
			return errs
		}
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("%d: %s", resp.StatusCode, string(body))
	}

	if target == nil {
		return nil
	}

	return jsoniter.Unmarshal(body, &target)
}

// Get performs a GET request to the URL with authenticated user.
func (c *Client) Get(url string) (*http.Response, error) {
	return c.GetWithHeaders(url, nil)
}

// GetWithHeaders performs a GET request to the URL with authenticated user and provided headers.
func (c *Client) GetWithHeaders(url string, header http.Header) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.addCookies(req)

	for name, values := range header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	return http.DefaultClient.Do(req)
}

// Post performs a POST request to the URL with authenticated user.
func (c *Client) Post(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	c.addCookies(req)

	return http.DefaultClient.Do(req)
}

func (c *Client) addCookies(req *http.Request) {
	req.AddCookie(c.sessionCookie)

	// Older versions of Sourcegraph require a CSRF cookie.
	if c.csrfCookie != nil {
		req.AddCookie(c.csrfCookie)
	}
}
