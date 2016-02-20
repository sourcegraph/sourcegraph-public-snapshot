package instrument

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"github.com/resonancelabs/go-pub/base"
)

// LogBuilder is a mechanism for constructing LogRecords. Each of its
// methods returns its receiver so that additional methods may be
// chained. For example,
//
//   instrument.Payload(req).CallDepth(1).Print("sending request"))
//
// A LogBuilder may be used as an argument to Logger.Log or converted
// to a LogRecord explicitly.
type LogBuilder struct {
	message     string
	eventName   string
	payload     interface{}
	fileName    string
	lineNumber  int
	stackFrames []string
	level       string
	isError     bool
}

// Print inserts a human-readable, unstructured string into the log
// treating its arguments like fmt.Print. It overrides previous calls
// to Print, Println, and Printf.
func (b *LogBuilder) Print(args ...interface{}) *LogBuilder {
	b.message = fmt.Sprint(args...)
	return b
}

// Println inserts a human-readable, unstructured string into the log
// treating its arguments like fmt.Println. It overrides previous calls
// to Print, Println, and Printf.
func (b *LogBuilder) Println(args ...interface{}) *LogBuilder {
	b.message = fmt.Sprintln(args...)
	return b
}

// Printf inserts a human-readable, unstructured string into the log
// treating its arguments like fmt.Printf. It overrides previous calls
// to Print, Println, and Printf.
func (b *LogBuilder) Printf(format string, args ...interface{}) *LogBuilder {
	b.message = fmt.Sprintf(format, args...)
	return b
}

// Payload inserts structured data into the log. These data may be
// truncated. It overrides previous calls to Payload.
func (b *LogBuilder) Payload(payload interface{}) *LogBuilder {
	b.payload = payload
	return b
}

// EventName sets a stable name for event that triggered this log
// record. It overrides previous calls to EventName.
func (b *LogBuilder) EventName(name string) *LogBuilder {
	b.eventName = name
	return b
}

// FileLine adds the file name and line number of the caller to the
// log record. It overrides previous calls to FileLine.
func (b *LogBuilder) FileLine(callDepth int) *LogBuilder {
	// XXX + 1?
	_, file, line, ok := runtime.Caller(callDepth + 1)
	if !ok {
		file = "???" // "???" is what glog prints when the filename is missing
		line = 1
	} else {
		// TODO: Do we want to retain the full file path?
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	b.fileName = file
	b.lineNumber = line
	b.stackFrames = nil
	return b
}

// CallStack adds call stack information of the caller to the log
// record. It overrides previous calls to FileLine and CallStack.
func (b *LogBuilder) CallStack(callDepth int) *LogBuilder {
	// XXX + 1?
	_, file, line, ok := runtime.Caller(callDepth + 1)
	if !ok {
		file = "???" // "???" is what glog prints when the filename is missing
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	b.fileName = file
	b.lineNumber = line

	// Grab the stack and then "manually" split it into an array of frames.
	// The splitting is a bit fragile, but Go doesn't expose an API to get
	// the frames separately (runtime.Caller does include the arguments).
	stackBuffer := make([]byte, 1024)
	runtime.Stack(stackBuffer, false)

	// The return stack information is two-lines per frame
	lines := strings.Split(string(stackBuffer), "\n")
	b.stackFrames = make([]string, 0, len(lines)/2+1-callDepth)

	// "goroutine <N> [running]:"
	b.stackFrames = append(b.stackFrames, lines[0])

	// <function>(<args)
	//		<filename>:<line> +<iptr>
	for i := 1 + callDepth*2; i+1 < len(lines); i += 2 {
		b.stackFrames = append(b.stackFrames, lines[i]+"\n"+lines[i+1])
	}
	return b
}

// RawLevel sets the severity of a log record. It overrides previous calls to
// RawLevel. The argument is expected to be a string of length one. By
// convention, the following strings are used:
//
//  "I" - informational
//  "W" - warning
//  "E" - error
//  "F" - fatal
//
// Both "E" and "F" are considered erroneous and will be highlighted
// in various presentations.
func (b *LogBuilder) RawLevel(lvl string) *LogBuilder {
	b.level = lvl
	switch lvl {
	case "E", "F":
		b.isError = true
	default:
		b.isError = false
	}
	return b
}

// InfoLog sets the log record severity to "I" and overrides previous calls to
// Info()/Warning()/Error()/RawLevel().
func (b *LogBuilder) Info() *LogBuilder {
	return b.RawLevel("I")
}

// InfoLog sets the log record severity to "W" and overrides previous calls to
// Info()/Warning()/Error()/RawLevel().
func (b *LogBuilder) Warning() *LogBuilder {
	return b.RawLevel("W")
}

// InfoLog sets the log record severity to "E" and overrides previous calls to
// Info()/Warning()/Error()/RawLevel().  Log records marked as errors are
// considered exceptional and are thus highlighted in some presentations.
func (b *LogBuilder) Error() *LogBuilder {
	return b.RawLevel("E")
}

// LogRecord is a raw format of a single log. LogRecords may be
// constructed explicitly or using a LogBuilder.
type LogRecord struct {
	TimestampMicros base.Micros
	Message         string
	EventName       string
	Payload         interface{}
	Level           string
	FileName        string
	LineNumber      int
	StackFrames     []string
	IsError         bool
}

// LogRecord returns the data represented by this builder.
func (b *LogBuilder) LogRecord() *LogRecord {
	return &LogRecord{
		Message:     b.message,
		EventName:   b.eventName,
		Payload:     b.payload,
		Level:       b.level,
		FileName:    b.fileName,
		LineNumber:  b.lineNumber,
		StackFrames: b.stackFrames,
		IsError:     b.isError,
	}
}

func (l *LogRecord) String() string {
	var buf bytes.Buffer
	buf.WriteString("LogRecord:{")
	if l.TimestampMicros != 0 {
		buf.WriteString(l.TimestampMicros.ToTime().Format("0102 15:04:05.000000"))
		buf.WriteString(" ")
	}
	if l.Message != "" {
		buf.WriteString(fmt.Sprintf("%q", l.Message))
		buf.WriteString(" ")
	}
	if l.EventName != "" {
		buf.WriteString("(")
		buf.WriteString(l.EventName)
		buf.WriteString(") ")
	}
	if buf.Len() > 1 {
		// remove the last space
		buf.Truncate(buf.Len() - 1)
	}
	buf.WriteString("}")
	return buf.String()
}
