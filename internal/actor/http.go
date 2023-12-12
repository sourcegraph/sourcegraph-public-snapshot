package actor

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/cookie"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

const (
	// headerKeyActorUID is the header key for the actor's user ID.
	headerKeyActorUID = "X-Sourcegraph-Actor-UID"

	// headerKeyAnonymousActorUID is an optional header to propagate the
	// anonymous UID of an unauthenticated actor.
	headerKeyActorAnonymousUID = "X-Sourcegraph-Actor-Anonymous-UID"
)

const (
	// headerValueInternalActor indicates the request uses an internal actor.
	headerValueInternalActor = "internal"
	// headerValueNoActor indicates the request has no actor.
	headerValueNoActor = "none"
)

const (
	// metricActorTypeUser is a label indicating a request was in the context of a user.
	// We do not record actual user IDs as metric labels to limit cardinality.
	metricActorTypeUser = "user"
	// metricTypeUserActor is a label indicating a request was in the context of an internal actor.
	metricActorTypeInternal = headerValueInternalActor
	// metricActorTypeNone is a label indicating a request was in the context of an internal actor.
	metricActorTypeNone = headerValueNoActor
	// metricActorTypeInvalid is a label indicating a request was in the context of an internal actor.
	metricActorTypeInvalid = "invalid"
)

var (
	metricIncomingActors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_actors_incoming_requests",
		Help: "Total number of actors set from incoming requests by actor type.",
	}, []string{"actor_type", "path"})

	metricOutgoingActors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_actors_outgoing_requests",
		Help: "Total number of actors set on outgoing requests by actor type.",
	}, []string{"actor_type", "path"})
)

// HTTPTransport is a roundtripper that sets actors within request context as headers on
// outgoing requests. The attached headers can be picked up and attached to incoming
// request contexts with actor.HTTPMiddleware.
//
// 🚨 SECURITY: Wherever possible, prefer to act in the context of a specific user rather
// than as an internal actor, which can grant a lot of access in some cases.
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

	actor := FromContext(req.Context())
	path := getCondensedURLPath(req.URL.Path)
	switch {
	// Indicate this is an internal user
	case actor.IsInternal():
		req.Header.Set(headerKeyActorUID, headerValueInternalActor)
		metricOutgoingActors.WithLabelValues(metricActorTypeInternal, path).Inc()

	// Indicate this is an authenticated user
	case actor.IsAuthenticated():
		req.Header.Set(headerKeyActorUID, actor.UIDString())
		metricOutgoingActors.WithLabelValues(metricActorTypeUser, path).Inc()

	// Indicate no authenticated actor is associated with request
	default:
		req.Header.Set(headerKeyActorUID, headerValueNoActor)
		if actor.AnonymousUID != "" {
			req.Header.Set(headerKeyActorAnonymousUID, actor.AnonymousUID)
		}
		metricOutgoingActors.WithLabelValues(metricActorTypeNone, path).Inc()
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
func HTTPMiddleware(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		uidStr := req.Header.Get(headerKeyActorUID)
		path := getCondensedURLPath(req.URL.Path)
		switch uidStr {
		// Request associated with internal actor - add internal actor to context
		//
		// 🚨 SECURITY: Wherever possible, prefer to set the actor ID explicitly through
		// actor.HTTPTransport or similar, since assuming internal actor grants a lot of
		// access in some cases.
		case headerValueInternalActor:
			ctx = WithInternalActor(ctx)
			metricIncomingActors.WithLabelValues(metricActorTypeInternal, path).Inc()

		// Request not associated with an authenticated user
		case "", headerValueNoActor:
			// Even though the current user is not authenticated, we may still have an
			// anonymous UID to propagate.
			if anonymousUID := req.Header.Get(headerKeyActorAnonymousUID); anonymousUID != "" {
				ctx = WithActor(ctx, FromAnonymousUser(anonymousUID))
			}
			metricIncomingActors.WithLabelValues(metricActorTypeNone, path).Inc()

		// Request associated with authenticated user - add user actor to context
		default:
			uid, err := strconv.Atoi(uidStr)
			if err != nil {
				trace.Logger(ctx, logger).
					Warn("invalid user ID in request",
						log.Error(err),
						log.String("uid", uidStr))
				metricIncomingActors.WithLabelValues(metricActorTypeInvalid, path).Inc()

				// Do not proceed with request
				rw.WriteHeader(http.StatusForbidden)
				_, _ = rw.Write([]byte(fmt.Sprintf("%s was provided, but the value was invalid", headerKeyActorUID)))
				return
			}

			// Valid user, add to context
			actor := FromUser(int32(uid))
			ctx = WithActor(ctx, actor)
			metricIncomingActors.WithLabelValues(metricActorTypeUser, path).Inc()
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

// AnonymousUIDMiddleware sets the actor to an unauthenticated actor with an anonymousUID
// from the cookie if it exists. It will not overwrite an existing actor.
func AnonymousUIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var anonymousUID string

		// Get from cookie if available, otherwise get from header
		if cookieAnonymousUID, ok := cookie.AnonymousUID(req); ok {
			anonymousUID = cookieAnonymousUID
		} else if headerAnonymousUID := req.Header.Get(headerKeyActorAnonymousUID); headerAnonymousUID != "" {
			anonymousUID = headerAnonymousUID
		}

		// Don't clobber an existing authenticated actor
		a := FromContext(req.Context())
		if !a.IsAuthenticated() && !a.IsInternal() {
			// If we found an anonymous UID, use that as the actor context
			ctx := req.Context()
			if anonymousUID != "" {
				ctx = WithActor(ctx, FromAnonymousUser(anonymousUID))
			}
			next.ServeHTTP(rw, req.WithContext(ctx))
			return
		}

		// Otherwise, update the current actor. This won't overwrite an authenticated actor.
		if anonymousUID != "" {
			a.AnonymousUID = anonymousUID
		}

		next.ServeHTTP(rw, req)
	})
}
