package background

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Metrics struct {
	numSomething prometheus.Counter
}

func NewMetrics(observationCtx *observation.Context) *Metrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	// TODO
	numSomething := counter(
		"src_codeintel_background_TODO_total",
		"TODO",
	)

	return &Metrics{
		numSomething: numSomething,
	}
}
