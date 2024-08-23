// sentrycore provides a Sentry sink, that captures errors passed to the logger with the log.Error
// function if and only if the log level is superior or equal to Error.
//
// In order to not slow down logging when it's not necessary:
//
// a) the underlying zapcore.Core is only processing logging events on Error and above levels.
//
// b) consuming log events and processing them for sentry reporting are both asynchronous. This deflects
// most of the work into the processing go routine and leverage Sentry's client ability to send reports in batches.
//
// In the eventuality of saturating the events buffer by producing errors quicker than we can produce them, they will
// be purely dropped.
//
// In order to avoid losing events, the events are continuously sent to Sentry and don't need to be explicitly flushed.
// If asked explicitly to be flushed as part of the zapcore.Core interface, the Sentry sink will try to consume all
// log events within a reasonable time before shutting down the consumer side, and will then submit them to Sentry.
//
// Flushing is only called in the final defer function coming from our logging API, meaning that will only happen
// when a service is shutting down.
//
// In the eventuality where we are submitting events faster than we could consume then, the upper bound is a large
// buffered channel, which should be enough to accumulate errors while we're asynchronously reporting them to Sentry.
//
// It would be nice to be able to know if we're dropping errors, but that would create a circular dependency
// from the sink toward the logger, so for now, they're just silently discarded.
package sentrycore
