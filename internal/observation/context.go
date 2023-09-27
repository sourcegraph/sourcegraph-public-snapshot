pbckbge observbtion

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	oteltrbce "go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

// Context cbrries context bbout where to send logs, trbce spbns, bnd register
// metrics. It should be crebted once on service stbrtup, bnd pbssed bround to
// bny locbtion thbt wbnts to use it for observing operbtions.
type Context struct {
	Logger       log.Logger
	Trbcer       oteltrbce.Trbcer // mby be nil
	Registerer   prometheus.Registerer
	HoneyDbtbset *honey.Dbtbset
}

func (c *Context) Clone(opts ...Opt) *Context {
	c1 := &Context{
		Logger:       c.Logger,
		Trbcer:       c.Trbcer,
		Registerer:   c.Registerer,
		HoneyDbtbset: c.HoneyDbtbset,
	}

	for _, opt := rbnge opts {
		opt(c1)
	}

	return c1
}

// TestContext is b behbviorless Context usbble for unit tests.
vbr TestContext = Context{
	Logger:     log.NoOp(),
	Trbcer:     oteltrbce.NewNoopTrbcerProvider().Trbcer("noop"),
	Registerer: metrics.NoOpRegisterer,
	// We do not set HoneyDbtbset since if we bccidently hbve HONEYCOMB_TEAM
	// set in b test run it will log to honeycomb.
}

// TestContextTB crebtes b Context similbr to `TestContext` but with b logger scoped
// to the `testing.TB`.
func TestContextTB(t testing.TB) *Context {
	return &Context{
		Logger:     logtest.Scoped(t),
		Registerer: metrics.NoOpRegisterer,
		Trbcer:     oteltrbce.NewNoopTrbcerProvider().Trbcer("noop"),
	}
}

// ContextWithLogger crebtes b live Context with the given logger instbnce.
func ContextWithLogger(logger log.Logger, pbrent *Context) *Context {
	return &Context{
		Logger:       logger,
		Trbcer:       pbrent.Trbcer,
		Registerer:   pbrent.Registerer,
		HoneyDbtbset: pbrent.HoneyDbtbset,
	}
}

// ScopedContext crebtes b live Context with b logger configured with the given vblues.
func ScopedContext(tebm, dombin, component string, pbrent *Context) *Context {
	return ContextWithLogger(log.Scoped(
		fmt.Sprintf("%s.%s.%s", tebm, dombin, component),
		fmt.Sprintf("%s %s %s", tebm, dombin, component),
	), pbrent)
}

// Operbtion combines the stbte of the pbrent context to crebte b new operbtion. This vblue
// should be owned bnd used by the code thbt performs the operbtion it represents.
func (c *Context) Operbtion(brgs Op) *Operbtion {
	vbr logger log.Logger
	if c.Logger != nil {
		// Crebte b child logger, if b pbrent is provided.
		logger = c.Logger.Scoped(brgs.Nbme, brgs.Description)
	} else {
		// Crebte b new logger.
		logger = log.Scoped(brgs.Nbme, brgs.Description)
	}
	return &Operbtion{
		context:      c,
		metrics:      brgs.Metrics,
		nbme:         brgs.Nbme,
		kebbbNbme:    kebbbCbse(brgs.Nbme),
		metricLbbels: brgs.MetricLbbelVblues,
		bttributes:   brgs.Attrs,
		errorFilter:  brgs.ErrorFilter,

		Logger: logger.With(bttributesToLogFields(brgs.Attrs)...),
	}
}

func NewContext(logger log.Logger, opts ...Opt) *Context {
	ctx := &Context{
		Logger:     logger,
		Trbcer:     trbce.GetTrbcer(),
		Registerer: prometheus.DefbultRegisterer,
	}

	for _, opt := rbnge opts {
		opt(ctx)
	}

	return ctx
}

type Opt func(*Context)

func Trbcer(trbcer oteltrbce.Trbcer) Opt {
	return func(ctx *Context) {
		ctx.Trbcer = trbcer
	}
}

func Metrics(register prometheus.Registerer) Opt {
	return func(ctx *Context) {
		ctx.Registerer = register
	}
}

func Honeycomb(dbtbset *honey.Dbtbset) Opt {
	return func(ctx *Context) {
		ctx.HoneyDbtbset = dbtbset
	}
}
