package httpapi

import "github.com/prometheus/client_golang/prometheus"

// These are different to the existing measurements in the language
// processors, since it will only measure actual user interactions.
var universeSuccessCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "httpapi",
	Name:      "universe_success",
	Help:      "Counts success of universe from user operations.",
}, []string{"method"})
var universeErrorCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "httpapi",
	Name:      "universe_error",
	Help:      "Counts errors of universe from user operations",
}, []string{"method"})

func universeObserve(method string, err error) {
	if err == nil {
		universeSuccessCounter.WithLabelValues(method).Inc()
	} else {
		universeErrorCounter.WithLabelValues(method).Inc()
	}
}
