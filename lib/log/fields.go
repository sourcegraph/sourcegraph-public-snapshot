package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// A Field is a marshaling operation used to add a key-value pair to a logger's context.
//
// Field is an aliased import that is intentionally restricted so as to not allow overly
// liberal use of log fields, namely 'Any()'.
type Field = zapcore.Field

var (
	String   = zap.String
	Int      = zap.Int
	Ints     = zap.Ints
	Float64  = zap.Float64
	Duration = zap.Duration
	Time     = zap.Time

	Error      = zap.Error
	NamedError = zap.NamedError

	Namespace = zap.Namespace
)
