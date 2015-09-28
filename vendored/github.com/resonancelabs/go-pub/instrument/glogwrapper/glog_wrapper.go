// A package meant as a drop-in API replacement for the golang team's `glog`.
//
// Log messages are sent both to the "real" glog package and also translated to
// something reasonable in the crouton universe.
package glogwrapper

import (
	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"

	"github.com/golang/glog"
)

type Level int32

func V(level Level) Level {
	return level
}

// ---------------------------
// UNFORMATTED LOG STATEMENTS
// ---------------------------
func Info(args ...interface{}) {
	instrument.Log(instrument.FileLine(2).Info().Print(args...))
	glog.Info(args...)
}

func Warning(args ...interface{}) {
	instrument.Log(instrument.FileLine(2).Warning().Print(args...))
	glog.Warning(args...)
}

func Error(args ...interface{}) {
	instrument.Log(instrument.FileLine(2).Error().Print(args...))
	glog.Error(args...)
}

func Fatal(args ...interface{}) {
	instrument.Log(instrument.FileLine(2).RawLevel("F").Print(args...))
	glog.Fatal(args...)
}

// ------------------------
// FORMATTED LOG STATEMENTS
// ------------------------

func Infof(format string, args ...interface{}) {
	instrument.Log(instrument.FileLine(2).Info().Printf(format, args...))
	glog.Infof(format, args...)
}

func Warningf(format string, args ...interface{}) {
	instrument.Log(instrument.FileLine(2).Warning().Printf(format, args...))
	glog.Warningf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	instrument.Log(instrument.FileLine(2).Error().Printf(format, args...))
	glog.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	instrument.Log(instrument.FileLine(2).RawLevel("F").Printf(format, args...))
	glog.Fatalf(format, args...)
}

// --------------------------------
// VERBOSITY FLAGGED LOG STATEMENTS
// --------------------------------

func (level Level) Info(args ...interface{}) {
	instrument.Log(instrument.FileLine(2).Info().Print(args...))
	glog.V(glog.Level(level)).Info(args...)
}

func (level Level) Infoln(args ...interface{}) {
	instrument.Log(instrument.FileLine(2).Info().Println(args...))
	glog.V(glog.Level(level)).Infoln(args...)
}

func (level Level) Infof(format string, args ...interface{}) {
	instrument.Log(instrument.FileLine(2).Info().Printf(format, args...))
	glog.V(glog.Level(level)).Infof(format, args...)
}
