package background

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Metrics struct {
	numTodo prometheus.Counter
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

	numTodo := counter(
		"src_codeintel_sentinel_todo_total",
		"TODO",
	)

	return &Metrics{
		numTodo: numTodo,
	}
}
