// A package with an API that serves as a drop-in replacement for the default
// golang `log` package.
//
// All logging statements are tee'd to the instrumentation logger
// before being passed through to the normal golang log package.
//
// This ends up being less trivial than one would hope... TL;DR: Doing an
// interposition library is gross when the underlying library depends on the
// number of stack frames between the user code and the library code. I had to
// copy-pasta more of the golang log.go source than I had hoped, though it's
// all just glue. The core function (Output()) is thankfully
// stackdepth-independent.
package logwrapper

import (
	"log" // the "actual" golang logging library we mimic here

	"fmt"
	"io"
	"os"

	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
)

//////////////////
// THE LOGGER TYPE
//////////////////

// See golang's log.Logger.
type Logger struct {
	*log.Logger
	runtime instrument.Runtime
}

// For compatibility (just passthroughs, obviously).
const (
	Ldate         = log.Ldate
	Ltime         = log.Ltime
	Lmicroseconds = log.Lmicroseconds
	Llongfile     = log.Llongfile
	Lshortfile    = log.Lshortfile
	LstdFlags     = log.LstdFlags
)

// New creates a new Logger.   The out variable sets the
// destination to which log data will be written.
// The prefix appears at the beginning of each generated log line.
// The flag argument defines the logging properties.
func New(out io.Writer, prefix string, flag int) *Logger {
	return NewForRuntime(instrument.DefaultRuntime(), out, prefix, flag)
}

// For users of custom/proxy runtimes (just in case).
func NewForRuntime(r instrument.Runtime, out io.Writer, prefix string, flag int) *Logger {
	return &Logger{
		runtime: r,
		Logger:  log.New(out, prefix, flag),
	}
}

func (l *Logger) Output(calldepth int, s string) error {
	l.runtime.Log(instrument.FileLine(calldepth + 1).Info().Print(s))
	return l.Logger.Output(calldepth+1, s)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.runtime.Log(instrument.FileLine(2).Info().Printf(format, v...))
	l.Logger.Output(2, fmt.Sprintf(format, v...))
}

func (l *Logger) Print(v ...interface{}) {
	l.runtime.Log(instrument.FileLine(2).Info().Print(v...))
	l.Logger.Output(2, fmt.Sprint(v...))
}

func (l *Logger) Println(v ...interface{}) {
	l.runtime.Log(instrument.FileLine(2).Info().Println(v...))
	l.Logger.Output(2, fmt.Sprintln(v...))
}

func (l *Logger) Fatal(v ...interface{}) {
	l.runtime.Log(instrument.FileLine(2).RawLevel("F").Print(v...))
	l.Logger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.runtime.Log(instrument.FileLine(2).RawLevel("F").Printf(format, v...))
	l.Logger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l *Logger) Fatalln(v ...interface{}) {
	l.runtime.Log(instrument.FileLine(2).RawLevel("F").Println(v...))
	l.Logger.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.runtime.Log(instrument.FileLine(2).RawLevel("F").Print(v...))
	l.Logger.Output(2, s)
	panic(s)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprint(v...)
	l.runtime.Log(instrument.FileLine(2).RawLevel("F").Printf(format, v...))
	l.Logger.Output(2, s)
	panic(s)
}

func (l *Logger) Panicln(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.runtime.Log(instrument.FileLine(2).RawLevel("F").Println(v...))
	l.Logger.Output(2, s)
	panic(s)
}

///////////////////////////
// EXPORTED FUNCTIONS
///////////////////////////

// Same params as in the log source:
var std = New(os.Stderr, "", LstdFlags)

func SetOutput(w io.Writer)   { log.SetOutput(w) }
func Flags() int              { return std.Flags() }
func SetFlags(flag int)       { std.SetFlags(flag) }
func Prefix() string          { return std.Prefix() }
func SetPrefix(prefix string) { std.SetPrefix(prefix) }

// NOTE: the following functions defer to log (via std.Output(...))
// for the actual logging but call Exit/panic/etc themselves. The
// implementations are adapted from the golang log.go source. It would
// look cleaner to defer to the Logger implementation above, but doing
// so would get the stack depth parameter wrong.

func Print(v ...interface{}) {
	std.Output(2, fmt.Sprint(v...))
}
func Printf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(format, v...))
}
func Println(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...))
}
func Fatal(v ...interface{}) {
	std.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}
func Fatalf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}
func Fatalln(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}
func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	std.Output(2, s)
	panic(s)
}
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	std.Output(2, s)
	panic(s)
}
func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	std.Output(2, s)
	panic(s)
}
