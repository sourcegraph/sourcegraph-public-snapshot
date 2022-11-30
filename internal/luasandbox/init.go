package luasandbox

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func GetService() *Service {
	svc, _ := initServiceMemo.Init()
	return svc
}

var initServiceMemo = memo.NewMemoizedConstructor(func() (*Service, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("luasandbox", ""),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	return newService(observationContext), nil
})
