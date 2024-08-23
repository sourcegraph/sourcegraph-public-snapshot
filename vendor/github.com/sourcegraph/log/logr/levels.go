package logr

import "go.uber.org/zap/zapcore"

// Zap levels are int8 - make sure we stay in bounds.  logr itself should
// ensure we never get negative values.
//
// Source: https://github.com/go-logr/zapr/blob/48df242fffb25049c72e208aea4826177ff5fe8e/zapr.go#L196
func toZapLevel(lvl int) zapcore.Level {
	if lvl > 127 {
		lvl = 127
	}
	// zap levels are inverted.
	return 0 - zapcore.Level(lvl)
}
