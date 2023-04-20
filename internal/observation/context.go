package observation

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Context carries context about where to send logs, trace spans, and register
// metrics. It should be created once on service startup, and passed around to
// any location that wants to use it for observing operations.
type Context struct {
	Logger       log.Logger
	Tracer       *trace.Tracer
	Registerer   prometheus.Registerer
	HoneyDataset *honey.Dataset
}

func (c *Context) Clone(opts ...Opt) *Context {
	c1 := &Context{
		Logger:       c.Logger,
		Tracer:       c.Tracer,
		Registerer:   c.Registerer,
		HoneyDataset: c.HoneyDataset,
	}

	for _, opt := range opts {
		opt(c1)
	}

	return c1
}

// TestContext is a behaviorless Context usable for unit tests.
var TestContext = Context{Logger: log.NoOp(), Registerer: metrics.TestRegisterer}

// TestContextTB creates a Context similar to `TestContext` but with a logger scoped
// to the `testing.TB`.
func TestContextTB(t testing.TB) *Context {
	return &Context{
		Logger:     logtest.Scoped(t),
		Registerer: metrics.TestRegisterer,
	}
}

// ContextWithLogger creates a live Context with the given logger instance.
func ContextWithLogger(logger log.Logger, parent *Context) *Context {
	return &Context{
		Logger:       logger,
		Tracer:       parent.Tracer,
		Registerer:   parent.Registerer,
		HoneyDataset: parent.HoneyDataset,
	}
}

// ScopedContext creates a live Context with a logger configured with the given values.
func ScopedContext(team, domain, component string, parent *Context) *Context {
	return ContextWithLogger(log.Scoped(
		fmt.Sprintf("%s.%s.%s", team, domain, component),
		fmt.Sprintf("%s %s %s", team, domain, component),
	), parent)
}

// Operation combines the state of the parent context to create a new operation. This value
// should be owned and used by the code that performs the operation it represents.
func (c *Context) Operation(args Op) *Operation {
	var logger log.Logger
	if c.Logger != nil {
		// Create a child logger, if a parent is provided.
		logger = c.Logger.Scoped(args.Name, args.Description)
	} else {
		// Create a new logger.
		logger = log.Scoped(args.Name, args.Description)
	}
	return &Operation{
		context:      c,
		metrics:      args.Metrics,
		name:         args.Name,
		kebabName:    kebabCase(args.Name),
		metricLabels: args.MetricLabelValues,
		logFields:    args.LogFields,
		errorFilter:  args.ErrorFilter,

		Logger: logger.With(toLogFields(args.LogFields)...),
	}
}

func NewContext(logger log.Logger, opts ...Opt) *Context {
	ctx := &Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	for _, opt := range opts {
		opt(ctx)
	}

	return ctx
}

type Opt func(*Context)

func Tracer(provider oteltrace.TracerProvider) Opt {
	return func(ctx *Context) {
		ctx.Tracer = &trace.Tracer{TracerProvider: provider}
	}
}

func Metrics(register prometheus.Registerer) Opt {
	return func(ctx *Context) {
		ctx.Registerer = register
	}
}

func Honeycomb(dataset *honey.Dataset) Opt {
	return func(ctx *Context) {
		ctx.HoneyDataset = dataset
	}
}
