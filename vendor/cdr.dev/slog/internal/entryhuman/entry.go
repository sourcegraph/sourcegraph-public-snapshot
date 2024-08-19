// Package entryhuman contains the code to format slog.SinkEntry
// for humans.
package entryhuman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/xerrors"

	"cdr.dev/slog"
)

// StripTimestamp strips the timestamp from entry and returns
// it and the rest of the entry.
func StripTimestamp(ent string) (time.Time, string, error) {
	ts := ent[:len(TimeFormat)]
	rest := ent[len(TimeFormat):]
	et, err := time.Parse(TimeFormat, ts)
	return et, rest, err
}

// TimeFormat is a simplified RFC3339 format.
const TimeFormat = "2006-01-02 15:04:05.000"

func c(w io.Writer, attrs ...color.Attribute) *color.Color {
	c := color.New(attrs...)
	c.DisableColor()
	if shouldColor(w) {
		c.EnableColor()
	}
	return c
}

// Fmt returns a human readable format for ent.
//
// We never return with a trailing newline because Go's testing framework adds one
// automatically and if we include one, then we'll get two newlines.
// We also do not indent the fields as go's test does that automatically
// for extra lines in a log so if we did it here, the fields would be indented
// twice in test logs. So the Stderr logger indents all the fields itself.
func Fmt(w io.Writer, ent slog.SinkEntry) string {
	ents := c(w, color.Reset).Sprint("")
	ts := ent.Time.Format(TimeFormat)
	ents += ts + " "

	level := "[" + ent.Level.String() + "]"
	level = c(w, levelColor(ent.Level)).Sprint(level)
	ents += fmt.Sprintf("%v\t", level)

	if len(ent.LoggerNames) > 0 {
		loggerName := "(" + quoteKey(strings.Join(ent.LoggerNames, ".")) + ")"
		loggerName = c(w, color.FgMagenta).Sprint(loggerName)
		ents += fmt.Sprintf("%v\t", loggerName)
	}

	hpath, hfn := humanPathAndFunc(ent.File, ent.Func)
	loc := fmt.Sprintf("<%v:%v>\t%v", hpath, ent.Line, hfn)
	loc = c(w, color.FgCyan).Sprint(loc)
	ents += fmt.Sprintf("%v\t", loc)

	var multilineKey string
	var multilineVal string
	msg := strings.TrimSpace(ent.Message)
	if strings.Contains(msg, "\n") {
		multilineKey = "msg"
		multilineVal = msg
		msg = "..."
	}
	msg = quote(msg)
	ents += msg

	if ent.SpanContext != (trace.SpanContext{}) {
		ent.Fields = append(slog.M(
			slog.F("trace", ent.SpanContext.TraceID),
			slog.F("span", ent.SpanContext.SpanID),
		), ent.Fields...)
	}

	for i, f := range ent.Fields {
		if multilineVal != "" {
			break
		}

		var s string
		switch v := f.Value.(type) {
		case string:
			s = v
		case error, xerrors.Formatter:
			s = fmt.Sprintf("%+v", v)
		}
		s = strings.TrimSpace(s)
		if !strings.Contains(s, "\n") {
			continue
		}

		// Remove this field.
		ent.Fields = append(ent.Fields[:i], ent.Fields[i+1:]...)
		multilineKey = f.Name
		multilineVal = s
	}

	if len(ent.Fields) > 0 {
		// No error is guaranteed due to slog.Map handling errors itself.
		fields, _ := json.MarshalIndent(ent.Fields, "", "")
		fields = bytes.ReplaceAll(fields, []byte(",\n"), []byte(", "))
		fields = bytes.ReplaceAll(fields, []byte("\n"), []byte(""))
		fields = formatJSON(w, fields)
		ents += "\t" + string(fields)
	}

	if multilineVal != "" {
		if msg != "..." {
			ents += " ..."
		}

		// Proper indentation.
		lines := strings.Split(multilineVal, "\n")
		for i, line := range lines[1:] {
			if line != "" {
				lines[i+1] = c(w, color.Reset).Sprint("") + strings.Repeat(" ", len(multilineKey)+4) + line
			}
		}
		multilineVal = strings.Join(lines, "\n")

		multilineKey = c(w, color.FgBlue).Sprintf(`"%v"`, multilineKey)
		ents += fmt.Sprintf("\n%v: %v", multilineKey, multilineVal)
	}

	return ents
}

func levelColor(level slog.Level) color.Attribute {
	switch level {
	case slog.LevelDebug:
		return color.Reset
	case slog.LevelInfo:
		return color.FgBlue
	case slog.LevelWarn:
		return color.FgYellow
	case slog.LevelError:
		return color.FgRed
	default:
		return color.FgHiRed
	}
}

var forceColorWriter = io.Writer(&bytes.Buffer{})

// isTTY checks whether the given writer is a *os.File TTY.
func isTTY(w io.Writer) bool {
	if w == forceColorWriter {
		return true
	}
	f, ok := w.(interface {
		Fd() uintptr
	})
	return ok && terminal.IsTerminal(int(f.Fd()))
}

func shouldColor(w io.Writer) bool {
	return isTTY(w) || os.Getenv("FORCE_COLOR") != ""
}

// quotes quotes a string so that it is suitable
// as a key for a map or in general some output that
// cannot span multiple lines or have weird characters.
func quote(key string) string {
	// strconv.Quote does not quote an empty string so we need this.
	if key == "" {
		return `""`
	}

	quoted := strconv.Quote(key)
	// If the key doesn't need to be quoted, don't quote it.
	// We do not use strconv.CanBackquote because it doesn't
	// account tabs.
	if quoted[1:len(quoted)-1] == key {
		return key
	}
	return quoted
}

func quoteKey(key string) string {
	// Replace spaces in the map keys with underscores.
	return strings.ReplaceAll(key, " ", "_")
}

var mainPackagePath string
var mainModulePath string

func init() {
	// Unfortunately does not work for tests yet :(
	// See https://github.com/golang/go/issues/33976
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	mainPackagePath = bi.Path
	mainModulePath = bi.Main.Path
}

// humanPathAndFunc takes the absolute path to a file and an absolute module path to a
// function in that file and returns the module path to the file. It also returns
// the path to the function stripped of its module prefix.
//
// If the file is in the main Go module then its path is returned
// relative to the main Go module's root.
//
// fn is from https://pkg.go.dev/runtime#Func.Name
func humanPathAndFunc(filename, fn string) (hpath, hfn string) {
	// pkgDir is the dir of the pkg.
	//   e.g. cdr.dev/slog/internal
	// base is the package name and the function name separated by a period.
	//   e.g. entryhuman.humanPathAndFunc
	// There can be multiple periods when methods of types are involved.
	pkgDir, base := filepath.Split(fn)
	s := strings.Split(base, ".")
	pkg := s[0]
	hfn = strings.Join(s[1:], ".")

	if pkg == "main" {
		// This happens with go build main.go
		if mainPackagePath == "command-line-arguments" {
			// Without a real mainPath, we can't find the path to the file
			// relative to the module. So we just return the base.
			return filepath.Base(filename), hfn
		}
		// Go doesn't return the full main path in runtime.Func.Name()
		// It just returns the path "main"
		// Only runtime.ReadBuildInfo returns it so we have to check and replace.
		pkgDir = mainPackagePath
		// pkg main isn't reflected on the disk so we should not add it
		// into the pkgpath.
		pkg = ""
	}

	hpath = filepath.Join(pkgDir, pkg, filepath.Base(filename))

	if mainModulePath != "" {
		relhpath, err := filepath.Rel(mainModulePath, hpath)
		if err == nil {
			hpath = "./" + relhpath
		}
	}

	return hpath, hfn
}
