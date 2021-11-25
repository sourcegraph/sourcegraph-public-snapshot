package actor

import (
	"net/http"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// headerActorUID is the header key for the actor's user ID.
const headerActorUID = "X-Sourcegraph-Actor-UID"

const (
	// internalActorHeaderValue indicates the request uses an internal actor.
	internalActorHeaderValue = "internal"
	// noActorHeaderValue indicates the request has no actor.
	noActorHeaderValue = "none"
)

// HTTPTransport is a roundtripper that sets actors within request context as headers on
// outgoing requests.
type HTTPTransport struct {
	RoundTripper http.RoundTripper
}

var _ http.RoundTripper = &HTTPTransport{}

var metricOutgoingActors = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_actors_on_outgoing_request",
	Help: "Total number of actors set on outgoing requests.",
}, []string{"actor"})

func (t *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefaultTransport
	}

	actor := FromContext(req.Context())
	if actor.IsInternal() {
		// Indicate this is an internal user
		req.Header.Set(headerActorUID, internalActorHeaderValue)
		metricOutgoingActors.WithLabelValues(internalActorHeaderValue).Inc()
	} else if actor.IsAuthenticated() {
		// Indicate this is an authenticated user
		req.Header.Set(headerActorUID, actor.UIDString())
		metricOutgoingActors.WithLabelValues("user").Inc()
	} else {
		// Indicate no actor is associated with request
		req.Header.Set(headerActorUID, noActorHeaderValue)
		metricOutgoingActors.WithLabelValues(noActorHeaderValue).Inc()
	}

	return t.RoundTripper.RoundTrip(req)
}

var metricIncomingActors = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_actors_from_incoming_request",
	Help: "Total number of actors set from incoming requests.",
}, []string{"actor"})

// HTTPMiddleware wraps the given handle func and attaches the actor indicated in incoming
// requests to the request header.
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
			} else {
				// Valid user, add to context
				actor := FromUser(int32(uid))
				ctx = WithActor(ctx, actor)
				metricIncomingActors.WithLabelValues("user").Inc()
			}
		}

		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
