package actor

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	// headerKeyActorUID is the header key for the actor's user ID.
	headerKeyActorUID = "X-Sourcegraph-Actor-UID"
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
	}, []string{"actor_type"})

	metricOutgoingActors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_actors_outgoing_requests",
		Help: "Total number of actors set on outgoing requests by actor type.",
	}, []string{"actor_type"})
)

// HTTPTransport is a roundtripper that sets actors within request context as headers on
// outgoing requests. The attached headers can be picked up and attached to incoming
// request contexts with actor.HTTPMiddleware.
//
// 🚨 SECURITY: Wherever possible, prefer to act in the context of a specific user rather
// than as an internal actor, which can grant a lot of access in some cases.
type HTTPTransport struct {
	RoundTripper http.RoundTripper
}

var _ http.RoundTripper = &HTTPTransport{}

func (t *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefaultTransport
	}

	actor := FromContext(req.Context())
	switch {
	// Indicate this is an internal user
	case actor.IsInternal():
		req.Header.Set(headerKeyActorUID, headerValueInternalActor)
		metricOutgoingActors.WithLabelValues(metricActorTypeInternal).Inc()

	// Indicate this is an authenticated user
	case actor.IsAuthenticated():
		req.Header.Set(headerKeyActorUID, actor.UIDString())
		metricOutgoingActors.WithLabelValues(metricActorTypeUser).Inc()

	// Indicate no actor is associated with request
	default:
		req.Header.Set(headerKeyActorUID, headerValueNoActor)
		metricOutgoingActors.WithLabelValues(metricActorTypeNone).Inc()
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
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		uidStr := req.Header.Get(headerKeyActorUID)
		switch uidStr {
		// Request associated with internal actor - add internal actor to context
		//
		// 🚨 SECURITY: Wherever possible, prefer to set the actor ID explicitly through
		// actor.HTTPTransport or similar, since assuming internal actor grants a lot of
		// access in some cases.
		case headerValueInternalActor:
			ctx = WithInternalActor(ctx)
			metricIncomingActors.WithLabelValues(metricActorTypeInternal).Inc()

		// Request not associated with any actor
		case "", headerValueNoActor:
			metricIncomingActors.WithLabelValues(metricActorTypeNone).Inc()

		// Request associated with authenticated user - add user actor to context
		default:
			uid, err := strconv.Atoi(uidStr)
			if err != nil {
				log15.Warn("invalid user ID in request",
					"error", err,
					"uid", uidStr)
				metricIncomingActors.WithLabelValues(metricActorTypeInvalid).Inc()

				// Do not proceed with request
				rw.WriteHeader(http.StatusForbidden)
				_, _ = rw.Write([]byte(fmt.Sprintf("%s was provided, but the value was invalid", headerKeyActorUID)))
				return
			}

			// Valid user, add to context
			actor := FromUser(int32(uid))
			ctx = WithActor(ctx, actor)
			metricIncomingActors.WithLabelValues(metricActorTypeUser).Inc()
		}

		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
