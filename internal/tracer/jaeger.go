package tracer

import (
	"fmt"
	"io"
	"reflect"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegermetrics "github.com/uber/jaeger-lib/metrics"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newJaegerTracer creates an Jaeger tracer that serves as the underlying default tracer
// when using opentracing.
func newJaegerTracer(logger log.Logger, opts *options) (opentracing.Tracer, io.Closer, error) {
	cfg, err := jaegercfg.FromEnv()
	cfg.ServiceName = opts.resource.Name
	if err != nil {
		return nil, nil, errors.Wrap(err, "jaegercfg.FromEnv failed")
	}
	cfg.Tags = append(cfg.Tags,
		opentracing.Tag{Key: "service.version", Value: opts.resource.Version},
		opentracing.Tag{Key: "service.env", Value: opts.resource.Namespace})
	if reflect.DeepEqual(cfg.Sampler, &jaegercfg.SamplerConfig{}) {
		// Default sampler configuration for when it is not specified via
		// JAEGER_SAMPLER_* env vars. In most cases, this is sufficient
		// enough to connect Sourcegraph to Jaeger without any env vars.
		cfg.Sampler.Type = jaeger.SamplerTypeConst
		cfg.Sampler.Param = 1
	}
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jaegerLoggerShim{logger: logger.Scoped("jaeger", "Jaeger tracer")}),
		jaegercfg.Metrics(jaegermetrics.NullFactory),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "jaegercfg.NewTracer failed")
	}
	return tracer, closer, err
}

type jaegerLoggerShim struct {
	logger log.Logger
}

func (l jaegerLoggerShim) Error(msg string) { l.logger.Error(msg) }

func (l jaegerLoggerShim) Infof(msg string, args ...any) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}
