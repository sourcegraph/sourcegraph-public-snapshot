package overhead

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// HTTPMiddleware wraps all downstream handlers and exposes a histogram measuring the difference between the Cody Gateway API latency and the backend API latency.
func HTTPMiddleware(apiHistogram metric.Int64Histogram, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		r = r.WithContext(context.WithValue(r.Context(), overheadKey, &RequestLatencyData{}))

		handler.ServeHTTP(w, r)

		latency := time.Since(start)
		o := FromContext(r.Context())
		if string(o.Feature) == "" {
			// don't report overhead for requests that don't have a feature
			return
		}
		overheadMs := latency.Milliseconds() - o.UpstreamLatency.Milliseconds()
		apiHistogram.Record(context.Background(), overheadMs, metric.WithAttributeSet(attribute.NewSet(
			attribute.String("provider", o.Provider),
			attribute.String("feature", string(o.Feature)),
			attribute.String("stream", strconv.FormatBool(o.Stream)))))
	})
}

type RequestLatencyData struct {
	UpstreamLatency time.Duration
	Feature         codygateway.Feature
	Provider        string
	Stream          bool
}

type contextKey int

const overheadKey contextKey = iota

func FromContext(ctx context.Context) *RequestLatencyData {
	return ctx.Value(overheadKey).(*RequestLatencyData)
}
