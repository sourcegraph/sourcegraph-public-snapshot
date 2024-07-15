package tenant

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/trace"
)

const (
	// headerKeyTenantID is the header key for the tenant ID.
	headerKeyTenantID = "X-Sourcegraph-Tenant-ID"
)

const (
	// headerValueNoTenant indicates the request has no tenant.
	headerValueNoTenant = "none"
)

var (
	metricIncomingTenants = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_tenants_incoming_requests",
		Help: "Total number of tenants set from incoming requests by tenant type.",
	}, []string{"tenant_type", "path"})

	metricOutgoingTenants = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_tenants_outgoing_requests",
		Help: "Total number of tenants set on outgoing requests by tenant type.",
	}, []string{"tenant_type", "path"})
)

// HTTPTransport is a roundtripper that sets tenants within request context as headers on
// outgoing requests. The attached headers can be picked up and attached to incoming
// request contexts with tenant.HTTPMiddleware.
//
// TODO(@bobheadxi): Migrate to httpcli.Doer and httpcli.Middleware
type HTTPTransport struct {
	RoundTripper http.RoundTripper
}

var _ http.RoundTripper = &HTTPTransport{}

// 🚨 SECURITY: Do not send any PII here.
func (t *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefaultTransport
	}

	// RoundTripper should not modify original request. All the code paths
	// below set a header, so we clone the request immediately.
	req = req.Clone(req.Context())

	tenant := FromContext(req.Context())

	path := getCondensedURLPath(req.URL.Path)
	switch {
	// no tenant set
	case tenant.ID() == 0:
		req.Header.Set(headerKeyTenantID, headerValueNoTenant)
		metricOutgoingTenants.WithLabelValues(headerValueNoTenant, path).Inc()

	default:
		req.Header.Set(headerKeyTenantID, strconv.Itoa(tenant.ID()))
		metricOutgoingTenants.WithLabelValues("tenant", path).Inc()
	}

	return t.RoundTripper.RoundTrip(req)
}

// HTTPMiddleware wraps the given handle func and attaches the actor indicated in incoming
// requests to the request header. This should only be used to wrap internal handlers for
// communication between Sourcegraph services.
//
// 🚨 SECURITY: This should *never* be called to wrap externally accessible handlers (i.e.
// only use for internal endpoints), because internal requests can bypass repository
// permissions checks.
func InternalHTTPMiddleware(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		path := getCondensedURLPath(req.URL.Path)

		idStr := req.Header.Get(headerKeyTenantID)
		switch idStr {
		// Request not associated with a tenant
		case "", headerValueNoTenant:
			metricIncomingTenants.WithLabelValues(headerValueNoTenant, path).Inc()

		// Request associated with authenticated user - add user actor to context
		default:
			uid, err := strconv.Atoi(idStr)
			if err != nil {
				trace.Logger(ctx, logger).
					Warn("invalid user ID in request",
						log.Error(err),
						log.String("uid", idStr))
				metricIncomingTenants.WithLabelValues("invalid", path).Inc()

				// Do not proceed with request
				rw.WriteHeader(http.StatusForbidden)
				_, _ = rw.Write([]byte(fmt.Sprintf("%s was provided, but the value was invalid", headerKeyTenantID)))
				return
			}

			// Valid user
			ctx = withTenant(ctx, uid)
			metricIncomingTenants.WithLabelValues("tenant", path).Inc()
		}

		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// getCondensedURLPath truncates known high-cardinality paths to be used as metric labels in order to reduce the
// label cardinality. This can and should be expanded to include other paths as necessary.
func getCondensedURLPath(urlPath string) string {
	if strings.HasPrefix(urlPath, "/.internal/git/") {
		return "/.internal/git/..."
	}
	if strings.HasPrefix(urlPath, "/git/") {
		return "/git/..."
	}
	return urlPath
}

func ExternalTenantFromHostnameMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		host := strings.SplitN(req.Host, ":", 2)[0]
		tenantID, ok := knownHosts[host]
		if !ok {
			http.Error(rw, fmt.Sprintf("tenant %q not known", host), http.StatusBadRequest)
			return
		}

		ctx = withTenant(ctx, tenantID)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

var knownHosts = map[string]int{
	"erik.sourcegraph.test": 1,
	"sourcegraph.test":      2,
}
