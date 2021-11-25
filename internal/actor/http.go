package actor

import (
	"net/http"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// headerActorUID is the header key for the actor's user ID.
const headerActorUID = "X-Sourcegraph-Actor-ID"

const (
	// internalActorHeaderValue indicates the request uses an internal actor.
	internalActorHeaderValue = "internal"
	// noActorHeaderValue indicates the request has no actor.
	noActorHeaderValue = "none"
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
// outgoing requests.
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
		req.Header.Set(headerActorUID, internalActorHeaderValue)
		metricOutgoingActors.WithLabelValues(internalActorHeaderValue).Inc()

	// Indicate this is an authenticated user
	case actor.IsAuthenticated():
		req.Header.Set(headerActorUID, actor.UIDString())
		metricOutgoingActors.WithLabelValues("user").Inc()

	// Indicate no actor is associated with request
	default:
		req.Header.Set(headerActorUID, noActorHeaderValue)
		metricOutgoingActors.WithLabelValues(noActorHeaderValue).Inc()
	}

	return t.RoundTripper.RoundTrip(req)
}

// HTTPMiddleware wraps the given handle func and attaches the actor indicated in incoming
// requests to the request header.
//
// 🚨 SECURITY: This should *never* be called to wrap externally accessible handlers (i.e.
// only use for internal endpoints), because internal requests can bypass repository
// permissions checks.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		uidStr := req.Header.Get(headerActorUID)
		switch uidStr {
		// Request associated with internal actor - add internal actor to context
		case internalActorHeaderValue:
			ctx = WithInternalActor(ctx)
			metricIncomingActors.WithLabelValues(internalActorHeaderValue).Inc()

		// Request not associated with any actor
		case "", noActorHeaderValue:
			metricIncomingActors.WithLabelValues(noActorHeaderValue).Inc()

		// Request associated with authenticated user - add user actor to context
		default:
			uid, err := strconv.Atoi(uidStr)
			if err != nil {
				log15.Warn("invalid user ID in request",
					"error", err,
					"uid", uidStr)
				metricIncomingActors.WithLabelValues("invalid").Inc()
				break
			}

			// Valid user, add to context
			actor := FromUser(int32(uid))
			ctx = WithActor(ctx, actor)
			metricIncomingActors.WithLabelValues("user").Inc()
		}

		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
