package observation

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Context carries context about where to send logs, trace spans, and register
// metrics. It should be created once on service startup, and passed around to
// any location that wants to use it for observing operations.
type Context struct {
	Logger       log.Logger
	Tracer       oteltrace.Tracer // may be nil
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
var TestContext = Context{
	Logger:     log.NoOp(),
	Tracer:     noop.NewTracerProvider().Tracer("noop"),
	Registerer: metrics.NoOpRegisterer,
	// We do not set HoneyDataset since if we accidently have HONEYCOMB_TEAM
	// set in a test run it will log to honeycomb.
}

// TestContextTB creates a Context similar to `TestContext` but with a logger scoped
// to the `testing.TB` and a pedantic Registerer.
func TestContextTB(t testing.TB) *Context {
	return &Context{
		Logger:     logtest.Scoped(t),
		Registerer: prometheus.NewPedanticRegistry(),
		Tracer:     noop.NewTracerProvider().Tracer("noop"),
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
	), parent)
}

// Operation combines the state of the parent context to create a new operation. This value
// should be owned and used by the code that performs the operation it represents.
func (c *Context) Operation(args Op) *Operation {
	var logger log.Logger
	if c.Logger != nil {
		// Create a child logger, if a parent is provided.
		logger = c.Logger.Scoped(args.Name)
	} else {
		// Create a new logger.
		logger = log.Scoped(args.Name)
	}
	return &Operation{
		context:      c,
		metrics:      args.Metrics,
		name:         args.Name,
		kebabName:    kebabCase(args.Name),
		metricLabels: args.MetricLabelValues,
		attributes:   args.Attrs,
		errorFilter:  args.ErrorFilter,

		Logger: logger.With(attributesToLogFields(args.Attrs)...),
	}
}

func NewContext(logger log.Logger, opts ...Opt) *Context {
	ctx := &Context{
		Logger:     logger,
		Tracer:     trace.GetTracer(),
		Registerer: prometheus.DefaultRegisterer,
	}

	for _, opt := range opts {
		opt(ctx)
	}

	return ctx
}

type Opt func(*Context)

func Tracer(tracer oteltrace.Tracer) Opt {
	return func(ctx *Context) {
		ctx.Tracer = tracer
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
