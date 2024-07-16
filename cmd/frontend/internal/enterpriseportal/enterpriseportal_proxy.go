package enterpriseportal

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/std"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/audit"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

var (
	EnterprisePortalProd  = mustParseURL("https://enterprise-portal.sourcegraph.com")
	EnterprisePortalDev   = mustParseURL("https://enterprise-portal.sgdev.org")
	EnterprisePortalLocal = mustParseURL("http://127.0.0.1:6081")
)

type SiteAdminProxy struct {
	db    database.DB
	proxy http.Handler
}

type SAMSConfig struct {
	sams.ConnConfig
	ClientID     string
	ClientSecret string
	Scopes       []scopes.Scope
}

var _ http.Handler = (*SiteAdminProxy)(nil)

// NewSiteAdminProxy allows Sourcegraph.com to proxy requests to Enterprise Portal
// on behalf of site admins.
//
// When https://linear.app/sourcegraph/project/kr-p1-streamlined-role-assignment-via-sams-and-entitle-2f118b3f9d4c/overview
// is shipped, we will be able to use SAMS as the source-of-truth for who is
// a site admin in Sourcegraph.com. Then, we can allow Enterprise Portal to
// accept SAMS cookie auth directly, and remove this proxy.
func NewSiteAdminProxy(
	logger log.Logger,
	db database.DB,
	samsConfig SAMSConfig,
	pathPrefix string,
	target *url.URL,
) *SiteAdminProxy {
	if samsConfig.ClientID == "" && samsConfig.ClientSecret == "" {
		logger.Info("proxy disabled")
		return &SiteAdminProxy{
			db: db,
			proxy: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("not configured"))
				w.WriteHeader(http.StatusNotImplemented)
			}),
		}
	}

	logger = logger.With(
		log.String("sams.url", pointers.Deref(samsConfig.APIURL, samsConfig.ExternalURL)),
		log.String("sams.clientID", pointers.Deref(samsConfig.APIURL, samsConfig.ClientID)),
	)
	clientCredentials := sams.ClientCredentialsTokenSource(
		samsConfig.ConnConfig,
		samsConfig.ClientID,
		samsConfig.ClientSecret,
		samsConfig.Scopes,
	)
	if _, err := clientCredentials.Token(); err != nil {
		// Only log the error, as it may be transient, but good to have a record
		// of a potential configuration failure.
		logger.Error("token healthcheck failed",
			log.Error(err),
			log.Strings("scopes", scopes.ToStrings(samsConfig.Scopes)))
	}
	return newSiteAdminProxy(
		logger,
		db,
		clientCredentials,
		pathPrefix,
		target,
	)
}

// newSiteAdminProxy is used for testing the proxy, and accepts interfaces instead.
func newSiteAdminProxy(
	logger log.Logger,
	db database.DB,
	clientCredentials oauth2.TokenSource,
	pathPrefix string,
	target *url.URL,
) *SiteAdminProxy {
	transport := httpcli.UncachedExternalClient.Transport

	// We support targeting a local instance in dev.
	if target.Hostname() == "127.0.0.1" {
		transport = httpcli.InternalClient.Transport
	}

	return &SiteAdminProxy{
		db: db,
		proxy: &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				// We must set the Host header to the target host, otherwise
				// Cloudflare might block us.
				req.Host = target.Host

				// Rewrite the URL to point to the target.
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path = strings.TrimPrefix(req.URL.Path, pathPrefix)
				req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, pathPrefix)

				// Do not pass along tokens or cookies - we add new auth via
				// the SAMS client credentials oauth2 transport.
				req.Header.Del("authorization")
				req.Header.Del("cookie")
			},
			Transport: &oauth2.Transport{
				Source: clientCredentials, // authenticate with SAMS client credentials
				Base:   transport,
			},

			ErrorLog: std.NewLogger(logger, log.LevelWarn),
			ModifyResponse: func(resp *http.Response) error {
				audit.Log(resp.Request.Context(), logger, audit.Record{
					Entity: "enterpriseportal-proxy",
					Action: "access",
					Fields: []log.Field{
						log.String("proxied.scheme", resp.Request.URL.Scheme),
						log.String("proxied.host", resp.Request.URL.Host),
						log.String("proxied.path", resp.Request.URL.Path),
						log.Int("response.statusCode", resp.StatusCode),
					},
				})
				return nil
			},
		},
	}
}

func (p *SiteAdminProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 🚨 SECURITY: Only allow site admins to access Enterprise Portal through
	// the shared credentials proxy.
	act := actor.FromContext(r.Context())
	if err := auth.CheckCurrentActorIsSiteAdmin(act, p.db); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	p.proxy.ServeHTTP(w, r)
}

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
