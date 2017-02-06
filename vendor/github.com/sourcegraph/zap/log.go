package zap

import (
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/experimental_level"
	"github.com/go-kit/kit/log/term"
)

var logLevelConfig level.Config

func init() {
	logLevel := os.Getenv("LOGLEVEL")
	switch logLevel {
	case "", "debug":
		logLevelConfig.Allowed = level.AllowAll()
	case "info":
		logLevelConfig.Allowed = level.AllowInfoAndAbove()
	case "warn":
		logLevelConfig.Allowed = level.AllowWarnAndAbove()
	case "error":
		logLevelConfig.Allowed = level.AllowErrorOnly()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown log level %q (valid levels are: %v)\n", logLevel, level.AllowAll())
		os.Exit(1)
	}
}

func (s *Server) baseLogger() *log.Context {
	colorFn := func(keyvals ...interface{}) term.FgBgColor {
		for i := 0; i < len(keyvals)-1; i += 2 {
			if keyvals[i] != "level" {
				continue
			}
			switch keyvals[i+1] {
			case "debug":
				return term.FgBgColor{Fg: term.DarkGray}
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
	logger0 = level.New(logger0, logLevelConfig)
	logger1 := log.NewContext(logger0)
	// logger1 = logger1.With("ts", log.DefaultTimestampUTC)
	if s.ID != "" {
		logger1 = logger1.With("server", s.ID)
	}
	return logger1
}

func abbrevGitOID(oid string) string {
	if len(oid) == 40 {
		return oid[:6]
	}
	return oid
}
