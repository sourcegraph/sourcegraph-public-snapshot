package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	// The metric generated here will be named as "src_gitlab_requests_total".
	requestCounter = metrics.NewRequestMeter("gitlab", "Total number of requests sent to the GitLab API.")

	// Whether debug logging is turned on
	traceEnabled int32 = 0
)

func init() {
	go func() {
		conf.Watch(func() {
			exp := conf.Get().ExperimentalFeatures
			if exp == nil {
				atomic.StoreInt32(&traceEnabled, 0)
				return
			}
			if debugLog := exp.DebugLog; debugLog == nil || !debugLog.ExtsvcGitlab {
				atomic.StoreInt32(&traceEnabled, 0)
				return
			}
			atomic.StoreInt32(&traceEnabled, 1)
		})
	}()
}

func trace(msg string, ctx ...any) {
	if atomic.LoadInt32(&traceEnabled) == 1 {
		log15.Info(fmt.Sprintf("TRACE %s", msg), ctx...)
	}
}

// TokenType is the type of an access token.
type TokenType string

const (
	TokenTypePAT   TokenType = "pat"   // "pat" represents personal access token.
	TokenTypeOAuth TokenType = "oauth" // "oauth" represents OAuth token.
)

// ClientProvider creates GitLab API clients. Each client has separate authentication creds and a
// separate cache, but they share an underlying HTTP client and rate limiter. Callers who want a simple
// unauthenticated API client should use `NewClientProvider(baseURL, transport).GetClient()`.
type ClientProvider struct {
	// The URN of the external service that the client is derived from.
	urn string

	// baseURL is the base URL of GitLab; e.g., https://gitlab.com or https://gitlab.example.com
	baseURL *url.URL

	// httpClient is the underlying the HTTP client to use.
	httpClient httpcli.Doer

	gitlabClients   map[string]*Client
	gitlabClientsMu sync.Mutex

	tokenRefresher oauthutil.TokenRefresher
}

type CommonOp struct {
	// NoCache, if true, will bypass any caching done in this package
	NoCache bool
}

func NewClientProvider(urn string, baseURL *url.URL, cli httpcli.Doer, tokenRefresher oauthutil.TokenRefresher) *ClientProvider {
	if cli == nil {
		cli = httpcli.ExternalDoer
	}
	cli = requestCounter.Doer(cli, func(u *url.URL) string {
		// The 3rd component of the Path (/api/v4/XYZ) mostly maps to the type of API
		// request we are making.
		var category string
		if parts := strings.SplitN(u.Path, "/", 3); len(parts) >= 4 {
			category = parts[3]
		}
		return category
	})

	return &ClientProvider{
		urn:            urn,
		baseURL:        baseURL.ResolveReference(&url.URL{Path: path.Join(baseURL.Path, "api/v4") + "/"}),
		httpClient:     cli,
		gitlabClients:  make(map[string]*Client),
		tokenRefresher: tokenRefresher,
	}
}

// GetAuthenticatorClient returns a client authenticated by the given
// authenticator.
func (p *ClientProvider) GetAuthenticatorClient(a auth.Authenticator) *Client {
	return p.getClient(a)
}

// GetPATClient returns a client authenticated by the personal access token.
func (p *ClientProvider) GetPATClient(personalAccessToken, sudo string) *Client {
	if personalAccessToken == "" {
		return p.getClient(nil)
	}
	return p.getClient(&SudoableToken{Token: personalAccessToken, Sudo: sudo})
}

// GetOAuthClient returns a client authenticated by the OAuth token.
func (p *ClientProvider) GetOAuthClient(oauthToken string) *Client {
	if oauthToken == "" {
		return p.getClient(nil)
	}
	return p.getClient(&auth.OAuthBearerToken{Token: oauthToken})
}

// GetClient returns an unauthenticated client.
func (p *ClientProvider) GetClient() *Client {
	return p.getClient(nil)
}

func (p *ClientProvider) getClient(a auth.Authenticator) *Client {
	p.gitlabClientsMu.Lock()
	defer p.gitlabClientsMu.Unlock()

	key := "<nil>"
	if a != nil {
		key = a.Hash()
	}
	if c, ok := p.gitlabClients[key]; ok {
		return c
	}

	c := p.newClient(a)
	p.gitlabClients[key] = c
	return c
}

// Client is a GitLab API client. Clients are associated with a particular user
// identity, which is defined by the Auth implementation. In addition to the
// generic types provided by the auth package, Client also supports
// SudoableToken: if this is used and its Sudo field is non-empty, then the user
// identity will be the user ID specified by Sudo (rather than the user that
// owns the token).
//
// The Client's cache is keyed by Auth.Hash(). It is NOT keyed by the actual
// user ID that is defined by the authentication method. So if an OAuth token
// and personal access token belong to the same user and there are two
// corresponding Client instances, those Client instances will NOT share the
// same cache. However, two Client instances sharing the exact same values for
// those fields WILL share a cache.
type Client struct {
	// The URN of the external service that the client is derived from.
	urn string

	baseURL          *url.URL
	httpClient       httpcli.Doer
	projCache        *rcache.Cache
	Auth             auth.Authenticator
	rateLimitMonitor *ratelimit.Monitor
	rateLimiter      *ratelimit.InstrumentedLimiter // Our internal rate limiter

	tokenRefresher oauthutil.TokenRefresher
}

// newClient creates a new GitLab API client with an optional personal access token to authenticate requests.
//
// The URL must point to the base URL of the GitLab instance. This is https://gitlab.com for GitLab.com and
// http[s]://[gitlab-hostname] for self-hosted GitLab instances.
//
// See the docstring of Client for the meaning of the parameters.
func (p *ClientProvider) newClient(a auth.Authenticator) *Client {
	return p.NewClientWithTokenRefresher(a, p.tokenRefresher)
}

func (p *ClientProvider) NewClientWithTokenRefresher(a auth.Authenticator, refresher oauthutil.TokenRefresher) *Client {
	// Cache for GitLab project metadata.
	var cacheTTL time.Duration
	if isGitLabDotComURL(p.baseURL) && a == nil {
		cacheTTL = 10 * time.Minute // cache for longer when unauthenticated
	} else {
		cacheTTL = 30 * time.Second
	}
	key := "gl_proj:"
	var tokenHash string
	if a != nil {
		tokenHash = a.Hash()
		key += tokenHash
	}
	projCache := rcache.NewWithTTL(key, int(cacheTTL/time.Second))

	rl := ratelimit.DefaultRegistry.Get(p.urn)
	rlm := ratelimit.DefaultMonitorRegistry.GetOrSet(p.baseURL.String(), tokenHash, "rest", &ratelimit.Monitor{})

	return &Client{
		urn:              p.urn,
		baseURL:          p.baseURL,
		httpClient:       p.httpClient,
		projCache:        projCache,
		Auth:             a,
		rateLimiter:      rl,
		rateLimitMonitor: rlm,
		tokenRefresher:   refresher,
	}
}

func isGitLabDotComURL(baseURL *url.URL) bool {
	hostname := strings.ToLower(baseURL.Hostname())
	return hostname == "gitlab.com" || hostname == "www.gitlab.com"
}

// do is the default method for making API requests and will prepare the correct
// base path.
func (c *Client) do(ctx context.Context, req *http.Request, result any) (responseHeader http.Header, responseCode int, err error) {
	req.URL = c.baseURL.ResolveReference(req.URL)
	return c.doWithBaseURL(ctx, getOAuthContext(c.baseURL.String()), req, result)
}

// doWithBaseURL doesn't amend the request URL. When an OAuth Bearer token is
// used for authentication, it will try to refresh the token and retry the same
// request when the token has expired.
func (c *Client) doWithBaseURL(ctx context.Context, oauthContext *oauthutil.OAuthContext, req *http.Request, result any) (responseHeader http.Header, responseCode int, err error) {
	var responseStatus string

	span, ctx := ot.StartSpanFromContext(ctx, "GitLab")
	span.SetTag("URL", req.URL.String())
	defer func() {
		if err != nil {
			span.SetTag("error", err.Error())
		}
		if responseStatus != "" {
			span.SetTag("status", responseStatus)
		}
		span.Finish()
	}()

	if c.rateLimiter != nil {
		err = c.rateLimiter.Wait(ctx)
		if err != nil {
			return nil, 0, errors.Wrap(err, "rate limit")
		}
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var code int
	var header http.Header
	var body []byte

	oauthAuther, ok := c.Auth.(*auth.OAuthBearerToken)
	if ok && oauthContext != nil {
		code, header, body, err = oauthutil.DoRequest(ctx, c.httpClient, req, oauthAuther, c.tokenRefresher, *oauthContext)
		if err != nil {
			trace("GitLab API error", "method", req.Method, "url", req.URL.String(), "err", err)
			return nil, 0, errors.Wrap(err, "do request with retry and refresh")
		}
	} else {
		if c.Auth != nil {
			if err := c.Auth.Authenticate(req); err != nil {
				return nil, 0, errors.Wrap(err, "authenticating request")
			}
		}

		resp, err := c.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			trace("GitLab API error", "method", req.Method, "url", req.URL.String(), "err", err)
			return nil, 0, errors.Wrap(err, "do request")
		}
		defer func() { _ = resp.Body.Close() }()

		code = resp.StatusCode
		header = resp.Header

		// We swallow the error here, because we don't want to fail. Parsing the body
		// is just optional to provide some more context.
		body, _ = io.ReadAll(resp.Body)
	}
	trace("GitLab API", "method", req.Method, "url", req.URL.String(), "respCode", code)

	if code < 200 || code >= 400 {
		err := NewHTTPError(code, body)
		return nil, code, errors.Wrap(err, fmt.Sprintf("unexpected response from GitLab API (%s)", req.URL))
	}

	return header, code, json.Unmarshal(body, result)
}

// RateLimitMonitor exposes the rate limit monitor.
func (c *Client) RateLimitMonitor() *ratelimit.Monitor {
	return c.rateLimitMonitor
}

func (c *Client) WithAuthenticator(a auth.Authenticator) *Client {
	tokenHash := a.Hash()

	cc := *c
	cc.rateLimiter = ratelimit.DefaultRegistry.Get(c.urn)
	cc.rateLimitMonitor = ratelimit.DefaultMonitorRegistry.GetOrSet(cc.baseURL.String(), tokenHash, "rest", &ratelimit.Monitor{})
	cc.Auth = a

	return &cc
}

func (c *Client) ValidateToken(ctx context.Context) error {
	req, err := http.NewRequest(http.MethodGet, "user", nil)
	if err != nil {
		return err
	}
	v := struct{}{}
	_, _, err = c.do(ctx, req, &v)
	return err
}

func (c *Client) GetAuthenticatedUserOAuthScopes(ctx context.Context) ([]string, error) {
	// The oauth token info path is non standard so we need to build it manually
	// without the default `/api/v4` prefix
	u, _ := url.Parse(c.baseURL.String())
	u.Path = "oauth/token/info"

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	v := struct {
		Scopes []string `json:"scopes,omitempty"`
	}{}

	_, _, err = c.doWithBaseURL(ctx, getOAuthContext(c.baseURL.String()), req, &v)
	if err != nil {
		return nil, errors.Wrap(err, "getting oauth scopes")
	}
	return v.Scopes, nil
}

type HTTPError struct {
	code int
	body []byte
}

func NewHTTPError(code int, body []byte) HTTPError {
	return HTTPError{
		code: code,
		body: body,
	}
}

func (err HTTPError) Code() int {
	return err.code
}

func (err HTTPError) Message() string {
	var errBody struct {
		Message string `json:"message"`
	}
	// Swallow error, decoding the body as
	_ = json.Unmarshal(err.body, &errBody)
	return errBody.Message
}

func (err HTTPError) Error() string {
	return fmt.Sprintf("HTTP error status %d", err.code)
}

func (err HTTPError) Unauthorized() bool {
	return err.code == http.StatusUnauthorized
}

func (err HTTPError) Forbidden() bool {
	return err.code == http.StatusForbidden
}

func (err HTTPError) IsTemporary() bool {
	return err.code == http.StatusTooManyRequests
}

// HTTPErrorCode returns err's HTTP status code, if it is an HTTP error from
// this package. Otherwise it returns 0.
func HTTPErrorCode(err error) int {
	var e HTTPError
	if errors.As(err, &e) {
		return e.Code()
	}

	return 0
}

// IsNotFound reports whether err is a GitLab API error of type NOT_FOUND, the equivalent cached
// response error, or HTTP 404.
func IsNotFound(err error) bool {
	return errors.HasType(err, &ProjectNotFoundError{}) ||
		errors.Is(err, ErrMergeRequestNotFound) ||
		HTTPErrorCode(err) == http.StatusNotFound
}

// ErrMergeRequestNotFound is when the requested GitLab merge request is not found.
var ErrMergeRequestNotFound = errors.New("GitLab merge request not found")

// ErrProjectNotFound is when the requested GitLab project is not found.
var ErrProjectNotFound = &ProjectNotFoundError{}

// ProjectNotFoundError is when the requested GitHub repository is not found.
type ProjectNotFoundError struct {
	Name string
}

func (e ProjectNotFoundError) Error() string {
	if e.Name == "" {
		return "GitLab project not found"
	}
	return fmt.Sprintf("GitLab project %q not found", e.Name)
}

func (e ProjectNotFoundError) NotFound() bool { return true }

var MockGetOAuthContext func() *oauthutil.OAuthContext

// getOAuthContext matches the corresponding auth provider using the given
// baseURL and returns the oauthutil.OAuthContext of it.
func getOAuthContext(baseURL string) *oauthutil.OAuthContext {
	if MockGetOAuthContext != nil {
		return MockGetOAuthContext()
	}

	for _, authProvider := range conf.SiteConfig().AuthProviders {
		if authProvider.Gitlab != nil {
			p := authProvider.Gitlab
			glURL := strings.TrimSuffix(p.Url, "/")
			if !strings.HasPrefix(baseURL, glURL) {
				continue
			}

			return &oauthutil.OAuthContext{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  glURL + "/oauth/authorize",
					TokenURL: glURL + "/oauth/token",
				},
				Scopes: RequestedOAuthScopes(p.ApiScope, nil),
			}
		}
	}
	return nil
}

// ProjectArchivedError is returned when a request cannot be performed due to the
// GitLab project being archived.
type ProjectArchivedError struct{ Name string }

func (ProjectArchivedError) Archived() bool { return true }

func (e ProjectArchivedError) Error() string {
	if e.Name == "" {
		return "GitLab project is archived"
	}
	return fmt.Sprintf("GitLab project %q is archived", e.Name)
}

func (ProjectArchivedError) NonRetryable() bool { return true }
