package luasandbox

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	svc     *Service
	svcOnce sync.Once
)

func GetService() *Service {
	svcOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("luasandbox", ""),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		svc = newService(
			observationContext,
		)
	})

	return svc
}
