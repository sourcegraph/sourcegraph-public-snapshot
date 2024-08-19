// Package slog implements minimal structured logging.
//
// See https://cdr.dev/slog for overview docs and a comparison with existing libraries.
//
// The examples are the best way to understand how to use this library effectively.
//
// The Logger type implements a high level API around the Sink interface.
// Logger implements Sink as well to allow composition.
//
// Implementations of the Sink interface are available in the sloggers subdirectory.
package slog // import "cdr.dev/slog"

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"go.opencensus.io/trace"
)

var defaultExitFn = os.Exit

// Sink is the destination of a Logger.
//
// All sinks must be safe for concurrent use.
type Sink interface {
	LogEntry(ctx context.Context, e SinkEntry)
	Sync()
}

// Log logs the given entry with the context to the
// underlying sinks.
//
// It extends the entry with the set fields and names.
func (l Logger) Log(ctx context.Context, e SinkEntry) {
	if e.Level < l.level {
		return
	}

	e.Fields = l.fields.append(e.Fields)
	e.LoggerNames = appendNames(l.names, e.LoggerNames...)

	for _, s := range l.sinks {
		s.LogEntry(ctx, e)
	}
}

// Sync calls Sync on all the underlying sinks.
func (l Logger) Sync() {
	for _, s := range l.sinks {
		s.Sync()
	}
}

// Logger wraps Sink with a nice API to log entries.
//
// Logger is safe for concurrent use.
type Logger struct {
	sinks []Sink
	level Level

	names  []string
	fields Map

	skip int
	exit func(int)
}

// Make creates a logger that writes logs to the passed sinks at LevelInfo.
func Make(sinks ...Sink) Logger {
	return Logger{
		sinks: sinks,
		level: LevelInfo,

		exit: os.Exit,
	}
}

// Debug logs the msg and fields at LevelDebug.
func (l Logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelDebug, msg, fields)
}

// Info logs the msg and fields at LevelInfo.
func (l Logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelInfo, msg, fields)
}

// Warn logs the msg and fields at LevelWarn.
func (l Logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelWarn, msg, fields)
}

// Error logs the msg and fields at LevelError.
//
// It will then Sync().
func (l Logger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelError, msg, fields)
	l.Sync()
}

// Critical logs the msg and fields at LevelCritical.
//
// It will then Sync().
func (l Logger) Critical(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelCritical, msg, fields)
	l.Sync()
}

// Fatal logs the msg and fields at LevelFatal.
//
// It will then Sync() and os.Exit(1).
func (l Logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelFatal, msg, fields)
	l.Sync()

	if l.exit == nil {
		l.exit = defaultExitFn
	}

	l.exit(1)
}

// With returns a Logger that prepends the given fields on every
// logged entry.
//
// It will append to any fields already in the Logger.
func (l Logger) With(fields ...Field) Logger {
	l.fields = l.fields.append(fields)
	return l
}

// Named appends the name to the set names
// on the logger.
func (l Logger) Named(name string) Logger {
	l.names = appendNames(l.names, name)
	return l
}

// Leveled returns a Logger that only logs entries
// equal to or above the given level.
func (l Logger) Leveled(level Level) Logger {
	l.level = level
	l.sinks = append([]Sink(nil), l.sinks...)
	return l
}

// AppendSinks appends the sinks to the set sink
// targets on the logger.
func (l Logger) AppendSinks(s ...Sink) Logger {
	l.sinks = append(l.sinks, s...)
	return l
}

func (l Logger) log(ctx context.Context, level Level, msg string, fields Map) {
	ent := l.entry(ctx, level, msg, fields)
	l.Log(ctx, ent)
}

func (l Logger) entry(ctx context.Context, level Level, msg string, fields Map) SinkEntry {
	ent := SinkEntry{
		Time:        time.Now().UTC(),
		Level:       level,
		Message:     msg,
		Fields:      fieldsFromContext(ctx).append(fields),
		SpanContext: trace.FromContext(ctx).SpanContext(),
	}
	ent = ent.fillLoc(l.skip + 3)
	return ent
}

var helpers sync.Map

// Helper marks the calling function as a helper
// and skips it for source location information.
// It's the slog equivalent of testing.TB.Helper().
func Helper() {
	_, _, fn := location(1)
	helpers.LoadOrStore(fn, struct{}{})
}

func (ent SinkEntry) fillFromFrame(f runtime.Frame) SinkEntry {
	ent.Func = f.Function
	ent.File = f.File
	ent.Line = f.Line
	return ent
}

func (ent SinkEntry) fillLoc(skip int) SinkEntry {
	// Copied from testing.T
	const maxStackLen = 50
	var pc [maxStackLen]uintptr

	// Skip two extra frames to account for this function
	// and runtime.Callers itself.
	n := runtime.Callers(skip+2, pc[:])
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		_, helper := helpers.Load(frame.Function)
		if !helper || !more {
			// Found a frame that wasn't a helper function.
			// Or we ran out of frames to check.
			return ent.fillFromFrame(frame)
		}
	}
}

func location(skip int) (file string, line int, fn string) {
	pc, file, line, _ := runtime.Caller(skip + 1)
	f := runtime.FuncForPC(pc)
	return file, line, f.Name()
}

func appendNames(names []string, names2 ...string) []string {
	if len(names2) == 0 {
		return names
	}
	names3 := make([]string, 0, len(names)+len(names2))
	names3 = append(names3, names...)
	names3 = append(names3, names2...)
	return names3
}

// Field represents a log field.
type Field struct {
	Name  string
	Value interface{}
}

// F is a convenience constructor for Field.
func F(name string, value interface{}) Field {
	return Field{Name: name, Value: value}
}

// M is a convenience constructor for Map
func M(fs ...Field) Map {
	return fs
}

// Error is the standard key used for logging a Go error value.
func Error(err error) Field {
	return F("error", err)
}

type fieldsKey struct{}

func fieldsWithContext(ctx context.Context, fields Map) context.Context {
	return context.WithValue(ctx, fieldsKey{}, fields)
}

func fieldsFromContext(ctx context.Context) Map {
	l, _ := ctx.Value(fieldsKey{}).(Map)
	return l
}

// With returns a context that contains the given fields.
//
// Any logs written with the provided context will have the given logs prepended.
//
// It will append to any fields already in ctx.
func With(ctx context.Context, fields ...Field) context.Context {
	f1 := fieldsFromContext(ctx)
	f2 := f1.append(fields)
	return fieldsWithContext(ctx, f2)
}

// SinkEntry represents the structure of a log entry.
// It is the argument to the sink when logging.
type SinkEntry struct {
	Time time.Time

	Level   Level
	Message string

	LoggerNames []string

	Func string
	File string
	Line int

	SpanContext trace.SpanContext

	Fields Map
}

// Level represents a log level.
type Level int

// The supported log levels.
//
// The default level is Info.
const (
	// LevelDebug is used for development and debugging messages.
	LevelDebug Level = iota

	// LevelInfo is used for normal informational messages.
	LevelInfo

	// LevelWarn is used when something has possibly gone wrong.
	LevelWarn

	// LevelError is used when something has certainly gone wrong.
	LevelError

	// LevelCritical is used when when something has gone wrong and should
	// be immediately investigated.
	LevelCritical

	// LevelFatal is used when the process is about to exit due to an error.
	LevelFatal
)

var levelStrings = map[Level]string{
	LevelDebug:    "DEBUG",
	LevelInfo:     "INFO",
	LevelWarn:     "WARN",
	LevelError:    "ERROR",
	LevelCritical: "CRITICAL",
	LevelFatal:    "FATAL",
}

// String implements fmt.Stringer.
func (l Level) String() string {
	s, ok := levelStrings[l]
	if !ok {
		return fmt.Sprintf("slog.Level(%v)", int(l))
	}
	return s
}
