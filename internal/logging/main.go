package logging

import (
	"bytes"
	"fmt"
	"io"
	"log" //nolint:logging // Expected usage; inits old loggers for interop
	"os"

	"github.com/fatih/color"
	"github.com/inconshreveable/log15" //nolint:logging // Expected usage; inits old loggers for interop

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	logColors = map[log15.Lvl]color.Attribute{
		log15.LvlCrit:  color.FgRed,
		log15.LvlError: color.FgRed,
		log15.LvlWarn:  color.FgYellow,
		log15.LvlInfo:  color.FgCyan,
		log15.LvlDebug: color.Faint,
	}
	// We'd prefer these in caps, not lowercase, and don't need the 4-character alignment
	logLabels = map[log15.Lvl]string{
		log15.LvlCrit:  "CRITICAL",
		log15.LvlError: "ERROR",
		log15.LvlWarn:  "WARN",
		log15.LvlInfo:  "INFO",
		log15.LvlDebug: "DEBUG",
	}
)

func condensedFormat(r *log15.Record) []byte {
	colorAttr := logColors[r.Lvl]
	text := logLabels[r.Lvl]
	var msg bytes.Buffer
	if env.LogSourceLink {
		// Link to open the file:line in VS Code.
		url := "vscode://file/" + fmt.Sprintf("%#v", r.Call)

		// Constructs an escape sequence that iTerm recognizes as a link.
		// See https://iterm2.com/documentation-escape-codes.html
		link := fmt.Sprintf("\x1B]8;;%s\x07%s\x1B]8;;\x07", url, "src")

		fmt.Fprint(&msg, color.New(color.Faint).Sprint(link)+" ")
	}
	if colorAttr != 0 {
		fmt.Fprint(&msg, color.New(colorAttr).Sprint(text)+" ")
	}
	fmt.Fprint(&msg, r.Msg)
	if len(r.Ctx) > 0 {
		for i := 0; i < len(r.Ctx); i += 2 {
			// not as smart about printing things as log15's internal magic
			fmt.Fprintf(&msg, ", %s: %v", r.Ctx[i].(string), r.Ctx[i+1])
		}
	}
	msg.WriteByte('\n')
	return msg.Bytes()
}

// Options control the behavior of a tracer.
//
// Deprecated: All logging should use github.com/sourcegraph/log instead. See https://docs-legacy.sourcegraph.com/dev/how-to/add_logging
type Options struct {
	filters     []func(*log15.Record) bool
	serviceName string
}

// If this idiom seems strange:
// https://github.com/tmrts/go-patterns/blob/master/idiom/functional-options.md
//
// Deprecated: All logging should use github.com/sourcegraph/log instead. See https://docs-legacy.sourcegraph.com/dev/how-to/add_logging
type Option func(*Options)

// Deprecated: All logging should use github.com/sourcegraph/log instead. See https://docs-legacy.sourcegraph.com/dev/how-to/add_logging
func ServiceName(s string) Option {
	return func(o *Options) {
		o.serviceName = s
	}
}

// Deprecated: All logging should use github.com/sourcegraph/log instead. See https://docs-legacy.sourcegraph.com/dev/how-to/add_logging
func Filter(f func(*log15.Record) bool) Option {
	return func(o *Options) {
		o.filters = append(o.filters, f)
	}
}

func init() {
	// Enable colors by default but support https://no-color.org/
	color.NoColor = env.Get("NO_COLOR", "", "Disable colored output") != ""
}

// For severity field, see https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
//
// Deprecated: All logging should use github.com/sourcegraph/log instead. See https://docs-legacy.sourcegraph.com/dev/how-to/add_logging
func LogEntryLevelString(l log15.Lvl) string {
	switch l {
	case log15.LvlDebug:
		return "DEBUG"
	case log15.LvlInfo:
		return "INFO"
	case log15.LvlWarn:
		return "WARNING"
	case log15.LvlError:
		return "ERROR"
	case log15.LvlCrit:
		return "CRITICAL"
	default:
		return "INVALID"
	}
}

// Init initializes log15's root logger based on Sourcegraph-wide logging configuration
// variables. See https://sourcegraph.com/docs/admin/observability/logs
//
// Deprecated: All logging should use github.com/sourcegraph/log instead. See https://docs-legacy.sourcegraph.com/dev/how-to/add_logging
func Init(options ...Option) {
	opts := &Options{}
	for _, setter := range options {
		setter(opts)
	}
	if opts.serviceName == "" {
		opts.serviceName = env.MyName
	}
	var handler log15.Handler
	switch env.LogFormat {
	case "condensed":
		handler = log15.StreamHandler(os.Stderr, log15.FormatFunc(condensedFormat))
	case "json":
		// for these uses: https://cloud.google.com/run/docs/logging#log-resource
		jsonFormatHandler := log15.StreamHandler(os.Stderr, log15.JsonFormat())
		handler = log15.FuncHandler(func(r *log15.Record) error {
			r.Ctx = append(r.Ctx, "severity", LogEntryLevelString(r.Lvl))
			return jsonFormatHandler.Log(r)
		})
	case "logfmt":
		fallthrough
	default:
		handler = log15.StreamHandler(os.Stderr, log15.LogfmtFormat())
	}
	for _, filter := range opts.filters {
		handler = log15.FilterHandler(filter, handler)
	}
	// Filter log output by level.
	lvl, err := log15.LvlFromString(env.LogLevel)
	if err == nil {
		handler = log15.LvlFilterHandler(lvl, handler)
	}
	if env.LogLevel == "none" {
		handler = log15.DiscardHandler()
		log.SetOutput(io.Discard)
	}
	log15.Root().SetHandler(log15.LvlFilterHandler(lvl, handler))
}
