package tracer

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

func newConfWatcher(
	logger log.Logger,
	c conftypes.SiteConfigQuerier,
	otelProvider *switchableOtelTracerProvider,
	otTracer *switchableOTTracer,
	initialOpts options,
) func() {
	// always keep a reference to our existing options to determine if an update is needed
	oldOpts := initialOpts

	// return function to be called on every conf update
	return func() {
		var (
			siteConfig     = c.SiteConfig()
			tracingConfig  = siteConfig.ObservabilityTracing
			previousPolicy = policy.GetTracePolicy()
			setTracerType  = None
			debug          = false
		)

		// If 'observability.tracing: {}', try to enable tracing by default
		if tracingConfig != nil {
			// If sampling policy is set, update the strategy and set a default TracerType
			var newPolicy policy.TracePolicy
			switch p := policy.TracePolicy(tracingConfig.Sampling); p {
			case policy.TraceNone, policy.TraceAll, policy.TraceSelective:
				// Set supported policy types
				newPolicy = p
			default:
				// Default to selective
				newPolicy = policy.TraceSelective
			}

			// Set and log our new trace policy
			if newPolicy != previousPolicy {
				policy.SetTracePolicy(newPolicy)
				logger.Info("updated TracePolicy",
					log.String("previous", string(previousPolicy)),
					log.String("new", string(newPolicy)))
			}

			// If the tracer type is configured, also set the tracer type. Otherwise, set
			// a default tracer type, unless the desired policy is none.
			if t := TracerType(tracingConfig.Type); t.isSetByUser() {
				setTracerType = t
			} else if newPolicy != policy.TraceNone {
				setTracerType = DefaultTracerType
			}

			// Configure debug mode
			debug = tracingConfig.Debug
		}

		// collect options
		opts := options{
			TracerType:  setTracerType,
			externalURL: siteConfig.ExternalURL,
			debug:       debug,
			// Stays the same
			resource: oldOpts.resource,
		}
		if opts == oldOpts {
			// Nothing changed
			return
		}

		// update old opts for comparison
		oldOpts = opts

		// create new tracer providers
		tracerLogger := logger.With(
			log.String("tracerType", string(opts.TracerType)),
			log.Bool("debug", opts.debug))
		otImpl, otelImpl, closer, err := newTracer(tracerLogger, &opts)
		if err != nil {
			tracerLogger.Warn("failed to initialize tracer", log.Error(err))
			// do not return - we still want to update tracers
		}

		// update global tracers
		otelProvider.set(otelImpl, closer, opts.debug)
		otTracer.set(tracerLogger, otImpl, opts.debug)
	}
}
