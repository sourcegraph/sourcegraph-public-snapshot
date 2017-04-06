package server

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log/term"
)

var logLevelOpt level.Option

func init() {
	logLevel := os.Getenv("LOGLEVEL")
	switch logLevel {
	case "debug":
		logLevelOpt = level.AllowAll()
	case "info":
		logLevelOpt = level.AllowInfo()
	case "warn":
		logLevelOpt = level.AllowWarn()
	case "", "error":
		logLevelOpt = level.AllowError()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown log level %q (valid levels are: debug, info, warn, error)\n", logLevel)
		os.Exit(1)
	}
}

func (s *Server) BaseLogger() log.Logger {
	colorFn := func(keyvals ...interface{}) term.FgBgColor {
		for i := 0; i < len(keyvals)-1; i += 2 {
			if keyvals[i] != "level" {
				continue
			}
			lvl, ok := keyvals[i+1].(level.Value)
			if !ok {
				// If this isn't a level.Value, it means
				// go-kit/log has changed. This wouldn't be
				// the first time, so rather just do not
				// color.
				break
			}
			switch lvl.String() {
			case "debug":
				return term.FgBgColor{Fg: term.Blue}
			case "info":
				return term.FgBgColor{Fg: term.Gray}
			case "warn":
				return term.FgBgColor{Fg: term.Yellow}
			case "error":
				return term.FgBgColor{Fg: term.Red}
			default:
				return term.FgBgColor{}
			}
		}
		return term.FgBgColor{}
	}

	w := s.LogWriter
	if w == nil {
		w = os.Stderr
	}

	logger0 := term.NewLogger(w, log.NewLogfmtLogger, colorFn)
	logger0 = level.NewFilter(logger0, logLevelOpt)
	logger1 := log.With(logger0)
	if v, _ := strconv.ParseBool(os.Getenv("LOGTIMESTAMP")); os.Getenv("LOGTIMESTAMP") == "" || v {
		// By default include timestamps, but adjust behaviour if LOGTIMESTAMP is specified.
		logger1 = log.With(logger1, "ts", log.DefaultTimestampUTC)
	}
	if s.ID != "" {
		logger1 = log.With(logger1, "server", s.ID)
	}
	return logger1
}
