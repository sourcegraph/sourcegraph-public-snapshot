// sentrycore provides a few additional cores that adds side effects to logging.
//
// See SentryCore for more information about turning log messages whose level is equal or above zapcore.WarnLevel
// into Sentry reports if there is an error attached with log.Error(err).
package sentrycore

import (
	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"go.uber.org/zap/zapcore"
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
	c.Key = b.Key
	c.Scope = b.Scope
	c.Level = b.Level
	c.Message = b.Message
	c.Fields = make([]zapcore.Field, len(b.Fields))
	copy(c.Fields, b.Fields)
	return &c
}

// ErrorContext is an error and its associated context that is accumulated during the core lifetime.
type ErrorContext struct {
	baseContext
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
	return &Core{
		w: &worker{
			hub:  hub.Clone(), // Avoid accidental side effects if the hub is modified elsewhere.
			C:    make(chan *Core, 512),
			done: make(chan struct{}),
		},
	}
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
// modified without affecting other core instances.
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
	// Clone ourselves, so we don't affect other loggers
	c = c.clone()
	c.base.Scope = entry.LoggerName
	c.base.Message = entry.Message
	c.base.Level = entry.Level

	sentryFields := make([]zapcore.Field, 0, len(fields))
	for _, f := range fields {
		if f.Type == zapcore.ErrorType {
			if enc, ok := f.Interface.(*encoders.ErrorEncoder); ok {
				// If we find one of our errors, we remove it from the fields so our error reports are not including
				// their own error as an attribute, which would a useless repetition.
				c.errs = append(c.errs, enc.Source)
				continue
			}
		}
		sentryFields = append(sentryFields, f)
	}
	c.base.Fields = append(sentryFields, c.base.Fields...)

	// c.w.in.Add(1)
	select {
	case c.w.C <- c:
	default: // if we can't queue, just drop the errors.
	}
	return nil
}

// Starts launches the go routine responsible for consuming ErrorContext that needs to be submitted to Sentry.
func (c *Core) Start() {
	c.w.start()
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

func (c *Core) Stop() {
	// close(s.worker.done)
}
