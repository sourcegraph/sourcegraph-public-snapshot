package tracer

import (
	"log"
	"reflect"

	"github.com/opentracing/opentracing-go"
	sglog "github.com/sourcegraph/log"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegermetrics "github.com/uber/jaeger-lib/metrics"
)

func configureJaeger(resource sglog.Resource) (opentracing.Tracer, error) {
	cfg, err := jaegercfg.FromEnv()
	cfg.ServiceName = resource.Name
	if err != nil {
		return nil, err
	}
	cfg.Tags = append(
		cfg.Tags,
		opentracing.Tag{Key: "service.version", Value: resource.Version},
		opentracing.Tag{Key: "service.instance.id", Value: resource.InstanceID},
	)
	if reflect.DeepEqual(cfg.Sampler, &jaegercfg.SamplerConfig{}) {
		// Default sampler configuration for when it is not specified via
		// JAEGER_SAMPLER_* env vars. In most cases, this is sufficient
		// enough to connect to Jaeger without any env vars.
		cfg.Sampler.Type = jaeger.SamplerTypeConst
		cfg.Sampler.Param = 1 // 1 => enabled
	}
	tracer, _, err := cfg.NewTracer(
		jaegercfg.Logger(&jaegerLogger{}),
		jaegercfg.Metrics(jaegermetrics.NullFactory),
	)
	if err != nil {
		return nil, err
	}
	return tracer, nil
}

type jaegerLogger struct{}

func (l *jaegerLogger) Error(msg string) {
	log.Printf("ERROR: %s", msg)
}

// Infof logs a message at info priority
func (l *jaegerLogger) Infof(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}
