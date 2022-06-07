// sentrycore provides a Sentry sink, that captures errors passed to the logger with the log.Error
// function if and only if the log level is superior or equal to Warn.
//
// In order to not slow down logging when it's not necessary:
// - the underlying zapcore.Core is only processing logging events on Warn and above levels.
// - consuming log events and processing them for sentry reporting are both asynchronous. This deflects
//   most of the work into the processing go routine and leverage Sentry's client ability to send reports in batches.
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

import (
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
)

const (
	// bufferSize defines how many errors the buffer can accumulate. After this limit, errors are discarded.
	bufferSize = 512
)

// baseContext contains the data surrounding an error, that is shared by all errors attached to the current core.
type baseContext struct {
	Key     string
	Scope   string
	Level   zapcore.Level
	Message string

	Fields []zapcore.Field
}

// clone safely copies a baseContext
func (b *baseContext) clone() *baseContext {
	c := *b
	c.Fields = make([]zapcore.Field, len(b.Fields))
	copy(c.Fields, b.Fields)
	return &c
}

// errorContext is an error and its associated context that is accumulated during the core lifetime.
type errorContext struct {
	*baseContext
	Error error
}

// Core turns any log message that comes with at least one error into one or more error reports. All
// error reports will share the same metadata, with the exception of those attached onto the errors themselves.
type Core struct {
	base baseContext
	// errs accumulates the errors fed to the core as attributes.
	errs []error
	w    *worker
}

var _ zapcore.Core = &Core{}

// NewCore returns a new SentryCore with a ready to use worker. It should be called only once, when attaching
// this core onto the global logger that is then used to create scoped loggers in other parts of the codebase.
func NewCore(hub *sentry.Hub) *Core {
	w := &worker{
		hub:  sentryHub{hub: hub.Clone()}, // Avoid accidental side effects if the hub is modified elsewhere.
		C:    make(chan *errorContext, bufferSize),
		done: make(chan struct{}),
	}
	w.start()
	return &Core{w: w}
}

// Core returns the underlying zapcore.
func (c *Core) Core() zapcore.Core {
	return c
}

// SetHub replaces the sentry.Hub used to submit sentry error reports.
func (c *Core) SetHub(hub *sentry.Hub) {
	c.w.setHub(hub)
}

// clone returns a copy of the core, carrying all previously accumulated context, but that can be safely be
// modified without affecting other core instances, with the exception of the worker reference which is always
// the same across cores.
func (c *Core) clone() *Core {
	clo := Core{
		w:    c.w,
		base: *c.base.clone(),
		errs: make([]error, len(c.errs)),
	}
	copy(clo.errs, c.errs)

	return &clo
}

// With stores fields passed to the core into a new core that will be then used to contruct the final error report.
//
// It does not capture errors, because we may get additional context in a subsequent With or Write call
// that will also need to be included.
func (c *Core) With(fields []zapcore.Field) zapcore.Core {
	c = c.clone()
	for _, f := range fields {
		if f.Type == zapcore.ErrorType {
			// Get original error, which we wrap on ErrorEncoder in log.Error
			if enc, ok := f.Interface.(*encoders.ErrorEncoder); ok {
				c.errs = append(c.errs, enc.Source)
				continue
			}
		}
		c.base.Fields = append(c.base.Fields, f)
	}
	return c
}

// Check inspects e to see if it needs to be sent to Sentry.
func (c *Core) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(e.Level) {
		return ce.AddCore(e, c)
	} else {
		return ce
	}
}

// Write will asynchronoulsy send out all errors and the fields that have been accumulated during the
// lifetime of the core.
func (c *Core) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Clone the base context, so we don't affect other cores
	base := c.base.clone()
	base.Scope = entry.LoggerName
	base.Message = entry.Message
	base.Level = entry.Level

	errs := make([]error, len(c.errs))
	copy(errs, c.errs)

	for _, f := range fields {
		if f.Type == zapcore.ErrorType {
			if enc, ok := f.Interface.(*encoders.ErrorEncoder); ok {
				// If we find one of our errors, we remove it from the fields so our error reports are not including
				// their own error as an attribute, which would a useless repetition.
				errs = append(errs, enc.Source)
				continue
			}
		}
		base.Fields = append(base.Fields, f)
	}

	for _, err := range errs {
		errC := errorContext{baseContext: base, Error: err}
		select {
		case c.w.C <- &errC:
		default: // if we can't queue, just drop the errors.
		}
	}
	return nil
}

// Enabled returns false when the log level is below the Warn level.
func (c *Core) Enabled(level zapcore.Level) bool {
	return level >= zapcore.WarnLevel
}

// Sync ensure that the remaining event are flushed, but has a hard limit of TODO seconds
// after which it will stop blocking to avoid interruping application shutdown.
func (c *Core) Sync() error {
	return c.w.Flush()
}

// Stop permanently shuts down the core. Only for testing purposes.
func (c *Core) Stop() {
	c.w.stop()
}
