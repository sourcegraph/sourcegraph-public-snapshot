package zap

import (
	"os"

	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/experimental_level"
	"github.com/go-kit/kit/log/term"
)

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
	logger0 := term.NewLogger(os.Stderr, log.NewLogfmtLogger, colorFn)
	logger0 = level.New(logger0, level.Config{Allowed: level.AllowAll()})
	logger1 := log.NewContext(logger0)
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
