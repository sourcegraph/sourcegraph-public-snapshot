package tracer

import (
	"sync/atomic"

	"github.com/sourcegraph/log"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// newConfWatcher creates a callback that can be used on subscription to changes in site
// configuration via conf.Watch(). The callback is stateful, compares the new state of
// configuration with previous known state on each call, and propagates any changes to the
// provider and debugMode references.
func newConfWatcher(
	logger log.Logger,
	c ConfigurationSource,
	// provider will be updated with the appropriate span processors.
	provider *oteltracesdk.TracerProvider,
	// spanProcessorBuilder is used to create span processors to configure on the provider
	// based on the given options.
	spanProcessorBuilder func(logger log.Logger, opts options, debug bool) (oteltracesdk.SpanProcessor, error),
	// debugMode is a shared reference that can be updated with the latest debug state.
	debugMode *atomic.Bool,
) func() {
	// always keep a reference to our existing options to determine if an update is needed
	oldOpts := options{
		// Default options
		TracerType:  None,
		externalURL: "",
	}
	var oldProcessor oteltracesdk.SpanProcessor

	// return function to be called on every conf update
	return func() {
		var (
			siteConfig     = c.Config()
			tracingConfig  = siteConfig.ObservabilityTracing
			previousPolicy = policy.GetTracePolicy()
			setTracerType  = None
			debugChanged   bool
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
				logger.Debug("updated TracePolicy",
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
			debugChanged = debugMode.CompareAndSwap(debugMode.Load(), tracingConfig.Debug)
		} else {
			debugChanged = debugMode.CompareAndSwap(debugMode.Load(), false)
		}

		// collect options
		opts := options{
			TracerType:  setTracerType,
			externalURL: siteConfig.ExternalURL,
			// Stays the same
			resource: oldOpts.resource,
		}
		if opts == oldOpts && !debugChanged {
			// Nothing changed
			return
		}

		// update old opts for comparison
		oldOpts = opts

		// create new span processor
		debug := debugMode.Load()
		tracerLogger := logger.With(
			log.String("tracerType", string(opts.TracerType)),
			log.Bool("debug", debug))
		processor, err := spanProcessorBuilder(logger, opts, debug)
		if err != nil {
			tracerLogger.Warn("failed to build updated processors", log.Error(err))
			// continue with handling, do not fail fast
		}

		// add the new processor. we do this before adding the new processor to
		// ensure we don't have any gaps where spans are being dropped.
		if processor != nil {
			provider.RegisterSpanProcessor(processor)
		}

		// remove the pre-existing processor - this shuts it down and prevents
		// newer traces from going to it. we do this regardless of processor
		// creation error
		if oldProcessor != nil {
			provider.UnregisterSpanProcessor(oldProcessor)
		}

		// update reference
		oldProcessor = processor
	}
}
